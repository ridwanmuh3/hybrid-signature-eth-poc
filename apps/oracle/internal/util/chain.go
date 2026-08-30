package util

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/spf13/viper"

	transactions "github.com/ridwanmuh3/hybrid-pqc-poc/apps/oracle/artifacts"
	"github.com/ridwanmuh3/hybrid-pqc-poc/apps/oracle/internal/model"
	"github.com/ridwanmuh3/hybrid-pqc-poc/apps/oracle/internal/signature"
)

func SetupOracleAuth(viperConfig *viper.Viper) (*bind.TransactOpts, *big.Int, error) {
	oraclePrivKeyHex, err := hexutil.Decode(viperConfig.GetString("ORACLE_PRIVATE_KEY"))
	if err != nil {
		return nil, nil, fmt.Errorf("error loading oracle key: %w", err)
	}

	oracleSk, err := crypto.ToECDSA(oraclePrivKeyHex)
	if err != nil {
		return nil, nil, fmt.Errorf("error converting oracle key: %w", err)
	}

	chainID := big.NewInt(viperConfig.GetInt64("CHAIN_ID"))
	auth, err := bind.NewKeyedTransactorWithChainID(oracleSk, chainID)
	if err != nil {
		return nil, nil, fmt.Errorf("error authenticate oracle: %w", err)
	}

	auth.Value = big.NewInt(0)

	return auth, chainID, nil
}

func ValidateOracleAuth(contractInstance *transactions.Transactions, auth *bind.TransactOpts) error {
	onChainOracle, err := contractInstance.ORACLE(&bind.CallOpts{})
	if err != nil {
		return fmt.Errorf("error reading contract oracle: %w", err)
	}

	if onChainOracle != auth.From {
		return fmt.Errorf(
			"oracle signer mismatch: contract ORACLE is %s but ORACLE_PRIVATE_KEY resolves to %s",
			onChainOracle.Hex(),
			auth.From.Hex(),
		)
	}

	return nil
}

func GetNextNonce(contractInstance *transactions.Transactions, sender common.Address) (*big.Int, error) {
	currentNonce, err := contractInstance.GetUserNonce(&bind.CallOpts{}, sender)
	if err != nil {
		return nil, err
	}
	return new(big.Int).Add(currentNonce, big.NewInt(1)), nil
}

func SendToBlockchain(contractInstance *transactions.Transactions, auth *bind.TransactOpts, nonce *big.Int, sender, receiver common.Address, req *model.SendAndSignTxRequest, amount *big.Int, ecdsaSig, pqSig []byte) (*types.Transaction, error) {
	pqSigToSend := pqSig
	if req.Mode != "hybrid" && req.Algorithm == "ECDSA" {
		pqSigToSend = []byte{}
	}

	tx, err := contractInstance.SendTransaction(auth, transactions.TransactionsTxParams{
		Nonce:            nonce,
		Sender:           sender,
		Receiver:         receiver,
		SigningAlgorithm: req.Algorithm,
		SigningMode:      req.Mode,
		Message:          req.Message,
		Amount:           amount,
		EcdsaSignature:   ecdsaSig,
		PqSignature:      pqSigToSend,
	})

	if err != nil {
		log.Infof("error sending tx to blockchain: %v", err)
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Blockchain execution failed: "+err.Error())
	}

	return tx, nil
}

func BuildSignatureBytes(params *transactions.TransactionsTxParams) []byte {
	if params.SigningMode == "hybrid" {
		// Reconstruct TLV finalSig from the raw on-chain components.
		// EcdsaSignature and PqSignature are stored as raw bytes by the contract;
		// the off-chain verifier expects the TLV-encoded form.
		return signature.EncodeSigTLV(
			params.SigningAlgorithm,
			params.EcdsaSignature,
			params.PqSignature,
		)
	}
	if len(params.PqSignature) > 0 {
		return params.PqSignature
	}
	return params.EcdsaSignature
}

func BuildTransactionResponse(txHash string, receipt *types.Receipt, tx *types.Transaction, relayer common.Address, params *transactions.TransactionsTxParams) fiber.Map {
	isHybrid := params.SigningMode == "hybrid"
	fullSignature := BuildSignatureBytes(params)

	var ecdsaSigHex, pqSigHex string
	if isHybrid {
		if len(params.EcdsaSignature) > 0 {
			ecdsaSigHex = hexutil.Encode(params.EcdsaSignature)
		}
		if len(params.PqSignature) > 0 {
			pqSigHex = hexutil.Encode(params.PqSignature)
		}
	} else if len(params.PqSignature) > 0 {
		pqSigHex = hexutil.Encode(params.PqSignature)
	} else {
		ecdsaSigHex = hexutil.Encode(params.EcdsaSignature)
	}
	resp := fiber.Map{
		"tx_hash":           txHash,
		"status":            receipt.Status,
		"block_number":      receipt.BlockNumber.Uint64(),
		"relayer_addr":      relayer.Hex(),
		"relayer_nonce":     tx.Nonce(),
		"user_nonce":        params.Nonce.String(),
		"sender":            params.Sender.Hex(),
		"receiver":          params.Receiver.Hex(),
		"signing_algorithm": params.SigningAlgorithm,
		"signing_mode":      params.SigningMode,
		"amount_wei":        params.Amount.String(),
		"message":           params.Message,
		"is_hybrid":         isHybrid,
		"signature_full":    hexutil.Encode(fullSignature),
		"signature_ecdsa":   ecdsaSigHex,
		"signature_pqs":     pqSigHex,
		"gas_price":         tx.GasPrice().Int64(),
		"gas_used":          receipt.GasUsed,
	}
	log.Infof("%#v", resp)
	return resp
}
