// Copyright 2023 Kapeta Inc.
// SPDX-License-Identifier: MIT

package sdkgoconfig

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/kapetacom/sdk-go-config/providers"
	"gopkg.in/yaml.v3"
)

type InstanceValue struct {
	ID string `yaml:"id"`
}

type Config struct {
	provider  providers.ConfigProvider
	callbacks []func(providers.ConfigProvider)
	once      sync.Once
}

// TODO: See if we can remove this global variable
var CONFIG Config

const (
	kapetaSystemType = "KAPETA_SYSTEM_TYPE"
	kapetaSystemID   = "KAPETA_SYSTEM_ID"
	kapetaBlockRef   = "KAPETA_BLOCK_REF"
	kapetaInstanceID = "KAPETA_INSTANCE_ID"

	defaultSystemType = "development"
	defaultSystemID   = ""
	defaultInstanceID = ""
)

func getEnvOrDefault(envVarName, defaultValue string) string {
	if value, exists := os.LookupEnv(envVarName); exists {
		return value
	}
	return defaultValue
}

func (c *Config) OnReady(callback func(providers.ConfigProvider)) {
	c.once.Do(func() {
		if c.provider != nil {
			callback(c.provider)
			return
		}
		c.callbacks = append(c.callbacks, callback)
	})
}

func (c *Config) IsReady() bool {
	return c.provider != nil
}

func (c *Config) GetProvider() providers.ConfigProvider {
	if c.provider == nil {
		panic("Configuration not yet initialized")
	}
	return c.provider
}

func (c *Config) Get(path string) interface{} {
	return c.GetProvider().Get(path)
}

func (c *Config) GetOrDefault(path string, defaultValue interface{}) interface{} {
	return c.GetProvider().GetOrDefault(path, defaultValue)
}

func (c *Config) GetAsInstanceHost(path string, defaultValue string) (string, error) {
	instance := c.Get(path).(*InstanceValue)
	if instance == nil {
		return defaultValue, nil
	}
	return c.getInstanceHost(instance.ID)
}

func (c *Config) getInstanceHost(instanceID string) (string, error) {
	return c.GetProvider().GetInstanceHost(instanceID)
}

// Init initializes the configuration provider based on the kapeta.yml file in the given block directory
func Init(blockDir string) (providers.ConfigProvider, error) {
	if CONFIG.provider != nil {
		return CONFIG.provider, nil
	}

	blockDefinition := map[string]interface{}{}

	if configContent, exists := os.LookupEnv("TEST_KAPETA_BLOCK_CONFIG_FILE"); exists {
		err := yaml.Unmarshal([]byte(configContent), &blockDefinition)
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling block config from test config: %s", err)
		}
	} else {
		blockYMLPath := filepath.Join(blockDir, "kapeta.yml")

		if _, err := os.Stat(blockYMLPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("kapeta.yml file not found in path: %s. Path must be absolute and point to a folder with a valid block definition", blockDir)
		}

		blockYMLContent, err := os.ReadFile(blockYMLPath)
		if err != nil {
			return nil, fmt.Errorf("error reading kapeta.yml file: %v", err)
		}

		if err := yaml.Unmarshal(blockYMLContent, &blockDefinition); err != nil {
			return nil, fmt.Errorf("error parsing kapeta.yml: %v", err)
		}
	}
	metadata := blockDefinition["metadata"]
	metadataMap := metadata.(map[string]interface{})
	if metadataMap["name"] == nil {
		return nil, fmt.Errorf("kapeta.yml file contained invalid YML: %s", blockDir)
	}

	blockRefLocal := fmt.Sprintf("%s:local", metadataMap["name"])

	systemType := strings.ToLower(getEnvOrDefault(kapetaSystemType, defaultSystemType))
	systemID := getEnvOrDefault(kapetaSystemID, defaultSystemID)
	instanceID := getEnvOrDefault(kapetaInstanceID, defaultInstanceID)
	blockRef := getEnvOrDefault(kapetaBlockRef, blockRefLocal)

	var provider providers.ConfigProvider

	switch systemType {
	case "k8s", "kubernetes":
		provider = providers.NewKubernetesConfigProvider(blockRef, systemID, instanceID, blockDefinition)

	case "development", "dev", "local":
		localProvider := providers.NewLocalConfigProvider(blockRef, systemID, instanceID, blockDefinition)
		// Only relevant locally
		if err := localProvider.RegisterInstanceWithLocalClusterService(); err != nil {
			return nil, err
		}
		provider = localProvider

	default:
		return nil, fmt.Errorf("unknown environment: %s", systemType)
	}

	CONFIG.provider = provider

	for _, callback := range CONFIG.callbacks {
		callback(provider)
	}

	return provider, nil
}

func Transcode(in, out interface{}) error {
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(in)
	if err != nil {
		return err
	}
	err = json.NewDecoder(buf).Decode(out)
	if err != nil {
		return err
	}
	return nil
}
