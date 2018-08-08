package utils

import "testing"

func TestDeployCsi(t *testing.T) {
	cs, _ := GetClientSet()
	DeployCsiPlugin(cs)
}
