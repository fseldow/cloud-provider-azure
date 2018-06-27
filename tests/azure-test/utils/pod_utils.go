package utils

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
