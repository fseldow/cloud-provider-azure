package utils

import (
	"io/ioutil"
	"os"
	"testing"

	"k8s.io/kubernetes/cmd/kubeadm/app/util"
)

func TestReadYaml(t *testing.T) {
	file := "attacher.yaml"
	yamlFile, err := os.Open(file)
	if err != nil {
		return err
	}
	defer yamlFile.Close()
	byteValue, _ := ioutil.ReadAll(yamlFile)
	map := util.SplitYAMLDocuments(byteValue)
	fmt.Print("a")
	//util.UnmarshalFromYaml()
}
