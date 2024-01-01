package main

import (
	"fmt"

	"github.com/google/go-containerregistry/pkg/name"

	"github.com/google/go-containerregistry/pkg/v1/remote"
	"log"
)

func main() {
	//alpine:3.12
	// 这里输入镜像
	ref, err := name.ParseReference("alpine:3.12")
	if err != nil {
		log.Fatalln(err)
	}
	// 获取镜像描述信息
	des, err := remote.Get(ref)
	if err != nil {
		log.Fatalln(err)
	}

	idx, err := des.ImageIndex()
	if err != nil {
		log.Fatalln(err)
	}
	mf, err := idx.IndexManifest()

	if err != nil {
		log.Fatalln(err)
	}
	for _, d := range mf.Manifests {
		// 使用 digest 可以得到 Image 对象，digest 本质就是 hash
		img, err := idx.Image(d.Digest)
		if err != nil {
			log.Fatalln(err)
		}

		cf, err := img.ConfigFile()
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Println(cf.OS, "/", cf.Architecture, ":", cf.Config.Entrypoint, cf.Config.Cmd)
	}

}
