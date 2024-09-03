package main

import (
	"github.com/janicaleksander/StocksHelp/db"
	"github.com/janicaleksander/StocksHelp/external"
	"github.com/janicaleksander/StocksHelp/httpapi"
	"github.com/janicaleksander/StocksHelp/market"
	"github.com/janicaleksander/StocksHelp/stockapi"
	"github.com/joho/godotenv"
	"log"
)

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	databaseAPI, err := db.NewDB()
	databaseAPI.Init()

	hub := stockapi.NewHub(databaseAPI)
	mockExchange := external.NewMockExchange(databaseAPI)
	m := market.NewMarket("market1", mockExchange)
	hub.SubscribeMarket(m)

	go hub.Run()
	go mockExchange.MockGenerate()

	server := httpapi.NewServer(":5050", hub)
	server.Run()

}
