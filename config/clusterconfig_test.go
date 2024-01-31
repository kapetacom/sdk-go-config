// Copyright 2023 Kapeta Inc.
// SPDX-License-Identifier: MIT

package config

import (
	_ "embed"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClusterConfig(t *testing.T) {
	c := NewClusterConfig()
	assert.NotNil(t, c)
}

func TestGetClusterServicePort(t *testing.T) {
	os.Setenv("KAPETA_LOCAL_CLUSTER_PORT", "8080")
	c := NewClusterConfig()
	assert.Equal(t, "8080", c.getClusterServicePort())
	os.Unsetenv("KAPETA_LOCAL_CLUSTER_PORT")
}

func TestGetClusterServiceHost(t *testing.T) {
	os.Setenv("KAPETA_LOCAL_CLUSTER_HOST", "10.0.0.1")
	c := NewClusterConfig()
	assert.Equal(t, "10.0.0.1", c.getClusterServiceHost())
	os.Unsetenv("KAPETA_LOCAL_CLUSTER_HOST")
}

func TestGetKapetaBasedir(t *testing.T) {
	c := NewClusterConfig()
	home := os.Getenv("HOME")
	assert.Equal(t, home+"/.kapeta", c.getKapetaBasedir())
}

func setKAPETA_HOME() {
	wd, _ := os.Getwd()
	// remove the config path from the wd
	wd = wd[:len(wd)-len("/config")]
	os.Setenv("KAPETA_HOME", wd+"/.kapeta")
}

func TestGetClusterConfigFile(t *testing.T) {
	setKAPETA_HOME()
	defer os.Unsetenv("KAPETA_HOME")
	c := NewClusterConfig()
	assert.Equal(t, os.Getenv("KAPETA_HOME")+"/cluster-service.yml", c.getClusterConfigFile())
}

func TestGetClusterServiceAddress(t *testing.T) {
	c := NewClusterConfig()
	assert.Equal(t, "http://127.0.0.1:35100", c.GetClusterServiceAddress())
}
