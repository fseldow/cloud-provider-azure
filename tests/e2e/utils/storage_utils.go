package utils

import (
	"k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
)

// DeleteStorageClass deletes a storage class
func DeleteStorageClass(cs clientset.Interface, scName string) error {
	Logf("Deleting storage class: %s", scName)
	err := cs.StorageV1().StorageClasses().Delete(scName, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return wait.PollImmediate(poll, deletionTimeout, func() (bool, error) {
		if _, err := cs.StorageV1().StorageClasses().Get(scName, metav1.GetOptions{}); err != nil {
			if apierrs.IsNotFound(err) {
				return true, nil
			}
			Logf("Error while waiting for storage class to be deleted: %v", err)
			return false, nil
		}
		return false, nil
	})
}

// DeletePersistentVolumeClaim deletes a presistence volume claim
func DeletePersistentVolumeClaim(cs clientset.Interface, namespace string, pvcName string) error {
	Logf("Deleting persistent volume claim %s", pvcName)
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

// GetPersistentVolumeList is a wapper around listing pods
func GetPersistentVolumeList(cs clientset.Interface) (*v1.PersistentVolumeList, error) {
	var pvList *v1.PersistentVolumeList
	var err error
	Logf("Fetching persistent volume list")
	if wait.PollImmediate(poll, singleCallTimeout, func() (bool, error) {
		pvList, err = cs.CoreV1().PersistentVolumes().List(metav1.ListOptions{})
		if err != nil {
			if isRetryableAPIError(err) {
				return false, nil
			}
			return false, err
		}
		return true, nil
	}) != nil {
		return pvList, err
	}
	return pvList, nil
}

// DeletePersistentVolume deletes a presistence volume claim
func DeletePersistentVolume(cs clientset.Interface, namespace string, pvName string) error {
	Logf("Deleting persistent volume %s", pvName)
	err := cs.CoreV1().PersistentVolumes().Delete(pvName, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return wait.PollImmediate(poll, deletionTimeout, func() (bool, error) {
		if _, err := cs.CoreV1().PersistentVolumes().Get(pvName, metav1.GetOptions{}); err != nil {
			if apierrs.IsNotFound(err) {
				return true, nil
			}
			Logf("Error while waiting for persistent volume to be deleted: %v", err)
			return false, nil
		}
		return false, nil
	})
}
