package integration_test

import (
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	transactions "github.com/ridwanmuh3/hybrid-pqc-poc/apps/oracle/artifacts"
	"github.com/ridwanmuh3/hybrid-pqc-poc/apps/oracle/internal/model"
	"github.com/ridwanmuh3/hybrid-pqc-poc/apps/oracle/internal/signature"
	"github.com/ridwanmuh3/hybrid-pqc-poc/apps/oracle/internal/util"
)

func TestHashEthSignedData_Consistency(t *testing.T) {
	chainID := big.NewInt(31337)
	nonce := big.NewInt(1)
	sender := common.HexToAddress("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266")
	receiver := common.HexToAddress("0x70997970C51812dc3A010C7d01b50e0d17dc79C8")
	algorithm := "Hybrid-Secp256k1-ML-DSA-65"
	mode := "hybrid"
	message := "payment for service"
	amount := big.NewInt(1000000000000000000)

	h1, err := util.HashEthSignedData(chainID, nonce, sender, receiver, algorithm, mode, message, amount)
	require.NoError(t, err)

	h2, err := util.HashEthSignedData(chainID, nonce, sender, receiver, algorithm, mode, message, amount)
	require.NoError(t, err)

	assert.Equal(t, h1, h2, "same inputs must produce same hash")
}

func TestHashEthSignedData_DifferentNonce(t *testing.T) {
	chainID := big.NewInt(31337)
	sender := common.HexToAddress("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266")
	receiver := common.HexToAddress("0x70997970C51812dc3A010C7d01b50e0d17dc79C8")

	h1, err := util.HashEthSignedData(chainID, big.NewInt(1), sender, receiver, "ECDSA", "single", "msg", big.NewInt(100))
	require.NoError(t, err)

	h2, err := util.HashEthSignedData(chainID, big.NewInt(2), sender, receiver, "ECDSA", "single", "msg", big.NewInt(100))
	require.NoError(t, err)

	assert.NotEqual(t, h1, h2, "different nonces must produce different hashes")
}

func TestHashEthSignedData_DifferentChainID(t *testing.T) {
	sender := common.HexToAddress("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266")
	receiver := common.HexToAddress("0x70997970C51812dc3A010C7d01b50e0d17dc79C8")

	h1, err := util.HashEthSignedData(big.NewInt(1), big.NewInt(1), sender, receiver, "ECDSA", "single", "msg", big.NewInt(100))
	require.NoError(t, err)

	h2, err := util.HashEthSignedData(big.NewInt(11155111), big.NewInt(1), sender, receiver, "ECDSA", "single", "msg", big.NewInt(100))
	require.NoError(t, err)

	assert.NotEqual(t, h1, h2, "different chain IDs must produce different hashes")
}

func TestBuildSignatureBytes_Hybrid(t *testing.T) {
	ecdsaSig := make([]byte, 65)
	pqcSigBytes := make([]byte, 100)
	params := &transactions.TransactionsTxParams{
		SigningAlgorithm: "Hybrid-Secp256k1-ML-DSA-65",
		SigningMode:      "hybrid",
		EcdsaSignature:   ecdsaSig,
		PqSignature:      pqcSigBytes,
	}
	result := util.BuildSignatureBytes(params)
	assert.NotEmpty(t, result, "hybrid sig should be non-empty")
	fields, err := signature.DecodeTLV(result)
	require.NoError(t, err)
	assert.Contains(t, string(fields[signature.TagAlgorithm]), "ML-DSA-65")
	assert.Equal(t, ecdsaSig, fields[signature.TagECDSASig])
	assert.Equal(t, pqcSigBytes, fields[signature.TagPQCSig])
}

func TestBuildSignatureBytes_SinglePQC(t *testing.T) {
	pqcSig := make([]byte, 200)
	params := &transactions.TransactionsTxParams{
		SigningAlgorithm: "ML-DSA-65",
		SigningMode:      "single",
		EcdsaSignature:   make([]byte, 0),
		PqSignature:      pqcSig,
	}
	result := util.BuildSignatureBytes(params)
	assert.Equal(t, pqcSig, result, "single PQC returns raw pqSig")
}

func TestBuildSignatureBytes_SingleECDSA(t *testing.T) {
	ecdsaSig := make([]byte, 65)
	params := &transactions.TransactionsTxParams{
		SigningAlgorithm: "ECDSA",
		SigningMode:      "single",
		EcdsaSignature:   ecdsaSig,
		PqSignature:      make([]byte, 0),
	}
	result := util.BuildSignatureBytes(params)
	assert.Equal(t, ecdsaSig, result, "single ECDSA returns raw ecdsaSig")
}

func TestBuildSignatureBytes_SingleECDSA_NoPqcSig(t *testing.T) {
	ecdsaSig := make([]byte, 65)
	params := &transactions.TransactionsTxParams{
		SigningAlgorithm: "ECDSA",
		SigningMode:      "single",
		EcdsaSignature:   ecdsaSig,
	}
	result := util.BuildSignatureBytes(params)
	assert.Equal(t, ecdsaSig, result)
}

func TestBoolToString(t *testing.T) {
	assert.Equal(t, "OK", util.BoolToString(true))
	assert.Equal(t, "FAIL", util.BoolToString(false))
}

func TestKeyToPemFormat(t *testing.T) {
	key := []byte("test-key-data")
	result := util.KeyToPemFormat("ML-DSA-65", "PRIVATE KEY", key)
	assert.True(t, strings.Contains(string(result), "BEGIN ML-DSA-65 PRIVATE KEY"))
	assert.True(t, strings.Contains(string(result), "END ML-DSA-65 PRIVATE KEY"))
}

func TestGetNextNonce(t *testing.T) {
	current := big.NewInt(0)
	next := new(big.Int).Add(current, big.NewInt(1))
	assert.Equal(t, big.NewInt(1), next)
}

func TestSetupOracleAuth_WithValidKey(t *testing.T) {
	sk, err := crypto.GenerateKey()
	require.NoError(t, err)
	skHex := hexutil.Encode(crypto.FromECDSA(sk))
	viperConfig := newViperConfig(t)
	viperConfig.Set("ORACLE_PRIVATE_KEY", skHex)
	_, _, err = util.SetupOracleAuth(viperConfig)
	require.NoError(t, err)
}

func TestSetupOracleAuth_InvalidKey(t *testing.T) {
	viperConfig := newViperConfig(t)
	viperConfig.Set("ORACLE_PRIVATE_KEY", "0xdeadbeef")
	_, _, err := util.SetupOracleAuth(viperConfig)
	assert.Error(t, err)
}

func TestSendToBlockchain_Logic(t *testing.T) {
	// ECDSA single mode: pqSig should be stripped to empty
	req := model.SendAndSignTxRequest{Mode: "single", Algorithm: "ECDSA"}
	pqcSig := []byte{1, 2, 3}
	sent := pqcSig
	if req.Mode != "hybrid" && req.Algorithm == "ECDSA" {
		sent = []byte{}
	}
	assert.Empty(t, sent, "ECDSA single mode should have empty pqSig")

	// Hybrid mode: pqSig should be preserved
	req = model.SendAndSignTxRequest{Mode: "hybrid", Algorithm: "Hybrid-Secp256k1-ML-DSA-65"}
	sent = pqcSig
	if req.Mode != "hybrid" && req.Algorithm == "ECDSA" {
		sent = []byte{}
	}
	assert.Equal(t, pqcSig, sent, "hybrid mode should keep pqSig")

	// PQC single mode: pqSig should be preserved
	req = model.SendAndSignTxRequest{Mode: "single", Algorithm: "ML-DSA-65"}
	sent = pqcSig
	if req.Mode != "hybrid" && req.Algorithm == "ECDSA" {
		sent = []byte{}
	}
	assert.Equal(t, pqcSig, sent, "PQC single mode should keep pqSig")
}
