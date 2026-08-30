package util

import "encoding/pem"

func KeyToPemFormat(algorithm string, keyType string, key []byte) []byte {
	keyPemBlock := &pem.Block{
		Type:  algorithm + " " + keyType,
		Bytes: key,
	}

	return pem.EncodeToMemory(keyPemBlock)
}
