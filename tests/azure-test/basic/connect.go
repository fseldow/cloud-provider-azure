package main

import (
	"fmt"
	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
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

func getServerVersion(filename string) (bool, error) {
	c := clientcmd.GetConfigFromFileOrDie(filename)
	restConfig, _ := clientcmd.NewDefaultClientConfig(*c, &clientcmd.ConfigOverrides{ClusterInfo: clientcmdapi.Cluster{Server: ""}}).ClientConfig()
	clientSet, _ := clientset.NewForConfig(restConfig)
	serverVersion, _ := clientSet.Discovery().ServerVersion()
	return serverVersion == nil, nil
}

func findExistingKubeConfig(file string) string {
	// The user did provide a --kubeconfig flag. Respect that and threat it as an
	// explicit path without building a DefaultClientConfigLoadingRules object.
	defaultKubeConfig := "aa"
	if file != defaultKubeConfig {
		return file
	}
	// The user did not provide a --kubeconfig flag. Find a config in the standard
	// locations using DefaultClientConfigLoadingRules, but also consider `defaultKubeConfig`.
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	rules.Precedence = append(rules.Precedence, defaultKubeConfig)
	return rules.GetDefaultFilename()
}

func main() {
	var a string
	a = "aa"
	a = findExistingKubeConfig(a)

	_, e := getServerVersion("C:\\Users\\t-xinhli\\.kube\\config")
	if e != nil {
		fmt.Sprint(e)
	}
}
