// Copyright 2023 Kapeta Inc.
// SPDX-License-Identifier: MIT

package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	KAPETA_CLUSTER_SERVICE_CONFIG_FILE  = "cluster-service.yml"
	KAPETA_CLUSTER_SERVICE_DEFAULT_PORT = "35100"
	KAPETA_CLUSTER_SERVICE_DEFAULT_HOST = "127.0.0.1"
)

type ClusterConfig struct {
	Cluster *Cluster `json:"cluster,omitempty"`
}
type Cluster struct {
	Port string `yaml:"port"`
	Host string `yaml:"host"`
}

// NewClusterConfig creates a new instance of ClusterConfig
func NewClusterConfig() *ClusterConfig {
	return &ClusterConfig{}
}

func (c *ClusterConfig) getClusterServicePort() string {
	if envPort := os.Getenv("KAPETA_LOCAL_CLUSTER_PORT"); envPort != "" {
		return envPort
	}

	return c.GetClusterConfig().Cluster.Port
}

func (c *ClusterConfig) getClusterServiceHost() string {
	if envHost := os.Getenv("KAPETA_LOCAL_CLUSTER_HOST"); envHost != "" {
		return envHost
	}
	return c.GetClusterConfig().Cluster.Host
}

func (c *ClusterConfig) getKapetaBasedir() string {
	return getKapetaDir()
}

func (c *ClusterConfig) getClusterConfigFile() string {
	return filepath.Join(c.getKapetaBasedir(), KAPETA_CLUSTER_SERVICE_CONFIG_FILE)
}

func (c *ClusterConfig) GetClusterConfig() *ClusterConfig {
	if os.Getenv("TEST_KAPETA_CLUSTER_CONFIG_FILE") != "" {
		err := yaml.Unmarshal([]byte(os.Getenv("TEST_KAPETA_CLUSTER_CONFIG_FILE")), &c)
		if err != nil {
			fmt.Printf("Error unmarshalling cluster config from test config: %s\n", err)
			return nil
		}
	} else {
		if _, err := os.Stat(c.getClusterConfigFile()); err == nil {
			rawYAML, err := os.ReadFile(c.getClusterConfigFile())
			if err != nil {
				fmt.Printf("Error reading cluster config file: %s\n", err)
				return nil
			}

			err = yaml.Unmarshal(rawYAML, &c)
			if err != nil {
				fmt.Printf("Error unmarshalling cluster config: %s\n", err)
				return nil
			}
		}
	}
	if c == nil {
		c = &ClusterConfig{}
	}

	if c.Cluster == nil {
		c.Cluster = &Cluster{}
	}

	if c.Cluster.Port == "" {
		c.Cluster.Port = KAPETA_CLUSTER_SERVICE_DEFAULT_PORT
	}

	if c.Cluster.Host == "" {
		c.Cluster.Host = KAPETA_CLUSTER_SERVICE_DEFAULT_HOST
	}

	fmt.Printf("Read cluster config from file: %s\n", c.getClusterConfigFile())

	return c
}

func (c *ClusterConfig) GetClusterServiceAddress() string {
	clusterPort := c.getClusterServicePort()
	host := c.getClusterServiceHost()
	return fmt.Sprintf("http://%s:%s", host, clusterPort)
}

func getKapetaDir() string {
	kapetaDir := os.Getenv("KAPETA_HOME")
	if kapetaDir == "" {
		kapetaDir = filepath.Join(os.Getenv("HOME"), ".kapeta")
	}
	return kapetaDir
}
