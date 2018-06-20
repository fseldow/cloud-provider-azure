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
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
)

const (
	pauseImage = "k8s.gcr.io/pause:3.1"
)

//DefaultPod returns a default pod representation
func DefaultPod(name string) (result *v1.Pod) {
	result = &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "container",
					Image: pauseImage,
				},
			},
		},
	}
	return
}

// WaitListPods is a wapper around listing pods
func WaitListPods(c clientset.Interface, ns string) (*v1.PodList, error) {
	var pods *v1.PodList
	var err error
	if wait.PollImmediate(Poll, SingleCallTimeout, func() (bool, error) {
		pods, err = c.CoreV1().Pods(ns).List(metav1.ListOptions{})
		if err != nil {
			if isRetryableAPIError(err) {
				return false, nil
			}
			return false, err
		}
		return true, nil
	}) != nil {
		return pods, err
	}
	return pods, nil
}
