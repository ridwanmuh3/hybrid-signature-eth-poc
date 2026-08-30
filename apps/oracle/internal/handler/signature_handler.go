package handler

import (
	"context"
	"errors"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/v2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/spf13/viper"

	transactions "github.com/ridwanmuh3/hybrid-pqc-poc/apps/oracle/artifacts"
	"github.com/ridwanmuh3/hybrid-pqc-poc/apps/oracle/internal/chain"
	"github.com/ridwanmuh3/hybrid-pqc-poc/apps/oracle/internal/model"
	"github.com/ridwanmuh3/hybrid-pqc-poc/apps/oracle/internal/signature"
	"github.com/ridwanmuh3/hybrid-pqc-poc/apps/oracle/internal/util"
)

// senderLocks serialises /api/sign requests per sender. Concurrent requests
// for the same sender would otherwise read the same on-chain nonce and all
// sign for the same nonce, causing all-but-the-first to revert with
// "Nonce value too low" once mined.
var senderLocks sync.Map // map[common.Address]*sync.Mutex

func lockSender(addr common.Address) func() {
	mIface, _ := senderLocks.LoadOrStore(addr, &sync.Mutex{})
	m := mIface.(*sync.Mutex)
	m.Lock()
	return m.Unlock
}

// waitMinedTimeout caps how long a single /api/sign request may block waiting
// for inclusion. Long enough for a slow Anvil block, short enough to keep the
// HTTP layer responsive if the upstream node is unhealthy.
const waitMinedTimeout = 60 * time.Second

func GenerateKeypairHandler(c *fiber.Ctx) error {
	req := new(model.GenerateKeypairRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": "error", "message": "Invalid request body: " + err.Error(),
		})
	}
	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": "error", "message": err.Error(),
		})
	}

	generator := &signature.KeyPairGenerator{}
	var privateKey, publicKey []byte
	var algorithm string
	var err error

	switch req.Mode {
	case "single":
		privateKey, publicKey, algorithm, err = generator.GenerateSingle(req)
	case "hybrid":
		privateKey, publicKey, algorithm, err = generator.GenerateHybrid(req)
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": "error", "message": "Invalid mode. Use 'single' or 'hybrid'.",
		})
	}

	if err != nil {
		log.Infof("key generation failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status": "error", "message": "Key generation failed",
		})
	}

	return c.JSON(fiber.Map{
		"status":      "success",
		"mode":        req.Mode,
		"algorithm":   algorithm,
		"private_key": hexutil.Encode(privateKey),
		"public_key":  hexutil.Encode(publicKey),
	})
}

func SendAndSignTransactionHandler(viperConfig *viper.Viper, client *ethclient.Client, contractInstance *transactions.Transactions) fiber.Handler {
	return func(c *fiber.Ctx) error {
		req := new(model.SendAndSignTxRequest)
		if err := c.BodyParser(req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status": "error", "message": "Invalid request body: " + err.Error(),
			})
		}
		if err := req.Validate(); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status": "error", "message": err.Error(),
			})
		}

		amountBig, ok := new(big.Int).SetString(req.Amount, 10)
		if !ok {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status": "error", "message": "Invalid amount format",
			})
		}

		auth, _, err := util.SetupOracleAuth(viperConfig)
		if err != nil {
			log.Infof("error setup oracle auth: %v", err)
			return fiber.ErrInternalServerError
		}

		sender := common.HexToAddress(req.Sender)
		receiver := common.HexToAddress(req.Receiver)

		// Per-sender mutex prevents two concurrent requests from reading the
		// same on-chain nonce and producing colliding submissions.
		unlock := lockSender(sender)
		defer unlock()

		txNonce, err := util.GetNextNonce(contractInstance, sender)
		if err != nil {
			log.Infof("error fetching user nonce: %v", err)
			return fiber.ErrInternalServerError
		}

		chainID, err := client.ChainID(c.Context())
		if err != nil {
			log.Infof("error fetching chain ID: %v", err)
			return fiber.ErrInternalServerError
		}

		hashData, err := util.HashEthSignedData(chainID, txNonce, sender, receiver, req.Algorithm, req.Mode, req.Message, amountBig)
		if err != nil {
			log.Infof("error hashing data: %v", err)
			return fiber.ErrInternalServerError
		}

		privateKey, err := hexutil.Decode(req.PrivateKey)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid private key hex")
		}

		signer := signature.NewTransactionSigner(req.Algorithm, req.Mode)
		ecdsaSig, pqSig, finalSig, err := signer.Sign(hashData, privateKey)
		if err != nil {
			log.Infof("signing error: %v", err)
			return fiber.ErrInternalServerError
		}

		tx, err := util.SendToBlockchain(contractInstance, auth, txNonce, sender, receiver, req, amountBig, ecdsaSig, pqSig)
		if err != nil {
			return err
		}

		// Bound the wait so a stalled or partitioned upstream node does not
		// pin a goroutine and Fiber connection indefinitely.
		waitCtx, cancel := context.WithTimeout(c.Context(), waitMinedTimeout)
		defer cancel()
		receipt, err := bind.WaitMined(waitCtx, client, tx.Hash())
		if err != nil {
			log.Infof("error retrieving transaction: %v", err)
			if errors.Is(err, context.DeadlineExceeded) {
				return fiber.NewError(fiber.StatusGatewayTimeout, "Timed out waiting for transaction inclusion")
			}
			return fiber.ErrInternalServerError
		}

		if receipt.Status == ethtypes.ReceiptStatusFailed {
			log.Infof("on-chain ECDSA verification failed, tx reverted: %s", tx.Hash().Hex())
			return fiber.NewError(fiber.StatusBadRequest, "On-chain ECDSA verification failed: transaction reverted")
		}

		return c.JSON(fiber.Map{
			"status":            "success",
			"signing_algorithm": req.Algorithm,
			"transaction_hash":  tx.Hash().Hex(),
			"user_nonce_used":   txNonce.Int64(),
			"sender":            req.Sender,
			"receiver":          req.Receiver,
			"amount_wei":        req.Amount,
			"signature":         hexutil.Encode(finalSig),
			"signature_length":  len(finalSig),
			"gas_price":         tx.GasPrice().Int64(),
			"gas_used":          int64(receipt.GasUsed),
		})
	}
}

func VerifyTransactionByHashHandler(client *ethclient.Client) fiber.Handler {		return func(c *fiber.Ctx) error {
			req := new(model.VerifyByHashRequest)
			if err := c.BodyParser(req); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"status": "error", "message": "Invalid request body: " + err.Error(),
				})
			}
			if err := req.Validate(); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"status": "error", "message": err.Error(),
				})
			}

		publicKeyBytes, err := hexutil.Decode(req.PublicKey)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid public key hex")
		}

		txHash := common.HexToHash(req.TxHash)
		tx, isPending, err := client.TransactionByHash(c.Context(), txHash)
		if err != nil {
			if errors.Is(err, ethereum.NotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"valid":          false,
					"verify_message": "Transaction not found",
					"note":           "No transaction with the given hash exists on-chain",
				})
			}
			log.Infof("error fetching tx: %v", err)
			return fiber.ErrInternalServerError
		}

		if isPending {
			return fiber.NewError(fiber.StatusBadRequest, "Transaction is still pending")
		}

		receipt, err := client.TransactionReceipt(c.Context(), txHash)
		if err != nil {
			log.Infof("error getting receipt: %v", err)
			return fiber.ErrInternalServerError
		}

		decoder, err := chain.NewTransactionDecoder()
		if err != nil {
			log.Infof("decoder error: %v", err)
			return fiber.ErrInternalServerError
		}

		params, err := decoder.Decode(tx)
		if err != nil {
			log.Infof("decode error: %v", err)
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		hashData, err := util.HashEthSignedData(
			tx.ChainId(),
			params.Nonce, params.Sender, params.Receiver,
			params.SigningAlgorithm, params.SigningMode,
			params.Message, params.Amount,
		)
		if err != nil {
			log.Infof("hash error: %v", err)
			return fiber.ErrInternalServerError
		}

		sigBytes := util.BuildSignatureBytes(params)

		msgFrom, err := ethtypes.Sender(ethtypes.LatestSignerForChainID(tx.ChainId()), tx)
		if err != nil {
			log.Infof("error getting relayer: %v", err)
			return fiber.ErrInternalServerError
		}

		txResponse := util.BuildTransactionResponse(req.TxHash, receipt, tx, msgFrom, params)

		verifier := signature.NewSignatureVerifier(params.SigningAlgorithm, params.SigningMode)
		isValid, note := verifier.Verify(hashData, sigBytes, publicKeyBytes, params.Sender)

		txResponse["valid"] = isValid
		txResponse["note"] = note
		if isValid {
			txResponse["verify_message"] = "Verification Success"
			return c.JSON(txResponse)
		}
		txResponse["verify_message"] = "Verification Failed"
		return c.Status(fiber.StatusUnauthorized).JSON(txResponse)
	}
}

func GetTransactionByHashHandler(client *ethclient.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		txHashHex := c.Params("txhash")
		txHash := common.HexToHash(txHashHex)

		tx, isPending, err := client.TransactionByHash(c.Context(), txHash)
		if err != nil {
			log.Infof("transaction not found")
			return fiber.NewError(fiber.StatusNotFound, "Transaction not found")
		}

		if isPending {
			log.Infof("transaction is pending")
			return fiber.NewError(fiber.StatusBadRequest, "Transaction is pending")
		}

		receipt, err := client.TransactionReceipt(c.Context(), txHash)
		if err != nil {
			log.Infof("error getting receipt: %v", err)
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to get receipt")
		}

		decoder, err := chain.NewTransactionDecoder()
		if err != nil {
			log.Infof("decoder error: %v", err)
			return fiber.ErrInternalServerError
		}

		params, err := decoder.Decode(tx)
		if err != nil {
			log.Infof("decode error: %v", err)
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		msgFrom, err := ethtypes.Sender(ethtypes.LatestSignerForChainID(tx.ChainId()), tx)
		if err != nil {
			log.Infof("error getting sender: %v", err)
			return fiber.ErrInternalServerError
		}

		response := util.BuildTransactionResponse(txHashHex, receipt, tx, msgFrom, params)
		return c.JSON(response)
	}
}
