package utils

import (
	"k8s.io/api/apps/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
)

// CreateStatefulSet creates a stateful set
func CreateStatefulSet(cs clientset.Interface, statefulSet *v1.StatefulSet) {

}

// DeleteStatefulSet deletes a stateful set
func DeleteStatefulSet(cs clientset.Interface, namespace string, statefulSetName string) error {
	Logf("Deleting statful set %s in namespace %s", statefulSetName, namespace)
	err := cs.AppsV1().StatefulSets(namespace).Delete(statefulSetName, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return wait.PollImmediate(poll, deletionTimeout, func() (bool, error) {
		if _, err := cs.AppsV1().StatefulSets(namespace).Get(statefulSetName, metav1.GetOptions{}); err != nil {
			if apierrs.IsNotFound(err) {
				return true, nil
			}
			Logf("Error while waiting for stateful set to be deleted: %v", err)
			return false, nil
		}
		return false, nil
	})
}

// CreateDaemonSet creates a daemonset
func CreateDaemonSet(cs clientset.Interface, statefulSet *v1.DaemonSet) {

}

// DeleteDaemonSet deletes a daemonset
func DeleteDaemonSet(cs clientset.Interface, namespace string, daemonSetName string) error {
	Logf("Deleting daemon set %s in namespace %s", daemonSetName, namespace)
	err := cs.AppsV1().DaemonSets(namespace).Delete(daemonSetName, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return wait.PollImmediate(poll, deletionTimeout, func() (bool, error) {
		if _, err := cs.AppsV1().DaemonSets(namespace).Get(daemonSetName, metav1.GetOptions{}); err != nil {
			if apierrs.IsNotFound(err) {
				return true, nil
			}
			Logf("Error while waiting for daemon set to be deleted: %v", err)
			return false, nil
		}
		return false, nil
	})
}
