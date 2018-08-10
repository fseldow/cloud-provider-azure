package utils

import (
	"os"

	apierrs "k8s.io/apimachinery/pkg/api/errors"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/cmd/kubeadm/app/util"
)

// DeployCsiPlugin depolys the csi plugin
func DeployCsiPlugin(cs clientset.Interface) error {
	Logf("Deploying the csi plugins")
	ns, err := CreateNamespace(cs, csiNamespace)
	if err != nil {
		if apierrs.IsAlreadyExists(err) {
			// Assume already built the csi plugins
			// and there will be no other namespace use the same name
			Logf("Already deployed CSI plugins, skip deployment.")
			return nil
		}
		return err
	}

	// Directly read service account from environment variable
	Logf("creating secret")
	envData := map[string][]byte{
		"tenantId":     []byte(os.Getenv("K8S_AZURE_TENANTID")),
		"clientId":     []byte(os.Getenv("K8S_AZURE_SPID")),
		"clientSecret": []byte(os.Getenv("K8S_AZURE_SPSEC")),
	}

	secret := createSecretManifest(envData)
	if _, err := cs.CoreV1().Secrets(csiNamespace).Create(secret); err != nil {
		return err
	}

	storageClass := createStorageClassManifest()
	if _, err = cs.StorageV1().StorageClasses().Create(storageClass); err != nil {
		return err
	}

	if err := createPlugins(cs, ns.Name); err != nil {
		return err
	}
	return nil
}

//CleanupCsiPlugins cleans up the csi plugins
func CleanupCsiPlugins(cs clientset.Interface) error {
	Logf("Cleaning up the csi plugins")
	if err := DeleteNameSpace(cs, csiNamespace); err != nil {
		return err
	}
	if err := DeleteClusterRole(cs, attacherClusterRoleName); err != nil {
		return err
	}
	if err := DeleteClusterRole(cs, pluginClusterRoleName); err != nil {
		return err
	}
	if err := DeleteClusterRole(cs, provisionerClusterRoleName); err != nil {
		return err
	}
	if err := DeleteClusterRoleBinding(cs, attacherClusterRoleBindingName); err != nil {
		return err
	}
	if err := DeleteClusterRoleBinding(cs, pluginClusterRoleBindingName); err != nil {
		return err
	}
	if err := DeleteClusterRoleBinding(cs, provisionerClusterRoleBindingName); err != nil {
		return err
	}
	if err := DeleteStorageClass(cs, storageClassName); err != nil {
		return err
	}
	return nil
}

func createPlugins(cs clientset.Interface, namespace string) error {
	util.SplitYAMLDocuments()
	util.UnmarshalFromYaml()
	/*
		funcs := []func(clientset.Interface, string) error{
			createAttacherServiceAccount,
			createAttacherClusterRole,
			createAttacherClusterRoleBinding,
			createAttacherStatefulSet,
			createPluginServiceAccount,
			createPluginClusterRole,
			createPluginClusterRoleBinding,
			createPluginDaemonSet,
			createProvisionerServiceAccount,
			createProvisionerClusterRole,
			createProvisionerClusterRoleBinding,
			createProvisionerStatefulSet,
		}
		for _, fun := range funcs {
			if err := fun(cs, namespace); err != nil {
				return err
			}
		}
	*/
	return nil
}

func createAttacherServiceAccount(cs clientset.Interface, namespace string) error {
	obj := createAttacherServiceAccountManifest()
	Logf("Creating attacher service account")
	_, err := cs.CoreV1().ServiceAccounts(namespace).Create(obj)
	return err
}

func createAttacherClusterRole(cs clientset.Interface, namespace string) error {
	obj := createAttacherClusterRoleManifest()
	Logf("Creating attacher cluster role")
	_, err := cs.RbacV1().ClusterRoles().Create(obj)
	return err
}

func createAttacherClusterRoleBinding(cs clientset.Interface, namespace string) error {
	obj := createAttacherClusterRoleBindingManifest()
	Logf("Creating attacher cluster role binding")
	_, err := cs.RbacV1().ClusterRoleBindings().Create(obj)
	return err
}

func createAttacherStatefulSet(cs clientset.Interface, namespace string) error {
	obj := createAttacherStatefulSetManifest()
	Logf("Creating attacher stateful set")
	_, err := cs.AppsV1beta1().StatefulSets(namespace).Create(obj)
	return err
}

func createPluginServiceAccount(cs clientset.Interface, namespace string) error {
	obj := createPluginServiceAccountManifest()
	Logf("Creating plugin service account")
	_, err := cs.CoreV1().ServiceAccounts(namespace).Create(obj)
	return err
}

func createPluginClusterRole(cs clientset.Interface, namespace string) error {
	obj := createPluginClusterRoleManifest()
	Logf("Creating plugin cluster role")
	_, err := cs.RbacV1().ClusterRoles().Create(obj)
	return err
}

func createPluginClusterRoleBinding(cs clientset.Interface, namespace string) error {
	obj := createPluginClusterRoleBindingManifest()
	Logf("Creating plugin cluster role binding")
	_, err := cs.RbacV1().ClusterRoleBindings().Create(obj)
	return err
}

func createPluginDaemonSet(cs clientset.Interface, namespace string) error {
	obj := createPluginDaemonSetManifest()
	Logf("Creating plugin daemon set")
	_, err := cs.AppsV1().DaemonSets(namespace).Create(obj)
	return err
}

func createProvisionerServiceAccount(cs clientset.Interface, namespace string) error {
	obj := createProvisionerServiceAccountManifest()
	Logf("Creating provisioner service account")
	_, err := cs.CoreV1().ServiceAccounts(namespace).Create(obj)
	return err
}

func createProvisionerClusterRole(cs clientset.Interface, namespace string) error {
	obj := createProvisionerClusterRoleManifest()
	Logf("Creating provisioner cluster role")
	_, err := cs.RbacV1().ClusterRoles().Create(obj)
	return err
}

func createProvisionerClusterRoleBinding(cs clientset.Interface, namespace string) error {
	obj := createProvisionerClusterRoleBindingManifest()
	Logf("Creating provisioner cluster role binding")
	_, err := cs.RbacV1().ClusterRoleBindings().Create(obj)
	return err
}

func createProvisionerStatefulSet(cs clientset.Interface, namespace string) error {
	obj := createProvisionerStatefulSetManifest()
	Logf("Creating provisioner stateful set")
	_, err := cs.AppsV1beta1().StatefulSets(namespace).Create(obj)
	return err
}
