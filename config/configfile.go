package config

import (
	"encoding/json"
	"io"
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

	// Read the file's content
	byteValue, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	// Unmarshal the JSON data into the map
	err = json.Unmarshal(byteValue, &out)
	if err != nil {
		return out, err
	}

	return out, nil
}
