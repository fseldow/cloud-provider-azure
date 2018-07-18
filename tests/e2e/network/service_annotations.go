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
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	testutils "k8s.io/cloud-provider-azure/tests/e2e/utils"
	"k8s.io/kubernetes/pkg/cloudprovider/providers/azure"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	nginxPort       = 80
	nginxStatusCode = 200
	callPoll        = 20 * time.Second
	callTimeout     = 10 * time.Minute
)

var _ = Describe("Service with annotation", func() {
	basename := "service"
	serviceName := "annotation-test"

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
		deployment := portDeployment(serviceName, labels)
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
	/*
		It("can be accessed by domain name", func() {
			By("Create service")
			serviceDomainNamePrefix := serviceName + string(uuid.NewUUID())

			annotation := map[string]string{
				azure.ServiceAnnotationDNSLabelName: serviceDomainNamePrefix,
			}

			_, err := createLoadBalancerService(cs, serviceName, annotation, labels, ns.Name, ports)
			Expect(err).NotTo(HaveOccurred())
			testutils.Logf("Successfully created LoadBalancer service " + serviceName + " in namespace " + ns.Name)

			defer func() {
				By("Cleaning up")
				err = cs.CoreV1().Services(ns.Name).Delete(serviceName, nil)
				Expect(err).NotTo(HaveOccurred())

			}()

			By("Waiting for service exposure")
			_, err = testutils.WaitServiceExposure(cs, ns.Name, serviceName)
			Expect(err).NotTo(HaveOccurred())

			By("Validating External domain name")
			var code int
			serviceDomainName := testutils.GetServiceDomainName(serviceDomainNamePrefix)
			url := fmt.Sprintf("http://%s:%v", serviceDomainName, ports[0].Port)
			for i := 1; i <= 30; i++ {
				resp, err := http.Get(url)
				if err == nil {
					defer func() {
						if resp != nil {
							resp.Body.Close()
						}
					}()
					code = resp.StatusCode
					if resp.StatusCode == nginxStatusCode {
						break
					}
				}
				time.Sleep(20 * time.Second)
			}
			Expect(err).NotTo(HaveOccurred())
			Expect(code).To(Equal(nginxStatusCode), "Fail to get response from the domain name")
		})

		It("can be bound to an internal load balancer", func() {
			annotation := map[string]string{
				azure.ServiceAnnotationLoadBalancerInternal: "true",
			}

			_, err := createLoadBalancerService(cs, serviceName, annotation, labels, ns.Name, ports)
			Expect(err).NotTo(HaveOccurred())
			testutils.Logf("Successfully created LoadBalancer service " + serviceName + " in namespace " + ns.Name)

			defer func() {
				By("Cleaning up")
				err = cs.CoreV1().Services(ns.Name).Delete(serviceName, nil)
				Expect(err).NotTo(HaveOccurred())
			}()

			By("Waiting for service exposure")
			ip, err := testutils.WaitServiceExposure(cs, ns.Name, serviceName)
			Expect(err).NotTo(HaveOccurred())

			url := fmt.Sprintf("%s:%v", ip, ports[0].Port)
			err = validateInternalLoadBalancer(cs, ns.Name, url)
		})
	*/
	It("can specify which subnet the internal load balancer should be bound to", func() {
		By("Obtain VNet property")
		By("Create service")

		annotation := map[string]string{
			azure.ServiceAnnotationLoadBalancerInternal:       "true",
			azure.ServiceAnnotationLoadBalancerInternalSubnet: "10.240.0.0",
		}

		_, err := createLoadBalancerService(cs, serviceName, annotation, labels, ns.Name, ports)
		Expect(err).NotTo(HaveOccurred())
		defer func() {
			By("Cleaning up")
			err = cs.CoreV1().Services(ns.Name).Delete(serviceName, nil)
			Expect(err).NotTo(HaveOccurred())
		}()

		By("Waiting for service exposure")
		ip, err := testutils.WaitServiceExposure(cs, ns.Name, serviceName)
		Expect(err).NotTo(HaveOccurred())

		url := fmt.Sprintf("%s:%v", ip, ports[0].Port)
		testutils.Logf(url)
	})

	It("should be bound to the load balancer from any available set with minimum rules in auto mode", func() {

	})

	It("should be bound to the load balancer among specific sets with minimum rules in {name1},{name2} mode", func() {

	})
})

func createLoadBalancerService(c clientset.Interface, name string, annotation map[string]string, labels map[string]string, namespace string, ports []v1.ServicePort) (*v1.Service, error) {
	service := v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Annotations: annotation,
		},
		Spec: v1.ServiceSpec{
			Selector: labels,
			Ports:    ports,
			Type:     "LoadBalancer",
		},
	}
	return c.CoreV1().Services(namespace).Create(&service)
}

// DefaultDeployment returns a defualt deplotment
func portDeployment(name string, labels map[string]string) (result *v1beta1.Deployment) {
	var replicas int32
	replicas = 5
	result = &v1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1beta1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: v1.PodSpec{
					Hostname: name,
					Containers: []v1.Container{
						{
							Name:            "test-app",
							Image:           "nginx:1.15",
							ImagePullPolicy: "Always",
							Ports: []v1.ContainerPort{
								{
									ContainerPort: nginxPort,
								},
							},
						},
					},
				},
			},
		},
	}
	return
}

// As an ILB, two stuff require validationi:
// 1. external IP cannot be public
// 2. internal source can access to it
func validateInternalLoadBalancer(c clientset.Interface, ns string, url string) error {
	// create a pod to access to the service
	podName := "front-pod"
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: podName,
		},
		Spec: v1.PodSpec{
			Hostname: podName,
			Containers: []v1.Container{
				{
					Name:            "test-app",
					Image:           "nginx:1.15",
					ImagePullPolicy: "Always",
					Command: []string{
						"/bin/sh",
						"code=0",
						"while [ $code != 200 ]; do code=$(curl -s -o /dev/null -w \"%{http_code}\" " + url + "); echo $code; sleep 1; done",
					},
					Ports: []v1.ContainerPort{
						{
							ContainerPort: nginxPort,
						},
					},
				},
			},
		},
	}
	_, err := c.CoreV1().Pods(ns).Create(pod)
	if err != nil {
		return err
	}
	defer func() {
		err = testutils.WaitDeletePod(c, ns, podName)
	}()

	var publicFlag, internalFlag bool
	wait.PollImmediate(callPoll, callTimeout, func() (bool, error) {
		if !publicFlag {
			resp, err := http.Get(url)
			defer func() {
				if resp != nil {
					resp.Body.Close()
				}
			}()
			if err == nil {
				return false, fmt.Errorf("The load balancer is unexpectly external")
			}
			if !testutils.JudgeRetryable(err) {
				publicFlag = true
			}
		}

		if !internalFlag {
			// get pod command result
			request := c.CoreV1().Pods(ns).GetLogs(pod.Name, nil)
			s, _ := request.Stream()
			fmt.Println(s)
			if s != nil {
				internalFlag = true
			}
		}
		return publicFlag && internalFlag, nil
	})
	return err
}
