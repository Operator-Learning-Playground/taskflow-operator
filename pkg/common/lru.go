package common

import (
	lru "github.com/hashicorp/golang-lru"
	"log"
)

// ImageCache 镜像缓存
var ImageCache *lru.Cache

// InitImageCache 初始化LRU缓存
func InitImageCache(size int) {
	c, err := lru.New(size)
	if err != nil {
		log.Fatalln(err)
	}
	ImageCache = c
}
