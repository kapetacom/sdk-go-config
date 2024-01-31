// Copyright 2023 Kapeta Inc.
// SPDX-License-Identifier: MIT

package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	cfg "github.com/kapetacom/sdk-go-config/config"
)

type RequestOptions struct {
	headers map[string]string
	url     string
}

const (
	KAPETA_ENVIRONMENT_TYPE   = "KAPETA_ENVIRONMENT_TYPE"
	HEADER_KAPETA_BLOCK       = "X-Kapeta-Block"
	HEADER_KAPETA_SYSTEM      = "X-Kapeta-System"
	HEADER_KAPETA_INSTANCE    = "X-Kapeta-Instance"
	HEADER_KAPETA_ENVIRONMENT = "X-Kapeta-Environment"
)

// LocalConfigProvider struct represents the local config provider
type LocalConfigProvider struct {
	AbstractConfigProvider
	configuration map[string]interface{}
	cfg           *cfg.ClusterConfig
}

// NewLocalConfigProvider creates an instance of LocalConfigProvider
func NewLocalConfigProvider(blockRef, systemID, instanceID string, blockDefinition map[string]interface{}) *LocalConfigProvider {
	return &LocalConfigProvider{
		AbstractConfigProvider: AbstractConfigProvider{
			BlockRef:        blockRef,
			SystemID:        systemID,
			InstanceID:      instanceID,
			BlockDefinition: blockDefinition,
		},
		configuration: make(map[string]interface{}),
		cfg:           cfg.NewClusterConfig(),
	}

}

// ResolveIdentity resolves and verifies system and instance ID
func (l *LocalConfigProvider) ResolveIdentity() error {
	fmt.Printf("Resolving identity for block: %s\n", l.BlockRef)

	url := l.getIdentityURL()
	identity, err := l.getIdentity(url)
	if err != nil {
		return fmt.Errorf("failed to resolve identity: %w", err)
	}

	fmt.Printf("Identity resolved:\n - System ID: %s\n - Instance ID: %s\n", identity.SystemID, identity.InstanceID)

	l.setIdentity(identity.SystemID, identity.InstanceID)

	if err := l.loadConfiguration(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	return nil
}

// LoadConfiguration loads the configuration for the instance
func (l *LocalConfigProvider) loadConfiguration() error {
	configuration, err := l.getInstanceConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	l.configuration = configuration
	return nil
}

// GetServerPort gets the port to listen on for the current instance
func (l *LocalConfigProvider) GetServerPort(portType string) (string, error) {
	if portType == "" {
		portType = DEFAULT_SERVER_PORT_TYPE
	}

	envVar := fmt.Sprintf("KAPETA_LOCAL_SERVER_PORT_%s", strings.ToUpper(portType))
	if port, ok := os.LookupEnv(envVar); ok {
		return port, nil
	}

	url := l.getProviderPortURL(portType)
	port, err := l.getString(url)
	if err != nil {
		return "", fmt.Errorf("failed to resolve server port for type %s: %w", portType, err)
	}

	return port, nil
}

func getEnvWithDefault(envVar, defaultValue string) string {
	if value, ok := os.LookupEnv(envVar); ok {
		return value
	}
	return defaultValue
}

// GetServerHost gets the server host for the current instance
func (l *LocalConfigProvider) GetServerHost() (string, error) {
	return getEnvWithDefault("KAPETA_LOCAL_SERVER", "127.0.0.1"), nil
}

// RegisterInstanceWithLocalClusterService registers the instance with the cluster service
func (l *LocalConfigProvider) RegisterInstanceWithLocalClusterService() error {
	url := l.getInstanceURL()
	body := map[string]interface{}{
		"pid": os.Getpid(),
	}
	_, err := l.sendRequest(http.MethodPut, url, body, nil)
	if err != nil {
		return fmt.Errorf("failed to register instance: %w", err)
	}
	exitHandler := func() {
		l.InstanceStopped()
		os.Exit(0)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		exitHandler()
	}()
	return nil
}

// InstanceStopped notifies the cluster service that the instance has stopped
func (l *LocalConfigProvider) InstanceStopped() {
	url := l.getInstanceURL()
	_, err := l.sendRequest(http.MethodDelete, url, nil, nil)
	if err != nil {
		fmt.Printf("failed to notify instance stopped: %s\n", err)
	}
}

// GetServiceAddress gets the service address for the specified resource and port type
func (l *LocalConfigProvider) GetServiceAddress(resourceName, portType string) (string, error) {
	url := l.getServiceClientURL(resourceName, portType)
	return l.getString(url)
}

// GetResourceInfo gets the resource information for the specified resource type, port type, and resource name
func (l *LocalConfigProvider) GetResourceInfo(resourceType, portType, resourceName string) (*ResourceInfo, error) {
	url := l.getResourceInfoURL(resourceType, portType, resourceName)

	resourceInfo := &ResourceInfo{}
	d, err := l.getRequestRaw(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource info: %w", err)
	}
	err = json.Unmarshal(d, resourceInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}
	return resourceInfo, nil
}

// GetInstanceHost gets the host for the specified instance ID
func (l *LocalConfigProvider) GetInstanceHost(instanceID string) (string, error) {
	url := l.getInstanceHostURL(instanceID)
	return l.getString(url)
}

// GetConfig gets the configuration value for the specified path
func (l *LocalConfigProvider) GetConfig(path string) interface{} {
	return l.configuration[path]
}

// GetOrDefault gets the configuration value for the specified path, or a default value if not found
func (l *LocalConfigProvider) GetOrDefault(path string, defaultValue interface{}) interface{} {
	if value, ok := l.configuration[path]; ok {
		return value
	}
	return defaultValue
}

func (l *LocalConfigProvider) getInstanceConfig() (map[string]interface{}, error) {
	url := l.getInstanceConfigURL()

	configuration := map[string]interface{}{}
	d, err := l.getRequestRaw(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance configuration: %w", err)
	}
	err = json.Unmarshal(d, &configuration)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}
	return configuration, nil
}

func (l *LocalConfigProvider) getIdentityURL() string {
	return l.getConfigBaseURL() + "/identity"
}

func (l *LocalConfigProvider) getInstanceConfigURL() string {
	return l.getConfigBaseURL() + "/instance"
}

func (l *LocalConfigProvider) getConfigBaseURL() string {
	return l.getClusterServiceBaseURL() + "/config"
}

func (l *LocalConfigProvider) getProviderPortURL(serviceType string) string {
	return l.getConfigBaseURL() + fmt.Sprintf("/provides/%s", l.encode(serviceType))
}

func (l *LocalConfigProvider) getServiceClientURL(resourceName, serviceType string) string {
	return l.getConfigBaseURL() + fmt.Sprintf("/consumes/%s/%s", l.encode(resourceName), l.encode(serviceType))
}

func (l *LocalConfigProvider) getResourceInfoURL(operatorType, portType, resourceName string) string {
	return l.getConfigBaseURL() + fmt.Sprintf("/consumes/resource/%s/%s/%s", l.encode(operatorType), l.encode(portType), l.encode(resourceName))
}

func (l *LocalConfigProvider) getInstanceHostURL(instanceID string) string {
	elements := []string{l.encode(l.SystemID), l.encode(instanceID), "address", "public"}
	subPath := strings.Join(elements, "/")
	return l.getInstanceURL() + "/" + subPath
}

func (l *LocalConfigProvider) getInstanceURL() string {
	return l.getClusterServiceBaseURL() + "/instances"
}

func (l *LocalConfigProvider) getClusterServiceBaseURL() string {

	return l.cfg.GetClusterServiceAddress()
}

func (l *LocalConfigProvider) sendRequest(method, url string, body interface{}, headers map[string]string) (*http.Response, error) {
	reqBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

func (l *LocalConfigProvider) getRequest(url string) (string, error) {
	opts := RequestOptions{
		headers: map[string]string{
			HEADER_KAPETA_ENVIRONMENT: "process",
			HEADER_KAPETA_BLOCK:       l.BlockRef,
			HEADER_KAPETA_SYSTEM:      l.SystemID,
			HEADER_KAPETA_INSTANCE:    l.InstanceID,
		},
		url: url,
	}
	// override environment type if set in environment variable this is used when running in a container
	if os.Getenv(KAPETA_ENVIRONMENT_TYPE) != "" {
		opts.headers[HEADER_KAPETA_ENVIRONMENT] = os.Getenv(KAPETA_ENVIRONMENT_TYPE)
	}

	resp, err := l.sendRequest(http.MethodGet, opts.url, nil, opts.headers)
	if err != nil {
		return "", fmt.Errorf("failed to send GET request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", nil
	}

	if resp.StatusCode > 399 {
		return "", fmt.Errorf("request failed - Status: %d", resp.StatusCode)
	}

	d, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}
	return string(d), nil

}

func (l *LocalConfigProvider) getRequestRaw(url string) ([]byte, error) {
	opts := RequestOptions{
		headers: map[string]string{
			HEADER_KAPETA_ENVIRONMENT: "process",
			HEADER_KAPETA_BLOCK:       l.BlockRef,
			HEADER_KAPETA_SYSTEM:      l.SystemID,
			HEADER_KAPETA_INSTANCE:    l.InstanceID,
		},
		url: url,
	}

	resp, err := l.sendRequest(http.MethodGet, opts.url, nil, opts.headers)
	if err != nil {
		return nil, fmt.Errorf("failed to send GET request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode > 399 {
		return nil, fmt.Errorf("request failed - Status: %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func (l *LocalConfigProvider) getString(url string) (string, error) {
	result, err := l.getRequest(url)
	if err != nil {
		return "", fmt.Errorf("failed to send GET request: %w", err)
	}
	return result, nil
}

func (l *LocalConfigProvider) getIdentity(url string) (*Identity, error) {
	result := &Identity{}
	d, err := l.getRequestRaw(url)
	if err != nil {
		return nil, fmt.Errorf("failed to send GET request: %w", err)
	}
	err = json.Unmarshal(d, result)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}
	return result, nil

}

func (l *LocalConfigProvider) setIdentity(systemID, instanceID string) {
	l.SystemID = systemID
	l.InstanceID = instanceID
}

func (l *LocalConfigProvider) encode(text string) string {
	return url.QueryEscape(strings.ToLower(text))
}

func (l *LocalConfigProvider) GetProviderId() string {
	return "local"
}

func (l *LocalConfigProvider) Get(path string) interface{} {
	return l.GetConfig(path)
}