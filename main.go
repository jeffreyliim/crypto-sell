package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/adshao/go-binance/v2"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
	"sync"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func initBinanceClient() *binance.Client {
	return binance.NewClient(os.Getenv("BINANCE_API_KEY"), os.Getenv("BINANCE_SECRET_KEY"))
}

func main() {
	client := initBinanceClient()
	accSvc, err := client.NewGetAccountService().Do(context.Background())
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	wg := sync.WaitGroup{}
	balances := accSvc.Balances
	for _, bal := range balances {
		wg.Add(1)
		go func(bal binance.Balance) {
			defer wg.Done()
			free, err := strconv.ParseFloat(bal.Free, 64)
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			// can sell anything except for USDT
			if bal.Asset != "USDT" && free != 0 {
				marketSellMaxQty(client, bal)
			}
		}(bal)
	}
	wg.Wait()
}

func marketSellMaxQty(client *binance.Client, bal binance.Balance) {
	coin := bal.Asset
	orderSvc := client.NewCreateOrderService()
	res, err := orderSvc.NewOrderRespType(binance.NewOrderRespTypeFULL).
		Quantity(bal.Free).
		Side(binance.SideTypeSell).
		Type(binance.OrderTypeMarket).
		Symbol(coin + "USDT").
		Do(context.Background())
	if err != nil {
		fmt.Println(err.Error() + " " + coin)
		return
	}

	b, _ := json.Marshal(res)
	fmt.Println(string(b))
	fmt.Println("sold this coin", coin)
}
