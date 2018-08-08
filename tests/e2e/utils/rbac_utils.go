package utils

import (
	"k8s.io/api/rbac/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
)

func CreateSecret() {

}

func DeleteSecret() {

}

// CreateClusterRole creates a ClusterRole
func CreateClusterRole(cs clientset.Interface, clusterRole *v1.ClusterRole) (cr *v1.ClusterRole, err error) {
	Logf("Creating cluster role %s", clusterRole.Name)
	cr, err = cs.Rbac().ClusterRoles().Create(clusterRole)
	if apierrs.IsAlreadyExists(err) {
		err = DeleteClusterRoleBinding(cs, clusterRole.Name)
		if err != nil {
			return
		}
		cr, err = cs.Rbac().ClusterRoles().Create(clusterRole)
		if err != nil {
			return
		}
	} else if err != nil {
		return
	}
	err = wait.PollImmediate(poll, deletionTimeout, func() (bool, error) {
		if cr, err = cs.Rbac().ClusterRoles().Get(clusterRole.Name, metav1.GetOptions{}); err != nil {
			if apierrs.IsNotFound(err) {
				return true, nil
			}
			Logf("Error while waiting for namespace to be terminated: %v", err)
			return false, nil
		}
		return false, nil
	})
	return
}

// DeleteClusterRole deletes a ClusterRole
func DeleteClusterRole(cs clientset.Interface, clusterRoleName string) error {
	Logf("Deleting cluster role %s", clusterRoleName)
	cs.Rbac().ClusterRoles().Delete(clusterRoleName, &metav1.DeleteOptions{})
	return wait.PollImmediate(poll, deletionTimeout, func() (bool, error) {
		if _, err := cs.Rbac().ClusterRoles().Get(clusterRoleName, metav1.GetOptions{}); err != nil {
			if apierrs.IsNotFound(err) {
				return true, nil
			}
			Logf("Error while waiting for cluster role name to be deleted: %v", err)
			return false, nil
		}
		return false, nil
	})
}

// CreateClusterRoleBinding creates ClusterRoleBinding according to the manifest
func CreateClusterRoleBinding(cs clientset.Interface, clusterRoleBinding *v1.ClusterRoleBinding) (crb *v1.ClusterRoleBinding, err error) {
	Logf("Creating cluster role binding %s", clusterRoleBinding.Name)
	crb, err = cs.Rbac().ClusterRoleBindings().Create(clusterRoleBinding)
	if apierrs.IsAlreadyExists(err) {
		err = DeleteClusterRoleBinding(cs, clusterRoleBinding.Name)
		if err != nil {
			return
		}
		crb, err = cs.Rbac().ClusterRoleBindings().Create(clusterRoleBinding)
		if err != nil {
			return
		}
	} else if err != nil {
		return
	}
	err = wait.PollImmediate(poll, deletionTimeout, func() (bool, error) {
		if crb, err = cs.Rbac().ClusterRoleBindings().Get(clusterRoleBinding.Name, metav1.GetOptions{}); err != nil {
			if apierrs.IsNotFound(err) {
				return true, nil
			}
			Logf("Error while waiting for namespace to be terminated: %v", err)
			return false, nil
		}
		return false, nil
	})
	return
}

// DeleteClusterRoleBinding deletes ClusterRoleBinding
func DeleteClusterRoleBinding(cs clientset.Interface, clusterRoleBindingName string) error {
	Logf("Deleting cluster role binding %s", clusterRoleBindingName)
	cs.Rbac().ClusterRoleBindings().Delete(clusterRoleBindingName, &metav1.DeleteOptions{})
	return wait.PollImmediate(poll, deletionTimeout, func() (bool, error) {
		if _, err := cs.Rbac().ClusterRoleBindings().Get(clusterRoleBindingName, metav1.GetOptions{}); err != nil {
			if apierrs.IsNotFound(err) {
				return true, nil
			}
			Logf("Error while waiting for clusterRoleBinding to be deleted: %v", err)
			return false, nil
		}
		return false, nil
	})
}
