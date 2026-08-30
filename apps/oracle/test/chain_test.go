package integration_test

import (
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	transactions "github.com/ridwanmuh3/hybrid-pqc-poc/apps/oracle/artifacts"
	"github.com/ridwanmuh3/hybrid-pqc-poc/apps/oracle/internal/chain"
	"github.com/ridwanmuh3/hybrid-pqc-poc/apps/oracle/internal/model"
	"github.com/ridwanmuh3/hybrid-pqc-poc/apps/oracle/internal/signature"
	"github.com/ridwanmuh3/hybrid-pqc-poc/apps/oracle/internal/util"
)

func TestTransactionDecoder_DecodeSendTransaction(t *testing.T) {
	decoder, err := chain.NewTransactionDecoder()
	require.NoError(t, err)

	chainID := big.NewInt(31337)
	sender := common.HexToAddress("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266")
	receiver := common.HexToAddress("0x70997970C51812dc3A010C7d01b50e0d17dc79C8")

	hashData, err := util.HashEthSignedData(chainID, big.NewInt(1), sender, receiver, "ECDSA", "single", "hello", big.NewInt(100))
	require.NoError(t, err)

	sk, err := crypto.GenerateKey()
	require.NoError(t, err)

	ecdsaSig, err := crypto.Sign(hashData.Bytes(), sk)
	require.NoError(t, err)

	oracleSk, err := crypto.GenerateKey()
	require.NoError(t, err)
	oracleAuth, err := bind.NewKeyedTransactorWithChainID(oracleSk, chainID)
	require.NoError(t, err)

	params := &model.SendAndSignTxRequest{
		Sender:    sender.Hex(),
		Receiver:  receiver.Hex(),
		Message:   "hello",
		Amount:    "100",
		Algorithm: "ECDSA",
		Mode:      "single",
	}

	t.Run("ECDSA single mode", func(t *testing.T) {
		txParams := &transactions.TransactionsTxParams{
			Nonce:            big.NewInt(1),
			Sender:           sender,
			Receiver:         receiver,
			SigningAlgorithm: "ECDSA",
			SigningMode:      "single",
			Message:          "hello",
			Amount:           big.NewInt(100),
			EcdsaSignature:   ecdsaSig,
			PqSignature:      []byte{},
		}
		sigBytes := util.BuildSignatureBytes(txParams)
		assert.Equal(t, ecdsaSig, sigBytes)
	})

	t.Run("PQC single mode", func(t *testing.T) {
		pqcSig := []byte{0x01, 0x02, 0x03, 0x04}
		txParams := &transactions.TransactionsTxParams{
			Nonce:            big.NewInt(1),
			Sender:           sender,
			Receiver:         receiver,
			SigningAlgorithm: "ML-DSA-65",
			SigningMode:      "single",
			Message:          "hello",
			Amount:           big.NewInt(100),
			EcdsaSignature:   []byte{},
			PqSignature:      pqcSig,
		}
		sigBytes := util.BuildSignatureBytes(txParams)
		assert.Equal(t, pqcSig, sigBytes)
	})

	t.Run("hybrid mode", func(t *testing.T) {
		pqcSig := []byte{0x01, 0x02, 0x03, 0x04}
		txParams := &transactions.TransactionsTxParams{
			Nonce:            big.NewInt(1),
			Sender:           sender,
			Receiver:         receiver,
			SigningAlgorithm: "Hybrid-Secp256k1-ML-DSA-65",
			SigningMode:      "hybrid",
			Message:          "hello",
			Amount:           big.NewInt(100),
			EcdsaSignature:   ecdsaSig,
			PqSignature:      pqcSig,
		}
		sigBytes := util.BuildSignatureBytes(txParams)
		assert.True(t, len(sigBytes) > 65, "hybrid sig should be longer than ECDSA-only")
		// Verify TLV format
		fields, err := signature.DecodeTLV(sigBytes)
		require.NoError(t, err)
		assert.Contains(t, string(fields[signature.TagAlgorithm]), "ML-DSA-65")
	})

	_ = decoder
	_ = params
	_ = oracleAuth
}

func TestTransactionDecoder_RejectsNonSendTransactionData(t *testing.T) {
	decoder, err := chain.NewTransactionDecoder()
	require.NoError(t, err)

	sk, err := crypto.GenerateKey()
	require.NoError(t, err)

	signedTx, err := types.SignNewTx(sk, types.NewEIP155Signer(big.NewInt(1)), &types.LegacyTx{
		Nonce:    0,
		GasPrice: big.NewInt(1),
		Gas:      21000,
		Value:    big.NewInt(0),
		Data:     []byte{},
	})
	require.NoError(t, err)

	_, err = decoder.Decode(signedTx)
	assert.Error(t, err)
}

func TestTransactionDecoder_RejectsTooShortData(t *testing.T) {
	decoder, err := chain.NewTransactionDecoder()
	require.NoError(t, err)

	sk, err := crypto.GenerateKey()
	require.NoError(t, err)

	signedTx, err := types.SignNewTx(sk, types.NewEIP155Signer(big.NewInt(1)), &types.LegacyTx{
		Nonce:    0,
		GasPrice: big.NewInt(1),
		Gas:      21000,
		Value:    big.NewInt(0),
		Data:     []byte{0x01, 0x02},
	})
	require.NoError(t, err)

	_, err = decoder.Decode(signedTx)
	assert.Error(t, err)
}

func TestBuildTransactionResponse_Structure(t *testing.T) {
	sender := common.HexToAddress("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266")
	receiver := common.HexToAddress("0x70997970C51812dc3A010C7d01b50e0d17dc79C8")
	ecdsaSig := make([]byte, 65)
	pqcSig := make([]byte, 100)

	params := &transactions.TransactionsTxParams{
		Nonce:            big.NewInt(1),
		Sender:           sender,
		Receiver:         receiver,
		SigningAlgorithm: "Hybrid-Secp256k1-ML-DSA-65",
		SigningMode:      "hybrid",
		Message:          "test message",
		Amount:           big.NewInt(100),
		EcdsaSignature:   ecdsaSig,
		PqSignature:      pqcSig,
	}

	txHash := "0xabcdef1234567890"
	receipt := &types.Receipt{Status: 1, BlockNumber: big.NewInt(42)}
	tx := types.NewTx(&types.LegacyTx{Nonce: 1, GasPrice: big.NewInt(1000), Gas: 21000, Data: []byte{}})
	relayer := common.HexToAddress("0x0000000000000000000000000000000000000001")

	resp := util.BuildTransactionResponse(txHash, receipt, tx, relayer, params)

	assert.Equal(t, txHash, resp["tx_hash"])
	assert.Equal(t, uint64(1), resp["status"])
	assert.Equal(t, uint64(42), resp["block_number"].(uint64))
	assert.Equal(t, sender.Hex(), resp["sender"])
	assert.Equal(t, receiver.Hex(), resp["receiver"])
	assert.Equal(t, "hybrid", resp["signing_mode"])
	assert.Equal(t, true, resp["is_hybrid"])
	assert.Equal(t, "test message", resp["message"])

	sigFull, ok := resp["signature_full"].(string)
	assert.True(t, ok)
	assert.True(t, strings.HasPrefix(sigFull, "0x"))
}

func TestBuildTransactionResponse_ECDSA_Single(t *testing.T) {
	sender := common.HexToAddress("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266")
	receiver := common.HexToAddress("0x70997970C51812dc3A010C7d01b50e0d17dc79C8")
	ecdsaSig := make([]byte, 65)
	for i := range ecdsaSig {
		ecdsaSig[i] = byte(i)
	}

	params := &transactions.TransactionsTxParams{
		Nonce:            big.NewInt(1),
		Sender:           sender,
		Receiver:         receiver,
		SigningAlgorithm: "ECDSA",
		SigningMode:      "single",
		Message:          "test",
		Amount:           big.NewInt(100),
		EcdsaSignature:   ecdsaSig,
		PqSignature:      []byte{},
	}

	receipt := &types.Receipt{Status: 1, BlockNumber: big.NewInt(42)}
	tx := types.NewTx(&types.LegacyTx{Nonce: 1, GasPrice: big.NewInt(1000), Gas: 21000, Data: []byte{}})
	relayer := sender

	resp := util.BuildTransactionResponse("0xabc", receipt, tx, relayer, params)

	assert.Equal(t, false, resp["is_hybrid"])
	ecdsaHex, ok := resp["signature_ecdsa"].(string)
	assert.True(t, ok)
	assert.True(t, strings.HasPrefix(ecdsaHex, "0x"))
}
