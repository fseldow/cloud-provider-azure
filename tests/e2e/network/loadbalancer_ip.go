package network

import (
	"time"

	aznetwork "github.com/Azure/azure-sdk-for-go/services/network/mgmt/2017-09-01/network"
	"github.com/Azure/go-autorest/autorest/to"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	clientset "k8s.io/client-go/kubernetes"
	testutils "k8s.io/cloud-provider-azure/tests/e2e/utils"
	"k8s.io/kubernetes/pkg/cloudprovider/providers/azure"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ServiceLoadBalancer", func() {
	basename := "service-lb"
	serviceName := "servicelb-test"

	var cs clientset.Interface
	var ns *v1.Namespace

	labels := map[string]string{
		"app": serviceName,
	}
	ports := []v1.ServicePort{{
		Port:       nginxPort,
		TargetPort: intstr.FromInt(nginxPort),
	}}

	BeforeEach(func() {
		var err error
		cs, err = testutils.GetClientSet()
		Expect(err).NotTo(HaveOccurred())

		ns, err = testutils.CreateTestingNameSpace(basename, cs)
		Expect(err).NotTo(HaveOccurred())

		testutils.Logf("Creating deployment " + serviceName)
		deployment := defaultDeployment(serviceName, labels)
		_, err = cs.Extensions().Deployments(ns.Name).Create(deployment)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := cs.Extensions().Deployments(ns.Name).Delete(serviceName, nil)
		Expect(err).NotTo(HaveOccurred())

		err = testutils.DeleteNameSpace(cs, ns.Name)
		Expect(err).NotTo(HaveOccurred())

		cs = nil
		ns = nil
	})

	It("should support assigning to specific IP when updating public service", func() {
		annotation := map[string]string{
			azure.ServiceAnnotationLoadBalancerInternal: "true",
		}

		service := loadBalancerService(cs, serviceName, annotation, labels, ns.Name, ports)
		_, err := cs.CoreV1().Services(ns.Name).Create(service)
		Expect(err).NotTo(HaveOccurred())
		testutils.Logf("Successfully created LoadBalancer service " + serviceName + " in namespace " + ns.Name)

		defer func() {
			By("Cleaning up")
			err = cs.CoreV1().Services(ns.Name).Delete(serviceName, nil)
			Expect(err).NotTo(HaveOccurred())
		}()
		By("Waiting for service exposure")
		ip1, err := testutils.WaitServiceExposure(cs, ns.Name, serviceName)
		Expect(err).NotTo(HaveOccurred())
		testutils.Logf("Get Externel IP: %s", ip1)

		targetIP := "10.240.1.13"

		service, err = cs.CoreV1().Services(ns.Name).Get(serviceName, metav1.GetOptions{})
		service.Spec.LoadBalancerIP = targetIP
		service.ObjectMeta.Annotations = annotation

		_, err = cs.CoreV1().Services(ns.Name).Update(service)
		Expect(err).NotTo(HaveOccurred())

		ip2, err := testutils.WaitUpdateServiceExposure(cs, ns.Name, serviceName, targetIP)
		Expect(err).NotTo(HaveOccurred())
		testutils.Logf("Get Externel IP: %s", ip2)
		Expect(ip2).To(Equal(targetIP))

		time.Sleep(time.Minute)
	})

	It("should update public IP without deleting the user's PIP resource", func() {
		annotation := map[string]string{
			azure.ServiceAnnotationLoadBalancerInternal: "false",
		}

		service := loadBalancerService(cs, serviceName, annotation, labels, ns.Name, ports)
		_, err := cs.CoreV1().Services(ns.Name).Create(service)
		Expect(err).NotTo(HaveOccurred())
		testutils.Logf("Successfully created LoadBalancer service " + serviceName + " in namespace " + ns.Name)

		defer func() {
			By("Cleaning up")
			err = cs.CoreV1().Services(ns.Name).Delete(serviceName, nil)
			Expect(err).NotTo(HaveOccurred())
		}()
		By("Waiting for service exposure")
		ip1, err := testutils.WaitServiceExposure(cs, ns.Name, serviceName)
		Expect(err).NotTo(HaveOccurred())
		testutils.Logf("Get Externel IP: %s", ip1)

		targetIP := "40.76.2.174"

		service, err = cs.CoreV1().Services(ns.Name).Get(serviceName, metav1.GetOptions{})
		service.Spec.LoadBalancerIP = ""
		_, err = cs.CoreV1().Services(ns.Name).Update(service)
		Expect(err).NotTo(HaveOccurred())

		ip2, err := testutils.WaitUpdateServiceExposure(cs, ns.Name, serviceName, targetIP)
		Expect(err).NotTo(HaveOccurred())
		testutils.Logf("Get Externel IP: %s", ip2)
		Expect(ip2).NotTo(Equal(targetIP))

		time.Sleep(time.Minute)
	})

	It("should support updating internal IP when updating internal service", func() {
		annotation := map[string]string{
			azure.ServiceAnnotationLoadBalancerInternal: "true",
		}

		service := loadBalancerService(cs, serviceName, annotation, labels, ns.Name, ports)
		_, err := cs.CoreV1().Services(ns.Name).Create(service)
		Expect(err).NotTo(HaveOccurred())
		testutils.Logf("Successfully created LoadBalancer service " + serviceName + " in namespace " + ns.Name)

		defer func() {
			By("Cleaning up")
			err = cs.CoreV1().Services(ns.Name).Delete(serviceName, nil)
			Expect(err).NotTo(HaveOccurred())
		}()
		By("Waiting for service exposure")
		ip1, err := testutils.WaitServiceExposure(cs, ns.Name, serviceName)
		Expect(err).NotTo(HaveOccurred())
		testutils.Logf("Get Externel IP: %s", ip1)

		targetIP := "10.240.2.17"

		service, err = cs.CoreV1().Services(ns.Name).Get(serviceName, metav1.GetOptions{})
		service.Spec.LoadBalancerIP = targetIP
		_, err = cs.CoreV1().Services(ns.Name).Update(service)
		Expect(err).NotTo(HaveOccurred())

		ip2, err := testutils.WaitUpdateServiceExposure(cs, ns.Name, serviceName, targetIP)
		Expect(err).NotTo(HaveOccurred())
		testutils.Logf("Get Externel IP: %s", ip2)
		Expect(ip2).To(Equal(targetIP))

		time.Sleep(time.Minute)
	})

	It("should support updating an internal service to a public service with assigned IP", func() {

	})
})

func judgeInternal(service v1.Service) bool {
	return service.Spec.LoadBalancerIP == "true"
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

func assignPublicIPResourceOnAzure() {

}

func defaultPublicIPAddress(ipName string) aznetwork.PublicIPAddress {
	return aznetwork.PublicIPAddress{
		Name:     to.StringPtr(ipName),
		Location: to.StringPtr(helpers.Location()),
		PublicIPAddressPropertiesFormat: &aznetwork.PublicIPAddressPropertiesFormat{
				PublicIPAllocationMethod: aznetwork.Static,
		},
	}
}
