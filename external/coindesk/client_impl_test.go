package coindesk

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type noopRT struct{}

func (rt *noopRT) RoundTrip(*http.Request) (*http.Response, error) {
	resp := &CurrentPriceResponse{
		Time: &TimeFormats{
			Updated:    "",
			UpdatedISO: "",
			UpdatedUK:  "",
		},
		ChartName: "test",
	}
	b, _ := json.Marshal(resp)
	return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewBufferString(string(b)))}, nil
}

type errRT struct{}

func (rt *errRT) RoundTrip(*http.Request) (*http.Response, error) {
	return &http.Response{StatusCode: 500, Body: io.NopCloser(nil)}, assert.AnError
}

func Test_clientImpl_CurrentPrice(t *testing.T) {
	testcases := []struct {
		name         string
		roundtripper http.RoundTripper
		expectErr    bool
	}{
		{
			name:         "happy",
			roundtripper: &noopRT{},
			expectErr:    false,
		},

		{
			name:         "failed - round tripper",
			roundtripper: &errRT{},
			expectErr:    true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			httpClient := &http.Client{Transport: tc.roundtripper}
			c := &clientImpl{
				baseURL:    "base",
				httpClient: httpClient,
			}
			_, err := c.CurrentPrice("BTC")
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
