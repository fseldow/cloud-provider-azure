package storage

import (
	"fmt"
	"strings"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/cloud-provider-azure/tests/e2e/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	nginxPort       = 80
	nginxStatusCode = 200
)

var _ = Describe("CSI plugin Storage", func() {
	basename := "azuredisk"

	var cs clientset.Interface
	var ns *v1.Namespace

	BeforeEach(func() {
		var err error
		cs, err = utils.GetClientSet()
		Expect(err).NotTo(HaveOccurred())

		ns, err = utils.CreateTestingNameSpace(basename, cs)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := utils.DeleteNameSpace(cs, ns.Name)
		Expect(err).NotTo(HaveOccurred())

		cs = nil
		ns = nil
	})

	It("should be able to bound a disk to pod", func() {
		// Assume plugin built
		//Create PVC
		pvcName := "csi-disk-pvc"
		podName := "claim"
		storageClassName := "azuredisk-csi"
		pvc := createPersistentVolumeClaimManifest(pvcName, storageClassName)
		_, err := cs.CoreV1().PersistentVolumeClaims(ns.Name).Create(pvc)
		Expect(err).NotTo(HaveOccurred())

		pod := createClaimPodManifest(podName, pvcName)
		_, err = cs.CoreV1().Pods(ns.Name).Create(pod)
		Expect(err).NotTo(HaveOccurred())

		err = validatePersistentVolumeAttachment(cs, ns.Name, pod.Name)
		Expect(err).NotTo(HaveOccurred())

	})
})

func createClaimPodManifest(podName string, pvcName string) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: podName,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Image: "nginx",
					Name:  "test-container",
					VolumeMounts: []v1.VolumeMount{
						{
							MountPath: "/af",
							Name:      "af-pvc",
						},
					},
					Command: []string{
						"/bin/sh",
						"-c",
						"df",
					},
				},
			},
			Volumes: []v1.Volume{
				{
					Name: "af-pvc",
					VolumeSource: v1.VolumeSource{
						PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
							ClaimName: pvcName,
						},
					},
				},
			},
		},
	}
}

func createPersistentVolumeClaimManifest(pvcName string, storageClassName string) *v1.PersistentVolumeClaim {
	return &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: pvcName,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteOnce,
			},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceName(v1.ResourceStorage): resource.MustParse("2Gi"),
				},
			},
			StorageClassName: &storageClassName,
		},
	}
}

func validatePersistentVolumeAttachment(cs clientset.Interface, ns string, podName string) error {
	pullInterval := 10 * time.Second
	pullTimeout := 10 * time.Minute
	err := wait.PollImmediate(pullInterval, pullTimeout, func() (bool, error) {
		utils.Logf("Still testing internal access")
		log, err := cs.CoreV1().Pods(ns).GetLogs(podName, &v1.PodLogOptions{}).Do().Raw()
		if err != nil {
			return false, nil
		}
		return strings.Contains(fmt.Sprintf("%s", log), "af"), nil
	})
	utils.Logf("validation finished")
	return err
}
