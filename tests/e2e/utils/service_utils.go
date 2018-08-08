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
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
)

// GetServiceDomainName cat prefix and azure suffix
func GetServiceDomainName(prefix string) (ret string) {
	suffix := extractSuffix()
	ret = prefix + suffix
	Logf("Get domain name: %s", ret)
	return
}

// WaitServiceExposure returns ip of ingress
func WaitServiceExposure(cs clientset.Interface, namespace string, name string) error {
	var service *v1.Service
	var err error

	if wait.PollImmediate(10*time.Second, 10*time.Minute, func() (bool, error) {
		service, err = cs.CoreV1().Services(namespace).Get(name, metav1.GetOptions{})
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
		return true, nil
	}) != nil {
		return err
	}
	return nil
}

// DeleteServiceAccount deletes a ServiceAccount
func DeleteServiceAccount(cs clientset.Interface, namespace string, serviceAccountName string) error {
	Logf("Deleting service account %s in namespace %s", namespace, serviceAccountName)
	err := cs.CoreV1().ServiceAccounts(namespace).Delete(serviceAccountName, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return wait.PollImmediate(poll, deletionTimeout, func() (bool, error) {
		if _, err := cs.CoreV1().ServiceAccounts(namespace).Get(serviceAccountName, metav1.GetOptions{}); err != nil {
			if apierrs.IsNotFound(err) {
				return true, nil
			}
			Logf("Error while waiting for service account to be deleted: %v", err)
			return false, nil
		}
		return false, nil
	})
}
