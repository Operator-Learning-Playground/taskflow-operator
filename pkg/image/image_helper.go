package builder

import (
	"fmt"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"k8s.io/klog/v2"
	"strings"
)

// ImageCommand 镜像内部 command args
type ImageCommand struct {
	Command []string // 对应 docker 的 container-agent
	Args    []string // 对应 docker 的 cmd
}

func (ic *ImageCommand) String() string {
	return fmt.Sprintf("Command: [%s],Args :[%s]",
		strings.Join(ic.Command, " "), strings.Join(ic.Args, " "))
}

// Image 镜像对象
type Image struct {
	Ref     name.Reference           // lru 缓存内部使用此对象作为 key
	Name    string                   // image 名称 ex: alpine:3.12
	Digest  v1.Hash                  // hash 值
	Command map[string]*ImageCommand // map key ex: Linux/amd64 储存不同架构的 args command
}

// addCommand 加入
func (i *Image) addCommand(os, arch string, cmds []string, args []string) {
	key := fmt.Sprintf("%s/%s", os, arch)

	i.Command[key] = &ImageCommand{
		Command: cmds,
		Args:    args,
	}
}

// newImage 初始化 Image 对象
func newImage(name string, digest v1.Hash, ref name.Reference) *Image {
	return &Image{
		Name:    name,
		Digest:  digest,
		Ref:     ref,
		Command: make(map[string]*ImageCommand),
	}
}

// ParseImage 解析 image 镜像, 获取到镜像的 command args 信息
func (im *ImageManager) ParseImage(imgName string) (*Image, error) {
	ref, err := name.ParseReference(imgName, name.WeakValidation)
	if err != nil {
		klog.Error("parse image reference error: ", err)
		return nil, err
	}

	des, err := remote.Get(ref) // 获取镜像描述信息
	if err != nil {
		klog.Error("remote get image error: ", err)
		return nil, err
	}

	// 初始化业务 Image 对象
	imgBuilder := newImage(imgName, des.Digest, ref)

	// image 区分为两种类型：image 类型 index 类型

	// image 类型
	if des.MediaType.IsImage() {
		klog.Infof("[%v] image is Image type image", imgName)

		img, err := des.Image()
		if err != nil {
			klog.Error("parse Image type image error: ", err)
			return nil, err
		}
		conf, err := img.ConfigFile()
		if err != nil {
			klog.Error("parse Image type image config error: ", err)
			return nil, err
		}
		imgBuilder.addCommand(conf.OS, conf.Architecture, conf.Config.Entrypoint, conf.Config.Cmd)
		klog.Infof("Image Name: [%v], type: Image, os: [%v], Architecture: [%v], Entrypoint: [%v], Cmd: [%v]",
			imgName, conf.OS, conf.Architecture, conf.Config.Entrypoint, conf.Config.Cmd)

		return imgBuilder, nil
	}

	klog.Infof("[%v] image is Index type image", imgName)
	// index 模式
	idx, err := des.ImageIndex()
	if err != nil {
		klog.Error("parse Index type image error: ", err)
		return nil, err
	}
	mf, err := idx.IndexManifest()
	if err != nil {
		klog.Error("parse IndexManifest Index type image error: ", err)
		return nil, err
	}
	for _, d := range mf.Manifests {
		img, err := idx.Image(d.Digest)
		if err != nil {
			klog.Error("parse Index type image error: ", err)
			return nil, err
		}
		conf, err := img.ConfigFile()
		if err != nil {
			klog.Error("parse Index type image config error: ", err)
			return nil, err
		}
		imgBuilder.addCommand(conf.OS, conf.Architecture, conf.Config.Entrypoint, conf.Config.Cmd)
		klog.Infof("Image Name: [%v], type: Index, os: [%v], Architecture: [%v], Entrypoint: [%v], Cmd: [%v]",
			imgName, conf.OS, conf.Architecture, conf.Config.Entrypoint, conf.Config.Cmd)
	}
	return imgBuilder, nil
}
