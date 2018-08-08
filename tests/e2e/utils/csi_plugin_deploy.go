package utils

import (
	"os"

	clientset "k8s.io/client-go/kubernetes"
)

// DeployCsiPlugin depolys the csi plugin
func DeployCsiPlugin(cs clientset.Interface) error {
	ns, err := CreateNamespace(cs, csiNamespace)
	if err != nil {
		return err
	}

	// Temp solution
	Logf("creating secret")
	envData := map[string][]byte{
		"tenantId":     []byte(os.Getenv("K8S_AZURE_TENANTID")),
		"clientId":     []byte(os.Getenv("K8S_AZURE_SPID")),
		"clientSecret": []byte(os.Getenv("K8S_AZURE_SPSEC")),
	}

	secret := createSecretManifest(envData)
	cs.CoreV1().Secrets(ns.Name).Create(secret)

	storageClass := createStorageClassManifest()
	cs.StorageV1().StorageClasses().Create(storageClass)

	createPlugins(cs, ns.Name)

	return nil
}

func createPlugins(cs clientset.Interface, namespace string) {
	attacherServiceAccount := createAttacherServiceAccountManifest()
	cs.CoreV1().ServiceAccounts(namespace).Create(attacherServiceAccount)
	attacherClusterRole := createAttacherClusterRoleManifest()
	cs.RbacV1().ClusterRoles().Create(attacherClusterRole)
	attacherClusterRoleBinding := createAttacherClusterRoleBindingManifest()
	cs.RbacV1().ClusterRoleBindings().Create(attacherClusterRoleBinding)
	attacherStatefulSetName := createAttacherStatefulSetManifest()
	cs.AppsV1beta1().StatefulSets(namespace).Create(attacherStatefulSetName)

	pluginServiceAccount := createPluginServiceAccountManifest()
	cs.CoreV1().ServiceAccounts(namespace).Create(pluginServiceAccount)
	pluginClusterRole := createPluginClusterRoleManifest()
	cs.RbacV1().ClusterRoles().Create(pluginClusterRole)
	pluginClusterRoleBinding := createPluginClusterRoleBindingManifest()
	cs.RbacV1().ClusterRoleBindings().Create(pluginClusterRoleBinding)
	pluginClusterDaemonSet := createPluginDeamonSetManifest()
	cs.AppsV1().DaemonSets(namespace).Create(pluginClusterDaemonSet)

	provisionerServiceAccount := createProvisionerServiceAccountManifest()
	cs.CoreV1().ServiceAccounts(namespace).Create(provisionerServiceAccount)
	provisionerClusterRole := createProvisionerClusterRoleManifest()
	cs.RbacV1().ClusterRoles().Create(provisionerClusterRole)
	provisionerClusterRoleBinding := createProvisionerClusterRoleBindingManifest()
	cs.RbacV1().ClusterRoleBindings().Create(provisionerClusterRoleBinding)
	provisionerStatefulSetName := createProvisionerStatefulSetManifest()
	cs.AppsV1beta1().StatefulSets(namespace).Create(provisionerStatefulSetName)
}
