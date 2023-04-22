package main

import (
	"fmt"
	"github.com/myoperator/cicdoperator/pkg/builder"
	"github.com/myoperator/cicdoperator/pkg/common"
)

func main() {
	// 初始化lru缓存
	common.InitImageCache(2)
	// image
	img1, _ := builder.ParseImage("alpine:3.12")
	img2, _ := builder.ParseImage("nginx:1.18-alpine")
	img3, _ := builder.ParseImage("mysql")

	// 加入缓存
	common.ImageCache.Add(img1.Ref, img1)
	common.ImageCache.Add(img2.Ref, img2)
	common.ImageCache.Add(img3.Ref, img3)

	for _, key := range common.ImageCache.Keys() {
		v, _ := common.ImageCache.Get(key)
		fmt.Println(v.(*builder.Image).Name)
	}

}
