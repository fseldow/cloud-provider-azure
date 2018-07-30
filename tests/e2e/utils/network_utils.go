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

// DeleteSubnetWithRetry tries to delete a subnet in 5 minutes
func DeleteSubnetWithRetry(azureTestClient *AzureTestClient, vnetName string, subnetName string) error {
	Logf("Deleting subnet named %s in vnet %s", subnetName, vnetName)
	subnetClient := azureTestClient.GetSubnetsClient()
	err := wait.PollImmediate(poll, singleCallTimeout, func() (bool, error) {
		_, err := subnetClient.Delete(context.Background(), GetResourceGroup(), vnetName, subnetName)
		if err != nil {
			return false, nil
		}
		return true, nil
	})
	return err
}

// WaitCreatePIP waits to create a public ip resource
func WaitCreatePIP(azureTestClient *AzureTestClient, ipName string, ipParameter aznetwork.PublicIPAddress) (aznetwork.PublicIPAddress, error) {
	Logf("Creating public IP resourc named %s", ipName)
	pipClient := azureTestClient.GetPublicIPAddressesClient()
	_, err := pipClient.CreateOrUpdate(context.Background(), GetResourceGroup(), ipName, ipParameter)
	var pip aznetwork.PublicIPAddress
	if err != nil {
		return pip, err
	}
	err = wait.PollImmediate(poll, singleCallTimeout, func() (bool, error) {
		pip, err = pipClient.Get(context.Background(), GetResourceGroup(), ipName, "")
		if err != nil {
			if !IsRetryableAPIError(err) {
				return false, err
			}
			return false, nil
		}
		return pip.IPAddress != nil, nil
	})
	return pip, err
}

// DeletePIPWithRetry tries to delete a pulic ip resourc
func DeletePIPWithRetry(azureTestClient *AzureTestClient, ipName string) error {
	Logf("Deleting public IP resourc named %s", ipName)
	pipClient := azureTestClient.GetPublicIPAddressesClient()
	err := wait.PollImmediate(poll, singleCallTimeout, func() (bool, error) {
		_, err := pipClient.Delete(context.Background(), GetResourceGroup(), ipName)
		if err != nil {
			return false, nil
		}
		return true, nil
	})
	return err
}

// WaitGetPIP waits to get a specific public ip resource
func WaitGetPIP(azureTestClient *AzureTestClient, ipName string) (err error) {
	pipClient := azureTestClient.GetPublicIPAddressesClient()
	err = wait.PollImmediate(poll, singleCallTimeout, func() (bool, error) {
		_, err = pipClient.Get(context.Background(), GetResourceGroup(), ipName, "")
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
