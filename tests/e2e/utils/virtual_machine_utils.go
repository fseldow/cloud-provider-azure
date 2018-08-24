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

	azcompute "github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2017-12-01/compute"
)

func (azureTestClient *AzureTestClient) getVirtualMachine(vmName string) (azcompute.VirtualMachine, error) {
	vmClient := azureTestClient.createVirtualMachinesClient()
	vm, err := vmClient.Get(context.Background(), getResourceGroup(), vmName, "")
	return vm, err
}
