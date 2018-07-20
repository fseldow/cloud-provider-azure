package main

import (
	"fmt"
	"net/http"
	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	testutils "k8s.io/cloud-provider-azure/tests/e2e/utils"
)

const (
	nginxPort       = 80
	nginxStatusCode = 200
	callPoll        = 20 * time.Second
	callTimeout     = 10 * time.Minute
)

func validateInternalLoadBalancer(c clientset.Interface, ns string, url string) error {
	// create a pod to access to the service
	podName := "front-pod"
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: podName,
		},
		Spec: v1.PodSpec{
			Hostname: podName,
			Containers: []v1.Container{
				{
					Name:            "test-app",
					Image:           "nginx:1.15",
					ImagePullPolicy: "Always",
					Command: []string{
						"/bin/sh",
						"code=0",
						"while [ $code != 200 ]; do code=$(curl -s -o /dev/null -w \"%{http_code}\" " + url + "); echo $code; sleep 1; done",
					},
					Ports: []v1.ContainerPort{
						{
							ContainerPort: nginxPort,
						},
					},
				},
			},
		},
	}
	_, err := c.CoreV1().Pods(ns).Create(pod)
	if err != nil {
		return err
	}
	defer func() {
		err = testutils.WaitDeletePod(c, ns, podName)
	}()

	var publicFlag, internalFlag bool
	wait.PollImmediate(1*time.Second, callTimeout, func() (bool, error) {
		if !publicFlag {
			resp, err := http.Get(url)
			defer func() {
				if resp != nil {
					resp.Body.Close()
				}
			}()
			if err == nil {
				return false, fmt.Errorf("The load balancer is unexpectly external")
			}
			if !testutils.JudgeRetryable(err) {
				publicFlag = true
			}
		}

		if !internalFlag {
			// get pod command result
			request := c.CoreV1().Pods(ns).GetLogs(pod.Name, nil)
			b, _ := request.Stream()
			fmt.Print(b)
			if true {
				internalFlag = true
			}
		}
		return publicFlag && internalFlag, nil
	})
	return err
}

func main() {
	url := "www.google.com"
	a := "while [ $code != 200 ]; do code=$(curl -s -o /dev/null -w \"%{http_code}\" " + url + "); echo $code; sleep 1; done"
	fmt.Print(a)
}
