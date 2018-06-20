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
	"k8s.io/apimachinery/pkg/fields"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
)

func isRetryableAPIError(err error) bool {
	// These errors may indicate a transient error that we can retry in tests.
	if apierrs.IsInternalError(err) || apierrs.IsTimeout(err) || apierrs.IsServerTimeout(err) ||
		apierrs.IsTooManyRequests(err) || utilnet.IsProbableEOF(err) || utilnet.IsConnectionReset(err) {
		return true
	}
	// If the error sends the Retry-After header, we respect it as an explicit confirmation we should retry.
	if _, shouldRetry := apierrs.SuggestsClientDelay(err); shouldRetry {
		return true
	}
	return false
}

// WaitListSchedulableNodes is a wapper around listing nodes
func WaitListSchedulableNodes(c clientset.Interface) (*v1.NodeList, error) {
	var nodes *v1.NodeList
	var err error
	if wait.PollImmediate(Poll, SingleCallTimeout, func() (bool, error) {
		nodes, err = c.CoreV1().Nodes().List(metav1.ListOptions{FieldSelector: fields.Set{
			"spec.unschedulable": "false",
		}.AsSelector().String()})
		if err != nil {
			if isRetryableAPIError(err) {
				return false, nil
			}
			return false, err
		}
		return true, nil
	}) != nil {
		return nodes, err
	}
	return nodes, nil
}

// WaitAutoScaleNodes returns nodes count after autoscaling in 30 minutes
func WaitAutoScaleNodes(c clientset.Interface, targetNodeCount int) error {
	Logf(fmt.Sprintf("waiting for auto-scaling the node... Target node count: %v", targetNodeCount))
	var nodes *v1.NodeList
	var err error
	poll := 20 * time.Second
	autoScaleTimeOut := 30 * time.Minute
	if wait.PollImmediate(poll, autoScaleTimeOut, func() (bool, error) {
		nodes, err = c.CoreV1().Nodes().List(metav1.ListOptions{FieldSelector: fields.Set{
			"spec.unschedulable": "false",
		}.AsSelector().String()})
		if err != nil {
			if isRetryableAPIError(err) {
				return false, nil
			}
			return false, err
		}
		if nodes == nil {
			err = fmt.Errorf("Unexpected nil node list")
			return false, err
		}
		return targetNodeCount == len(nodes.Items), nil
	}) == wait.ErrWaitTimeout {
		return fmt.Errorf("There should be %v nodes after autoscaling, but only get %v", targetNodeCount, len(nodes.Items))
	} else {
		return err
	}
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
	Logf("Deleting node: %s", name)
	if err := cs.CoreV1().Nodes().Delete(name, nil); err != nil {
		return err
	}

	// wait for namespace to delete or timeout.
	err := wait.PollImmediate(2*time.Second, DeleteNSTimeout, func() (bool, error) {
		if _, err := cs.CoreV1().Nodes().Get(name, metav1.GetOptions{}); err != nil {
			return apierrs.IsNotFound(err), nil
		}
		return false, nil
	})
	return err
}
