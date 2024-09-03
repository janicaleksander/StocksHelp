package stockapi

import (
	"errors"
	"fmt"
	"github.com/janicaleksander/StocksHelp/db"
	"github.com/janicaleksander/StocksHelp/market"
	"log"
	"sync"
	"time"
)

// Channels :
//		c1 : buy requests
//		c2 : sell requests
//		c3 : check requests

// c1 : buy requests
// c1 : buy requests

type Payload struct {
	Address string
	Action  string
	Data    interface{}
	Date    time.Time
}

type Response struct {
	Name interface{}
	Data interface{}
	Date time.Time
}

type Hub struct {
	MarketConns    map[*market.Market]bool
	InputChannels  map[string]chan Payload
	OutputChannels map[string]chan Response
	Filters        map[string]func(interface{}) bool
	wg             sync.WaitGroup
	Storage        db.Storage
}

func NewHub(db db.Storage) *Hub {
	return &Hub{
		MarketConns:    make(map[*market.Market]bool),
		InputChannels:  make(map[string]chan Payload),
		OutputChannels: make(map[string]chan Response),
		Filters:        make(map[string]func(interface{}) bool),
		wg:             sync.WaitGroup{},
		Storage:        db,
	}
}

func (h *Hub) Run() {
	log.Print("Running hub")
	for m, ok := range h.MarketConns {
		if ok {
			log.Printf("starting %v", m.Name)
			go m.Run() // run market
		}
	}

	// setup input/output channels
	h.InputChannels["c1"] = make(chan Payload, 1024)
	h.InputChannels["c2"] = make(chan Payload, 1024)
	h.InputChannels["c3"] = make(chan Payload, 1024)

	h.PrepareOutput()
	go func() {
		for {
			for inputName, inputChan := range h.InputChannels {
				var sent bool
				var done bool
				var price float64
				select {
				case receive, ok := <-inputChan:
					if !ok {
						log.Printf("Problem with channel %v", inputChan)
						continue
					}
					//do actions with recieve
					for markets, ok := range h.MarketConns {
						if ok && !done {
							h.wg.Add(1)
							go func() {
								defer h.wg.Done()
								if markets.Name == receive.Address {
									markets.InputChan <- receive.Data.(string)
									price = <-markets.OutputChan
									done = true
								}

							}()
						}
					}
					h.wg.Wait()
					done = false
					for outputName, outputChan := range h.OutputChannels {
						filter, ok := h.Filters[outputName]
						if !ok {
							log.Printf("not have a filter")
						}
						if ok {
							if filter(inputName) {
								sent = true
								resp := Response{Name: receive.Data, Data: price, Date: time.Now().UTC()}
								outputChan <- resp
							}
						}

					}
					if !sent {
						fmt.Println("No channel found for message:")
					}
				default:
				}

			}
		}

	}()
}
func (h *Hub) SubscribeMarket(m *market.Market) error {
	h.MarketConns[m] = true
	return nil

}

func (h *Hub) MakeCurrencyRequest(marketName, action, name string) (Response, error) {
	switch action {
	case "CHECK":
		h.InputChannels["c1"] <- Payload{
			Address: marketName,
			Action:  "CHECK",
			Data:    name,
			Date:    time.Now(),
		}
		resp := <-h.OutputChannels["c1"]
		return resp, nil
	case "BUY":
		h.InputChannels["c2"] <- Payload{
			Address: marketName,
			Action:  "BUY",
			Data:    name,
			Date:    time.Now(),
		}
		resp := <-h.OutputChannels["c2"]
		return resp, nil
	case "SELL":
		h.InputChannels["c3"] <- Payload{
			Address: marketName,
			Action:  "SELL",
			Data:    name,
			Date:    time.Now(),
		}
		resp := <-h.OutputChannels["c3"]
		return resp, nil
	}

	return Response{}, errors.New("Wrong action")
}

func (h *Hub) PrepareOutput() {

	h.OutputChannels["c1"] = make(chan Response)
	h.OutputChannels["c2"] = make(chan Response)
	h.OutputChannels["c3"] = make(chan Response)

	h.Filters["c1"] = func(inputName interface{}) bool {
		if inputName.(string) == "c1" {
			return true
		} else {
			return false
		}
	}
	h.Filters["c2"] = func(inputName interface{}) bool {
		if inputName.(string) == "c2" {
			return true
		} else {
			return false
		}
	}
	h.Filters["c3"] = func(inputName interface{}) bool {
		if inputName.(string) == "c3" {
			return true
		} else {
			return false
		}
	}
	// filter ic1 -> oc1

}
