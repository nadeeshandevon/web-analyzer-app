package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetupServers_DefaultValues(t *testing.T) {
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("METRICS_PORT")
	os.Unsetenv("ENABLE_PPROF")
	os.Unsetenv("LOG_LEVEL")

	_, mainServer, metricsServer := setupServers()

	assert.NotNil(t, mainServer)
	assert.NotNil(t, metricsServer)
	assert.Equal(t, ":8081", mainServer.Addr)
	assert.Equal(t, ":9090", metricsServer.Addr)
}

func TestSetupServers_CustomValues(t *testing.T) {
	os.Setenv("SERVER_PORT", "9091")
	os.Setenv("METRICS_PORT", "10010")
	os.Setenv("ENABLE_PPROF", "true")
	os.Setenv("LOG_LEVEL", "debug")

	defer func() {
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("METRICS_PORT")
		os.Unsetenv("ENABLE_PPROF")
		os.Unsetenv("LOG_LEVEL")
	}()

	log, mainServer, metricsServer := setupServers()

	assert.NotNil(t, log)
	assert.NotNil(t, mainServer)
	assert.NotNil(t, metricsServer)
	assert.Equal(t, ":9091", mainServer.Addr)
	assert.Equal(t, ":10010", metricsServer.Addr)
}
