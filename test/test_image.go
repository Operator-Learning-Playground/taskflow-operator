package main

import (
	"fmt"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/myoperator/cicdoperator/pkg/image"
	"k8s.io/klog/v2"
)

func parseImage(img string) (*builder.Image, error) {
	ref, err := name.ParseReference(img)
	if err != nil {
		return nil, err
	}

	des, err := remote.Get(ref) // 获取镜像描述信息
	if err != nil {
		return nil, err
	}

	// 初始化我们的业务 Image 对象
	imgBuilder := builder.NewImage(img, des.Digest, ref)

	// image 类型
	if des.MediaType.IsImage() {
		img, err := des.Image()
		if err != nil {
			return nil, err
		}
		conf, err := img.ConfigFile()
		if err != nil {
			return nil, err
		}
		imgBuilder.AddCommand(conf.OS, conf.Architecture, conf.Config.Entrypoint, conf.Config.Cmd)
		return imgBuilder, nil
	}

	// index 模式
	idx, err := des.ImageIndex()
	if err != nil {
		return nil, err
	}
	mf, err := idx.IndexManifest()

	if err != nil {
		return nil, err
	}
	for _, d := range mf.Manifests {
		img, err := idx.Image(d.Digest)
		if err != nil {
			return nil, err
		}
		conf, err := img.ConfigFile()
		if err != nil {
			return nil, err
		}
		imgBuilder.AddCommand(conf.OS, conf.Architecture, conf.Config.Entrypoint, conf.Config.Cmd)
		//fmt.Println(conf.OS, "/", conf.Architecture, ":", conf.Config.Entrypoint, conf.Config.Cmd)
	}
	return imgBuilder, nil
}

func main() {

	img, err := parseImage("try:v1")
	if err != nil {
		klog.Fatal(err)
	}
	fmt.Println(img)

}
