package signature

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/open-quantum-safe/liboqs-go/oqs"

	"github.com/ridwanmuh3/hybrid-pqc-poc/apps/oracle/internal/model"
)

type KeyPairGenerator struct{}

func (g *KeyPairGenerator) GenerateSingle(req *model.GenerateKeypairRequest) ([]byte, []byte, string, error) {
	if req.Algorithm == "ECDSA" {
		return g.GenerateECDSA()
	}
	return g.GeneratePQC(req.Algorithm)
}

func (g *KeyPairGenerator) GenerateECDSA() ([]byte, []byte, string, error) {
	ecdsaKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, nil, "", err
	}

	privateKey := crypto.FromECDSA(ecdsaKey)
	publicKey := crypto.FromECDSAPub(&ecdsaKey.PublicKey)

	return privateKey, publicKey, "ECDSA", nil
}

func (g *KeyPairGenerator) GeneratePQC(algorithm string) ([]byte, []byte, string, error) {
	signer := oqs.Signature{}

	algo := algorithm
	if algo == "" {
		algo = "ML-DSA-65"
	}

	if err := signer.Init(algo, nil); err != nil {
		return nil, nil, "", fmt.Errorf("OQS init failed: %w", err)
	}

	defer signer.Clean()

	pubKeyRaw, err := signer.GenerateKeyPair()
	if err != nil {
		return nil, nil, "", err
	}

	privKeyRaw := signer.ExportSecretKey()

	privateKey := make([]byte, len(privKeyRaw))
	copy(privateKey, privKeyRaw)

	publicKey := make([]byte, len(pubKeyRaw))
	copy(publicKey, pubKeyRaw)

	return privateKey, publicKey, algo, nil
}

func (g *KeyPairGenerator) GenerateHybrid(req *model.GenerateKeypairRequest) ([]byte, []byte, string, error) {
	ecdsaSk, err := g.GetOrCreateECDSAKey(req.EcdsaPrivateKey)
	if err != nil {
		return nil, nil, "", err
	}

	pqcPrivateKey, pqcPublicKey, algo, err := g.GeneratePQC(req.Algorithm)
	if err != nil {
		return nil, nil, "", err
	}

	ecdsaPriv := crypto.FromECDSA(ecdsaSk)
	ecdsaPub := crypto.FromECDSAPub(&ecdsaSk.PublicKey)
	algorithmName := fmt.Sprintf("Hybrid-Secp256k1-%s", algo)

	privateKey := EncodePrivKeyTLV(algorithmName, ecdsaPriv, pqcPrivateKey)
	publicKey := EncodePubKeyTLV(algorithmName, ecdsaPub, pqcPublicKey)

	return privateKey, publicKey, algorithmName, nil
}

func (g *KeyPairGenerator) GetOrCreateECDSAKey(existingKey string) (*ecdsa.PrivateKey, error) {
	if existingKey != "" {
		decodedKey, err := hexutil.Decode(existingKey)
		if err != nil {
			return nil, fmt.Errorf("invalid existing ECDSA key: %w", err)
		}
		return crypto.ToECDSA(decodedKey)
	}
	return crypto.GenerateKey()
}
