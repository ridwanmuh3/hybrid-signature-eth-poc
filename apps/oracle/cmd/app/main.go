package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	transactions "github.com/ridwanmuh3/hybrid-pqc-poc/apps/oracle/artifacts"
	"github.com/ridwanmuh3/hybrid-pqc-poc/apps/oracle/internal/config"
	"github.com/ridwanmuh3/hybrid-pqc-poc/apps/oracle/internal/route"
	"github.com/ridwanmuh3/hybrid-pqc-poc/apps/oracle/internal/util"
)

func main() {
	viperConfig := config.NewViper()
	app := config.NewFiber()

	client := config.NewGethClient(viperConfig)
	contractAddress := common.HexToAddress(viperConfig.GetString("CONTRACT_ADDRESS"))

	contractInstance, err := transactions.NewTransactions(contractAddress, client)
	if err != nil {
		log.Fatalf("failed to connect access smart contract: %v", err)
	}

	auth, _, err := util.SetupOracleAuth(viperConfig)
	if err != nil {
		log.Fatalf("failed to setup oracle auth: %v", err)
	}
	if err := util.ValidateOracleAuth(contractInstance, auth); err != nil {
		log.Fatalf("failed oracle auth validation: %v", err)
	}

	// middleware
	app.Use(recover.New())
	app.Use(cors.New(cors.ConfigDefault))
	app.Use(logger.New())

	route.Setup(app, viperConfig, client, contractInstance)

	// Graceful shutdown on SIGINT / SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("Shutting down oracle server...")
		if err := app.Shutdown(); err != nil {
			log.Printf("server shutdown error: %v", err)
		}
	}()

	log.Println("Oracle server listening on :9000")
	if err = app.Listen(":9000"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
