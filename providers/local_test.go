// Copyright 2023 Kapeta Inc.
// SPDX-License-Identifier: MIT

package providers

import (
	"encoding/json"
	"fmt"
	"github.com/kapetacom/schemas/packages/go/model"
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

	srv := setupLocalTestServer(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/instance") {
			_, _ = w.Write([]byte("{\"id\": \"instanceID\", \"host\": \"bla\"}"))
			return
		}
		_, _ = w.Write([]byte("40004"))
	})
	defer srv.Close()

	host, port := hostAndFromURL(srv.URL)
	os.Setenv("KAPETA_LOCAL_CLUSTER_HOST", host)
	os.Setenv("KAPETA_LOCAL_CLUSTER_PORT", port)
	defer os.Unsetenv("KAPETA_LOCAL_CLUSTER_HOST")
	defer os.Unsetenv("KAPETA_LOCAL_CLUSTER_PORT")

	provider := NewLocalConfigProvider("kapeta/block-type-gateway-http", "systemID", "instanceID", map[string]interface{}{})
	serverPort, err := provider.GetServerPort("http")
	assert.NoError(t, err)
	assert.Equal(t, "40004", serverPort)

}
func TestLocalCreateLocalConfigProvider(t *testing.T) {
	blockRef := "block-ref"
	systemID := "system-id"
	instanceID := "instance-id"
	blockDefinition := map[string]interface{}{
		"type": "local",
	}

	srv := setupLocalTestServer(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("40004"))
	})
	defer srv.Close()

	host, port := hostAndFromURL(srv.URL)
	os.Setenv("KAPETA_LOCAL_CLUSTER_HOST", host)
	os.Setenv("KAPETA_LOCAL_CLUSTER_PORT", port)
	defer os.Unsetenv("KAPETA_LOCAL_CLUSTER_HOST")
	defer os.Unsetenv("KAPETA_LOCAL_CLUSTER_PORT")
	provider := NewLocalConfigProvider(blockRef, systemID, instanceID, blockDefinition)

	assert.Equal(t, blockRef, provider.GetBlockReference())
	assert.Equal(t, systemID, provider.GetSystemId())
	assert.Equal(t, instanceID, provider.GetInstanceId())
	assert.Equal(t, "local", provider.GetProviderId())
}

func TestLocalResolveIdentity(t *testing.T) {
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

	srv := setupLocalTestServer(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("40004"))
	})
	defer srv.Close()

	host, port := hostAndFromURL(srv.URL)
	os.Setenv("KAPETA_LOCAL_CLUSTER_HOST", host)
	os.Setenv("KAPETA_LOCAL_CLUSTER_PORT", port)
	defer os.Unsetenv("KAPETA_LOCAL_CLUSTER_HOST")
	defer os.Unsetenv("KAPETA_LOCAL_CLUSTER_PORT")
	provider := NewLocalConfigProvider("block-ref", "system-id", "instance-id", map[string]interface{}{
		"type": "local",
	})

	assert.Equal(t, "system-id", provider.GetSystemId())
	assert.Equal(t, "instance-id", provider.GetInstanceId())
}

func TestLocalLoadConfiguration(t *testing.T) {
	// create test server that return the correct values
	srv := setupLocalTestServer(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/instance") {
			_, _ = w.Write([]byte("{\"id\": \"instanceID\", \"host\": \"bla\"}"))
			return
		}
		if strings.HasSuffix(r.URL.Path, "/rest") {
			_, _ = w.Write([]byte("8080"))
		} else if strings.HasSuffix(r.URL.Path, "/grpc") {
			_, _ = w.Write([]byte("8081"))
		}
	})
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

func TestLocalGetServiceAddress1(t *testing.T) {
	// create test server that return the correct values
	srv := setupLocalTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

func TestLocalGetInstanceHost(t *testing.T) {
	// create test server that return the correct values
	srv := setupLocalTestServer(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/instance") {
			_, _ = w.Write([]byte("{\"id\": \"instanceID\", \"host\": \"bla\"}"))
			return
		}
		if strings.Contains(r.URL.Path, "system-id/unknown-instance-id") {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if strings.Contains(r.URL.Path, "/instances/system-id/") {
			_, _ = w.Write([]byte("10.0.0.1"))
			return
		}
	})
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

func TestLocalGetInstanceOperator(t *testing.T) {
	// Create a test server
	srv := setupLocalTestServer(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/config/operator/") {
			io := InstanceOperator{
				Hostname: "testhost",
				Ports: map[string]InstanceOperatorPort{
					"rest": {Protocol: "http", Port: 8080},
					"grpc": {Protocol: "grpc", Port: 8081},
				},
			}
			data, err := json.Marshal(io)
			if err != nil {
				t.Fatal(err)
			}
			_, _ = w.Write(data)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})
	defer srv.Close()

	provider := NewLocalConfigProvider("block-ref", "system-id", "instance-id", map[string]interface{}{
		"type": "local",
	})

	// Test the GetInstanceOperator method
	instanceID := "test-instance-id"
	operator, err := provider.GetInstanceOperator(instanceID)
	if err != nil {
		t.Errorf("GetInstanceOperator returned an error: %v", err)
	}

	assert.Equal(t, "testhost", operator.Hostname)
	assert.Len(t, operator.Ports, 2)

	if port, ok := operator.Ports["rest"]; ok {
		assert.Equal(t, "http", port.Protocol)
		assert.Equal(t, 8080, port.Port)
	} else {
		t.Errorf("REST port not found")
	}

	if port, ok := operator.Ports["grpc"]; ok {
		assert.Equal(t, "grpc", port.Protocol)
		assert.Equal(t, 8081, port.Port)
	} else {
		t.Errorf("gRPC port not found")
	}
}

func TestLocalGetInstanceForConsumer(t *testing.T) {
	resourceName := "test-resource"
	// Mock data
	mockPlan := &model.Plan{
		Spec: model.PlanSpec{
			Connections: []model.Connection{
				{
					Consumer: model.Endpoint{
						BlockId:      "instance-id",
						ResourceName: resourceName,
					},
					Provider: model.Endpoint{
						BlockId: "provider-block-id",
					},
				},
			},
			Blocks: []model.BlockInstance{
				{
					Id: "provider-block-id",
					Block: model.AssetReference{
						Ref: "provider-ref",
					},
				},
			},
		},
	}

	mockBlock := &model.Kind{
		// fill with mock Kind data
	}

	srv := setupLocalTestServer(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/config/operator/") {
			io := InstanceOperator{
				Hostname: "testhost",
				Ports: map[string]InstanceOperatorPort{
					"rest": {Protocol: "http", Port: 8080},
					"grpc": {Protocol: "grpc", Port: 8081},
				},
			}
			data, err := json.Marshal(io)
			if err != nil {
				t.Fatal(err)
			}
			_, _ = w.Write(data)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})
	defer srv.Close()

	provider := NewLocalConfigProvider("block-ref", "system-id", "instance-id", map[string]interface{}{
		"type": "local",
	})

	// Mock GetPlan method
	provider.GetPlan = func() (*model.Plan, error) {
		return mockPlan, nil
	}

	// Mock GetKind method
	provider.GetKind = func(ref string) (*model.Kind, error) {
		if ref == "provider-ref" {
			return mockBlock, nil
		}
		return nil, fmt.Errorf("block not found")
	}

	// Test the GetInstanceForConsumer method

	instanceDetails, err := provider.GetInstanceForConsumer(resourceName)
	assert.NoError(t, err)

	// Validate the returned data
	assert.Equal(t, "provider-block-id", instanceDetails.InstanceId)
	assert.Len(t, instanceDetails.Connections, 1)
	assert.Equal(t, resourceName, instanceDetails.Connections[0].Consumer.ResourceName)
	assert.Equal(t, mockBlock, instanceDetails.Block)
}

func TestLocalGetInstancesForProvider(t *testing.T) {
	resourceName := "test-resource"
	mockPlan := &model.Plan{
		Spec: model.PlanSpec{
			Connections: []model.Connection{
				{
					Provider: model.Endpoint{
						BlockId:      "instance-id",
						ResourceName: resourceName,
					},
					Consumer: model.Endpoint{
						BlockId: "consumer-block-id-1",
					},
				},
				{
					Provider: model.Endpoint{
						BlockId:      "instance-id",
						ResourceName: resourceName,
					},
					Consumer: model.Endpoint{
						BlockId: "consumer-block-id-2",
					},
				},
			},
			Blocks: []model.BlockInstance{
				{
					Id: "consumer-block-id-1",
					Block: model.AssetReference{
						Ref: "block-ref-1",
					},
				},
				{
					Id: "consumer-block-id-2",
					Block: model.AssetReference{
						Ref: "block-ref-2",
					},
				},
			},
		},
	}

	mockBlocks := map[string]*model.Kind{
		"block-ref-1": {},
		"block-ref-2": {},
	}

	srv := setupLocalTestServer(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/config/operator/") {
			io := InstanceOperator{
				Hostname: "testhost",
				Ports: map[string]InstanceOperatorPort{
					"rest": {Protocol: "http", Port: 8080},
					"grpc": {Protocol: "grpc", Port: 8081},
				},
			}
			data, err := json.Marshal(io)
			if err != nil {
				t.Fatal(err)
			}
			_, _ = w.Write(data)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})
	defer srv.Close()

	provider := NewLocalConfigProvider("my-block-ref", "system-id", "instance-id", map[string]interface{}{
		"type": "local",
	})

	// Mock GetPlan method
	provider.GetPlan = func() (*model.Plan, error) {
		return mockPlan, nil
	}

	// Mock GetKind method
	provider.GetKind = func(ref string) (*model.Kind, error) {
		if kind, ok := mockBlocks[ref]; ok {
			return kind, nil
		}
		return nil, fmt.Errorf("block not found")
	}

	instances, err := provider.GetInstancesForProvider(resourceName)
	assert.NoError(t, err)

	assert.Len(t, instances, 2)

	for _, instance := range instances {
		assert.NotNil(t, instance.Block)
		assert.Len(t, instance.Connections, 1)
	}
}

func setupLocalTestServer(handler http.HandlerFunc) *httptest.Server {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/identity") {
			_, _ = w.Write([]byte("{\"systemId\": \"system-id\", \"instanceId\": \"instance-id\"}"))
			return
		}

		if strings.HasSuffix(r.URL.Path, "/instance") {
			_, _ = w.Write([]byte("{\"id\": \"instanceID\", \"host\": \"bla\"}"))
			return
		}

		// Is hit when self-registering with local cluster service
		if r.Method == "PUT" && strings.HasSuffix(r.URL.Path, "/instances") {
			_, _ = w.Write([]byte("{}"))
			return
		}
		handler(w, r)
	}))

	host, port := hostAndFromURL(srv.URL)
	os.Setenv("KAPETA_LOCAL_CLUSTER_HOST", host)
	os.Setenv("KAPETA_LOCAL_CLUSTER_PORT", port)
	return srv
}
