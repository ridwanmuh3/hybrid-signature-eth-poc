package model

import (
	"fmt"
	"strings"
)

// SupportedAlgorithms is the allowlist of PQC algorithms the oracle accepts.
var SupportedAlgorithms = map[string]bool{
	"ML-DSA-44": true, "ML-DSA-65": true, "ML-DSA-87": true,
	"Falcon-512": true, "Falcon-1024": true,
	"MAYO-1": true, "MAYO-3": true, "MAYO-5": true,
	"SNOVA_24_5_4": true, "SNOVA_37_8_4": true, "SNOVA_60_10_4": true,
}

type GenerateKeypairRequest struct {
	Algorithm       string `json:"algorithm"`
	Mode            string `json:"mode"`
	EcdsaPrivateKey string `json:"ecdsa_private_key,omitempty"`
}

func (r *GenerateKeypairRequest) Validate() error {
	if r.Mode != "single" && r.Mode != "hybrid" {
		return fmt.Errorf("mode must be 'single' or 'hybrid', got %q", r.Mode)
	}
	if r.Mode == "hybrid" {
		if r.Algorithm == "" || r.Algorithm == "ECDSA" {
			return fmt.Errorf("hybrid mode requires a PQC algorithm, not %q", r.Algorithm)
		}
	}
	if r.Mode == "single" && r.Algorithm != "ECDSA" {
		if !SupportedAlgorithms[r.Algorithm] {
			return fmt.Errorf("unsupported algorithm %q", r.Algorithm)
		}
	}
	return nil
}

type SendAndSignTxRequest struct {
	Sender     string `json:"sender"`
	Receiver   string `json:"receiver"`
	Message    string `json:"message"`
	Amount     string `json:"amount"`
	PrivateKey string `json:"private_key"`
	Algorithm  string `json:"algorithm"`
	Mode       string `json:"mode"`
}

func (r *SendAndSignTxRequest) Validate() error {
	var errs []string
	if r.Sender == "" || !strings.HasPrefix(r.Sender, "0x") || len(r.Sender) != 42 {
		errs = append(errs, "sender must be a 0x-prefixed 40-hex-char address")
	}
	if r.Receiver == "" || !strings.HasPrefix(r.Receiver, "0x") || len(r.Receiver) != 42 {
		errs = append(errs, "receiver must be a 0x-prefixed 40-hex-char address")
	}
	if r.Amount == "" {
		errs = append(errs, "amount is required")
	}
	if r.PrivateKey == "" {
		errs = append(errs, "private_key is required")
	}
	if r.Mode != "single" && r.Mode != "hybrid" {
		errs = append(errs, "mode must be 'single' or 'hybrid'")
	}
	if r.Mode == "hybrid" {
		if !strings.HasPrefix(r.Algorithm, "Hybrid-Secp256k1-") {
			errs = append(errs, "hybrid mode requires algorithm prefixed with 'Hybrid-Secp256k1-'")
		}
	} else if r.Algorithm == "" {
		errs = append(errs, "algorithm is required")
	}
	if len(errs) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(errs, "; "))
	}
	return nil
}

type VerifyByHashRequest struct {
	TxHash    string `json:"tx_hash"`
	PublicKey string `json:"public_key"`
}

func (r *VerifyByHashRequest) Validate() error {
	var errs []string
	if r.TxHash == "" || !strings.HasPrefix(r.TxHash, "0x") {
		errs = append(errs, "tx_hash must be a 0x-prefixed hash")
	}
	if r.PublicKey == "" || !strings.HasPrefix(r.PublicKey, "0x") {
		errs = append(errs, "public_key must be a 0x-prefixed hex string")
	}
	if len(errs) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(errs, "; "))
	}
	return nil
}

type VerifyTxRequest struct {
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
	Amount   string `json:"amount"`
	Message  string `json:"message"`
	Nonce    uint64 `json:"nonce"`

	Signature string `json:"signature"`
	PublicKey string `json:"public_key"`

	Mode      string `json:"mode"`
	Algorithm string `json:"algorithm"`
}
