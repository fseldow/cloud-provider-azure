package utils

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	leastPrefixMask = 12
)

//ValidateIPFitPrefix validates whether certain ip fits prefix
func ValidateIPFitPrefix(ip, prefix string) error {
	ipInt, _, err := prefixString2intArray(ip)
	if err != nil {
		return err
	}
	prefixInt, mask, err := prefixString2intArray(prefix)
	if err != nil {
		return err
	}
	for i := 0; i < mask; i++ {
		if prefixInt[i] != ipInt[i] {
			return fmt.Errorf("%s not in prefix %s", ip, prefix)
		}
	}
	return nil
}

type ipNode struct {
	Occupied bool
	Usable   bool
	Depth    int
	Left     *ipNode // bit 0
	Right    *ipNode // bit 1
}

func newIPNode(depth int) *ipNode {
	root := &ipNode{
		Occupied: false,
		Usable:   true,
		Depth:    depth,
		Left:     nil,
		Right:    nil,
	}
	if depth < leastPrefixMask {
		root.Usable = false
	}
	return root
}

func initIPTreeRoot(depth int) *ipNode {
	if depth >= 32 {
		return nil
	}
	root := newIPNode(depth)
	root.Left = initIPTreeRoot(depth + 1)
	root.Right = initIPTreeRoot(depth + 1)
	return root
}

func setOccupiedByMask(root *ipNode, ip []int, prefixMask int) {
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
	if ip[root.Depth] == 1 {
		setOccupiedByMask(root.Right, ip, prefixMask)
	} else {
		setOccupiedByMask(root.Left, ip, prefixMask)
	}
}

func findNodeUsable(root *ipNode, ip []int) ([]int, int) {
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

func prefixIntArray2String(ret []int, prefixMask int) (ip string) {
	ip = ""
	if prefixMask < 0 {
		return ip
	}
	for i := 0; i < 4; i++ {
		temp := 0
		for j := 0; j < 8; j++ {
			temp += ret[i*8+j] << uint(7-j)
		}
		ip += strconv.Itoa(temp)
		if i != 3 {
			ip += "."
		}
	}
	ip += "/" + strconv.Itoa(prefixMask)
	return
}

func prefixString2intArray(ip string) (ret []int, prefixMask int, err error) {
	splitPos := strings.Index(ip, "/")
	if splitPos == -1 {
		prefixMask = -1
		splitPos = len(ip)
	} else {
		prefixMask, err = strconv.Atoi(ip[splitPos+1:])
	}
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
