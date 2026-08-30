package util

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func HashEthSignedData(
	chainID *big.Int,
	nonce *big.Int,
	sender common.Address,
	receiver common.Address,
	signingAlgorithm string,
	signingMode string,
	message string,
	amount *big.Int,
) (common.Hash, error) {
	args := abi.Arguments{
		{Type: abi.Type{T: abi.UintTy, Size: 256}},
		{Type: abi.Type{T: abi.UintTy, Size: 256}},
		{Type: abi.Type{T: abi.AddressTy}},
		{Type: abi.Type{T: abi.AddressTy}},
		{Type: abi.Type{T: abi.StringTy}},
		{Type: abi.Type{T: abi.StringTy}},
		{Type: abi.Type{T: abi.StringTy}},
		{Type: abi.Type{T: abi.UintTy, Size: 256}},
	}

	packed, err := args.Pack(
		chainID,
		nonce,
		sender,
		receiver,
		signingAlgorithm,
		signingMode,
		message,
		amount,
	)
	if err != nil {
		return common.Hash{}, err
	}

	messageHash := crypto.Keccak256Hash(packed)

	ethSignedMessageHash := crypto.Keccak256Hash(
		[]byte("\x19ETHEREUM-ORACLE-SIGNED:\n32"),
		messageHash.Bytes(),
	)

	return ethSignedMessageHash, nil
}
