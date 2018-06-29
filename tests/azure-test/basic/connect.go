package main

import (
	"fmt"
	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	clientset "k8s.io/client-go/kubernetes"
	testutils "k8s.io/cloud-provider-azure/tests/azure-test/utils"
	//imageutils "k8s.io/cloud-provider-azure/tests/azure-test/utils/image"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("test ", func() {
	var cs clientset.Interface
	basename := "test"
	var ns *v1.Namespace
	//configPath := "C:\\Users\\t-xinhli\\.kube\\config"
	var err error

	BeforeEach(func() {
		By("Creating a kubernetes client")
		cs, _ = testutils.GetClientSet()

		By("Creating namespace")
		ns, _ = testutils.CreateTestingNS(basename, cs, nil)

	})

	It("judge server version ", func() {
		By("Obtain server version")
		_, err := cs.Discovery().ServerVersion()
		Expect(err).NotTo(Equal(HaveOccurred()))
	})

	It("add a single pod", func() {
		//fmt.Printf(imageutils.GetPauseImageName())
		By("creating the pod")
		name := "pod-qos-class-" + string(uuid.NewUUID())
		pod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:  "nginx",
						Image: testutils.PodImage,
					},
				},
			},
		}
		pod, err = cs.CoreV1().Pods(ns.Name).Create(pod)
		Expect(err).NotTo(HaveOccurred())

		time.Sleep(10 * time.Second)
		pods, _ := testutils.WaitListPods(cs, ns.Name)
		fmt.Print(len(pods.Items))
		for _, p := range pods.Items {
			fmt.Printf(string(p.Status.Phase) + "\n")
		}
		//label := labels.SelectorFromSet(labels.Set(map[string]string{"app": "cassandra"}))
		//p1 := cs.CoreV1().Pods(ns.Name).Get(name, metav1.GetOptions{})
	})

	AfterEach(func() {
		testutils.DeleteNS(cs, ns.Name)
	})

})

func main() {
	var a string
	a = "aa"
	a = testutils.ExtractRegion()
	fmt.Printf(a)
}
