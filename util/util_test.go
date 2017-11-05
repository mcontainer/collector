package util

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestSetAggregatorEndpointWithNoEnv(t *testing.T) {
	a := "127.0.0.1:1000"
	SetAggregatorEndpoint(&a)
	assert.Equal(t, "127.0.0.1:1000", a)
}

func TestSetAggregatorEndpointWithEnv(t *testing.T) {
	const KEY = "AGGREGATOR"
	a := "127.0.0.1:1000"
	os.Setenv(KEY, "192.168.1.1:2000")
	SetAggregatorEndpoint(&a)
	assert.Equal(t, os.Getenv(KEY), a)
	os.Unsetenv(KEY)
}
