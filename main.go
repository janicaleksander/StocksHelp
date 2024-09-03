package main

import (
	"fmt"
	"github.com/janicaleksander/StocksHelp/db"
	"github.com/janicaleksander/StocksHelp/external"
	"github.com/janicaleksander/StocksHelp/httpapi"
	"github.com/janicaleksander/StocksHelp/market"
	"github.com/janicaleksander/StocksHelp/stockapi"
	"log"
	"os"
)

func main() {
	// Load environment variables if needed
	// err := godotenv.Load()
	// if err != nil {
	// 	log.Fatalf("Error loading .env file: %v", err)
	// }

	databaseAPI, err := db.NewDB()
	if err != nil {
		log.Fatal(err)
	}
	databaseAPI.Init()

	hub := stockapi.NewHub(databaseAPI)
	mockExchange := external.NewMockExchange(databaseAPI)
	m := market.NewMarket("market1", mockExchange)
	hub.SubscribeMarket(m)

	go hub.Run()
	go mockExchange.MockGenerate()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	address := fmt.Sprintf("0.0.0.0:%s", port)
	server := httpapi.NewServer(address, hub)
	server.Run()
}
