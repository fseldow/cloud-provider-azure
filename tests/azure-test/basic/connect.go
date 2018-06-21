package connect

import (
	"fmt"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	testutils "k8s.io/cloud-provider-azure/tests/azure-test/utils"
	imageutils "k8s.io/cloud-provider-azure/tests/azure-test/utils/image"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("test ", func() {
	var cs clientset.Interface
	basename := "test"
	var ns *v1.Namespace
	configPath := "C:\\Users\\t-xinhli\\.kube\\config"
	var err error

	BeforeEach(func() {
		By("Creating a kubernetes client")
		cs, _ = testutils.GetClientSet(configPath)

		By("Creating namespace")
		ns, _ = testutils.CreateTestingNS(basename, cs, nil)

	})

	It("judge server version ", func() {
		By("Obtain server version")
		_, err := cs.Discovery().ServerVersion()
		Expect(err).NotTo(Equal(HaveOccurred()))
	})

	It("add a single pod", func() {
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
						Image: imageutils.GetPauseImageName(),
					},
				},
			},
		}
		pod, err = cs.CoreV1().Pods(ns.Name).Create(pod)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		//testutils.DeleteNS(cs, ns.Name)
	})

})

func getServerVersion(filename string) (bool, error) {
	c := clientcmd.GetConfigFromFileOrDie(filename)
	restConfig, _ := clientcmd.NewDefaultClientConfig(*c, &clientcmd.ConfigOverrides{ClusterInfo: clientcmdapi.Cluster{Server: ""}}).ClientConfig()
	clientSet, _ := clientset.NewForConfig(restConfig)
	serverVersion, _ := clientSet.Discovery().ServerVersion()
	return serverVersion == nil, nil
}

func main() {
	_, e := getServerVersion("C:\\Users\\t-xinhli\\.kube\\config")
	if e != nil {
		fmt.Sprint(e)
	}
}
