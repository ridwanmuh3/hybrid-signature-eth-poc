package signature

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/open-quantum-safe/liboqs-go/oqs"
)

type TransactionSigner struct {
	algorithm string
	mode      string
}

func NewTransactionSigner(algorithm, mode string) *TransactionSigner {
	return &TransactionSigner{
		algorithm: algorithm,
		mode:      mode,
	}
}

func (s *TransactionSigner) Sign(hashData common.Hash, privateKey []byte) (ecdsaSig, pqSig, finalSig []byte, err error) {
	parsedAlgorithm := strings.TrimPrefix(s.algorithm, "Hybrid-Secp256k1-")

	switch s.mode {
	case "single":
		return s.SignSingle(hashData, privateKey, parsedAlgorithm)
	case "hybrid":
		return s.SignHybrid(hashData, privateKey, parsedAlgorithm)
	default:
		return nil, nil, nil, fmt.Errorf("invalid signing mode: %s", s.mode)
	}
}

func (s *TransactionSigner) SignSingle(hashData common.Hash, privateKey []byte, algorithm string) ([]byte, []byte, []byte, error) {
	if s.algorithm == "ECDSA" {
		ecdsaSig, err := s.SignECDSA(hashData, privateKey)
		return ecdsaSig, nil, ecdsaSig, err
	}

	if algorithm == "" {
		return nil, nil, nil, fmt.Errorf("algorithm required for PQC")
	}

	pqSig, err := s.SignPqcSignature(algorithm, hashData.Bytes(), privateKey)
	return nil, pqSig, pqSig, err
}

func (s *TransactionSigner) SignHybrid(hashData common.Hash, privateKey []byte, algorithm string) ([]byte, []byte, []byte, error) {
	ecdsaPart, pqcPart, err := parseHybridPrivateKey(privateKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed parsing hybrid private key: %w", err)
	}

	ecdsaSig, err := s.SignECDSA(hashData, ecdsaPart)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("ECDSA signing failed: %w", err)
	}

	pqSig, err := s.SignPqcSignature(algorithm, hashData.Bytes(), pqcPart)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("PQC signing failed: %w", err)
	}

	finalSig := EncodeSigTLV(s.algorithm, ecdsaSig, pqSig)
	return ecdsaSig, pqSig, finalSig, nil
}

func (s *TransactionSigner) SignECDSA(hashData common.Hash, privateKey []byte) ([]byte, error) {
	ecdsaSk, err := crypto.ToECDSA(privateKey)
	if err != nil {
		return nil, fmt.Errorf("invalid ECDSA private key: %w", err)
	}

	return crypto.Sign(hashData.Bytes(), ecdsaSk)
}

func (s *TransactionSigner) SignPqcSignature(algorithm string, message []byte, privateKey []byte) ([]byte, error) {
	signer := oqs.Signature{}

	// Copy before Init: oqs.Signature.Clean() zeroes secretKey in-place via
	// MemCleanse, which would corrupt the caller's slice if we passed it directly.
	privCopy := make([]byte, len(privateKey))
	copy(privCopy, privateKey)

	if err := signer.Init(algorithm, privCopy); err != nil {
		slog.Error("OQS init error", "err", err)
		return nil, err
	}

	defer signer.Clean()

	pqcSig, err := signer.Sign(message)
	if err != nil {
		return nil, err
	}

	sig := make([]byte, len(pqcSig))
	copy(sig, pqcSig)

	return sig, nil
}

func parseHybridPrivateKey(privateKey []byte) (ecdsaPart, pqcPart []byte, err error) {
	fields, err := decodeTLV(privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("hybrid private key TLV: %w", err)
	}
	ecdsaPart, err = requireField(fields, TagECDSAPriv)
	if err != nil {
		return nil, nil, err
	}
	pqcPart, err = requireField(fields, TagPQCPriv)
	if err != nil {
		return nil, nil, err
	}
	return ecdsaPart, pqcPart, nil
}
