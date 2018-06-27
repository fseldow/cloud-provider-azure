package utils

import (
	"fmt"
	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
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
		},
		Spec: v1.ServiceSpec{
			Selector: labels,
			Ports:    ports,
			Type:     "LoadBalancer",
		},
	}
	return c.CoreV1().Services(namespace).Create(service)
}

func WaitExternelIP(c clientset.Interface, namespace string, name string) error {
	var service *v1.Service
	var err error
	if wait.PollImmediate(10*time.Second, 10*time.Minute, func() (bool, error) {
		service, err = c.CoreV1().Services(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		IngressList := service.Status.LoadBalancer.Ingress
		if len(IngressList) == 0 {
			return false, fmt.Errorf("Cannot find Ingress")
		}
		ExternalIP := IngressList[0].IP
		fmt.Printf(ExternalIP)
		return true, nil
	}) != nil {
		return err
	}
	return nil
}
