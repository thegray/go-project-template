package shared

type Money struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

func NewMoney(amount int64, currency string) Money {
	if currency == "" {
		currency = "USD"
	}
	return Money{Amount: amount, Currency: currency}
}
