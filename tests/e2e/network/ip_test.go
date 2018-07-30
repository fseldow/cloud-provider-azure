package network

import (
	"fmt"
	"strconv"
	"testing"

	"testing"

	"github.com/Azure/go-autorest/autorest/to"
	"github.com/stretchr/testify/assert"
	testutils "k8s.io/cloud-provider-azure/tests/e2e/utils"

	"github.com/stretchr/testify/assert"
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
	para := defaultPublicIPAddress("PIP-test2")
	tc, _ := testutils.ObtainAzureTestClient()
	pip, err := testutils.WaitCreatePIP(tc, "PIP-test2", para)
	a := to.String(pip.IPAddress)
	fmt.Print(a)
	fmt.Print(err)
}

func TestPrivateIPSelection(t *testing.T) {
	tc, _ := testutils.ObtainAzureTestClient()
	IP, err := selectAvailablePrivateIP(tc)
	fmt.Print(IP)
	assert.NoError(t, err, "find no private ip")
}

func TestMine(t *testing.T) {
	a := strconv.Itoa(test())
	fmt.Print(a)
}
func test() (i int) {
	i = 0
	defer func() {
		fmt.Print(i)
		i++
	}()
	i = 10
	return
}
