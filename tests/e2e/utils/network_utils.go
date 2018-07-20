package utils

import (
	"context"

	aznetwork "github.com/Azure/azure-sdk-for-go/services/network/mgmt/2017-09-01/network"
	"k8s.io/apimachinery/pkg/util/wait"
)

// WaitGetVirtualNetworkList is a wapper around listing VirtualNetwork
func WaitGetVirtualNetworkList(azureTestClient *AzureTestClient) (result aznetwork.VirtualNetworkListResultPage, err error) {
	Logf("Getting virtural network list")
	vNetClient := aznetwork.VirtualNetworksClient{BaseClient: azureTestClient.BaseClient}
	err = wait.PollImmediate(poll, singleCallTimeout, func() (bool, error) {
		result, err = vNetClient.List(context.Background(), GetResourceGroup())
		if err != nil {
			if !isRetryableAPIError(err) {
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
	subnetsClient := aznetwork.SubnetsClient{BaseClient: azureTestClient.BaseClient}
	_, err := subnetsClient.CreateOrUpdate(context.Background(), GetResourceGroup(), *vnet.Name, *subnetName, subnetParameter)
	return err
}

//func waitGetLoadBalancerRuleList(azureTestClient *AzureTestClient)
