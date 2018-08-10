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

package e2e

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/golang/glog"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	"k8s.io/cloud-provider-azure/tests/e2e/utils"

	_ "k8s.io/cloud-provider-azure/tests/e2e/autoscaling"
	_ "k8s.io/cloud-provider-azure/tests/e2e/network"
	_ "k8s.io/cloud-provider-azure/tests/e2e/storage"
)

func TestAzureTest(t *testing.T) {
	if err := createCsiPlugins(); err != nil {
		glog.Fatal("Failed to build the csi plugins: %v", err)
	}
	/*
		defer func() {
			if err := cleanupCsiPlugins(); err != nil {
				glog.Fatal("Failed to clean up the csi plugins: %v", err)
			}
		}()
	*/

	RegisterFailHandler(Fail)
	reportDir := "_report/"

	var r []Reporter
	if reportDir != "" {
		if err := os.MkdirAll(reportDir, 0755); err != nil {
			glog.Fatal("Failed creating report directory: %v", err)
		} else {
			r = append(r, reporters.NewJUnitReporter(path.Join(reportDir, fmt.Sprintf("junit_%02d.xml", config.GinkgoConfig.ParallelNode))))
		}
	}
	RunSpecsWithDefaultAndCustomReporters(t, "Cloud provider Azure e2e suite", r)
}

func createCsiPlugins() error {
	cs, err := utils.GetClientSet()
	if err != nil {
		return err
	}
	err = utils.DeployCsiPlugin(cs)
	return err
}

func cleanupCsiPlugins() error {
	cs, err := utils.GetClientSet()
	if err != nil {
		return err
	}
	err = utils.CleanupCsiPlugins(cs)
	return err
}
