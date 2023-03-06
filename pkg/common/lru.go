package common

import (
	lru "github.com/hashicorp/golang-lru"
	"log"
)

// 镜像缓存
var ImageCache *lru.Cache

func InitImageCache(size int) {
	c, err := lru.New(size)
	if err != nil {
		log.Fatalln(err)
	}
	ImageCache = c
}
