package network

import (
	aznetwork "github.com/Azure/azure-sdk-for-go/services/network/mgmt/2017-09-01/network"
	"github.com/Azure/go-autorest/autorest/to"
	"k8s.io/api/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/cloud-provider-azure/tests/e2e/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("node2", func() {
	basename := "service"

	var cs clientset.Interface
	var ns *v1.Namespace
	var azureTestClient *utils.AzureTestClient

	BeforeEach(func() {
		var err error
		cs, err = utils.CreateKubeClientSet()
		Expect(err).NotTo(HaveOccurred())

		ns, err = utils.CreateTestingNamespace(basename, cs)
		Expect(err).NotTo(HaveOccurred())

		azureTestClient, err = utils.CreateAzureTestClient()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := utils.DeleteNamespace(cs, ns.Name)
		Expect(err).NotTo(HaveOccurred())

		cs = nil
		ns = nil
		azureTestClient = nil
	})

	It("interface", func() {

		nodesList, err := utils.GetAgentNodes(cs)
		Expect(err).NotTo(HaveOccurred())

		Expect(len(nodesList)).NotTo(Equal(0))
		experimentNodeName := nodesList[0].Name
		experimentVMName := experimentNodeName

		pip, err := azureTestClient.CreatePublicIP(experimentVMName, createDynamicPublicIPAddressManifest(experimentVMName))
		Expect(err).NotTo(HaveOccurred())

		vmInterface, err := azureTestClient.GetAgentVirtualMachineInterface(experimentVMName)
		Expect(err).NotTo(HaveOccurred())
		for _, ipconfig := range *vmInterface.IPConfigurations {
			if *ipconfig.Primary {
				ipconfig.PublicIPAddress = &pip
				break
			}
		}
		err = azureTestClient.CreateOrUpdateInterface(to.String(vmInterface.Name), vmInterface)
		Expect(err).NotTo(HaveOccurred())
	})
})

func createDynamicPublicIPAddressManifest(ipName string) aznetwork.PublicIPAddress {
	return aznetwork.PublicIPAddress{
		Name:     to.StringPtr(ipName),
		Location: to.StringPtr("eastus"),
		PublicIPAddressPropertiesFormat: &aznetwork.PublicIPAddressPropertiesFormat{
			PublicIPAllocationMethod: aznetwork.Dynamic,
		},
	}
}
