package external

import (
	"fmt"
	"github.com/janicaleksander/StocksHelp/db"
	"log"
	"math/rand"
	"sync"
	"time"
)

type MockExchange struct {
	Values  map[string]float64
	Mu      sync.Mutex
	storage db.Storage
}

func NewMockExchange(s db.Storage) *MockExchange {
	return &MockExchange{
		Values:  make(map[string]float64),
		Mu:      sync.Mutex{},
		storage: s,
	}
}
func (m *MockExchange) MockGenerate() {
	symbols := []string{"ZORAX", "NUVEX", "RIVEX", "YOLAX", "QUFEX", "MIREX", "DORAX", "XALOR", "VIXOR", "JUMEX"}
	b, err := m.storage.CheckFirst()
	if err != nil {
		log.Fatal(err)
	}
	if b {
		for _, s := range symbols {
			m.Values[s] = 0.0
		}
		if err := m.storage.SetDefault(m.Values); err != nil {
			fmt.Println(err)
		}
	} else {
		data, err := m.storage.GetState()
		if err != nil {
			log.Fatal(err)
		}
		for key, val := range data {
			m.Values[key] = val
		}
	}
	for {
		// mock api
		m.Mu.Lock()
		m.ChangePrice()
		m.Mu.Unlock()
		//time.Sleep(time.Hour * 1)
		time.Sleep(time.Second * 10)
	}

}
func (m *MockExchange) ChangePrice() {
	sharescount := len(m.Values)
	keys := make([]string, 0, len(m.Values))
	for key := range m.Values {
		keys = append(keys, key)
	}
	a := keys[rand.Intn(sharescount)]
	newPrice := changePriceHelper(m.Values[a])
	m.Values[a] = newPrice
	m.storage.UpdatePrice(a, newPrice)
}
func changePriceHelper(price float64) float64 {
	strategy := rand.Intn(4)
	switch strategy {
	case 0:
		return 1 + (rand.Float64()-0.5)*0.02 + price*(1+(rand.Float64()-0.5)*0.02)
	case 1:
		return 1 + (rand.Float64()-0.5)*0.02 + price*(1+(rand.Float64()-0.5)*0.06)
	case 2:
		return 1 + (rand.Float64()-0.5)*0.02 + price*(1+rand.Float64()*0.05)
	case 3:
		return 1 + (rand.Float64()-0.5)*0.02 + price*(1-rand.Float64()*0.05)
	}
	return price
}
