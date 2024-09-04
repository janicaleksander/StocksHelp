package customType

import "time"

type ChartStockInfo struct {
	Name   string
	Price  float64
	TimeAt time.Time
}
