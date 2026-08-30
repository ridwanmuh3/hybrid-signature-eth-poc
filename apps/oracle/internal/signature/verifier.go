package signature

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/open-quantum-safe/liboqs-go/oqs"
)

type SignatureVerifier struct {
	algorithm string
	mode      string
}

func NewSignatureVerifier(algorithm, mode string) *SignatureVerifier {
	return &SignatureVerifier{
		algorithm: algorithm,
		mode:      mode,
	}
}

func (v *SignatureVerifier) Verify(hashData common.Hash, signature, publicKey []byte, sender common.Address) (bool, string) {
	switch v.mode {
	case "single":
		return v.VerifySingle(hashData, signature, publicKey, sender)
	case "hybrid":
		return v.VerifyHybrid(hashData, signature, publicKey, sender)
	default:
		return false, "Invalid verification mode"
	}
}

func (v *SignatureVerifier) VerifySingle(hashData common.Hash, signature, publicKey []byte, sender common.Address) (bool, string) {
	if v.algorithm == "ECDSA" {
		return v.VerifyECDSA(hashData, signature, publicKey, sender)
	}
	return v.VerifyPQC(hashData, signature, publicKey)
}

func (v *SignatureVerifier) VerifyECDSA(hashData common.Hash, signature, publicKey []byte, sender common.Address) (bool, string) {
	if len(signature) != 65 {
		return false, "Invalid Secp256k1 signature length"
	}

	sigPublicKey, err := crypto.SigToPub(hashData.Bytes(), signature)
	if err != nil {
		return false, "Secp256k1 recovery failed"
	}

	recoveredAddress := crypto.PubkeyToAddress(*sigPublicKey)

	expectedAddress, err := crypto.UnmarshalPubkey(publicKey)
	if err != nil {
		return false, "Invalid Secp256k1 public key"
	}
	derivedAddress := crypto.PubkeyToAddress(*expectedAddress)

	if recoveredAddress == sender && recoveredAddress == derivedAddress {
		return true, "Valid Standard Secp256k1 Signature (Matches Sender)"
	}

	return false, fmt.Sprintf("Invalid Secp256k1: Recovered %s != Sender %s", recoveredAddress.Hex(), sender.Hex())
}

func (v *SignatureVerifier) VerifyPQC(hashData common.Hash, signature, publicKey []byte) (bool, string) {
	if len(signature) == 0 || len(publicKey) == 0 {
		return false, "Empty PQC signature or public key"
	}

	verifier := oqs.Signature{}

	if err := verifier.Init(v.algorithm, nil); err != nil {
		return false, "OQS initialization failed"
	}

	defer verifier.Clean()

	details := verifier.Details()
	if len(publicKey) != details.LengthPublicKey {
		return false, fmt.Sprintf("Invalid PQC public key length: got %d, want %d", len(publicKey), details.LengthPublicKey)
	}
	if len(signature) > details.MaxLengthSignature {
		return false, fmt.Sprintf("Invalid PQC signature length: got %d, max %d", len(signature), details.MaxLengthSignature)
	}

	result, err := verifier.Verify(hashData.Bytes(), signature, publicKey)
	if err != nil {
		return false, "OQS verification error"
	}

	if result {
		return true, fmt.Sprintf("Valid Post-Quantum (%s) Signature", v.algorithm)
	}

	return false, "Invalid Post-Quantum Signature"
}

func (v *SignatureVerifier) VerifyHybrid(hashData common.Hash, signature, publicKey []byte, sender common.Address) (bool, string) {
	sigFields, err := decodeTLV(signature)
	if err != nil {
		return false, fmt.Sprintf("hybrid signature TLV: %v", err)
	}
	pubFields, err := decodeTLV(publicKey)
	if err != nil {
		return false, fmt.Sprintf("hybrid public key TLV: %v", err)
	}

	ecdsaSig, err := requireField(sigFields, TagECDSASig)
	if err != nil {
		return false, fmt.Sprintf("missing ECDSA sig: %v", err)
	}
	pqcSig, err := requireField(sigFields, TagPQCSig)
	if err != nil {
		return false, fmt.Sprintf("missing PQC sig: %v", err)
	}
	ecdsaPub, err := requireField(pubFields, TagECDSAPub)
	if err != nil {
		return false, fmt.Sprintf("missing ECDSA pub: %v", err)
	}
	pqcPub, err := requireField(pubFields, TagPQCPub)
	if err != nil {
		return false, fmt.Sprintf("missing PQC pub: %v", err)
	}

	// Algorithm binding: sig and pub blobs must declare the same algorithm,
	// and both must match what this verifier was configured with.
	sigAlgo := string(sigFields[TagAlgorithm])
	pubAlgo := string(pubFields[TagAlgorithm])
	if sigAlgo != "" && pubAlgo != "" && sigAlgo != pubAlgo {
		return false, fmt.Sprintf("algorithm mismatch: sig=%s pub=%s", sigAlgo, pubAlgo)
	}
	if sigAlgo != "" && sigAlgo != v.algorithm {
		return false, fmt.Sprintf("sig algorithm %q does not match verifier %q", sigAlgo, v.algorithm)
	}

	parsedAlgorithm := strings.TrimPrefix(v.algorithm, "Hybrid-Secp256k1-")

	ecdsaValid := v.VerifyHybridECDSA(hashData, ecdsaSig, ecdsaPub, sender)
	pqsValid := v.VerifyHybridPQC(hashData, pqcSig, pqcPub, parsedAlgorithm)

	if ecdsaValid && pqsValid {
		return true, fmt.Sprintf("Valid Hybrid Signature (Secp256k1 + %s)", parsedAlgorithm)
	}
	return false, fmt.Sprintf("Hybrid Failed -> ECDSA: %v, PQS: %v", ecdsaValid, pqsValid)
}

func (v *SignatureVerifier) VerifyHybridECDSA(hashData common.Hash, signature, publicKey []byte, sender common.Address) bool {
	sigPublicKey, err := crypto.SigToPub(hashData.Bytes(), signature)
	if err != nil {
		return false
	}

	recoveredAddress := crypto.PubkeyToAddress(*sigPublicKey)
	expectedAddress, err := crypto.UnmarshalPubkey(publicKey)
	if err != nil {
		return false
	}

	derivedAddress := crypto.PubkeyToAddress(*expectedAddress)
	return recoveredAddress == sender && recoveredAddress == derivedAddress
}

func (v *SignatureVerifier) VerifyHybridPQC(hashData common.Hash, signature, publicKey []byte, algorithm string) bool {
	if len(signature) == 0 || len(publicKey) == 0 {
		return false
	}

	verifier := oqs.Signature{}

	if err := verifier.Init(algorithm, nil); err != nil {
		return false
	}

	defer verifier.Clean()

	valid, err := verifier.Verify(hashData.Bytes(), signature, publicKey)
	return err == nil && valid
}
