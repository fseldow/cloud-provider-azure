package azuretest

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	_ "k8s.io/cloud-provider-azure/tests/azure-test/basic"
)

func TestAzureTest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AzureTest Suite")
}
