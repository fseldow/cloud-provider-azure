/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package autoscaling

import (
	"fmt"
	"strconv"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	clientset "k8s.io/client-go/kubernetes"
	testutils "k8s.io/cloud-provider-azure/tests/e2e/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("[Serial][Feature:Autoscaler] Cluster node autoscaling [Slow]", func() {
	basename := "autoscaler"

	var cs clientset.Interface
	var ns *v1.Namespace

	var initNodeCount int
	var initCoreCount int64

	var nodeCore int64      //Core capacity for each node, assume uniform
	var podSize int64 = 200 //podSize to create, 200m, 0.2 core
	var podCount int        // To make sure enough pods to exceed the temporary node volume

	var err error

	var namespacesToDelete []*v1.Namespace

	var initNodesNames []string

	BeforeEach(func() {
		By("Create test context")
		cs, err = testutils.GetClientSet()
		Expect(err).NotTo(HaveOccurred())

		ns, err = testutils.CreateTestingNS(basename, cs)
		Expect(err).NotTo(HaveOccurred())
		namespacesToDelete = append(namespacesToDelete, ns)

		nodes, err := testutils.WaitListSchedulableNodes(cs)
		Expect(err).NotTo(HaveOccurred())

		initNodeCount = len(nodes.Items)
		testutils.Logf("Initial number of schedulable nodes: %v", initNodeCount)
		if initNodeCount <= 1 {
			testutils.Skipf("there should be at least 1 agend node and 1 master node")
		}

		initCoreCount = 0
		for _, node := range nodes.Items {
			initNodesNames = append(initNodesNames, node.Name)
			quantity := node.Status.Capacity[v1.ResourceCPU]
			initCoreCount += quantity.Value()
		}
		testutils.Logf("Initial number of cores: %v", initCoreCount)

		nodeCore = initCoreCount / int64(initNodeCount)
		podCount = int(nodeCore * 1000 / podSize)
		testutils.Logf("will create %v pod, each %vm size", podCount, podSize)
	})

	AfterEach(func() {
		for _, nsToDel := range namespacesToDelete {
			testutils.DeleteNS(cs, nsToDel.Name)
			Expect(err).NotTo(HaveOccurred())
		}
		//delete extra nodes
		nodes, err := testutils.WaitListSchedulableNodes(cs)
		Expect(err).NotTo(HaveOccurred())
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
		testutils.Logf("creating pods")
		for i := 0; i < podCount; i++ {
			pod := testutils.DefaultPod(fmt.Sprintf("%s-pod-%v", basename, i))
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
			testutils.Logf("deleting pods")
			pods, _ := testutils.WaitListPods(cs, ns.Name)
			for _, p := range pods.Items {
				cs.CoreV1().Pods(ns.Name).Delete(p.Name, nil)
			}
		}()
		By("scale up")
		targetNodeCount := initNodeCount + 1
		err := testutils.WaitAutoScaleNodes(cs, targetNodeCount)
		Expect(err).NotTo(HaveOccurred())
	})

	It("when deployment (upscale + downscale)", func() {
		pod := testutils.DefaultPod(basename + "-pod")
		pod.Spec.Containers[0].Resources = v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU: resource.MustParse(
					strconv.FormatInt(podSize, 10) + "m"),
			},
		}

		testutils.Logf("Create deployment")
		deployment := testutils.DefaultDeployment(basename+"-deployment", map[string]string{basename: "true"})
		replicas := int32(podCount)
		deployment.Spec.Replicas = &replicas
		deployment.Spec.Template.Spec = pod.Spec
		_, err = cs.Extensions().Deployments(ns.Name).Create(deployment)
		Expect(err).NotTo(HaveOccurred())

		By("Scale up")
		targetNodeCount := initNodeCount + 1
		err := testutils.WaitAutoScaleNodes(cs, targetNodeCount)
		Expect(err).NotTo(HaveOccurred())

		By("Scale down")
		testutils.Logf("Delete Pods by replic=0")
		replicas = 0
		deployment.Spec.Replicas = &replicas
		_, err = cs.Extensions().Deployments(ns.Name).Update(deployment)
		Expect(err).NotTo(HaveOccurred())
		targetNodeCount = initNodeCount
		err = testutils.WaitAutoScaleNodes(cs, targetNodeCount)
		Expect(err).NotTo(HaveOccurred())
	})

})
