package coindesk

import (
	"encoding/json"
	"io"
	"net/http"
)

type clientImpl struct {
	httpClient *http.Client
	baseURL    string
}

var _ IClient = (*clientImpl)(nil)

// CurrentPrice retrieves the current price of target coin/token
func (c *clientImpl) CurrentPrice(target string) (*CurrentPriceResponse, error) {
	// param target does nothing for the purpose of the exercise

	rawResp, err := c.httpClient.Get(c.baseURL + "/bpi/currentprice.json")
	if err != nil {
		return nil, err
	}

	respData, err := io.ReadAll(rawResp.Body)
	if err != nil {
		return nil, err
	}

	resp := &CurrentPriceResponse{}
	err = json.Unmarshal(respData, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
