package config

import "github.com/kapetacom/sdk-go-config/providers"

// ConfigProviderMock is a mock implementation of providers.ConfigProvider.
// It can be used to mock the ConfigProvider interface.
// Example:
//
//	configProvider := &ConfigProviderMock{
//		GetResourceInfoFunc: func(resourceType, resourcePort, resourceName string) (*providers.ResourceInfo, error) {
//			return &providers.ResourceInfo{
//				Host: "localhost",
//				Port: "8080",
//			}, nil
//		},
//	}
//	configProvider.GetResourceInfo("service", "http", "service1")
//	// Output: &providers.ResourceInfo{Host: "localhost", Port: "8080"}
type ConfigProviderMock struct {
	GetResourceInfoFunc         func(resourceType, resourcePort, resourceName string) (*providers.ResourceInfo, error)
	GetFunc                     func(path string) interface{}
	GetBlockDefinitionFunc      func() interface{}
	GetBlockReferenceFunc       func() string
	GetInstanceForConsumerFunc  func(resourceName string) (*providers.BlockInstanceDetails, error)
	GetInstanceHostFunc         func(instanceID string) (string, error)
	GetInstanceIdFunc           func() string
	GetInstanceOperatorFunc     func(instanceId string) (*providers.InstanceOperator, error)
	GetInstancesForProviderFunc func(resourceName string) ([]*providers.BlockInstanceDetails, error)
	GetOrDefaultFunc            func(path string, defaultValue interface{}) interface{}
	GetProviderIdFunc           func() string
	GetServerHostFunc           func() (string, error)
	GetServerPortFunc           func(portType string) (string, error)
	GetServiceAddressFunc       func(serviceName string, portType string) (string, error)
	GetSystemIdFunc             func() string
}

func (c *ConfigProviderMock) Get(path string) interface{} {
	return c.GetFunc(path)
}

func (c *ConfigProviderMock) GetBlockDefinition() interface{} {
	return c.GetBlockDefinitionFunc()
}

func (c *ConfigProviderMock) GetBlockReference() string {
	return c.GetBlockReferenceFunc()
}

func (c *ConfigProviderMock) GetInstanceForConsumer(resourceName string) (*providers.BlockInstanceDetails, error) {
	return c.GetInstanceForConsumerFunc(resourceName)
}

func (c *ConfigProviderMock) GetInstanceHost(instanceID string) (string, error) {
	return c.GetInstanceHostFunc(instanceID)
}

func (c *ConfigProviderMock) GetInstanceId() string {
	return c.GetInstanceIdFunc()
}

func (c *ConfigProviderMock) GetInstanceOperator(instanceId string) (*providers.InstanceOperator, error) {
	return c.GetInstanceOperatorFunc(instanceId)
}

func (c *ConfigProviderMock) GetInstancesForProvider(resourceName string) ([]*providers.BlockInstanceDetails, error) {
	return c.GetInstancesForProviderFunc(resourceName)
}

func (c *ConfigProviderMock) GetOrDefault(path string, defaultValue interface{}) interface{} {
	return c.GetOrDefaultFunc(path, defaultValue)
}

func (c *ConfigProviderMock) GetProviderId() string {
	return c.GetProviderIdFunc()
}

func (c *ConfigProviderMock) GetResourceInfo(resourceType string, portType string, resourceName string) (*providers.ResourceInfo, error) {
	return c.GetResourceInfoFunc(resourceType, portType, resourceName)
}

func (c *ConfigProviderMock) GetServerHost() (string, error) {
	return c.GetServerHostFunc()
}

func (c *ConfigProviderMock) GetServerPort(portType string) (string, error) {
	return c.GetServerPortFunc(portType)
}

func (c *ConfigProviderMock) GetServiceAddress(serviceName string, portType string) (string, error) {
	return c.GetServiceAddressFunc(serviceName, portType)
}

func (c *ConfigProviderMock) GetSystemId() string {
	return c.GetSystemIdFunc()
}
