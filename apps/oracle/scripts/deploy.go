package main

import (
	"context"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	transactions "github.com/ridwanmuh3/hybrid-pqc-poc/apps/oracle/artifacts"
	"github.com/ridwanmuh3/hybrid-pqc-poc/apps/oracle/internal/config"
)

func main() {
	viperConfig := config.NewViper()

	client, err := ethclient.Dial(viperConfig.GetString("ETHEREUM_URL"))
	if err != nil {
		log.Fatalf("failed to dial ethereum: %v", err)
	}

	skBytes, err := hexutil.Decode(viperConfig.GetString("ORACLE_PRIVATE_KEY"))
	if err != nil {
		log.Fatalf("failed to decode oracle private key: %v", err)
	}

	privateKey, err := crypto.ToECDSA(skBytes)
	if err != nil {
		log.Fatalf("failed to convert ecdsa from hex: %v", err)
	}

	oracleAddress := common.HexToAddress(viperConfig.GetString("ORACLE_ADDRESS"))

	chainId := big.NewInt(viperConfig.GetInt64("CHAIN_ID"))
	nonce, err := client.PendingNonceAt(context.Background(), oracleAddress)
	if err != nil {
		log.Fatalf("failed to get pending nonce: %v", err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainId)
	if err != nil {
		log.Fatalf("failed to create transactor: %v", err)
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)

	address, _, _, err := transactions.DeployTransactions(auth, client, oracleAddress)
	if err != nil {
		log.Fatalf("failed to deploy contract: %v", err)
	}

	log.Printf("contact address: %v", address.Hex())
}
