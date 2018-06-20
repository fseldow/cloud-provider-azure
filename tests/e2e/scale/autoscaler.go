package scale

import (
	"fmt"
	"strconv"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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
			Expect(err).NotTo(HaveOccurred())
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

	It("when adding pod robustly (only upscale)", func() {
		By("creating pods")
		nodeCore := initCoreCount / int64(initNodeCount)
		var podSize int64
		podSize = 200
		podCount := int(nodeCore * 1000 / podSize)

		for i := 0; i < podCount; i++ {
			pod := testutils.DefaultPod(basename + "-pod" + string(uuid.NewUUID()))
			pod.Spec.Containers[0].Resources = v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU: resource.MustParse(
						strconv.FormatInt(podSize, 10) + "m"),
				},
			}
			_, err = cs.CoreV1().Pods(ns.Name).Create(pod)
			Expect(err).NotTo(HaveOccurred())
		}
		defer func() {
			By("delete pods")
			pods, _ := testutils.WaitListPods(cs, ns.Name)
			for _, p := range pods.Items {
				cs.CoreV1().Pods(ns.Name).Delete(p.Name, nil)
			}
		}()

		time.Sleep(10 * time.Second)
		pods, _ := testutils.WaitListPods(cs, ns.Name)
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

	It("when deployment (upscale + downscale)", func() {
		fmt.Printf("Start up scaling")
		By("Up scale")

		By("creating pods")
		nodeCore := initCoreCount / int64(initNodeCount)
		var podSize int64
		podSize = 200
		podCount := int(nodeCore * 1000 / podSize)

		pod := testutils.DefaultPod(basename + "-pod")
		pod.Spec.Containers[0].Resources = v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU: resource.MustParse(
					strconv.FormatInt(podSize, 10) + "m"),
			},
		}

		deployment := testutils.DefaultDeployment(basename+"-deployment", map[string]string{basename: "true"})
		replicas := int32(podCount)
		deployment.Spec.Replicas = &replicas
		deployment.Spec.Template.Spec = pod.Spec
		By("Create deployment ")
		_, err = cs.Extensions().Deployments(ns.Name).Create(deployment)
		Expect(err).NotTo(HaveOccurred())

		time.Sleep(10 * time.Second)
		pods, _ := testutils.WaitListPods(cs, ns.Name)
		for _, p := range pods.Items {
			fmt.Printf(string(p.Status.Phase) + "\n")
		}

		targetNodeCount := initNodeCount + 1
		By(fmt.Sprintf("scaling up the node... Target node count: %v", targetNodeCount))

		resultNodeCount, err := testutils.WaitAutoScaleNodes(cs, targetNodeCount)
		Expect(err).NotTo(HaveOccurred())
		By(fmt.Sprintf("Complete scaling up... Result node count: %v", resultNodeCount))
		Expect(resultNodeCount).To(Equal(targetNodeCount))

		fmt.Printf("start down scaling")

		By("Down scaling")
		By("Delete Pods by replic=0")
		replicas = 0
		deployment.Spec.Replicas = &replicas
		_, err = cs.Extensions().Deployments(ns.Name).Update(deployment)
		targetNodeCount = initNodeCount
		resultNodeCount, err = testutils.WaitAutoScaleNodes(cs, targetNodeCount)
		Expect(err).NotTo(HaveOccurred())
		By(fmt.Sprintf("Complete scaling up... Result node count: %v", resultNodeCount))
		Expect(resultNodeCount).To(Equal(targetNodeCount))
	})

})
