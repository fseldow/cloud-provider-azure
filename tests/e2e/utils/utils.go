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
	"strings"
	"time"

	"k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

const (
	DeleteNSTimeout   = 10 * time.Minute
	Poll              = 2 * time.Second
	SingleCallTimeout = 5 * time.Minute

	PodImage = "k8s.gcr.io/pause:3.1"
)

func findExistingKubeConfig() string {
	defaultKubeConfig := "/etc/kubernetes/admin.conf"
	// locations using DefaultClientConfigLoadingRules, but also consider `defaultKubeConfig`.
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	rules.Precedence = append(rules.Precedence, defaultKubeConfig)
	return rules.GetDefaultFilename()
}

// GetClientSet obtains the client set interface from Kubeconfig
func GetClientSet() (clientset.Interface, error) {
	//TODO: It should implement only once
	//rather than once per test
	Logf("Creating a kubernetes client")
	filename := findExistingKubeConfig()
	c := clientcmd.GetConfigFromFileOrDie(filename)
	restConfig, err := clientcmd.NewDefaultClientConfig(*c, &clientcmd.ConfigOverrides{ClusterInfo: clientcmdapi.Cluster{Server: ""}}).ClientConfig()
	if err != nil {
		return nil, err
	}
	clientSet, err := clientset.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	return clientSet, nil
}

// ExtractDNSPrefix obtains the cluster DNS prefix
func ExtractDNSPrefix() string {
	c := obtainConfig()
	return c.CurrentContext
}

// ExtractRegion obtains the cluster region
func ExtractRegion() string {
	c := obtainConfig()
	prefix := ExtractDNSPrefix()
	url := c.Clusters[prefix].Server
	domain := strings.Split(url, ".")
	if len(domain) < 4 {
		return "NaN"
	}
	return domain[len(domain)-4]
}

// ExtractSuffix obtains the server domain name suffix
func ExtractSuffix() string {
	c := obtainConfig()
	prefix := ExtractDNSPrefix()
	url := c.Clusters[prefix].Server
	domain := strings.Split(url, ".")
	suffix := url[len(domain[0]):len(url)]
	return suffix
}

// Load config from file
func obtainConfig() *clientcmdapi.Config {
	filename := findExistingKubeConfig()
	c := clientcmd.GetConfigFromFileOrDie(filename)
	return c
}

//CreateTestingNS builds namespace for each test
//baseName and labels determine name of the space
func CreateTestingNS(baseName string, c clientset.Interface) (*v1.Namespace, error) {
	Logf("Creating a test namespace")
	labels := map[string]string{}
	var runID = uuid.NewUUID()
	labels["e2e-run"] = string(runID)

	namespaceObj := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("e2e-tests-%v-", baseName),
			Namespace:    "",
			Labels:       labels,
		},
		Status: v1.NamespaceStatus{},
	}
	// Be robust about making the namespace creation call.
	var got *v1.Namespace
	if err := wait.PollImmediate(Poll, 30*time.Second, func() (bool, error) {
		var err error
		got, err = c.CoreV1().Namespaces().Create(namespaceObj)
		if err != nil {
			if isRetryableAPIError(err) {
				return false, nil
			}
			return false, err
		}
		return true, nil
	}); err != nil {
		return nil, err
	}
	return got, nil
}

// DeleteNS deletes the provided namespace, waits for it to be completely deleted, and then checks
// whether there are any pods remaining in a non-terminating state.
func DeleteNS(c clientset.Interface, namespace string) error {
	Logf("Deleting namespace %s", namespace)
	if err := c.CoreV1().Namespaces().Delete(namespace, nil); err != nil {
		return err
	}
	// wait for namespace to delete or timeout.
	err := wait.PollImmediate(2*time.Second, DeleteNSTimeout, func() (bool, error) {
		if _, err := c.CoreV1().Namespaces().Get(namespace, metav1.GetOptions{}); err != nil {
			if apierrs.IsNotFound(err) {
				return true, nil
			}
			Logf("Error while waiting for namespace to be terminated: %v", err)
			return false, nil
		}
		return false, nil
	})

	return err
}

// StringInSlice check if string in a list
func StringInSlice(s string, list []string) bool {
	for _, item := range list {
		if item == s {
			return true
		}
	}
	return false
}
