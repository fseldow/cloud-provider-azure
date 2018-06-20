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
	"fmt"
	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
)

const (
	serviceAnnotationDNSLabelName = "service.beta.kubernetes.io/azure-dns-label-name"
)

// DefaultService returns a default service representation
func DefaultService(name string) (result *v1.Service, labels map[string]string) {
	result = &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1.ServiceSpec{
			Selector: labels,
		},
	}
	return
}

// CreateLoadBalancerService creates a new service with service port
func CreateLoadBalancerService(c clientset.Interface, name string, labels map[string]string, namespace string, ports []v1.ServicePort) (*v1.Service, error) {
	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Annotations: map[string]string{
				serviceAnnotationDNSLabelName: name,
			},
		},
		Spec: v1.ServiceSpec{
			Selector: labels,
			Ports:    ports,
			Type:     "LoadBalancer",
		},
	}
	return c.CoreV1().Services(namespace).Create(service)
}

func getServiceDNS(service *v1.Service) string {
	suffix := ExtractSuffix()
	return service.Annotations[serviceAnnotationDNSLabelName] + suffix
}

// WaitExternalDNS returns ip of ingress
func WaitExternalDNS(c clientset.Interface, namespace string, name string) (string, error) {
	var service *v1.Service
	var err error
	var ExternalIP string
	var DNS string

	if wait.PollImmediate(10*time.Second, 10*time.Minute, func() (bool, error) {
		service, err = c.CoreV1().Services(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			if isRetryableAPIError(err) {
				return false, nil
			}
			return false, err
		}

		IngressList := service.Status.LoadBalancer.Ingress
		if IngressList == nil || len(IngressList) == 0 {
			err = fmt.Errorf("Cannot find Ingress")
			return false, nil
		}

		ExternalIP = IngressList[0].IP
		DNS = getServiceDNS(service)

		Logf(fmt.Sprintf("get external IP: %s", ExternalIP))
		Logf(fmt.Sprintf("Domain name: %s", DNS))
		return true, nil
	}) != nil {
		return DNS, err
	}
	return DNS, nil
}
