// Copyright 2023 Kapeta Inc.
// SPDX-License-Identifier: MIT

package providers

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

const DEFAULT_SERVER_PORT_TYPE = "rest"

func toEnvName(name string) string {
	return strings.ToUpper(strings.TrimSpace(strings.Map(func(r rune) rune {
		switch r {
		case '.', ',', '-':
			return '_'
		default:
			return r
		}
	}, name)))
}

// KubernetesConfigProvider implements the ConfigProvider interface
type KubernetesConfigProvider struct {
	AbstractConfigProvider
	configuration map[string]interface{}
	instanceHosts map[string]string
}

// NewKubernetesConfigProvider creates a new instance of KubernetesConfigProvider
func NewKubernetesConfigProvider(blockRef, systemID, instanceID string, blockDefinition map[string]interface{}) ConfigProvider {
	return &KubernetesConfigProvider{
		AbstractConfigProvider: AbstractConfigProvider{
			BlockRef:        blockRef,
			SystemID:        systemID,
			InstanceID:      instanceID,
			BlockDefinition: blockDefinition,
		},
		configuration: nil,
	}
}

// GetServerPort returns the port to listen on for the current instance
func (k *KubernetesConfigProvider) GetServerPort(portType string) (string, error) {
	if portType == "" {
		portType = DEFAULT_SERVER_PORT_TYPE
	}

	envVar := fmt.Sprintf("KAPETA_PROVIDER_PORT_%s", toEnvName(portType))
	if value, exists := os.LookupEnv(envVar); exists {
		return value, nil
	}

	return "80", nil // We default to port 80
}

// GetServerHost returns the host for the current process
func (k *KubernetesConfigProvider) GetServerHost() (string, error) {
	envVar := "KAPETA_PROVIDER_HOST"
	if value, exists := os.LookupEnv(envVar); exists {
		return value, nil
	}

	// Any host within the Docker container
	return "0.0.0.0", nil
}

// GetServiceAddress returns the service address for the given resource name and port type
func (k *KubernetesConfigProvider) GetServiceAddress(resourceName, portType string) (string, error) {
	envVar := fmt.Sprintf("KAPETA_CONSUMER_SERVICE_%s_%s", toEnvName(resourceName), toEnvName(portType))
	if value, exists := os.LookupEnv(envVar); exists {
		return value, nil
	}

	return "", fmt.Errorf("missing environment variable for internal resource: %s", envVar)
}

// GetResourceInfo returns the resource info for the given resource type, port type, and resource name
func (k *KubernetesConfigProvider) GetResourceInfo(resourceType, portType, resourceName string) (*ResourceInfo, error) {
	envVar := fmt.Sprintf("KAPETA_CONSUMER_RESOURCE_%s_%s", toEnvName(resourceName), toEnvName(portType))
	if value, exists := os.LookupEnv(envVar); exists {
		var resourceInfo ResourceInfo
		err := json.Unmarshal([]byte(value), &resourceInfo)
		if err != nil {
			return nil, fmt.Errorf("invalid JSON in environment variable: %s", envVar)
		}
		return &resourceInfo, nil
	}

	return nil, fmt.Errorf("missing environment variable for operator resource: %s", envVar)
}

// GetProviderId returns the identifier for the config provider
func (k *KubernetesConfigProvider) GetProviderId() string {
	return "kubernetes"
}

// getConfiguration is a private method to get the configuration value from the environment variable
func (k *KubernetesConfigProvider) getConfiguration(path string, defaultValue interface{}) interface{} {
	if k.configuration == nil {
		envVar := "KAPETA_INSTANCE_CONFIG"
		if value, exists := os.LookupEnv(envVar); exists {
			err := json.Unmarshal([]byte(value), &k.configuration)
			if err != nil {
				panic(fmt.Sprintf("Invalid JSON in environment variable: %s", envVar))
			}
		} else {
			fmt.Printf("Missing environment variable for instance configuration: %s\n", envVar)
			return defaultValue
		}

		if k.configuration == nil {
			k.configuration = make(map[string]interface{})
		}
	}

	result := k.configuration[path]
	if result == nil {
		return defaultValue
	}

	return result
}

// Get is an implementation of the ConfigProvider interface to get the configuration value from the object path
func (k *KubernetesConfigProvider) Get(path string) interface{} {
	return k.getConfiguration(path, nil)
}

// GetOrDefault is an implementation of the ConfigProvider interface to get the configuration value from the object path with a default value
func (k *KubernetesConfigProvider) GetOrDefault(path string, defaultValue interface{}) interface{} {
	return k.getConfiguration(path, defaultValue)
}

// GetInstanceHost returns the hostname for the given instance ID
func (k *KubernetesConfigProvider) GetInstanceHost(instanceID string) (string, error) {
	if k.instanceHosts == nil {
		if blockHosts, exists := os.LookupEnv("KAPETA_BLOCK_HOSTS"); exists {
			err := json.Unmarshal([]byte(blockHosts), &k.instanceHosts)
			if err != nil {
				panic("Invalid JSON in environment variable: KAPETA_BLOCK_HOSTS")
			}
		} else {
			return "", errors.New("environment variable KAPETA_BLOCK_HOSTS not found. Could not resolve instance host")
		}
	}

	if host, exists := k.instanceHosts[instanceID]; exists {
		return host, nil
	}

	return "", fmt.Errorf("unknown instance id when resolving host: %s", instanceID)
}
