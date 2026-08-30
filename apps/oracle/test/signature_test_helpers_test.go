package integration_test

import (
	"crypto/ecdsa"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"github.com/ridwanmuh3/hybrid-pqc-poc/apps/oracle/internal/model"
	"github.com/ridwanmuh3/hybrid-pqc-poc/apps/oracle/internal/signature"
)

var (
	fnTestHash = common.HexToHash("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
	hashA      = common.HexToHash("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	hashB      = common.HexToHash("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
)

type fnScenario struct {
	name      string
	algorithm string
	mode      string
}

var fnScenarios = []fnScenario{
	{"ECDSA (single)", "ECDSA", "single"},
	{"ML-DSA-44 (single)", "ML-DSA-44", "single"},
	{"Falcon-512 (single)", "Falcon-512", "single"},
	{"MAYO-1 (single)", "MAYO-1", "single"},
	{"SNOVA_24_5_4 (single)", "SNOVA_24_5_4", "single"},
	{"ML-DSA-65 (single)", "ML-DSA-65", "single"},
	{"MAYO-3 (single)", "MAYO-3", "single"},
	{"SNOVA_37_8_4 (single)", "SNOVA_37_8_4", "single"},
	{"ML-DSA-87 (single)", "ML-DSA-87", "single"},
	{"Falcon-1024 (single)", "Falcon-1024", "single"},
	{"MAYO-5 (single)", "MAYO-5", "single"},
	{"SNOVA_60_10_4 (single)", "SNOVA_60_10_4", "single"},
	{"Hybrid + ML-DSA-44", "Hybrid-Secp256k1-ML-DSA-44", "hybrid"},
	{"Hybrid + Falcon-512", "Hybrid-Secp256k1-Falcon-512", "hybrid"},
	{"Hybrid + MAYO-1", "Hybrid-Secp256k1-MAYO-1", "hybrid"},
	{"Hybrid + SNOVA_24_5_4", "Hybrid-Secp256k1-SNOVA_24_5_4", "hybrid"},
	{"Hybrid + ML-DSA-65", "Hybrid-Secp256k1-ML-DSA-65", "hybrid"},
	{"Hybrid + MAYO-3", "Hybrid-Secp256k1-MAYO-3", "hybrid"},
	{"Hybrid + SNOVA_37_8_4", "Hybrid-Secp256k1-SNOVA_37_8_4", "hybrid"},
	{"Hybrid + ML-DSA-87", "Hybrid-Secp256k1-ML-DSA-87", "hybrid"},
	{"Hybrid + Falcon-1024", "Hybrid-Secp256k1-Falcon-1024", "hybrid"},
	{"Hybrid + MAYO-5", "Hybrid-Secp256k1-MAYO-5", "hybrid"},
	{"Hybrid + SNOVA_60_10_4", "Hybrid-Secp256k1-SNOVA_60_10_4", "hybrid"},
}

type fnKeySet struct {
	priv   []byte
	pub    []byte
	sender common.Address
}

type hybridKeys struct {
	privateKey []byte
	publicKey  []byte
	sender     common.Address
}

type ecdsaKeys struct {
	privateKey []byte
	publicKey  []byte
	sender     common.Address
}

func hybridAlgorithm(pqAlgorithm string) string {
	return "Hybrid-Secp256k1-" + pqAlgorithm
}

func fnMakeKeySet(t testing.TB, sc fnScenario) fnKeySet {
	t.Helper()
	generator := &signature.KeyPairGenerator{}

	switch sc.mode {
	case "hybrid":
		pqAlgorithm := strings.TrimPrefix(sc.algorithm, hybridAlgorithm(""))
		keys := newECDSAKeys(t)
		priv, pub, _, err := generator.GenerateHybrid(&model.GenerateKeypairRequest{
			Algorithm:       pqAlgorithm,
			Mode:            "hybrid",
			EcdsaPrivateKey: hexutil.Encode(keys.privateKey),
		})
		require.NoError(t, err, "hybrid keygen for %s", pqAlgorithm)
		return fnKeySet{priv: priv, pub: pub, sender: keys.sender}
	case "single":
		if sc.algorithm == "ECDSA" {
			keys := newECDSAKeys(t)
			return fnKeySet{
				priv:   keys.privateKey,
				pub:    keys.publicKey,
				sender: keys.sender,
			}
		}
		priv, pub, _, err := generator.GenerateSingle(&model.GenerateKeypairRequest{
			Algorithm: sc.algorithm,
		})
		require.NoError(t, err, "single PQC keygen for %s", sc.algorithm)
		return fnKeySet{priv: priv, pub: pub}
	default:
		t.Fatalf("unknown mode: %s", sc.mode)
		return fnKeySet{}
	}
}

func newECDSAKeys(t testing.TB) ecdsaKeys {
	t.Helper()
	generator := &signature.KeyPairGenerator{}
	priv, pub, _, err := generator.GenerateECDSA()
	require.NoError(t, err, "ECDSA keygen")
	pubKey, err := crypto.UnmarshalPubkey(pub)
	require.NoError(t, err, "unmarshal ECDSA pub")
	return ecdsaKeys{
		privateKey: priv,
		publicKey:  pub,
		sender:     crypto.PubkeyToAddress(*pubKey),
	}
}

func newHybridKeys(t testing.TB, pqAlgorithm string) hybridKeys {
	t.Helper()
	ks := fnMakeKeySet(t, fnScenario{
		name:      hybridAlgorithm(pqAlgorithm),
		algorithm: hybridAlgorithm(pqAlgorithm),
		mode:      "hybrid",
	})
	return hybridKeys{
		privateKey: ks.priv,
		publicKey:  ks.pub,
		sender:     ks.sender,
	}
}

func newSinglePQCKeys(t testing.TB, algorithm string) (priv, pub []byte) {
	t.Helper()
	ks := fnMakeKeySet(t, fnScenario{
		name:      algorithm,
		algorithm: algorithm,
		mode:      "single",
	})
	return ks.priv, ks.pub
}

func hybridSign(t testing.TB, hash common.Hash, pqAlgorithm string, priv []byte) []byte {
	t.Helper()
	signer := signature.NewTransactionSigner(hybridAlgorithm(pqAlgorithm), "hybrid")
	_, _, finalSig, err := signer.Sign(hash, priv)
	require.NoError(t, err, "hybrid sign")
	return finalSig
}

func hybridVerify(hash common.Hash, pqAlgorithm string, sig, pub []byte, sender common.Address) (bool, string) {
	verifier := signature.NewSignatureVerifier(hybridAlgorithm(pqAlgorithm), "hybrid")
	return verifier.Verify(hash, sig, pub, sender)
}

func singlePQCSign(t testing.TB, hash common.Hash, algorithm string, priv []byte) []byte {
	t.Helper()
	signer := signature.NewTransactionSigner(algorithm, "single")
	_, _, sig, err := signer.Sign(hash, priv)
	require.NoError(t, err, "single PQC sign")
	return sig
}

func remintHybridForSender(t testing.TB, sc fnScenario, senderSk *ecdsa.PrivateKey) fnKeySet {
	t.Helper()
	generator := &signature.KeyPairGenerator{}
	pqAlgorithm := strings.TrimPrefix(sc.algorithm, hybridAlgorithm(""))
	priv, pub, _, err := generator.GenerateHybrid(&model.GenerateKeypairRequest{
		Algorithm:       pqAlgorithm,
		Mode:            "hybrid",
		EcdsaPrivateKey: hexutil.Encode(crypto.FromECDSA(senderSk)),
	})
	require.NoError(t, err, "[%s] hybrid remint", sc.name)
	return fnKeySet{
		priv:   priv,
		pub:    pub,
		sender: crypto.PubkeyToAddress(senderSk.PublicKey),
	}
}
