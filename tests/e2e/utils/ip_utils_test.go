package utils

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ssh"
)

func TestCIDRString2intArray(t *testing.T) {
	cidr := "10.240.0.0/16"
	intArray, prefix, err := cidrString2intArray(cidr)
	assert.Empty(t, err)
	assert.Equal(t, prefix, 16)
	intArraySuppose := []int{
		0, 0, 0, 0, 1, 0, 1, 0,
		1, 1, 1, 1, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
	}
	for i := range intArray {
		assert.Equal(t, intArray[i], intArraySuppose[i])
	}
}

func TestPrefixIntArray2String(t *testing.T) {
	prefix := 16
	intArray := []int{
		0, 0, 0, 0, 1, 0, 1, 0,
		1, 1, 1, 1, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
	}
	cidrIP := prefixIntArray2String(intArray, prefix)
	cidrSuppose := "10.240.0.0/16"
	assert.Equal(t, cidrIP, cidrSuppose)
}

func TestValidateIPInCIDR(t *testing.T) {
	cidr := "10.24.0.0/16"
	ip1 := "10.24.0.100"
	ip2 := "20.24.0.0"
	flag1, _ := ValidateIPInCIDR(ip1, cidr)
	assert.Equal(t, flag1, true)
	flag2, _ := ValidateIPInCIDR(ip2, cidr)
	assert.Equal(t, flag2, false)
}

func TestGetNextSubnet(t *testing.T) {
	vNetCIDR := "10.24.0.0/16"
	existSubnets := []string{
		"10.24.0.0/24",
		"10.24.1.0/24",
	}
	cidrResult := "10.24.2.0/24"
	cidr, err := getNextSubnet(vNetCIDR, existSubnets)
	assert.Empty(t, err)
	assert.Equal(t, cidrResult, cidr)
}

func TestAA(t *testing.T) {
	rsa := os.Getenv("K8S_AZURE_SSHPUB")
	signer, err := ssh.ParsePrivateKey([]byte(rsa))
	if err != nil {
		log.Fatalf("unable to parse private key: %v", err)
	}
	var hostKey ssh.PublicKey
	// An SSH client is represented with a ClientConn.
	//
	// To authenticate with the remote server you must pass at least one
	// implementation of AuthMethod via the Auth field in ClientConfig,
	// and provide a HostKeyCallback.
	config := &ssh.ClientConfig{
		User: "azureuser",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.FixedHostKey(hostKey),
	}
	client, err := ssh.Dial("tcp", "40.76.8.80:22", config)
	if err != nil {
		log.Fatal("Failed to dial: ", err)
	}

	// Each ClientConn can support multiple interactive sessions,
	// represented by a Session.
	session, err := client.NewSession()
	if err != nil {
		log.Fatal("Failed to create session: ", err)
	}
	defer session.Close()

	// Once a Session is created, you can execute a single command on
	// the remote side using the Run method.
	var b bytes.Buffer
	session.Stdout = &b
	if err := session.Run("/usr/bin/whoami"); err != nil {
		log.Fatal("Failed to run: " + err.Error())
	}
	fmt.Println(b.String())
}
