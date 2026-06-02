package fx

// Convert applies a static demo rate table (base USD).
func Convert(amount float64, from, to string) (float64, float64, error) {
	if from == to {
		return amount, 1.0, nil
	}
	rate := rateTable[from+"/"+to]
	if rate == 0 {
		rate = 1.0 / rateTable[to+"/"+from]
		if rate == 0 {
			rate = 1.0
		}
	}
	return amount * rate, rate, nil
}

var rateTable = map[string]float64{
	"USD/EUR": 0.92,
	"USD/INR": 83.0,
	"USD/GBP": 0.79,
	"EUR/USD": 1.09,
	"INR/USD": 0.012,
}
