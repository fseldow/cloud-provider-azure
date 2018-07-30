package network

import (
	"context"
	"fmt"
	"strconv"

	aznetwork "github.com/Azure/azure-sdk-for-go/services/network/mgmt/2017-09-01/network"
	"github.com/Azure/go-autorest/autorest/to"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/cloud-provider-azure/tests/e2e/utils"
	"k8s.io/kubernetes/pkg/cloudprovider/providers/azure"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ServiceLoadBalancer", func() {
	basename := "service-lb"
	serviceName := "servicelb-test"

	var cs clientset.Interface
	var ns *v1.Namespace
	var tc *utils.AzureTestClient

	labels := map[string]string{
		"app": serviceName,
	}
	ports := []v1.ServicePort{{
		Port:       nginxPort,
		TargetPort: intstr.FromInt(nginxPort),
	}}

	BeforeEach(func() {
		var err error
		cs, err = utils.GetClientSet()
		Expect(err).NotTo(HaveOccurred())

		ns, err = utils.CreateTestingNameSpace(basename, cs)
		Expect(err).NotTo(HaveOccurred())

		tc, err = utils.NewDefaultAzureTestClient()
		Expect(err).NotTo(HaveOccurred())

		utils.Logf("Creating deployment " + serviceName)
		deployment := defaultDeployment(serviceName, labels)
		_, err = cs.Extensions().Deployments(ns.Name).Create(deployment)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := cs.Extensions().Deployments(ns.Name).Delete(serviceName, nil)
		Expect(err).NotTo(HaveOccurred())

		err = utils.DeleteNameSpace(cs, ns.Name)
		Expect(err).NotTo(HaveOccurred())

		cs = nil
		ns = nil
		tc = nil
	})

	// Public w/o IP -> Public w/ IP
	It("should support assigning to specific IP when updating public service", func() {
		annotation := map[string]string{
			azure.ServiceAnnotationLoadBalancerInternal: "false",
		}
		ipName := basename + "-public-none-IP"

		service := loadBalancerService(cs, serviceName, annotation, labels, ns.Name, ports)
		_, err := cs.CoreV1().Services(ns.Name).Create(service)
		Expect(err).NotTo(HaveOccurred())
		utils.Logf("Successfully created LoadBalancer service " + serviceName + " in namespace " + ns.Name)

		pip, err := utils.WaitCreatePIP(tc, ipName, defaultPublicIPAddress(ipName))
		Expect(err).NotTo(HaveOccurred())
		targetIP := to.String(pip.IPAddress)
		utils.Logf("PIP to %s", targetIP)

		defer func() {
			By("Cleaning up")
			err = cs.CoreV1().Services(ns.Name).Delete(serviceName, nil)
			Expect(err).NotTo(HaveOccurred())
			err = utils.DeletePIPWithRetry(tc, ipName)
			Expect(err).NotTo(HaveOccurred())
		}()

		By("Waiting for exposure of the original service without assigned lb IP")
		ip1, err := utils.WaitServiceExposure(cs, ns.Name, serviceName)
		Expect(err).NotTo(HaveOccurred())

		Expect(ip1).NotTo(Equal(targetIP))

		By("Updating service to bound to specific public IP")
		utils.Logf("will update IP to %s", targetIP)
		service, err = cs.CoreV1().Services(ns.Name).Get(serviceName, metav1.GetOptions{})
		service = updateServiceBalanceIP(service, false, targetIP)

		_, err = cs.CoreV1().Services(ns.Name).Update(service)
		Expect(err).NotTo(HaveOccurred())

		err = utils.WaitUpdateServiceExposure(cs, ns.Name, serviceName, targetIP, true)
		Expect(err).NotTo(HaveOccurred())
	})

	// Public w/ IP -> Public w/o IP
	It("should update public IP without deleting the user's PIP resource", func() {
		annotation := map[string]string{
			azure.ServiceAnnotationLoadBalancerInternal: "false",
		}
		ipName := basename + "-public-IP-none"

		pip, err := utils.WaitCreatePIP(tc, ipName, defaultPublicIPAddress(ipName))
		Expect(err).NotTo(HaveOccurred())
		targetIP := to.String(pip.IPAddress)

		service := loadBalancerService(cs, serviceName, annotation, labels, ns.Name, ports)
		service = updateServiceBalanceIP(service, false, targetIP)
		_, err = cs.CoreV1().Services(ns.Name).Create(service)
		Expect(err).NotTo(HaveOccurred())
		utils.Logf("Successfully created LoadBalancer service " + serviceName + " in namespace " + ns.Name)

		defer func() {
			By("Cleaning up")
			err = cs.CoreV1().Services(ns.Name).Delete(serviceName, nil)
			Expect(err).NotTo(HaveOccurred())
			err = utils.DeletePIPWithRetry(tc, ipName)
			Expect(err).NotTo(HaveOccurred())
		}()

		By("Waiting for exposure of public service with assigned lb IP")
		err = utils.WaitUpdateServiceExposure(cs, ns.Name, serviceName, targetIP, true)
		Expect(err).NotTo(HaveOccurred())

		By("Updating the service to a public service without lb IP")
		service, err = cs.CoreV1().Services(ns.Name).Get(serviceName, metav1.GetOptions{})
		service = updateServiceBalanceIP(service, false, "")
		_, err = cs.CoreV1().Services(ns.Name).Update(service)
		Expect(err).NotTo(HaveOccurred())

		err = utils.WaitUpdateServiceExposure(cs, ns.Name, serviceName, targetIP, false)
		Expect(err).NotTo(HaveOccurred())

		By("Validate user's pulic IP resource exists")
		err = utils.WaitGetPIP(tc, ipName)
		Expect(err).NotTo(HaveOccurred())
	})

	// Internal w/ IP -> Internal w/ IP
	It("should support updating internal IP when updating internal service", func() {
		annotation := map[string]string{
			azure.ServiceAnnotationLoadBalancerInternal: "true",
		}
		ip1, err := selectAvailablePrivateIP(tc)
		Expect(err).NotTo(HaveOccurred())

		service := loadBalancerService(cs, serviceName, annotation, labels, ns.Name, ports)
		service = updateServiceBalanceIP(service, true, ip1)
		_, err = cs.CoreV1().Services(ns.Name).Create(service)
		Expect(err).NotTo(HaveOccurred())
		utils.Logf("Successfully created LoadBalancer service " + serviceName + " in namespace " + ns.Name)

		defer func() {
			By("Cleaning up")
			err = cs.CoreV1().Services(ns.Name).Delete(serviceName, nil)
			Expect(err).NotTo(HaveOccurred())
		}()
		By("Waiting for exposure of internal service with specific IP")
		err = utils.WaitUpdateServiceExposure(cs, ns.Name, serviceName, ip1, true)
		Expect(err).NotTo(HaveOccurred())

		ip2, err := selectAvailablePrivateIP(tc)
		Expect(err).NotTo(HaveOccurred())

		By("Updating internal service private IP")
		utils.Logf("will update IP to %s", ip2)
		service, err = cs.CoreV1().Services(ns.Name).Get(serviceName, metav1.GetOptions{})
		service = updateServiceBalanceIP(service, true, ip2)
		_, err = cs.CoreV1().Services(ns.Name).Update(service)
		Expect(err).NotTo(HaveOccurred())

		err = utils.WaitUpdateServiceExposure(cs, ns.Name, serviceName, ip2, true)
		Expect(err).NotTo(HaveOccurred())
	})

	// internal w/o IP -> public w/ IP
	It("should support updating an internal service to a public service with assigned IP", func() {
		annotation := map[string]string{
			azure.ServiceAnnotationLoadBalancerInternal: "true",
		}
		ipName := basename + "-internal-none-public-IP"

		service := loadBalancerService(cs, serviceName, annotation, labels, ns.Name, ports)
		_, err := cs.CoreV1().Services(ns.Name).Create(service)
		Expect(err).NotTo(HaveOccurred())
		utils.Logf("Successfully created LoadBalancer service " + serviceName + " in namespace " + ns.Name)

		pip, err := utils.WaitCreatePIP(tc, ipName, defaultPublicIPAddress(ipName))
		Expect(err).NotTo(HaveOccurred())
		targetIP := to.String(pip.IPAddress)

		defer func() {
			By("Cleaning up")
			err = cs.CoreV1().Services(ns.Name).Delete(serviceName, nil)
			Expect(err).NotTo(HaveOccurred())
			err = utils.DeletePIPWithRetry(tc, ipName)
			Expect(err).NotTo(HaveOccurred())
		}()

		By("Waiting for exposure of the original service without assigned lb private IP")
		ip1, err := utils.WaitServiceExposure(cs, ns.Name, serviceName)
		Expect(err).NotTo(HaveOccurred())
		Expect(ip1).NotTo(Equal(targetIP))

		By("Updating service to bound to specific public IP")
		utils.Logf("will update IP to %s, %v", targetIP, len(targetIP))
		service, err = cs.CoreV1().Services(ns.Name).Get(serviceName, metav1.GetOptions{})
		service = updateServiceBalanceIP(service, false, targetIP)

		_, err = cs.CoreV1().Services(ns.Name).Update(service)
		Expect(err).NotTo(HaveOccurred())

		err = utils.WaitUpdateServiceExposure(cs, ns.Name, serviceName, targetIP, true)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should have no operation since no change in service when update", func() {
		annotation := map[string]string{
			azure.ServiceAnnotationLoadBalancerInternal: "false",
		}
		ipName := basename + "-public-remain"
		pip, err := utils.WaitCreatePIP(tc, ipName, defaultPublicIPAddress(ipName))
		Expect(err).NotTo(HaveOccurred())
		targetIP := to.String(pip.IPAddress)

		service := loadBalancerService(cs, serviceName, annotation, labels, ns.Name, ports)
		service = updateServiceBalanceIP(service, false, targetIP)
		_, err = cs.CoreV1().Services(ns.Name).Create(service)
		Expect(err).NotTo(HaveOccurred())
		utils.Logf("Successfully created LoadBalancer service " + serviceName + " in namespace " + ns.Name)

		defer func() {
			By("Cleaning up")
			err = cs.CoreV1().Services(ns.Name).Delete(serviceName, nil)
			Expect(err).NotTo(HaveOccurred())
			err = utils.DeletePIPWithRetry(tc, ipName)
			Expect(err).NotTo(HaveOccurred())
		}()

		By("Waiting for exposure of the original service with assigned lb private IP")
		err = utils.WaitUpdateServiceExposure(cs, ns.Name, serviceName, targetIP, true)
		Expect(err).NotTo(HaveOccurred())

		By("Update without changing the service and wait for a while")
		utils.Logf("Exteral IP is now %s", targetIP)
		service, err = cs.CoreV1().Services(ns.Name).Get(serviceName, metav1.GetOptions{})
		_, err = cs.CoreV1().Services(ns.Name).Update(service)
		Expect(err).NotTo(HaveOccurred())

		//Wait for 10 minutes, there should return timeout err, since external ip should not change
		err = utils.WaitUpdateServiceExposure(cs, ns.Name, serviceName, targetIP, false /*expectSame*/)
		Expect(err).To(Equal(wait.ErrWaitTimeout))
	})
})

func judgeInternal(service v1.Service) bool {
	return service.Annotations[azure.ServiceAnnotationLoadBalancerInternal] == "true"
}

func updateServiceBalanceIP(service *v1.Service, isInternal bool, ip string) (result *v1.Service) {
	result = service
	if result == nil {
		return
	}
	result.Spec.LoadBalancerIP = ip
	if judgeInternal(*service) == isInternal {
		return
	}
	if isInternal {
		result.Annotations[azure.ServiceAnnotationLoadBalancerInternal] = "true"
	} else {
		delete(result.Annotations, azure.ServiceAnnotationLoadBalancerInternal)
	}
	return
}

func defaultPublicIPAddress(ipName string) aznetwork.PublicIPAddress {
	return aznetwork.PublicIPAddress{
		Name:     to.StringPtr(ipName),
		Location: to.StringPtr("eastus"),
		PublicIPAddressPropertiesFormat: &aznetwork.PublicIPAddressPropertiesFormat{
			PublicIPAllocationMethod: aznetwork.Static,
		},
	}
}

// select a private IP address in subnet 10.240.0.0/12
// select range from 10.240.1.0 ~ 10.240.1.100
func selectAvailablePrivateIP(tc *utils.AzureTestClient) (string, error) {
	vNet, err := getVNet(tc)
	vNetClient := tc.GetVirtualNetworksClient()
	if err != nil {
		return "", err
	}
	baseIP := "10.240.1."
	for i := 0; i <= 100; i++ {
		IP := baseIP + strconv.Itoa(i)
		ret, err := vNetClient.CheckIPAddressAvailability(context.Background(), utils.GetResourceGroup(), to.String(vNet.Name), IP)
		if err != nil {
			// just ignore
			continue
		}
		if ret.Available != nil && *ret.Available {
			return IP, nil
		}
	}
	return "", fmt.Errorf("Find no availabePrivateIP in range 10.240.1.0 ~ 10.240.1.100")
}
