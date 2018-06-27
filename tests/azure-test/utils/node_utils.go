package utils

import (
	"time"

	"k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
)

// WaitListSchedulableNodes is a wapper around listing nodes
func WaitListSchedulableNodes(c clientset.Interface) (*v1.NodeList, error) {
	var nodes *v1.NodeList
	var err error
	if wait.PollImmediate(Poll, SingleCallTimeout, func() (bool, error) {
		nodes, err = c.CoreV1().Nodes().List(metav1.ListOptions{FieldSelector: fields.Set{
			"spec.unschedulable": "false",
		}.AsSelector().String()})
		if err != nil {
			return false, err
		}
		return true, nil
	}) != nil {
		return nodes, err
	}
	return nodes, nil
}

// WaitAutoScaleNodes returns nodes count after autoscaling in 10 minutes
func WaitAutoScaleNodes(c clientset.Interface, targetNodeCount int) (int, error) {
	var nodes *v1.NodeList
	var err error
	if wait.PollImmediate(Poll, AutoScaleTimeOut, func() (bool, error) {
		nodes, err = c.CoreV1().Nodes().List(metav1.ListOptions{FieldSelector: fields.Set{
			"spec.unschedulable": "false",
		}.AsSelector().String()})
		if err != nil {
			return false, err
		}
		return targetNodeCount == len(nodes.Items), nil
	}) != nil {
		return len(nodes.Items), err
	}
	return len(nodes.Items), nil
}

// WaitNodeDeletion ensures a list of nodes to be deleted
func WaitNodeDeletion(cs clientset.Interface, names []string) error {
	for _, name := range names {
		if err := deleteNode(cs, name); err != nil {
			return err
		}
	}
	return nil
}

//deleteNodes deletes nodes according to names
func deleteNode(cs clientset.Interface, name string) error {
	if err := cs.CoreV1().Nodes().Delete(name, nil); err != nil {
		return err
	}

	// wait for namespace to delete or timeout.
	err := wait.PollImmediate(2*time.Second, DeleteNSTimeout, func() (bool, error) {
		if _, err := cs.CoreV1().Nodes().Get(name, metav1.GetOptions{}); err != nil {
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
