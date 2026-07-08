package main

import (
	"fmt"
	"os"
	"smlcloudplatform/docs"
	"smlcloudplatform/internal/config"

	"smlcloudplatform/pkg/microservice"

	purchaseorder_consumer "smlcloudplatform/internal/transaction/transactionconsumer/purchaseorder"
	purchasereceive_consumer "smlcloudplatform/internal/transaction/transactionconsumer/purchasereceive"
	purchaserequisition_consumer "smlcloudplatform/internal/transaction/transactionconsumer/purchaserequisition"
	rfq_consumer "smlcloudplatform/internal/transaction/transactionconsumer/rfq"

	"github.com/joho/godotenv"
)

// init loads .env files for local development.
// NOTE: This is a legacy consumer that still uses godotenv (not bootstrap.json).
// Keep godotenv for backward compatibility until migrated to bootstrap.json config.
func init() {
	env := os.Getenv("MODE")
	if env == "" {
		os.Setenv("MODE", "development")
		env = "development"
	}

	godotenv.Load(".env." + env + ".local")
	if env != "test" {
		godotenv.Load(".env.local")
	}
	godotenv.Load(".env." + env)
	godotenv.Load()
}

// @title           SML Cloud Platform API
// @version         1.0
// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @securityDefinitions.apikey  AccessToken
// @in                          header
// @name                        Authorization

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html
// @schemes http https
func main() {

	devApiMode := os.Getenv("DEV_API_MODE")
	host := os.Getenv("HOST_API")
	if host != "" {
		fmt.Printf("Host: %v\n", host)
		docs.SwaggerInfo.Host = host
	}

	cfg := config.NewConfig()
	ms, err := microservice.NewMicroservice(cfg)
	if err != nil {
		fmt.Printf("ERROR: Failed to initialize microservice: %v\n", err)
		os.Exit(1)
	}

	ms.HttpUsePrometheus()

	if devApiMode == "1" || devApiMode == "2" {

		ms.RegisterLivenessProbeEndpoint("/healthz")

		// purchase
		ms.RegisterConsumer(purchaseorder_consumer.InitPurchaseOrderTransactionConsumer(ms, cfg))
		ms.RegisterConsumer(purchaserequisition_consumer.InitPurchaseRequisitionTransactionConsumer(ms, cfg))
		ms.RegisterConsumer(rfq_consumer.InitRFQTransactionConsumer(ms, cfg))
		ms.RegisterConsumer(purchasereceive_consumer.InitPurchaseReceiveTransactionConsumer(ms, cfg))

	}

	ms.Start()
}
