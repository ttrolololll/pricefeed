package coindesk

import (
	"net/http"
	"sync"
)

const (
	defaultBaseURL = "https://api.coindesk.com/v1"
)

var (
	defaultClient IClient
	initOnce      = &sync.Once{}
)

//go:generate mockery --inpackage --name IClient
type IClient interface {
	CurrentPrice(target string) (*CurrentPriceResponse, error)
}

func Init(baseURL string, httpClient *http.Client) {
	initOnce.Do(func() {
		if baseURL == "" {
			baseURL = defaultBaseURL
		}
		if httpClient == nil {
			httpClient = http.DefaultClient
		}
		defaultClient = &clientImpl{
			baseURL:    baseURL,
			httpClient: httpClient,
		}
	})
}

func GetClient() IClient {
	return defaultClient
}
