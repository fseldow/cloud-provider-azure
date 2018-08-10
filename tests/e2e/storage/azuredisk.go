package storage

import (
	"fmt"
	"strings"
	"time"

	"k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
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
		pvcName := "csi-disk-pvc"
		podName := "claim"
		storageClassName := "azuredisk-csi"
		originalPersistentVolumesList, err := utils.GetPersistentVolumeList(cs) //to compare to obtain new pv name
		Expect(err).NotTo(HaveOccurred())

		By("Building environment")
		pvc := createPersistentVolumeClaimManifest(pvcName, storageClassName)
		utils.Logf("Creating persistent Volume Claim: %s", pvc.Name)
		_, err = cs.CoreV1().PersistentVolumeClaims(ns.Name).Create(pvc)
		Expect(err).NotTo(HaveOccurred())

		pod := createClaimPodManifest(podName, pvcName)
		utils.Logf("Creating pod: %s", pod.Name)
		_, err = cs.CoreV1().Pods(ns.Name).Create(pod)
		Expect(err).NotTo(HaveOccurred())

		By("Waiting persistent volume attach to the pod")
		err = waitPersistentVolumeAttachment(cs, ns.Name, pod.Name)
		Expect(err).NotTo(HaveOccurred())

		tempPersistentVolumesList, err := utils.GetPersistentVolumeList(cs)
		Expect(err).NotTo(HaveOccurred())
		// Obtain new persistent volume name
		Expect(len(tempPersistentVolumesList.Items)).To(Equal(len(originalPersistentVolumesList.Items)+1), "There should only add one single persistent volume")
		var freshPvName string
		for _, pv := range tempPersistentVolumesList.Items {
			isInOriginal := false
			for _, opv := range originalPersistentVolumesList.Items {
				if pv.Name == opv.Name {
					isInOriginal = true
					break
				}
			}
			if !isInOriginal {
				freshPvName = pv.Name
				break
			}
		}
		utils.Logf("Get new persistent volume name : %s", freshPvName)

		By("Cleaning up")
		Expect(utils.DeletePod(cs, ns.Name, pod.Name)).NotTo(HaveOccurred())
		Expect(utils.DeletePersistentVolumeClaim(cs, ns.Name, pvc.Name)).NotTo(HaveOccurred())

		By("Waiting for persistent volume releasing")
		Expect(waitPersistentVolumeRelease(cs, freshPvName)).NotTo(HaveOccurred())
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

func waitPersistentVolumeAttachment(cs clientset.Interface, ns string, podName string) error {
	pullInterval := 10 * time.Second
	pullTimeout := 10 * time.Minute
	err := wait.PollImmediate(pullInterval, pullTimeout, func() (bool, error) {
		utils.Logf("Waiting for persistent volume attachment")
		log, err := cs.CoreV1().Pods(ns).GetLogs(podName, &v1.PodLogOptions{}).Do().Raw()
		if err != nil {
			return false, nil
		}
		return strings.Contains(fmt.Sprintf("%s", log), "af"), nil
	})
	utils.Logf("validation finished")
	return err
}

func waitPersistentVolumeRelease(cs clientset.Interface, pvName string) error {
	pullInterval := 10 * time.Second
	pullTimeout := 20 * time.Minute
	err := wait.PollImmediate(pullInterval, pullTimeout, func() (bool, error) {
		utils.Logf("Waiting for persistent volume automatical relasement")
		if _, err := cs.CoreV1().PersistentVolumes().Get(pvName, metav1.GetOptions{}); err != nil {
			return apierrs.IsNotFound(err), nil
		}
		return false, nil
	})
	utils.Logf("finished")
	return err
}
