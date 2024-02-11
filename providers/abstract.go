// Copyright 2023 Kapeta Inc.
// SPDX-License-Identifier: MIT

package providers

import "github.com/kapetacom/schemas/packages/go/model"

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
	GetInstanceForConsumer(resourceName string) (*BlockInstanceDetails, error)
	GetInstanceOperator(instanceId string) (*InstanceOperator, error)
	GetInstancesForProvider(resourceName string) ([]*BlockInstanceDetails, error)
}

type DefaultCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type InstanceOperatorPort struct {
	Protocol string `json:"protocol"`
	Port     int    `json:"port"`
}

type InstanceOperator struct {
	Hostname    string                          `json:"hostname"`
	Ports       map[string]InstanceOperatorPort `json:"ports"`
	Path        string                          `json:"path,omitempty"`
	Query       string                          `json:"query,omitempty"`
	Hash        string                          `json:"hash,omitempty"`
	Credentials any                             `json:"credentials,omitempty"`
	Options     any                             `json:"options,omitempty"`
}

type BlockInstanceDetails struct {
	InstanceId  string              `json:"instanceId"`
	Block       *model.Kind         `json:"block"`
	Connections []*model.Connection `json:"connections"`
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
