package connect

import (
	"fmt"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	testutils "k8s.io/cloud-provider-azure/tests/azure-test/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pborman/uuid"
)

func GetClientSet(filename string) (clientset.Interface, error) {
	c := clientcmd.GetConfigFromFileOrDie(filename)
	restConfig, err := clientcmd.NewDefaultClientConfig(*c, &clientcmd.ConfigOverrides{ClusterInfo: clientcmdapi.Cluster{Server: ""}}).ClientConfig()
	if err != nil {
		return nil, err
	}
	clientSet, err := clientset.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	return clientSet, nil
}

var _ = Describe("test ", func() {
	var cs clientset.Interface
	basename := "test"
	var ns *v1.Namespace
	configPath := "C:\\Users\\t-xinhli\\.kube\\config"

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
				Labels: map[string]string{
					"name": name,
				},
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:  "nginx",
						Image: imageutils.GetE2EImage(imageutils.NginxSlim),
						Resources: v1.ResourceRequirements{
							Limits: v1.ResourceList{
								v1.ResourceCPU:    resource.MustParse("100m"),
								v1.ResourceMemory: resource.MustParse("100Mi"),
							},
							Requests: v1.ResourceList{
								v1.ResourceCPU:    resource.MustParse("100m"),
								v1.ResourceMemory: resource.MustParse("100Mi"),
							},
						},
					},
				},
			},
		}
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

func main() {
	_, e := getServerVersion("C:\\Users\\t-xinhli\\.kube\\config")
	if e != nil {
		fmt.Sprint(e)
	}
}
