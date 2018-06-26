package scale

import (
	"fmt"
	"strconv"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	clientset "k8s.io/client-go/kubernetes"
	testutils "k8s.io/cloud-provider-azure/tests/azure-test/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("[Serial][Feature:Autoscaler] Cluster node autoscaling [Slow]", func() {
	var cs clientset.Interface
	basename := "autoscaler"
	var ns *v1.Namespace
	//configPath := "C:\\Users\\t-xinhli\\.kube\\kubeconfig.eastus.json"

	var initNodeCount int
	var initCoreCount int64

	var err error

	var namespacesToDelete []*v1.Namespace

	var initNodesNames []string

	BeforeEach(func() {
		By("Creating a kubernetes client")
		cs, err = testutils.GetClientSet()
		Expect(err).To(BeNil())

		By("Creating namespace")
		ns, err = testutils.CreateTestingNS(basename, cs, nil)
		Expect(err).To(BeNil())
		namespacesToDelete = append(namespacesToDelete, ns)

		By("Obtaining node info")
		nodes, err := testutils.WaitListSchedulableNodes(cs)
		Expect(err).To(BeNil())

		initNodeCount = len(nodes.Items)
		By(fmt.Sprintf("Initial number of schedulable nodes: %v", initNodeCount))
		Expect(initNodeCount).NotTo(BeZero())

		initCoreCount = 0
		for _, node := range nodes.Items {
			initNodesNames = append(initNodesNames, node.Name)
			quentity := node.Status.Capacity[v1.ResourceCPU]
			initCoreCount += quentity.Value()
		}
		By(fmt.Sprintf("Initial number of cores: %v", initCoreCount))

	})

	AfterEach(func() {
		for _, nsToDel := range namespacesToDelete {
			testutils.DeleteNS(cs, nsToDel.Name)
		}
		//delete extra nodes
		nodes, err := testutils.WaitListSchedulableNodes(cs)
		Expect(err).To(BeNil())
		var nodesToDelete []string
		for _, n := range nodes.Items {
			if !testutils.StringInSlice(n.Name, initNodesNames) {
				nodesToDelete = append(nodesToDelete, n.Name)
			}
		}

		err = testutils.WaitNodeDeletion(cs, nodesToDelete)
		Expect(err).NotTo(HaveOccurred())

	})

	It("should add nodes if pending pod exceeds volumn. +1 node in test", func() {
		By("creating pods")
		nodeCore := initCoreCount / int64(initNodeCount)
		var podSize int64
		podSize = 200
		podCount := int(nodeCore * 1000 / podSize)

		for i := 0; i < podCount; i++ {
			name := "pod-new-" + string(uuid.NewUUID())
			pod := &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "container" + string(uuid.NewUUID()),
							Image: testutils.PodImage,
							Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceCPU: resource.MustParse(
										strconv.FormatInt(podSize, 10) + "m"),
								},
							},
						},
					},
				},
			}

			_, err = cs.CoreV1().Pods(ns.Name).Create(pod)
			Expect(err).NotTo(HaveOccurred())

		}
		time.Sleep(10 * time.Second)
		pods, _ := testutils.WaitListPods(cs, ns.Name)
		fmt.Print(len(pods.Items))
		for _, p := range pods.Items {
			fmt.Printf(string(p.Status.Phase) + "\n")
		}

		targetNodeCount := initNodeCount + 1
		By(fmt.Sprintf("scaling up the node... Target node count: %v", targetNodeCount))

		resultNodeCount, err := testutils.WaitAutoScaleNodes(cs, targetNodeCount)
		Expect(err).NotTo(HaveOccurred())
		By(fmt.Sprintf("Complete scaling up... Result node count: %v", resultNodeCount))
		Expect(resultNodeCount).To(Equal(targetNodeCount))
	})

	It("create two new empty nodes, should delete the nodes automaticaly", func() {
		By("creating two new nodes")

		for i := 0; i < 2; i++ {
			name := "k8s-agentpool-vmss-test-" + string(uuid.NewUUID())
			node := &v1.Node{
				ObjectMeta: metav1.ObjectMeta{Name: name},
				TypeMeta: metav1.TypeMeta{
					Kind:       "Node",
					APIVersion: "v1",
				},
			}
			_, err := cs.CoreV1().Nodes().Create(node)
			Expect(err).NotTo(HaveOccurred())
		}

		By("waiting for 2 nodes automatically deletion")
		resultNodeCount, err := testutils.WaitAutoScaleNodes(cs, initNodeCount)
		Expect(err).NotTo(HaveOccurred())
		By(fmt.Sprintf("Complete scaling up... Result node count: %v", resultNodeCount))
		Expect(resultNodeCount).To(Equal(initNodeCount))
	})

})
