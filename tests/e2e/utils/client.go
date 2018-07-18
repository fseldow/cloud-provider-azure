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
	testauth "k8s.io/cloud-provider-azure/tests/e2e/auth"
)

const (
	clusterLocationEnv = "K8S_AZURE_LOCATION"
)

// TestClient configs Azure specific clients
type TestClient struct {
	Subscription  string
	VNetClient    aznetwork.VirtualNetworksClient
	SubnetsClient aznetwork.SubnetsClient
}

// NewDefaultTestClient makes a new TestClient
func NewDefaultTestClient() (*TestClient, error) {
	authconfig := testauth.AzureAuthConfigFromTestProfile()
	servicePrincipleToken, err := testauth.GetServicePrincipalToken(authconfig, parseEnvFromLocation())
	if err != nil {
		return nil, err
	}
	baseClient := aznetwork.NewWithBaseURI(parseEnvFromLocation().TokenAudience, authconfig.SubscriptionID)
	baseClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipleToken)

	c := &TestClient{
		Subscription:  authconfig.SubscriptionID,
		VNetClient:    aznetwork.VirtualNetworksClient{BaseClient: baseClient},
		SubnetsClient: aznetwork.SubnetsClient{BaseClient: baseClient},
	}

	return c, nil
}

func parseEnvFromLocation() *azure.Environment {
	location := os.Getenv(clusterLocationEnv)
	if strings.Contains(location, "ch") {
		return &azure.ChinaCloud
	} else if strings.Contains(location, "ger") {
		return &azure.GermanCloud
	} else if strings.Contains(location, "gov") {
		return &azure.USGovernmentCloud
	} else {
		return &azure.PublicCloud
	}
}
