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
	"context"

	aznetwork "github.com/Azure/azure-sdk-for-go/services/network/mgmt/2017-09-01/network"
	"k8s.io/apimachinery/pkg/util/wait"
)

// WaitGetVirtualNetworkList is a wapper around listing VirtualNetwork
func WaitGetVirtualNetworkList(azureTestClient *AzureTestClient) (result aznetwork.VirtualNetworkListResultPage, err error) {
	Logf("Getting virtural network list")
	vNetClient := azureTestClient.GetVirtualNetworksClient()
	err = wait.PollImmediate(poll, singleCallTimeout, func() (bool, error) {
		result, err = vNetClient.List(context.Background(), GetResourceGroup())
		if err != nil {
			if !IsRetryableAPIError(err) {
				return false, err
			}
			return false, nil
		}
		return true, nil
	})
	return
}

// CreateNewSubnet will create a new subnet in certain virtual network
func CreateNewSubnet(azureTestClient *AzureTestClient, vnet aznetwork.VirtualNetwork, subnetName *string, prefix *string) error {
	Logf("creating a new subnet %s, %s", *subnetName, *prefix)
	subnetParameter := (*vnet.Subnets)[0]
	subnetParameter.Name = subnetName
	subnetParameter.AddressPrefix = prefix
	subnetsClient := azureTestClient.GetSubnetsClient()
	_, err := subnetsClient.CreateOrUpdate(context.Background(), GetResourceGroup(), *vnet.Name, *subnetName, subnetParameter)
	return err
}

// WaitDeleteSubnet tries to delete a subnet in 5 minutes
func WaitDeleteSubnet(azureTestClient *AzureTestClient, vnetName string, subnetName string) error {
	subnetClient := azureTestClient.GetSubnetsClient()
	return wait.PollImmediate(poll, singleCallTimeout, func() (bool, error) {
		_, err := subnetClient.Delete(context.Background(), GetResourceGroup(), vnetName, subnetName)
		if err != nil {
			return false, nil
		}
		return true, nil
	})
}
