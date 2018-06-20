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

package network

import (
	"fmt"
	"net/http"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	clientset "k8s.io/client-go/kubernetes"
	testutils "k8s.io/cloud-provider-azure/tests/e2e/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service Connection", func() {
	basename := "service"
	serviceName := "dns-service"

	var cs clientset.Interface
	var ns *v1.Namespace
	var err error
	var namespacesToDelete []*v1.Namespace

	labels := map[string]string{
		"app": serviceName,
	}
	ports := []v1.ServicePort{{
		Port:       8080,
		TargetPort: intstr.FromInt(8080),
	}}

	BeforeEach(func() {
		cs, err = testutils.GetClientSet()
		Expect(err).NotTo(HaveOccurred())

		ns, err = testutils.CreateTestingNS(basename, cs)
		Expect(err).NotTo(HaveOccurred())
		namespacesToDelete = append(namespacesToDelete, ns)
	})

	AfterEach(func() {
		for _, nsToDel := range namespacesToDelete {
			err = testutils.DeleteNS(cs, nsToDel.Name)
			Expect(err).NotTo(HaveOccurred())
		}
	})

	It("can connect via domain name", func() {
		testutils.Logf("Creating deployment " + serviceName)
		_, err = cs.Extensions().Deployments(ns.Name).Create(testutils.DefaultDeployment(serviceName, labels))
		Expect(err).NotTo(HaveOccurred())

		By("Create service")
		testutils.Logf("Name " + serviceName + " with type LoadBalancer in namespace " + ns.Name)
		_, err := testutils.CreateLoadBalancerService(cs, serviceName, labels, ns.Name, ports)
		Expect(err).NotTo(HaveOccurred())
		testutils.Logf("Service created successfully")

		defer func() {
			By("Cleaning up")
			err = cs.CoreV1().Services(ns.Name).Delete(serviceName, nil)
			Expect(err).NotTo(HaveOccurred())
			err = cs.Extensions().Deployments(ns.Name).Delete(serviceName, nil)
			Expect(err).NotTo(HaveOccurred())
		}()

		By("Wait for external domain name")
		serviceDomainName, err := testutils.WaitExternalDNS(cs, ns.Name, serviceName)
		Expect(err).NotTo(HaveOccurred())

		var resp *http.Response
		defer func() {
			if resp != nil {
				resp.Body.Close()
			}
		}()

		By("Validating External domain name")
		var code int
		for i := 1; i <= 30; i++ {
			url := fmt.Sprintf("http://%s:%v", serviceDomainName, ports[0].Port)
			resp, err = http.Get(url)
			if err == nil {
				defer func() {
					if resp != nil {
						resp.Body.Close()
					}
				}()
				code = resp.StatusCode
				if resp.StatusCode != 200 && i < 28 {
					i = 28
					continue
				} else if resp.StatusCode == 200 {
					break
				}
			}
			time.Sleep(20 * time.Second)
		}
		Expect(err).NotTo(HaveOccurred())
		Expect(code).To(Equal(200), "Fail to get response from the domain name")
	})
})
