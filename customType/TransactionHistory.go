package customType

import "time"

type TransactionHistory struct {
	Resource        string
	Quantity        float64
	PurchasePrice   float64
	SellingPrice    float64
	Purchase        bool
	Sale            bool
	TransactionTime time.Time
}
