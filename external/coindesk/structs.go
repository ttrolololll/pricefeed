package coindesk

type CurrentPriceResponse struct {
	Time      *TimeFormats                 `json:"time"`
	ChartName string                       `json:"chartName"`
	BPI       map[string]*CurrencyRateData `json:"bpi"`
}

type TimeFormats struct {
	Updated    string `json:"updated"`
	UpdatedISO string `json:"updatedISO"`
	UpdatedUK  string `json:"updateduk"`
}

type CurrencyRateData struct {
	Code        string  `json:"code"`
	Symbol      string  `json:"symbol"`
	Rate        string  `json:"rate"`
	Description string  `json:"description"`
	RateFloat   float64 `json:"rate_float"`
}
