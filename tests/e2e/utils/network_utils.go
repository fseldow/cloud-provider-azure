package utils

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	aznetwork "github.com/Azure/azure-sdk-for-go/services/network/mgmt/2017-09-01/network"
	"k8s.io/apimachinery/pkg/util/wait"
)

type IPNode struct {
	Occupied bool
	Usable   bool
	Depth    int
	Left     *IPNode // bit 0
	Right    *IPNode // bit 1
}

func NewIPNode(depth int) *IPNode {
	root := &IPNode{
		Occupied: false,
		Usable:   true,
		Depth:    depth,
		Left:     nil,
		Right:    nil,
	}
	return root
}

func InitIPTreeRoot(depth int) *IPNode {
	if depth >= 32 {
		return nil
	}
	root := NewIPNode(depth)
	root.Left = NewIPNode(depth + 1)
	root.Right = NewIPNode(depth + 1)
	return root
}

func setOccupiedByMask(root *IPNode, ip []int, prefixMask int) {
	if root == nil {
		return
	}
	root.Usable = false // Node passed cannot be used as subnet
	if root.Depth == prefixMask {
		root.Occupied = true
		return
	}
	if root.Depth > prefixMask {
		return
	}
	if ip[root.Depth+1] == 1 {
		setOccupiedByMask(root.Right, ip, prefixMask)
	} else {
		setOccupiedByMask(root.Left, ip, prefixMask)
	}
}

func findNodeUsable(root *IPNode, ip []int) ([]int, int) {
	if root == nil {
		return ip, -1
	}
	if root.Depth >= 32 || root.Occupied {
		return ip, -1 // at least remain 1 suffix for subnet
	}
	if root.Usable {
		return ip, root.Depth
	}
	var mask int
	ip[root.Depth] = 0
	ip, mask = findNodeUsable(root.Left, ip)
	if mask > 0 {
		return ip, mask
	}
	ip[root.Depth] = 1
	ip, mask = findNodeUsable(root.Right, ip)
	return ip, mask
}

func SubnetintArray2String(ret []int, prefixMask int) (ip string) {
	ip = ""
	for i := 0; i < 4; i++ {
		temp := 0
		for j := 0; j < 8; j++ {
			temp += ret[i*8+j] << uint(7-j)
		}
		ip += string(temp)
		if i != 3 {
			ip += "."
		}
	}
	ip += "/" + string(prefixMask)
	return
}

func SubnetString2intArray(ip string) (ret []int, prefixMask int, err error) {
	splitPos := strings.Index(ip, "/")
	prefixMask, err = strconv.Atoi(ip[splitPos+1:])
	if err != nil {
		return
	}
	for _, ipSection := range strings.Split(ip[0:splitPos], ".") {
		var section int
		section, err = strconv.Atoi(ipSection)
		if err != nil {
			return
		}
		for i := 7; i >= 0; i-- {
			ret = append(ret, section>>uint(i)&1)
		}
	}
	return
}

func GetAvailableSubnet(vnet aznetwork.VirtualNetwork) (string, error) {
	if len((*vnet.AddressSpace.AddressPrefixes)) == 0 {
		return "", fmt.Errorf("vNet has no prefix")
	}
	vnetPrefix := (*vnet.AddressSpace.AddressPrefixes)[0]
	intIPArray, vNetMask, err := SubnetString2intArray(vnetPrefix)
	if err != nil {
		return "", fmt.Errorf("Unexpected vnet address prefix")
	}
	root := InitIPTreeRoot(vNetMask)
	for _, subnet := range *vnet.Subnets {
		subnetIPArray, subnetMask, err := SubnetString2intArray(*subnet.AddressPrefix)
		if err != nil {
			return "", fmt.Errorf("Unexpected subnet address prefix")
		}
		setOccupiedByMask(root, subnetIPArray, subnetMask)
	}
	retArray, retMask := findNodeUsable(root, intIPArray)
	ret := SubnetintArray2String(retArray, retMask)
	return ret, nil
}

// WaitGetVirtualNetworkList is a wapper around listing VirtualNetwork
func WaitGetVirtualNetworkList() (result aznetwork.VirtualNetworkListResultPage, err error) {
	testClient, err := ObtainTestClient()
	if err != nil {
		return
	}

	err = wait.PollImmediate(poll, singleCallTimeout, func() (bool, error) {
		result, err = testClient.VNetClient.List(context.Background(), GetResourceGroup())
		if err != nil {
			if !isRetryableAPIError(err) {
				return false, err
			}
			return false, nil
		}
		return true, nil
	})
	return
}

// CreateNewSubnet will create a new subnet in certain virtual network
func CreateNewSubnet(tc TestClient, resourceGroupName string, vnet aznetwork.VirtualNetwork, subnetName string) error {
	subnetParameter := (*vnet.Subnets)[0]
	subnetParameter.Name = &subnetName
	address := "10.0.0.0/12"
	subnetParameter.AddressPrefix = &address
	_, err := tc.SubnetsClient.CreateOrUpdate(context.Background(), resourceGroupName, *vnet.Name, subnetName, subnetParameter)
	return err
}
