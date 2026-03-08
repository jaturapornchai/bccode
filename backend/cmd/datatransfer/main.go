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
	shopID         = flag.String("shopid", "", "shopID to transfer")
	toShopID       = flag.String("toshopid", "", "shopID to transfer")
	confirmTranser = flag.Bool("confirm", false, "confirm transfer")
)

func main() {

	// read shopid from std in
	godotenv.Load()

	flag.Parse()

	if *shopID == "" {
		panic("shopID is required")
	}

	if confirmTranser != nil && !*confirmTranser {

		reader := bufio.NewReader(os.Stdin)

		messageToShopDisplay := ""
		if *toShopID != "" {
			messageToShopDisplay = " to shopID: " + *toShopID
		}
		fmt.Println("Are you sure to transfer shopID: ", *shopID, messageToShopDisplay, " ? (y/n)")

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
	dbTransfer.BeginTransfer(*shopID, *toShopID)
}
