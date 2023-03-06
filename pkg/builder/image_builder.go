package builder

import (
	"fmt"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"strings"
)


type ImageCommand struct {
	Command []string // 对应docker 的entrypoint
	Args []string    // 对应docker 的cmd
}


func(ic *ImageCommand) String() string{
	return fmt.Sprintf("Command是:%s,Args是:%s", strings.Join(ic.Command," "), strings.Join(ic.Args," "))
}


// Image
type Image struct {
	// image名称 ex: alpine:3.12
	Name string
	// 唯一的hash
	Digest v1.Hash
	//  map key ex: Linux/amd64
	Command map[string]*ImageCommand
}


func(i *Image) AddCommand(os, arch string , cmds []string, args []string){
	key := fmt.Sprintf("%s/%s", os, arch)

	i.Command[key] = &ImageCommand{
		Command: cmds,
		Args: args,
	}
}


// NewImage 初始化函数
func NewImage(name string ,digest v1.Hash) *Image{
	return &Image {
		Name: name,
		Digest: digest,
		Command: make(map[string]*ImageCommand),
	}
}

// ParseImage 解析 image 镜像
func ParseImage(img string) (*Image,error) {

	ref, err := name.ParseReference(img)
	if err != nil {
		return nil, err
	}

	des, err := remote.Get(ref) // 获取镜像描述信息
	if err != nil {
		return nil, err
	}

	//初始化我们的业务 Image 对象
	imgBuilder := NewImage(img, des.Digest)
	//image 类型
	if des.MediaType.IsImage(){
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


	//下方是 index 模式
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
			return nil,err
		}
		conf, err := img.ConfigFile()
		if err != nil {
			return nil, err
		}
		imgBuilder.AddCommand(conf.OS, conf.Architecture, conf.Config.Entrypoint, conf.Config.Cmd)
		//fmt.Println(cf.OS,"/",cf.Architecture,":",cf.Config.Entrypoint,cf.Config.Cmd)
	}
	return  imgBuilder, nil
}

