package user

type Wallet struct {
	Money  float64
	Assets []struct {
		name     string
		quantity float64
		value    float64
	}
}
