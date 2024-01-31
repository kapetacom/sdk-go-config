// Copyright 2023 Kapeta Inc.
// SPDX-License-Identifier: MIT

package providers

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKubernetesConfigProvider(t *testing.T) {
	// Mock environment variables for testing
	os.Setenv("KAPETA_PROVIDER_PORT_REST", "8080")
	os.Setenv("KAPETA_PROVIDER_HOST", "localhost")
	os.Setenv("KAPETA_CONSUMER_SERVICE_TEST_SERVICE_REST", "http://test-service:8080")
	os.Setenv("KAPETA_CONSUMER_RESOURCE_TEST_RESOURCE_REST", `{"host": "test-resource", "port": 9090, "type": "test", "protocol": "http"}`)
	os.Setenv("KAPETA_INSTANCE_CONFIG", `{"exampleField": "exampleValue"}`)
	os.Setenv("KAPETA_BLOCK_HOSTS", `{"test-instance": "test-host"}`)

	// Create an instance of KubernetesConfigProvider
	configProvider := NewKubernetesConfigProvider("blockRef", "systemID", "instanceID", map[string]interface{}{})

	// Test GetServerPort
	serverPort, err := configProvider.GetServerPort("rest")
	assert.NoError(t, err)
	assert.Equal(t, "8080", serverPort)

	// Test GetServerHost
	serverHost, err := configProvider.GetServerHost()
	assert.NoError(t, err)
	assert.Equal(t, "localhost", serverHost)

	// Test GetServiceAddress
	serviceAddress, err := configProvider.GetServiceAddress("test-service", "rest")
	assert.NoError(t, err)
	assert.Equal(t, "http://test-service:8080", serviceAddress)

	// Test GetResourceInfo
	resourceInfo, err := configProvider.GetResourceInfo("test", "rest", "test-resource")
	assert.NoError(t, err)
	assert.NotNil(t, resourceInfo)
	assert.Equal(t, "test-resource", resourceInfo.Host)

	// Test Get and GetOrDefault
	value := configProvider.Get("exampleField")
	assert.Equal(t, "exampleValue", value)

	defaultValue := configProvider.GetOrDefault("nonexistentField", "default")
	assert.Equal(t, "default", defaultValue)

	// Test GetInstanceHost
	instanceHost, err := configProvider.GetInstanceHost("test-instance")
	assert.NoError(t, err)
	assert.Equal(t, "test-host", instanceHost)
}

func TestNewKubernetesConfigProvider(t *testing.T) {
	blockRef := "block-ref"
	systemID := "system-id"
	instanceID := "instance-id"
	blockDefinition := map[string]interface{}{
		"type": "kubernetes",
	}

	provider := NewKubernetesConfigProvider(blockRef, systemID, instanceID, blockDefinition)

	assert.Equal(t, blockRef, provider.GetBlockReference())
	assert.Equal(t, systemID, provider.GetSystemId())
	assert.Equal(t, instanceID, provider.GetInstanceId())
	assert.Equal(t, "kubernetes", provider.GetProviderId())
}

func TestGetServerPort(t *testing.T) {
	os.Setenv("KAPETA_PROVIDER_PORT_REST", "8080")
	os.Setenv("KAPETA_PROVIDER_PORT_GRPC", "8081")

	provider := NewKubernetesConfigProvider("block-ref", "system-id", "instance-id", map[string]interface{}{
		"type": "kubernetes",
	})

	port, err := provider.GetServerPort("rest")
	assert.NoError(t, err)
	assert.Equal(t, "8080", port)

	port, err = provider.GetServerPort("grpc")
	assert.NoError(t, err)
	assert.Equal(t, "8081", port)

	port, err = provider.GetServerPort("") // default to rest
	assert.NoError(t, err)
	assert.Equal(t, "8080", port)
}

func TestGetServerHost(t *testing.T) {
	os.Setenv("KAPETA_PROVIDER_HOST", "0.0.0.0")

	provider := NewKubernetesConfigProvider("block-ref", "system-id", "instance-id", map[string]interface{}{
		"type": "kubernetes",
	})

	host, err := provider.GetServerHost()
	assert.NoError(t, err)
	assert.Equal(t, "0.0.0.0", host)
}

func TestGetServiceAddress(t *testing.T) {
	os.Setenv("KAPETA_CONSUMER_SERVICE_FOO_REST", "10.0.0.1:8080")
	os.Setenv("KAPETA_CONSUMER_SERVICE_BAR_GRPC", "10.0.0.2:8081")

	provider := NewKubernetesConfigProvider("block-ref", "system-id", "instance-id", map[string]interface{}{
		"type": "kubernetes",
	})

	address, err := provider.GetServiceAddress("foo", "rest")
	assert.NoError(t, err)
	assert.Equal(t, "10.0.0.1:8080", address)

	address, err = provider.GetServiceAddress("bar", "grpc")
	assert.NoError(t, err)
	assert.Equal(t, "10.0.0.2:8081", address)

	_, err = provider.GetServiceAddress("baz", "rest")
	assert.Error(t, err)
	assert.Equal(t, "missing environment variable for internal resource: KAPETA_CONSUMER_SERVICE_BAZ_REST", err.Error())
}

func TestGetResourceInfo(t *testing.T) {
	os.Setenv("KAPETA_CONSUMER_RESOURCE_FOO_REST", "{\"host\": \"10.0.0.1\", \"port\": 8080}")
	os.Setenv("KAPETA_CONSUMER_RESOURCE_BAR_GRPC", "{\"host\": \"10.0.0.2\", \"port\": 8081}")

	provider := NewKubernetesConfigProvider("block-ref", "system-id", "instance-id", map[string]interface{}{
		"type": "kubernetes",
	})

	info, err := provider.GetResourceInfo("foo", "rest", "foo")
	assert.NoError(t, err)
	assert.Equal(t, "10.0.0.1", info.Host)
	assert.Equal(t, 8080, info.Port)

	info, err = provider.GetResourceInfo("foo", "grpc", "bar")
	assert.NoError(t, err)
	assert.Equal(t, "10.0.0.2", info.Host)
	assert.Equal(t, 8081, info.Port)

	_, err = provider.GetResourceInfo("foo", "rest", "baz")
	assert.Error(t, err)
	assert.Equal(t, "missing environment variable for operator resource: KAPETA_CONSUMER_RESOURCE_BAZ_REST", err.Error())
}

func TestGet(t *testing.T) {
	os.Setenv("KAPETA_INSTANCE_CONFIG", "{\"foo\": \"bar\"}")

	provider := NewKubernetesConfigProvider("block-ref", "system-id", "instance-id", map[string]interface{}{
		"type": "kubernetes",
	})

	value := provider.Get("foo")
	assert.Equal(t, "bar", value)

	value = provider.Get("baz")
	assert.Nil(t, value)
}

func TestGetOrDefault(t *testing.T) {
	os.Setenv("KAPETA_INSTANCE_CONFIG", "{\"foo\": \"bar\"}")

	provider := NewKubernetesConfigProvider("block-ref", "system-id", "instance-id", map[string]interface{}{
		"type": "kubernetes",
	})

	value := provider.GetOrDefault("foo", "baz")
	assert.Equal(t, "bar", value)

	value = provider.GetOrDefault("baz", "qux")
	assert.Equal(t, "qux", value)
}

func TestGetInstanceHost(t *testing.T) {
	os.Setenv("KAPETA_BLOCK_HOSTS", "{\"instance-id\": \"10.0.0.1\"}")

	provider := NewKubernetesConfigProvider("block-ref", "system-id", "instance-id", map[string]interface{}{
		"type": "kubernetes",
	})

	host, err := provider.GetInstanceHost("instance-id")
	assert.NoError(t, err)
	assert.Equal(t, "10.0.0.1", host)

	_, err = provider.GetInstanceHost("unknown-instance-id")
	assert.Error(t, err)
	assert.Equal(t, "unknown instance id when resolving host: unknown-instance-id", err.Error())
}
