package utils

import (
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
)

// CreatePersistentVolumeClaim creates a presistence volume claim
func CreatePersistentVolumeClaim() {

}

// DeletePersistentVolumeClaim deletes a presistence volume claim
func DeletePersistentVolumeClaim(cs clientset.Interface, namespace string, pvcName string) error {
	err := cs.CoreV1().PersistentVolumeClaims(namespace).Delete(pvcName, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return wait.PollImmediate(poll, deletionTimeout, func() (bool, error) {
		if _, err := cs.CoreV1().PersistentVolumeClaims(namespace).Get(pvcName, metav1.GetOptions{}); err != nil {
			if apierrs.IsNotFound(err) {
				return true, nil
			}
			Logf("Error while waiting for persistent volume claim to be deleted: %v", err)
			return false, nil
		}
		return false, nil
	})
}
