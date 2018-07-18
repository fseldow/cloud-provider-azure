package utils

import (
	"context"

	azureNetwork "github.com/Azure/azure-sdk-for-go/services/network/mgmt/2017-09-01/network"
	"k8s.io/apimachinery/pkg/util/wait"
)

// WaitGetVirtualNetworkList is a wapper around listing VirtualNetwork
func WaitGetVirtualNetworkList() (result azureNetwork.VirtualNetworkListResultPage, err error) {
	testClient, err := ObtainTestClient()
	if err != nil {
		return
	}

	err = wait.PollImmediate(poll, singleCallTimeout, func() (bool, error) {
		result, err = testClient.List(context.Background(), GetResourceGroup())
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
