package coindesk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Init(t *testing.T) {
	defaultClient = nil
	Init("", nil)
	assert.NotNil(t, defaultClient)
}

func Test_GetClient(t *testing.T) {
	defaultClient = &clientImpl{}
	c := GetClient()
	assert.NotNil(t, c)
}
