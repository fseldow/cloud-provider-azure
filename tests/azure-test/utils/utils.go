package utils

import (
	"fmt"
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
	DeleteNSTimeout = 10 * time.Minute

	// How often to Poll pods, nodes and claims.
	Poll = 2 * time.Second

	pollShortTimeout = 1 * time.Minute
	pollLongTimeout  = 5 * time.Minute
)

var RunId = uuid.NewUUID()

/*
	GetClientSet obtains the client set interface from Kubeconfig
*/
func GetClientSet(filename string) (clientset.Interface, error) {
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

/*
	CreateTestingNS build namespace for each test
	baseName and labels determine name of the space
*/
func CreateTestingNS(baseName string, c clientset.Interface, labels map[string]string) (*v1.Namespace, error) {
	if labels == nil {
		labels = map[string]string{}
	}
	labels["e2e-run"] = string(RunId)

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
			//Logf("Unexpected error while creating namespace: %v", err)
			return false, nil
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
	//startTime := time.Now()
	if err := c.CoreV1().Namespaces().Delete(namespace, nil); err != nil {
		return err
	}

	// wait for namespace to delete or timeout.
	err := wait.PollImmediate(2*time.Second, DeleteNSTimeout, func() (bool, error) {
		if _, err := c.CoreV1().Namespaces().Get(namespace, metav1.GetOptions{}); err != nil {
			if apierrs.IsNotFound(err) {
				return true, nil
			}
			//Logf("Error while waiting for namespace to be terminated: %v", err)
			return false, nil
		}
		return false, nil
	})

	return err
}
