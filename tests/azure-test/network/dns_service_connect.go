package network

import (
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	clientset "k8s.io/client-go/kubernetes"
	testutils "k8s.io/cloud-provider-azure/tests/azure-test/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("[Conformance]Service Connection ", func() {
	var cs clientset.Interface
	basename := "service"
	var ns *v1.Namespace
	var err error
	var namespacesToDelete []*v1.Namespace

	BeforeEach(func() {
		By("Creating a kubernetes client")
		cs, err = testutils.GetClientSet()
		Expect(err).To(BeNil())

		By("Creating namespace")
		ns, err = testutils.CreateTestingNS(basename, cs, nil)
		Expect(err).To(BeNil())
		namespacesToDelete = append(namespacesToDelete, ns)

	})

	AfterEach(func() {
		for _, nsToDel := range namespacesToDelete {
			err = testutils.DeleteNS(cs, nsToDel.Name)
			Expect(err).NotTo(HaveOccurred())
		}
	})

	It("can be done via DNS [Feature: DNS]", func() {
		serviceName := "god-lexue"
		labels := map[string]string{
			"app": "test-server",
		}
		ports := []v1.ServicePort{{
			Port:       8080,
			TargetPort: intstr.FromInt(8080),
		}}
		testutils.Logf("Service: %s, Port: %v", serviceName, ports[0].Port)

		By("Create deployment " + serviceName)
		cs.Extensions().Deployments(ns.Name).Create(testutils.DefaultDeployment(serviceName, labels))

		By("Create service " + serviceName + " with type LoadBalancer in namespace " + ns.Name)
		_, err := testutils.CreateLoadBalancerService(cs, serviceName, labels, ns.Name, ports)
		Expect(err).NotTo(HaveOccurred())
		testutils.Logf("Service created successfully")

		By("Wait for external IP")
		err = testutils.WaitExternelIP(cs, ns.Name, serviceName)
		Expect(err).NotTo(HaveOccurred())

		By("Cleaning up")
		err = cs.CoreV1().Services(ns.Name).Delete(serviceName, nil)
		Expect(err).NotTo(HaveOccurred())
		err = cs.Extensions().Deployments(ns.Name).Delete(serviceName, nil)
		Expect(err).NotTo(HaveOccurred())

		//time.Sleep(time.Minute * 10)
	})

})
