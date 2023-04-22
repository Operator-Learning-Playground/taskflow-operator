package main

import (
	"github.com/myoperator/cicdoperator/pkg/common"
	"github.com/myoperator/cicdoperator/pkg/k8sconfig"
	_ "k8s.io/code-generator"
)

func main() {
	common.InitImageCache(100)
	k8sconfig.InitManager()
}
