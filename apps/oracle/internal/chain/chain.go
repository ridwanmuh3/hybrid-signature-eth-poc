package chain

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	transactions "github.com/ridwanmuh3/hybrid-pqc-poc/apps/oracle/artifacts"
)

type TransactionDecoder struct {
	ContractABI abi.ABI
}

func NewTransactionDecoder() (*TransactionDecoder, error) {
	contractABI, err := abi.JSON(strings.NewReader(transactions.TransactionsABI))
	if err != nil {
		return nil, fmt.Errorf("invalid ABI: %w", err)
	}

	return &TransactionDecoder{ContractABI: contractABI}, nil
}

func (d *TransactionDecoder) Decode(tx *types.Transaction) (*transactions.TransactionsTxParams, error) {
	if len(tx.Data()) < 4 {
		return nil, fmt.Errorf("invalid transaction data")
	}

	method, err := d.ContractABI.MethodById(tx.Data()[:4])
	if err != nil {
		return nil, fmt.Errorf("method not found in ABI: %w", err)
	}

	if method.Name != "sendTransaction" {
		return nil, fmt.Errorf("transaction is not a sendTransaction call")
	}

	unpackedResult, err := method.Inputs.Unpack(tx.Data()[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack tx data: %w", err)
	}

	if len(unpackedResult) == 0 {
		return nil, fmt.Errorf("empty result")
	}

	return d.MapToParams(unpackedResult)
}

func (d *TransactionDecoder) MapToParams(unpackedResult []any) (*transactions.TransactionsTxParams, error) {
	params := &transactions.TransactionsTxParams{}

	if len(unpackedResult) >= 9 {
		params.Nonce = unpackedResult[0].(*big.Int)
		params.Sender = unpackedResult[1].(common.Address)
		params.Receiver = unpackedResult[2].(common.Address)
		params.SigningAlgorithm = unpackedResult[3].(string)
		params.SigningMode = unpackedResult[4].(string)
		params.Message = unpackedResult[5].(string)
		params.Amount = unpackedResult[6].(*big.Int)
		params.EcdsaSignature = unpackedResult[7].([]byte)
		params.PqSignature = unpackedResult[8].([]byte)
		return params, nil
	}

	tempJSON, err := json.Marshal(unpackedResult[0])
	if err != nil {
		return nil, fmt.Errorf("failed to marshal temp data: %w", err)
	}

	if err := json.Unmarshal(tempJSON, params); err != nil {
		return nil, fmt.Errorf("failed to map data structure: %w", err)
	}

	return params, nil
}
