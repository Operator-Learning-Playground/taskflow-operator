package builder

import (
	"fmt"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"k8s.io/klog/v2"
	"strings"
)

type ImageCommand struct {
	Command []string // 对应docker 的entrypoint
	Args    []string // 对应docker 的cmd
}

func (ic *ImageCommand) String() string {
	return fmt.Sprintf("Command是:%s,Args是:%s", strings.Join(ic.Command, " "), strings.Join(ic.Args, " "))
}

// Image
type Image struct {
	Ref     name.Reference           //增加了一个 。 缓存里直接用这个 作为key，更方便
	Name    string                   // image名称 ex: alpine:3.12
	Digest  v1.Hash                  // 唯一的hash
	Command map[string]*ImageCommand //  map key ex: Linux/amd64
}

func (i *Image) AddCommand(os, arch string, cmds []string, args []string) {
	key := fmt.Sprintf("%s/%s", os, arch)

	i.Command[key] = &ImageCommand{
		Command: cmds,
		Args:    args,
	}
}

// NewImage 初始化函数
func NewImage(name string, digest v1.Hash, ref name.Reference) *Image {
	return &Image{
		Name:    name,
		Digest:  digest,
		Ref:     ref,
		Command: make(map[string]*ImageCommand),
	}
}

// ParseImage 解析 image 镜像
func ParseImage(img string) (*Image, error) {

	ref, err := name.ParseReference(img, name.WeakValidation)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	des, err := remote.Get(ref) // 获取镜像描述信息
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	// 初始化业务 Image 对象
	imgBuilder := NewImage(img, des.Digest, ref)
	// image 类型
	if des.MediaType.IsImage() {
		img, err := des.Image()
		if err != nil {
			return nil, err
		}
		conf, err := img.ConfigFile()
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		imgBuilder.AddCommand(conf.OS, conf.Architecture, conf.Config.Entrypoint, conf.Config.Cmd)
		return imgBuilder, nil
	}

	// index 模式
	idx, err := des.ImageIndex()
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	mf, err := idx.IndexManifest()

	if err != nil {
		klog.Error(err)
		return nil, err
	}
	for _, d := range mf.Manifests {
		img, err := idx.Image(d.Digest)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		conf, err := img.ConfigFile()
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		imgBuilder.AddCommand(conf.OS, conf.Architecture, conf.Config.Entrypoint, conf.Config.Cmd)
		klog.Info("OS: ", conf.OS, " Architecture: ", conf.Architecture, " Entrypoint: ", conf.Config.Entrypoint,
			" Cmd: ", conf.Config.Cmd)
		//fmt.Println(cf.OS,"/",cf.Architecture,":",cf.Config.Entrypoint,cf.Config.Cmd)
	}
	return imgBuilder, nil
}
