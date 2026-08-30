package config

import (
	"log"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/spf13/viper"
)

func NewGethClient(config *viper.Viper) *ethclient.Client {
	client, err := ethclient.Dial(config.GetString("ETHEREUM_URL"))
	if err != nil {
		log.Fatalf("failed to dial ethereum connection: %v", err)
	}
	return client
}
