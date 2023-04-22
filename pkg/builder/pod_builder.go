package builder

import (
	"context"
	"fmt"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/myoperator/cicdoperator/pkg/apis/task/v1alpha1"
	"github.com/myoperator/cicdoperator/pkg/common"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
	"strings"
)

type PodBuilder struct {
	task   *v1alpha1.Task
	client client.Client
}

func NewPodBuilder(task *v1alpha1.Task, client client.Client) *PodBuilder {
	return &PodBuilder{task: task, client: client}
}

func (pb *PodBuilder) setInitContainer(pod *v1.Pod) {
	pod.Spec.InitContainers = []v1.Container{
		{
			Name:            pod.Name + "init",
			Image:           "shenyisyn/entrypoint:v1",
			ImagePullPolicy: v1.PullIfNotPresent,
			Command:         []string{"cp", "/app/entrypoint", "/entrypoint/bin"},
			VolumeMounts: []v1.VolumeMount{
				{
					Name:      "entrypoint-volume",
					MountPath: "/entrypoint/bin",
				},
			},
		},
	}
}

func (pb *PodBuilder) setPodVolumes(pod *v1.Pod) {
	volumes := []v1.Volume{
		{
			Name: "entrypoint-volume",
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: "podinfo",
			VolumeSource: v1.VolumeSource{
				DownwardAPI: &v1.DownwardAPIVolumeSource{
					Items: []v1.DownwardAPIVolumeFile{
						{
							Path: "order",
							FieldRef: &v1.ObjectFieldSelector{
								FieldPath: "metadata.annotations['taskorder']",
							},
						},
					},
				},
			},
		},
	}
	pod.Spec.Volumes = append(pod.Spec.Volumes, volumes...)
}

//设置 POD元信息 包含 注解
func (pb *PodBuilder) setPodMeta(pod *v1.Pod) {
	pod.Namespace = pb.task.Namespace
	pod.Name = "task-pod-" + pb.task.Name          // pod名称
	pod.Spec.RestartPolicy = v1.RestartPolicyNever // 设置永不重启
	pod.Annotations = map[string]string{
		"taskorder": "0",
	}
}

func (pb *PodBuilder) setContainer(index int, step v1alpha1.TaskStep) (v1.Container, error) {
	// 这里要强烈注意：step.Command必须要设置，如果没设置则通过http 去远程取。取不到直接报错
	command := step.Command // 取出它 原始的command ,是个 string切片
	if len(command) == 0 {  //没有写 command  . 需要从网上去解析
		ref, err := name.ParseReference(step.Image, name.WeakValidation)
		if err != nil {
			return step.Container, err
		}
		// 从缓存获取
		var getImage *Image

		// 如果缓存有
		if v, ok := common.ImageCache.Get(ref); ok {
			getImage = v.(*Image)
		} else { //缓存没有的情况下
			img, err := ParseImage(step.Image) //解析镜像
			if err != nil {
				return step.Container, err
			}
			common.ImageCache.Add(img.Ref, img) //加入缓存
			getImage = img
		}
		// 暂时先写死 OS=Linux/amd64
		tempOs := "linux/amd64"
		if imgObj, ok := getImage.Command[tempOs]; ok {
			command = imgObj.Command
			if len(step.Args) == 0 {  // 覆盖args （假设有的话)
				step.Args = imgObj.Args
			}
		} else {
			return step.Container, fmt.Errorf("error image command")
		}

	}

	args := step.Args //取出它原始的 args

	step.Container.ImagePullPolicy = v1.PullIfNotPresent //强迫设置拉取策略
	step.Container.Command = []string{"/entrypoint/bin/entrypoint"}
	step.Container.Args = []string{
		"--wait", "/etc/podinfo/order",
		"--waitcontent", strconv.Itoa(index + 1),
		"--out", "stdout", // entrypoint中写上stdout 就会定向到标准输出
		"--command",
	}
	// "sh -c"
	step.Container.Args = append(step.Container.Args, strings.Join(command, " "))
	step.Container.Args = append(step.Container.Args, args...)
	//设置挂载点
	step.Container.VolumeMounts = []v1.VolumeMount{
		{
			Name:      "entrypoint-volume",
			MountPath: "/entrypoint/bin",
		},
		{
			Name:      "podinfo",
			MountPath: "/etc/podinfo",
		},
	}
	return step.Container, nil
}

func (pb *PodBuilder) Build(ctx context.Context) error {

	newPod := &v1.Pod{}
	pb.setPodMeta(newPod)
	pb.setInitContainer(newPod)

	cList := make([]v1.Container, 0)
	for index, step := range pb.task.Spec.Steps {
		// 设置image拉起策略
		getContainer,err := pb.setContainer(index,step)
		if err != nil{
			fmt.Println("err: ", err)
			return err
		}
		cList = append(cList, getContainer)

	}

	newPod.Spec.Containers = cList
	pb.setPodVolumes(newPod) // 设置pod数据卷--重要，包含了downwardAPI 和emptyDir

	// 设置ownerReferences
	newPod.OwnerReferences = append(newPod.OwnerReferences, metav1.OwnerReference{
		APIVersion: pb.task.APIVersion,
		Kind:       pb.task.Kind,
		Name:       pb.task.Name,
		UID:        pb.task.UID,
	})

	return pb.client.Create(ctx, newPod)
}

func (pb *PodBuilder) setStep(pod *v1.Pod) {
	pod.Annotations = map[string]string{
		"taskorder": "0",
	}

}
