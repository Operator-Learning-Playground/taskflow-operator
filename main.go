package main

import (
	"github.com/myoperator/cicdoperator/pkg/k8sconfig"
	_ "k8s.io/code-generator"
)

func main() {
	k8sconfig.InitManager()
}
