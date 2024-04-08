// Copyright 2023 Kapeta Inc.
// SPDX-License-Identifier: MIT

package config

import (
	"encoding/json"
	"os"
)

// ReadConfigFile reads the environment configuration file and returns the map
func ReadConfigFile() (map[string]string, error) {
	out := make(map[string]string)
	kapetaConfigPath := os.Getenv("KAPETA_CONFIG_PATH")

	if kapetaConfigPath == "" {
		return out, nil
	}

	// Open the JSON file
	file, err := os.Open(kapetaConfigPath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	err = decoder.Decode(&out)

	if err != nil {
		return out, err
	}

	return out, nil
}
