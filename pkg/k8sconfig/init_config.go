package k8sconfig

import (
	"github.com/myoperator/cicdoperator/pkg/common"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"os"
)

func K8sRestConfigInPod() *rest.Config {
	config, err := rest.InClusterConfig()
	if err != nil {
		klog.Fatal("parse config in cluster err: ", err)
	}
	return config

}

func K8sRestConfig() *rest.Config {
	if os.Getenv("release") == "1" {
		klog.Info("run in cluster!")
		return K8sRestConfigInPod()
	}

	klog.Info("run outside cluster!")

	path := common.GetWd()
	config, err := clientcmd.BuildConfigFromFlags("", path + "/resources/config")
	config.Insecure = true
	if err != nil {
		klog.Fatal("parse config outside cluster err: ", err)

	}

	return config
}