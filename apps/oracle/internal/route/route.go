package route

import (
	"runtime"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/spf13/viper"

	transactions "github.com/ridwanmuh3/hybrid-pqc-poc/apps/oracle/artifacts"
	"github.com/ridwanmuh3/hybrid-pqc-poc/apps/oracle/internal/handler"
)

func Setup(app *fiber.App, viperConfig *viper.Viper, client *ethclient.Client, contractInstance *transactions.Transactions) {
	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "hybrid-pqc-oracle",
			"go":      runtime.Version(),
			"time":    time.Now().UTC().Format(time.RFC3339),
		})
	})

	// Rate limiter: 30 requests per second per IP, with burst up to 60
	api := app.Group("/api", limiter.New(limiter.Config{
		Max:        30,
		Expiration: 1 * time.Second,
	}))

	api.Post("/generate-key", handler.GenerateKeypairHandler)
	api.Post("/sign", handler.SendAndSignTransactionHandler(viperConfig, client, contractInstance))
	api.Get("/transactions/:txhash", handler.GetTransactionByHashHandler(client))
	api.Post("/verify", handler.VerifyTransactionByHashHandler(client))
}
