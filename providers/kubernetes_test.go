// Copyright 2023 Kapeta Inc.
// SPDX-License-Identifier: MIT

package providers

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKubernetesConfigProvider(t *testing.T) {
	// Mock environment variables for testing
	os.Setenv("KAPETA_PROVIDER_PORT_REST", "8080")
	os.Setenv("KAPETA_PROVIDER_HOST", "localhost")
	os.Setenv("KAPETA_CONSUMER_SERVICE_TEST_SERVICE_REST", "http://test-service:8080")
	os.Setenv("KAPETA_CONSUMER_RESOURCE_TEST_RESOURCE_REST", `{"host": "test-resource", "port": "9090", "type": "test", "protocol": "http"}`)
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

func TestK8sGetServerPort(t *testing.T) {
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

func TestK8sGetServerHost(t *testing.T) {
	os.Setenv("KAPETA_PROVIDER_HOST", "0.0.0.0")

	provider := NewKubernetesConfigProvider("block-ref", "system-id", "instance-id", map[string]interface{}{
		"type": "kubernetes",
	})

	host, err := provider.GetServerHost()
	assert.NoError(t, err)
	assert.Equal(t, "0.0.0.0", host)
}

func TestK8sGetServiceAddress(t *testing.T) {
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

func TestK8sGetResourceInfo(t *testing.T) {
	os.Setenv("KAPETA_CONSUMER_RESOURCE_FOO_REST", "{\"host\": \"10.0.0.1\", \"port\": \"8080\"}")
	os.Setenv("KAPETA_CONSUMER_RESOURCE_BAR_GRPC", "{\"host\": \"10.0.0.2\", \"port\": 8081}")

	provider := NewKubernetesConfigProvider("block-ref", "system-id", "instance-id", map[string]interface{}{
		"type": "kubernetes",
	})

	info, err := provider.GetResourceInfo("foo", "rest", "foo")
	assert.NoError(t, err)
	assert.Equal(t, "10.0.0.1", info.Host)
	assert.Equal(t, json.Number("8080"), info.Port)

	info, err = provider.GetResourceInfo("foo", "grpc", "bar")
	assert.NoError(t, err)
	assert.Equal(t, "10.0.0.2", info.Host)
	assert.Equal(t, json.Number("8081"), info.Port)

	_, err = provider.GetResourceInfo("foo", "rest", "baz")
	assert.Error(t, err)
	assert.Equal(t, "missing environment variable for operator resource: KAPETA_CONSUMER_RESOURCE_BAZ_REST", err.Error())
}

func TestK8sGet(t *testing.T) {
	os.Setenv("KAPETA_INSTANCE_CONFIG", "{\"foo\": \"bar\"}")

	provider := NewKubernetesConfigProvider("block-ref", "system-id", "instance-id", map[string]interface{}{
		"type": "kubernetes",
	})

	value := provider.Get("foo")
	assert.Equal(t, "bar", value)

	value = provider.Get("baz")
	assert.Nil(t, value)
}

func TestK8sGetOrDefault(t *testing.T) {
	os.Setenv("KAPETA_INSTANCE_CONFIG", "{\"foo\": \"bar\"}")

	provider := NewKubernetesConfigProvider("block-ref", "system-id", "instance-id", map[string]interface{}{
		"type": "kubernetes",
	})

	value := provider.GetOrDefault("foo", "baz")
	assert.Equal(t, "bar", value)

	value = provider.GetOrDefault("baz", "qux")
	assert.Equal(t, "qux", value)
}

func TestK8sGetInstanceHost(t *testing.T) {
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

func TestK8sGetInstanceForConsumer(t *testing.T) {
	envVar := "KAPETA_INSTANCE_FOR_CONSUMER_TESTRESOURCE"
	os.Setenv(envVar, "{\"instanceId\": \"instance-id\", \"block\": {\"ref\": \"block-ref\"}, \"connections\": []}")
	defer os.Unsetenv(envVar)

	provider := NewKubernetesConfigProvider("block-ref", "system-id", "instance-id", map[string]interface{}{"type": "kubernetes"})

	// Test with valid environment variable
	blockDetails, err := provider.GetInstanceForConsumer("TestResource")
	assert.NoError(t, err)
	assert.NotNil(t, blockDetails)
	assert.Equal(t, "instance-id", blockDetails.InstanceId)

	// Test with invalid JSON in environment variable
	os.Setenv(envVar, "invalid-json")
	_, err = provider.GetInstanceForConsumer("TestResource")
	assert.Error(t, err)

	// Test with missing environment variable
	os.Unsetenv(envVar)
	_, err = provider.GetInstanceForConsumer("TestResource")
	assert.Error(t, err)
}

func TestK8sGetInstanceOperator(t *testing.T) {
	envVar := "KAPETA_INSTANCE_OPERATOR_12E0023C_0814_402F_9C62_25A7C1FCD906"
	os.Setenv(envVar, "{\"hostname\": \"test-host\", \"ports\": {\"http\": {\"port\": 80}}}")
	defer os.Unsetenv(envVar)

	provider := NewKubernetesConfigProvider("block-ref", "system-id", "instance-id", map[string]interface{}{"type": "kubernetes"})

	// Test with valid environment variable
	instanceOperator, err := provider.GetInstanceOperator("12E0023C-0814-402F-9C62-25A7C1FCD906")
	assert.NoError(t, err)
	assert.NotNil(t, instanceOperator)
	assert.Equal(t, "test-host", instanceOperator.Hostname)
	assert.Equal(t, 80, instanceOperator.Ports["http"].Port)

	// Test with invalid JSON
	os.Setenv(envVar, "invalid-json")
	_, err = provider.GetInstanceOperator("instanceid")
	assert.Error(t, err)

	// Test with missing environment variable
	os.Unsetenv(envVar)
	_, err = provider.GetInstanceOperator("instanceid")
	assert.Error(t, err)
}

func TestK8sGetInstancesForProvider(t *testing.T) {
	envVar := "KAPETA_INSTANCES_FOR_PROVIDER_TESTRESOURCE"
	os.Setenv(envVar, "[{\"instanceId\": \"instance-id-1\", \"block\": {\"ref\": \"block-ref-1\"}}, {\"instanceId\": \"instance-id-2\", \"block\": {\"ref\": \"block-ref-2\"}}]")
	defer os.Unsetenv(envVar)

	provider := NewKubernetesConfigProvider("block-ref", "system-id", "instance-id", map[string]interface{}{"type": "kubernetes"})

	// Test with valid environment variable
	instances, err := provider.GetInstancesForProvider("TestResource")
	assert.NoError(t, err)
	assert.Len(t, instances, 2)

	// Test with invalid JSON
	os.Setenv(envVar, "invalid-json")
	_, err = provider.GetInstancesForProvider("TestResource")
	assert.Error(t, err)

	// Test with missing environment variable
	os.Unsetenv(envVar)
	_, err = provider.GetInstancesForProvider("TestResource")
	assert.Error(t, err)
}
