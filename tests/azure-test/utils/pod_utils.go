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
		//label := labels.SelectorFromSet(labels.Set(map[string]string{"app": "cassandra"}))
		pods, err = c.CoreV1().Pods(ns).List(metav1.ListOptions{})
		if err != nil {
			return false, err
		}
		return true, nil
	}) != nil {
		return pods, err
	}
	return pods, nil
}
