package utils

import (
	app_v1 "k8s.io/api/apps/v1"
	app_v1beta1 "k8s.io/api/apps/v1beta1"
	core_v1 "k8s.io/api/core/v1"
	rbac_v1 "k8s.io/api/rbac/v1"
	storage_v1 "k8s.io/api/storage/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	csiNamespace  = "csi-plugins-azuredisk"
	csiSecretName = "csi-azuredisk-secret"

	storageClassName = "azuredisk-csi"
	provisionerName  = "csi-azuredisk"

	socketMount     = "socket-dir"
	socketMountPath = "/var/lib/kubelet/plugins/csi-azuredisk"
	podMount        = "pods-mount-dir"
	podMountPath    = "/var/lib/kubelet/pods"
	devMount        = "dev-dir"
	devMountPath    = "/dev"

	attacherServiceAccountName     = "csi-attacher"
	attacherClusterRoleName        = "external-attacher-runner"
	attacherClusterRoleBindingName = "csi-attacher-role"
	attacherStatefulSetName        = "csi-attacher"

	pluginServiceAccountName     = "csi-plugin-azuredisk"
	pluginClusterRoleName        = "csi-plugin-azuredisk-runner"
	pluginClusterRoleBindingName = "csi-plugin-azuredisk-role"
	pluginDaemonSetName          = "csi-plugin-azuredisk"

	provisionerServiceAccountName     = "csi-provisioner"
	provisionerClusterRoleName        = "external-provisioner-runner"
	provisionerClusterRoleBindingName = "csi-provisioner-role"
	provisionerStatefulSetName        = "csi-provisioner"
)

var replicas = int32(1)
var privileged = true
var allowPrivilegeEscalation = true
var mountPropagationBidirectional = core_v1.MountPropagationBidirectional
var hostPathDirectoryOrCreate = core_v1.HostPathDirectoryOrCreate
var hostPathDirectory = core_v1.HostPathDirectory

func createSecretManifest(data map[string][]byte) *core_v1.Secret {
	return &core_v1.Secret{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: csiSecretName,
		},
		Data: data,
	}
}

func createStorageClassManifest() *storage_v1.StorageClass {
	persistentVolumeReclaimDelete := core_v1.PersistentVolumeReclaimDelete
	return &storage_v1.StorageClass{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: storageClassName,
		},
		Provisioner:   provisionerName,
		ReclaimPolicy: &persistentVolumeReclaimDelete,
		Parameters: map[string]string{
			"csiProvisionerSecretName":            csiSecretName,
			"csiProvisionerSecretNamespace":       csiNamespace,
			"csiControllerPublishSecretName":      csiSecretName,
			"csiControllerPublishSecretNamespace": csiNamespace,
		},
	}
}

// attacher
func createAttacherServiceAccountManifest() *core_v1.ServiceAccount {
	return &core_v1.ServiceAccount{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: attacherServiceAccountName,
		},
	}
}

func createAttacherClusterRoleManifest() *rbac_v1.ClusterRole {
	return &rbac_v1.ClusterRole{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: attacherClusterRoleName,
		},
		Rules: []rbac_v1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"persistentvolumes"},
				Verbs:     []string{"get", "list", "watch", "update"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"nodes"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"storage.k8s.io"},
				Resources: []string{"volumesattachments"},
				Verbs:     []string{"get", "list", "watch", "update"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"get"},
			},
		},
	}
}

func createAttacherClusterRoleBindingManifest() *rbac_v1.ClusterRoleBinding {
	return &rbac_v1.ClusterRoleBinding{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: attacherClusterRoleBindingName,
		},
		Subjects: []rbac_v1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      attacherServiceAccountName,
				Namespace: csiNamespace,
			},
		},
		RoleRef: rbac_v1.RoleRef{
			Kind:     "ClusterRole",
			Name:     attacherClusterRoleName,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
}

func createAttacherStatefulSetManifest() *app_v1beta1.StatefulSet {
	return &app_v1beta1.StatefulSet{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: attacherStatefulSetName,
		},
		Spec: app_v1beta1.StatefulSetSpec{
			ServiceName: attacherServiceAccountName,
			Replicas:    &replicas,
			Template: core_v1.PodTemplateSpec{
				ObjectMeta: meta_v1.ObjectMeta{
					Labels: map[string]string{
						"app": attacherServiceAccountName,
					},
				},
				Spec: core_v1.PodSpec{
					ServiceAccountName: attacherServiceAccountName,
					Containers: []core_v1.Container{
						{
							Name:  attacherServiceAccountName,
							Image: "quay.io/k8scsi/csi-attacher:v0.2.0",
							Args:  []string{"--v=5", "--csi-address=$(CSI_ENDPOINT)"},
							Env: []core_v1.EnvVar{
								{
									Name:  "CSI_ENDPOINT",
									Value: "/var/lib/kubelet/plugins/csi-azuredisk/csi.sock",
								},
							},
							ImagePullPolicy: core_v1.PullIfNotPresent,
							VolumeMounts: []core_v1.VolumeMount{
								{
									Name:      socketMount,
									MountPath: socketMountPath,
								},
							},
						},
						{
							Name: "plugin",
							//TODO:
							//should dynamic build the plugin image
							Image: "karataliu/csi-azuredisk:3",
							Env: []core_v1.EnvVar{
								{
									Name:  "CSI_ENDPOINT",
									Value: "/var/lib/kubelet/plugins/csi-azuredisk/csi.sock",
								},
								{
									Name:  "CSI_SERVICE_DISABLE_NODE",
									Value: "1",
								},
							},
							ImagePullPolicy: core_v1.PullAlways,
							VolumeMounts: []core_v1.VolumeMount{
								{
									Name:      socketMount,
									MountPath: socketMountPath,
								},
							},
						},
					},
					Volumes: []core_v1.Volume{
						{
							Name: socketMount,
							VolumeSource: core_v1.VolumeSource{
								EmptyDir: nil,
							},
						},
					},
				},
			},
		},
	}
}

// plungin
func createPluginServiceAccountManifest() *core_v1.ServiceAccount {
	return &core_v1.ServiceAccount{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: pluginServiceAccountName,
		},
	}
}

func createPluginClusterRoleManifest() *rbac_v1.ClusterRole {
	return &rbac_v1.ClusterRole{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: pluginClusterRoleName,
		},
		Rules: []rbac_v1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"persistentvolumes"},
				Verbs:     []string{"get", "list", "watch", "update"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"nodes"},
				Verbs:     []string{"get", "list", "update"},
			},
			{
				APIGroups: []string{"storage.k8s.io"},
				Resources: []string{"volumesattachments"},
				Verbs:     []string{"get", "list", "watch", "update"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"namespaces"},
				Verbs:     []string{"get", "list"},
			},
		},
	}
}

func createPluginClusterRoleBindingManifest() *rbac_v1.ClusterRoleBinding {
	return &rbac_v1.ClusterRoleBinding{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: pluginClusterRoleBindingName,
		},
		Subjects: []rbac_v1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      pluginServiceAccountName,
				Namespace: csiNamespace,
			},
		},
		RoleRef: rbac_v1.RoleRef{
			Kind:     "ClusterRole",
			Name:     pluginClusterRoleName,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
}

func createPluginDeamonSetManifest() *app_v1.DaemonSet {
	return &app_v1.DaemonSet{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: pluginDaemonSetName,
		},
		Spec: app_v1.DaemonSetSpec{
			Selector: &meta_v1.LabelSelector{
				MatchLabels: map[string]string{
					"app": pluginDaemonSetName,
				},
			},
			Template: core_v1.PodTemplateSpec{
				ObjectMeta: meta_v1.ObjectMeta{
					Labels: map[string]string{
						"app": pluginDaemonSetName,
					},
				},
				Spec: core_v1.PodSpec{
					ServiceAccountName: pluginServiceAccountName,
					Containers: []core_v1.Container{
						{
							Name:  "driver-registrar",
							Image: "quay.io/k8scsi/driver-registrar:v0.2.0",
							Args:  []string{"--v=5", "--csi-address=$(CSI_ENDPOINT)"},
							Env: []core_v1.EnvVar{
								{
									Name:  "CSI_ENDPOINT",
									Value: "/var/lib/kubelet/plugins/csi-azuredisk/csi.sock",
								},
								{
									Name: "KUBE_NODE_NAME",
									ValueFrom: &core_v1.EnvVarSource{
										FieldRef: &core_v1.ObjectFieldSelector{
											FieldPath: "spec.nodeName",
										},
									},
								},
							},
							VolumeMounts: []core_v1.VolumeMount{
								{
									Name:      socketMount,
									MountPath: socketMountPath,
								},
							},
						},
						{
							Name: "plugin",
							//TODO:
							//should dynamic build the plugin image
							ImagePullPolicy: core_v1.PullAlways,
							Image:           "karataliu/csi-azuredisk:3",
							SecurityContext: &core_v1.SecurityContext{
								Privileged: &privileged,
								Capabilities: &core_v1.Capabilities{
									Add: []core_v1.Capability{
										"SYS_ADMIN",
									},
								},
								AllowPrivilegeEscalation: &allowPrivilegeEscalation,
							},
							Env: []core_v1.EnvVar{
								{
									Name:  "CSI_ENDPOINT",
									Value: "/var/lib/kubelet/plugins/csi-azuredisk/csi.sock",
								},
								{
									Name:  "CSI_SERVICE_DISABLE_CONTROLLER",
									Value: "1",
								},
							},
							VolumeMounts: []core_v1.VolumeMount{
								{
									Name:      socketMount,
									MountPath: socketMountPath,
								},
								{
									Name:             podMount,
									MountPath:        podMountPath,
									MountPropagation: &mountPropagationBidirectional,
								},
								{
									Name:      devMount,
									MountPath: devMountPath,
								},
							},
						},
					},
					Volumes: []core_v1.Volume{
						{
							Name: socketMount,
							VolumeSource: core_v1.VolumeSource{
								HostPath: &core_v1.HostPathVolumeSource{
									Path: socketMountPath,
									//Type: &hostPathDirectoryOrCreate,
								},
							},
						},
						{
							Name: podMount,
							VolumeSource: core_v1.VolumeSource{
								HostPath: &core_v1.HostPathVolumeSource{
									Path: podMountPath,
									//Type: &hostPathDirectory,
								},
							},
						},
						{
							Name: devMount,
							VolumeSource: core_v1.VolumeSource{
								HostPath: &core_v1.HostPathVolumeSource{
									Path: devMountPath,
								},
							},
						},
					},
				},
			},
		},
	}
}

// provisioner
func createProvisionerServiceAccountManifest() *core_v1.ServiceAccount {
	return &core_v1.ServiceAccount{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: provisionerServiceAccountName,
		},
	}
}

func createProvisionerClusterRoleManifest() *rbac_v1.ClusterRole {
	return &rbac_v1.ClusterRole{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: provisionerClusterRoleName,
		},
		Rules: []rbac_v1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"persistentvolumes"},
				Verbs:     []string{"get", "list", "watch", "update", "delete"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"persistentvolumeclaims"},
				Verbs:     []string{"get", "list", "watch", "update"},
			},
			{
				APIGroups: []string{"storage.k8s.io"},
				Resources: []string{"storageclasses"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"events"},
				Verbs:     []string{"list", "watch", "create", "update", "patch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"get"},
			},
		},
	}
}

func createProvisionerClusterRoleBindingManifest() *rbac_v1.ClusterRoleBinding {
	return &rbac_v1.ClusterRoleBinding{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: provisionerClusterRoleBindingName,
		},
		Subjects: []rbac_v1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      provisionerServiceAccountName,
				Namespace: csiNamespace,
			},
		},
		RoleRef: rbac_v1.RoleRef{
			Kind:     "ClusterRole",
			Name:     provisionerClusterRoleName,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
}

func createProvisionerStatefulSetManifest() *app_v1beta1.StatefulSet {
	return &app_v1beta1.StatefulSet{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: provisionerStatefulSetName,
		},
		Spec: app_v1beta1.StatefulSetSpec{
			ServiceName: provisionerServiceAccountName,
			Replicas:    &replicas,
			Template: core_v1.PodTemplateSpec{
				ObjectMeta: meta_v1.ObjectMeta{
					Labels: map[string]string{
						"app": provisionerServiceAccountName,
					},
				},
				Spec: core_v1.PodSpec{
					ServiceAccountName: provisionerServiceAccountName,
					Containers: []core_v1.Container{
						{
							Name:  provisionerServiceAccountName,
							Image: "karataliu/csi-provisioner:bc0b671",
							Args:  []string{"--provisioner=csi-azuredisk", "--csi-address=$(CSI_ENDPOINT)", "--connection-timeout=2m", "--v=5"},
							Env: []core_v1.EnvVar{
								{
									Name:  "CSI_ENDPOINT",
									Value: "/var/lib/kubelet/plugins/csi-azuredisk/csi.sock",
								},
							},
							ImagePullPolicy: core_v1.PullIfNotPresent,
							VolumeMounts: []core_v1.VolumeMount{
								{
									Name:      socketMount,
									MountPath: socketMountPath,
								},
							},
						},
						{
							Name: "plugin",
							//TODO:
							//should dynamic build the plugin image
							Image: "karataliu/csi-azuredisk:3",
							Env: []core_v1.EnvVar{
								{
									Name:  "CSI_ENDPOINT",
									Value: "/var/lib/kubelet/plugins/csi-azuredisk/csi.sock",
								},
								{
									Name:  "CSI_SERVICE_DISABLE_NODE",
									Value: "1",
								},
							},
							ImagePullPolicy: core_v1.PullAlways,
							VolumeMounts: []core_v1.VolumeMount{
								{
									Name:      socketMount,
									MountPath: socketMountPath,
								},
							},
						},
					},
					Volumes: []core_v1.Volume{
						{
							Name: socketMount,
							VolumeSource: core_v1.VolumeSource{
								EmptyDir: nil,
							},
						},
					},
				},
			},
		},
	}
}
