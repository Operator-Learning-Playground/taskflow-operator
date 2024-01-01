package builder

import (
	lru "github.com/hashicorp/golang-lru"
	"log"
)

type ImageManager struct {
	ImageCache *lru.Cache
}

func NewImageManager(size int) *ImageManager {

	c, err := lru.New(size)
	if err != nil {
		log.Fatalln(err)
	}
	return &ImageManager{
		ImageCache: c,
	}
}
