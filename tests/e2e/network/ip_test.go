package network

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	testutils "k8s.io/cloud-provider-azure/tests/e2e/utils"
)

func TestValidation(t *testing.T) {
	ip1 := "10.0.10.11"
	ip2 := "11.0.0.0"
	prefix := "10.0.0.0/8"
	err := validateIPinPrefix(ip1, prefix)
	assert.NoError(t, err, ip1)
	err = validateIPinPrefix(ip2, prefix)
}

func TestUsable(t *testing.T) {
	tc, _ := testutils.ObtainAzureTestClient()
	vlist, _ := testutils.WaitGetVirtualNetworkList(tc)
	vNet := vlist.Values()[0]
	getAvailableSubnet(vNet)
}

func TestPIPCreation(t *testing.T) {
	para := defaultPublicIPAddress()
	tc, _ := testutils.ObtainAzureTestClient()
	err := testutils.WaitCreatePIP(tc, "PIP-test", para)
	fmt.Print(err)
}
