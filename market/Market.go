package market

import (
	"github.com/janicaleksander/StocksHelp/external"
	"time"
)

type Market struct {
	Name       string
	Ext        *external.MockExchange
	InputChan  chan string
	OutputChan chan float64
}

func NewMarket(name string, ext *external.MockExchange) *Market {
	return &Market{
		Name:       name,
		Ext:        ext,
		InputChan:  make(chan string, 1024),
		OutputChan: make(chan float64, 1024),
	}
}
func (m *Market) Run() {
	for {
		select {
		case cName := <-m.InputChan:
			go func() {
				m.Ext.Mu.Lock()
				price, ok := m.Ext.Values[cName]
				m.Ext.Mu.Unlock()
				if ok {
					m.OutputChan <- price
				}
				//m.OutputChan <- price
			}()
		default:
			time.Sleep(time.Millisecond * 100)
		}
	}
}
