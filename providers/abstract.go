// Copyright 2023 Kapeta Inc.
// SPDX-License-Identifier: MIT

package providers

type ConfigProvider interface {
	GetBlockDefinition() interface{}
	GetBlockReference() string
	GetSystemId() string
	GetInstanceId() string
	GetServerPort(portType string) (string, error)
	GetServiceAddress(serviceName, portType string) (string, error)
	GetResourceInfo(resourceType, portType, resourceName string) (*ResourceInfo, error)
	GetInstanceHost(instanceID string) (string, error)
	GetServerHost() (string, error)
	GetProviderId() string
	Get(path string) interface{}
	GetOrDefault(path string, defaultValue interface{}) interface{}
}

type InstanceValue struct {
	ID string `json:"id"`
}

type InstanceProviderValue struct {
	ID           string `json:"id"`
	PortType     string `json:"portType"`
	ResourceName string `json:"resourceName"`
}

// Identity struct represents the identity of a block
type Identity struct {
	SystemID   string `json:"systemId"`
	InstanceID string `json:"instanceId"`
}

// ResourceInfo struct represents information about a resource
type ResourceInfo struct {
	Host        string                 `json:"host"`
	Port        int                    `json:"port"`
	Type        string                 `json:"type"`
	Protocol    string                 `json:"protocol"`
	Options     map[string]interface{} `json:"options"`
	Credentials map[string]string      `json:"credentials"`
}

type AbstractConfigProvider struct {
	BlockRef        string                 `json:"blockRef"`
	SystemID        string                 `json:"systemId"`
	InstanceID      string                 `json:"instanceId"`
	BlockDefinition map[string]interface{} `json:"blockDefinition"`
}

func (a *AbstractConfigProvider) GetBlockDefinition() interface{} {
	return a.BlockDefinition
}

func (a *AbstractConfigProvider) GetBlockReference() string {
	return a.BlockRef
}

func (a *AbstractConfigProvider) GetSystemId() string {
	return a.SystemID
}

func (a *AbstractConfigProvider) GetInstanceId() string {
	return a.InstanceID
}

func (a *AbstractConfigProvider) SetIdentity(systemID, instanceID string) {
	a.SystemID = systemID
	a.InstanceID = instanceID
}
