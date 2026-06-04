package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	tf "smlcloudplatform/internal/datatransfer"
	"smlcloudplatform/pkg/microservice"

	"github.com/joho/godotenv"
)

var (
	holdingCode    = flag.String("holding_code", "", "holdingCode to transfer")
	toHoldingCode  = flag.String("toholding_code", "", "holdingCode to transfer")
	confirmTranser = flag.Bool("confirm", false, "confirm transfer")
)

func main() {

	// read holding_code from std in
	godotenv.Load()

	flag.Parse()

	if *holdingCode == "" {
		panic("holdingCode is required")
	}

	if confirmTranser != nil && !*confirmTranser {

		reader := bufio.NewReader(os.Stdin)

		messageToShopDisplay := ""
		if *toHoldingCode != "" {
			messageToShopDisplay = " to holdingCode: " + *toHoldingCode
		}
		fmt.Println("Are you sure to transfer holdingCode: ", *holdingCode, messageToShopDisplay, " ? (y/n)")

		text, _ := reader.ReadString('\n')

		if text != "y\n" {
			fmt.Println("Transfer is cancelled")
			return
		}
	}

	// // confirm for transfer
	sourceDBConfig := tf.SourceDatabaseConfig{}
	destinationDBConfig := tf.DestinationDatabaseConfig{}

	sourceDatabase := microservice.NewPersisterMongo(sourceDBConfig)
	targetDatabase := microservice.NewPersisterMongo(destinationDBConfig)

	dbTransfer := tf.NewDBTransfer(sourceDatabase, targetDatabase)
	dbTransfer.BeginTransfer(*holdingCode, *toHoldingCode)
}
