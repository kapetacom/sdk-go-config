// Copyright 2023 Kapeta Inc.
// SPDX-License-Identifier: MIT

package providers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func hostAndFromURL(url string) (string, string) {
	hostAndPort := url[7:]
	return strings.Split(hostAndPort, ":")[0], strings.Split(hostAndPort, ":")[1]
}
func TestLocal(t *testing.T) {
	// Mock environment variables for testing
	provider := NewLocalConfigProvider("kapeta/block-type-gateway-http", "systemID", "instanceID", map[string]interface{}{})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("40004"))
	}))
	defer srv.Close()

	host, port := hostAndFromURL(srv.URL)
	os.Setenv("KAPETA_LOCAL_CLUSTER_HOST", host)
	os.Setenv("KAPETA_LOCAL_CLUSTER_PORT", port)
	defer os.Unsetenv("KAPETA_LOCAL_CLUSTER_HOST")
	defer os.Unsetenv("KAPETA_LOCAL_CLUSTER_PORT")

	serverPort, err := provider.GetServerPort("http")
	assert.NoError(t, err)
	assert.Equal(t, "40004", serverPort)

}
func TestCreateLocalConfigProvider(t *testing.T) {
	blockRef := "block-ref"
	systemID := "system-id"
	instanceID := "instance-id"
	blockDefinition := map[string]interface{}{
		"type": "local",
	}

	provider := NewLocalConfigProvider(blockRef, systemID, instanceID, blockDefinition)

	assert.Equal(t, blockRef, provider.GetBlockReference())
	assert.Equal(t, systemID, provider.GetSystemId())
	assert.Equal(t, instanceID, provider.GetInstanceId())
	assert.Equal(t, "local", provider.GetProviderId())
}

func TestResolveIdentity(t *testing.T) {
	os.Setenv("KAPETA_ENVIRONMENT_TYPE", "process")
	os.Setenv("KAPETA_BLOCK", "block-ref")
	os.Setenv("KAPETA_SYSTEM", "system-id")
	os.Setenv("KAPETA_INSTANCE", "instance-id")
	defer func() {
		os.Unsetenv("KAPETA_ENVIRONMENT_TYPE")
		os.Unsetenv("KAPETA_BLOCK")
		os.Unsetenv("KAPETA_SYSTEM")
		os.Unsetenv("KAPETA_INSTANCE")
	}()
	provider := NewLocalConfigProvider("block-ref", "system-id", "instance-id", map[string]interface{}{
		"type": "local",
	})

	assert.Equal(t, "system-id", provider.GetSystemId())
	assert.Equal(t, "instance-id", provider.GetInstanceId())
}

func TestLoadConfiguration(t *testing.T) {
	// create test server that return the correct values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/rest") {
			_, _ = w.Write([]byte("8080"))
		} else if strings.HasSuffix(r.URL.Path, "/grpc") {
			_, _ = w.Write([]byte("8081"))
		}
	}))
	defer srv.Close()

	host, port := hostAndFromURL(srv.URL)
	os.Setenv("KAPETA_LOCAL_CLUSTER_HOST", host)
	os.Setenv("KAPETA_LOCAL_CLUSTER_PORT", port)

	defer os.Unsetenv("KAPETA_LOCAL_CLUSTER_HOST")
	defer os.Unsetenv("KAPETA_LOCAL_CLUSTER_PORT")

	provider := NewLocalConfigProvider("block-ref", "system-id", "instance-id", map[string]interface{}{
		"type": "local",
	})

	port, err := provider.GetServerPort("rest")
	assert.NoError(t, err)
	assert.Equal(t, "8080", port)

	port, err = provider.GetServerPort("grpc")
	assert.NoError(t, err)
	assert.Equal(t, "8081", port)

	testhost, err := provider.GetServerHost()
	assert.NoError(t, err)
	assert.Equal(t, "127.0.0.1", testhost)
}

func TestGetServiceAddress1(t *testing.T) {
	// create test server that return the correct values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "baz") {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/rest") {
			_, _ = w.Write([]byte("10.0.0.1:8080"))
			return
		}
		if strings.HasSuffix(r.URL.Path, "/grpc") {
			_, _ = w.Write([]byte("10.0.0.2:8081"))
			return
		}
	}))
	defer srv.Close()

	host, port := hostAndFromURL(srv.URL)
	os.Setenv("KAPETA_LOCAL_CLUSTER_HOST", host)
	os.Setenv("KAPETA_LOCAL_CLUSTER_PORT", port)

	defer os.Unsetenv("KAPETA_LOCAL_CLUSTER_HOST")
	defer os.Unsetenv("KAPETA_LOCAL_CLUSTER_PORT")

	provider := NewLocalConfigProvider("block-ref", "kapeta://soren_mathiasen/java-cloud-bucket:local", "instance-id", map[string]interface{}{
		"type": "local",
	})

	address, err := provider.GetServiceAddress("foo", "rest")
	assert.NoError(t, err)
	assert.Equal(t, "10.0.0.1:8080", address)

	address, err = provider.GetServiceAddress("bar", "grpc")
	assert.NoError(t, err)
	assert.Equal(t, "10.0.0.2:8081", address)

	_, err = provider.GetServiceAddress("baz", "rest")
	assert.Error(t, err)
	assert.Equal(t, "failed to send GET request: request failed - Status: 500", err.Error())
}

func TestGetInstanceHost1(t *testing.T) {
	// create test server that return the correct values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "system-id/unknown-instance-id") {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if strings.Contains(r.URL.Path, "/instances/system-id/") {
			_, _ = w.Write([]byte("10.0.0.1"))
			return
		}
	}))
	defer srv.Close()

	host, port := hostAndFromURL(srv.URL)
	os.Setenv("KAPETA_LOCAL_CLUSTER_HOST", host)
	os.Setenv("KAPETA_LOCAL_CLUSTER_PORT", port)

	defer os.Unsetenv("KAPETA_LOCAL_CLUSTER_HOST")
	defer os.Unsetenv("KAPETA_LOCAL_CLUSTER_PORT")

	os.Setenv("KAPETA_INSTANCE_CONFIG", "{\"foo\": \"bar\"}")
	defer os.Unsetenv("KAPETA_INSTANCE_CONFIG")
	provider := NewLocalConfigProvider("block-ref", "system-id", "instance-id", map[string]interface{}{
		"type": "local",
	})

	host, err := provider.GetInstanceHost("instance-id")
	assert.NoError(t, err)
	assert.Equal(t, "10.0.0.1", host)

	_, err = provider.GetInstanceHost("unknown-instance-id")
	assert.Error(t, err)
	assert.Equal(t, "failed to send GET request: request failed - Status: 500", err.Error())
}
