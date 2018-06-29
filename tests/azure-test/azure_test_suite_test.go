package azuretest

import (
	"fmt"
	"os"
	"path"
	"testing"

	testutils "k8s.io/cloud-provider-azure/tests/azure-test/utils"

	"github.com/golang/glog"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"

	//_ "k8s.io/cloud-provider-azure/tests/azure-test/basic"
	//_ "k8s.io/cloud-provider-azure/tests/azure-test/network"
	_ "k8s.io/cloud-provider-azure/tests/azure-test/scale"
)

func TestAzureTest(t *testing.T) {
	RegisterFailHandler(Fail)
	DNSPrefix := testutils.ExtractDNSPrefix()
	reportDir := DNSPrefix + "/report/"

	var r []Reporter
	if reportDir != "" {
		// TODO: we should probably only be trying to create this directory once
		// rather than once-per-Ginkgo-node.
		if err := os.MkdirAll(reportDir, 0755); err != nil {
			glog.Errorf("Failed creating report directory: %v", err)
		} else {
			r = append(r, reporters.NewJUnitReporter(path.Join(reportDir, fmt.Sprintf("junit_%v%02d.xml", "", config.GinkgoConfig.ParallelNode))))
		}
	}
	RunSpecsWithDefaultAndCustomReporters(t, "Azure Specific e2e Suite", r)
}
