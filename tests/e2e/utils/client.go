/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"os"
	"strings"

	aznetwork "github.com/Azure/azure-sdk-for-go/services/network/mgmt/2017-09-01/network"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
)

// AzureTestClient configs Azure specific clients
type AzureTestClient struct {
	networkClient aznetwork.BaseClient
}

// NewDefaultAzureTestClient makes a new AzureTestClient
func NewDefaultAzureTestClient() (*AzureTestClient, error) {
	authconfig := AzureAuthConfigFromTestProfile()
	servicePrincipleToken, err := GetServicePrincipalToken(authconfig, parseEnvFromLocation())
	if err != nil {
		return nil, err
	}
	baseClient := aznetwork.NewWithBaseURI(parseEnvFromLocation().TokenAudience, authconfig.SubscriptionID)
	baseClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipleToken)

	c := &AzureTestClient{
		networkClient: baseClient,
	}

	return c, nil
}

// GetSubnetsClient generates subnet client with the same baseclient as azure test client
func (tc *AzureTestClient) GetSubnetsClient() *aznetwork.SubnetsClient {
	return &aznetwork.SubnetsClient{BaseClient: tc.networkClient}
}

// GetVirtualNetworksClient generates virtual network client with the same baseclient as azure test client
func (tc *AzureTestClient) GetVirtualNetworksClient() *aznetwork.VirtualNetworksClient {
	return &aznetwork.VirtualNetworksClient{BaseClient: tc.networkClient}
}

// GetPublicIPAddressesClient generates virtual network client with the same baseclient as azure test client
func (tc *AzureTestClient) GetPublicIPAddressesClient() *aznetwork.PublicIPAddressesClient {
	return &aznetwork.PublicIPAddressesClient{BaseClient: tc.networkClient}
}

func parseEnvFromLocation() *azure.Environment {
	location := os.Getenv(clusterLocationEnv)
	if strings.Contains(location, "ch") {
		return &azure.ChinaCloud
	} else if strings.Contains(location, "ger") {
		return &azure.GermanCloud
	} else if strings.Contains(location, "gov") {
		return &azure.USGovernmentCloud
	}
	return &azure.PublicCloud
}
