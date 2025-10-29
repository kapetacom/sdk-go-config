// Copyright 2023 Kapeta Inc.
// SPDX-License-Identifier: MIT

package providers

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"

	cfg "github.com/kapetacom/sdk-go-config/config"
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
	muConfig      sync.Mutex
	configuration map[string]any
	muHosts       sync.Mutex
	instanceHosts map[string]string
}

// NewKubernetesConfigProvider creates a new instance of KubernetesConfigProvider
func NewKubernetesConfigProvider(blockRef, systemID, instanceID string, blockDefinition map[string]any) ConfigProvider {
	envConfig := cfg.ReadConfigFile()
	return &KubernetesConfigProvider{
		AbstractConfigProvider: AbstractConfigProvider{
			BlockRef:                 blockRef,
			SystemID:                 systemID,
			InstanceID:               instanceID,
			BlockDefinition:          blockDefinition,
			EnvironmentConfiguration: envConfig,
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
	if value, exists := k.LookupEnv(envVar); exists {
		return value, nil
	}

	return "80", nil // We default to port 80
}

// GetServerHost returns the host for the current process
func (k *KubernetesConfigProvider) GetServerHost() (string, error) {
	envVar := "KAPETA_PROVIDER_HOST"
	if value, exists := k.LookupEnv(envVar); exists {
		return value, nil
	}

	// Any host within the Docker container
	return "0.0.0.0", nil
}

// GetServiceAddress returns the service address for the given resource name and port type
func (k *KubernetesConfigProvider) GetServiceAddress(resourceName, portType string) (string, error) {
	envVar := fmt.Sprintf("KAPETA_CONSUMER_SERVICE_%s_%s", toEnvName(resourceName), toEnvName(portType))
	if value, exists := k.LookupEnv(envVar); exists {
		return value, nil
	}

	return "", fmt.Errorf("missing environment variable for internal resource: %s", envVar)
}

// GetResourceInfo returns the resource info for the given resource type, port type, and resource name
func (k *KubernetesConfigProvider) GetResourceInfo(resourceType, portType, resourceName string) (*ResourceInfo, error) {
	envVar := fmt.Sprintf("KAPETA_CONSUMER_RESOURCE_%s_%s", toEnvName(resourceName), toEnvName(portType))
	if value, exists := k.LookupEnv(envVar); exists {
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

// fixJSONEscapes normalizes common JSON escape issues found in environment variables.
func (k *KubernetesConfigProvider) fixJSONEscapes(jsonStr string) string {
	replacements := map[string]string{
		`\\n`:  "\n",
		`\\t`:  "\t",
		`\\r`:  "\r",
		`\'`:   `"`,
		`\\\"`: `"`,
		`\\'`:  `'`,
	}

	for old, newVal := range replacements {
		jsonStr = strings.ReplaceAll(jsonStr, old, newVal)
	}

	// Remove surrounding quotes that shells might add
	if len(jsonStr) >= 2 {
		first, last := jsonStr[0], jsonStr[len(jsonStr)-1]
		if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
			jsonStr = jsonStr[1 : len(jsonStr)-1]
		}
	}

	return jsonStr
}

// fixJSONSyntax removes common JSON syntax errors such as trailing commas.
func (k *KubernetesConfigProvider) fixJSONSyntax(jsonStr string) string {
	// Regex handles trailing commas before closing braces/brackets (even across newlines)
	re := regexp.MustCompile(`,\s*(\n\s*)?([\]}])`)
	return re.ReplaceAllString(jsonStr, "${1}${2}")
}

// parseJSONSafely attempts to parse the given JSON string, applying automatic cleanup if needed.
func (k *KubernetesConfigProvider) parseJSONSafely(value string) (map[string]any, bool) {
	var result map[string]any

	if json.Valid([]byte(value)) && json.Unmarshal([]byte(value), &result) == nil {
		return result, true
	}

	// Try with cleanup
	fixed := k.fixJSONSyntax(k.fixJSONEscapes(value))
	if !json.Valid([]byte(fixed)) {
		return nil, false
	}

	if err := json.Unmarshal([]byte(fixed), &result); err != nil {
		return nil, false
	}

	return result, true
}

// getConfiguration retrieves configuration from the environment, falling back to a default value.
func (k *KubernetesConfigProvider) getConfiguration(path string, defaultValue any) any {
	k.muConfig.Lock()
	defer k.muConfig.Unlock()

	if k.configuration == nil {
		envVar := "KAPETA_INSTANCE_CONFIG"
		value, exists := k.LookupEnv(envVar)
		if !exists {
			return defaultValue
		}

		value = strings.TrimSpace(value)
		config, ok := k.parseJSONSafely(value)
		if !ok {
			return defaultValue
		}
		k.configuration = config
	}

	if result, ok := k.configuration[path]; ok {
		return result
	}
	return defaultValue
}

// Get is an implementation of the ConfigProvider interface to get the configuration value from the object path
func (k *KubernetesConfigProvider) Get(path string) any {
	return k.getConfiguration(path, nil)
}

// GetOrDefault is an implementation of the ConfigProvider interface to get the configuration value from the object path with a default value
func (k *KubernetesConfigProvider) GetOrDefault(path string, defaultValue any) any {
	return k.getConfiguration(path, defaultValue)
}

// GetInstanceHost returns the hostname for the given instance ID
func (k *KubernetesConfigProvider) GetInstanceHost(instanceID string) (string, error) {
	k.muHosts.Lock()
	defer k.muHosts.Unlock()
	if k.instanceHosts == nil {
		if blockHosts, exists := k.LookupEnv("KAPETA_BLOCK_HOSTS"); exists {
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

func (k *KubernetesConfigProvider) GetInstanceForConsumer(resourceName string) (*BlockInstanceDetails, error) {
	envVar := fmt.Sprintf("KAPETA_INSTANCE_FOR_CONSUMER_%s", toEnvName(resourceName))
	if value, exists := k.LookupEnv(envVar); exists {
		blockDetails := &BlockInstanceDetails{}
		err := json.Unmarshal([]byte(value), blockDetails)
		if err != nil {
			return nil, fmt.Errorf("invalid JSON in environment variable: %s", envVar)
		}
		return blockDetails, nil
	}

	return nil, fmt.Errorf("missing environment variable for instance consumer: %s", envVar)
}

func (k *KubernetesConfigProvider) GetInstanceOperator(instanceId string) (*InstanceOperator, error) {
	envVar := fmt.Sprintf("KAPETA_INSTANCE_OPERATOR_%s", toEnvName(instanceId))
	if value, exists := k.LookupEnv(envVar); exists {
		instanceOperator := &InstanceOperator{}
		err := json.Unmarshal([]byte(value), instanceOperator)
		if err != nil {
			return nil, fmt.Errorf("invalid JSON in environment variable: %s", envVar)
		}
		return instanceOperator, nil
	}

	return nil, fmt.Errorf("missing environment variable for instance consumer: %s", envVar)
}

func (k *KubernetesConfigProvider) GetInstancesForProvider(resourceName string) ([]*BlockInstanceDetails, error) {
	envVar := fmt.Sprintf("KAPETA_INSTANCES_FOR_PROVIDER_%s", toEnvName(resourceName))
	if value, exists := k.LookupEnv(envVar); exists {
		instanceOperators := make([]*BlockInstanceDetails, 0)
		err := json.Unmarshal([]byte(value), &instanceOperators)
		if err != nil {
			return nil, fmt.Errorf("invalid JSON in environment variable: %s", envVar)
		}
		return instanceOperators, nil
	}

	return nil, fmt.Errorf("missing environment variable for instance consumer: %s", envVar)
}
