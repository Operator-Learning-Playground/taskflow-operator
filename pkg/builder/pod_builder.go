package builder

import (
	"context"
	"fmt"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/myoperator/cicdoperator/pkg/apis/task/v1alpha1"
	"github.com/myoperator/cicdoperator/pkg/common"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
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

// setInitContainer 设置InitContainer
func (pb *PodBuilder) setInitContainer(pod *v1.Pod) {
	// entrypoint容器
	pod.Spec.InitContainers = []v1.Container{
		{
			Name:            pod.Name + "init",
			Image:           EntryPointImage,
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

// setPodVolumes 设置挂载
// 设置pod数据卷: 包含downwardAPI 和emptyDir
func (pb *PodBuilder) setPodVolumes(pod *v1.Pod) {
	volumes := []v1.Volume{
		{
			Name: EntryPointVolume,
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: PodInfoVolume,
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
		{
			Name: ExecuteScriptsVolume,
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		},
	}
	pod.Spec.Volumes = append(pod.Spec.Volumes, volumes...)
}

const (
	EntryPointImage              = "docker.io/shenyisyn/entrypoint:v1.1"
	TaskPodPrefix                = "task-pod-"
	AnnotationTaskOrderKey       = "taskorder"
	AnnotationTaskOrderInitValue = "0"
	AnnotationExitOrder          = "-1" //退出step用的Order标识
)

// setPodMeta 设置Pod信息
func (pb *PodBuilder) setPodMeta(pod *v1.Pod) {
	pod.Namespace = pb.task.Namespace
	//pod.Name = TaskPodPrefix + pb.task.Name 	// pod名称
	// 如果要尾部增加随机字符串的使用方法，在k8s内部层面来保证唯一
	pod.GenerateName = TaskPodPrefix + pb.task.Name + "-"
	pod.Spec.RestartPolicy = v1.RestartPolicyNever // 设置永不重启

	pod.Labels = map[string]string{
		"type":     "taskpod",
		"taskname": pb.task.Name,
	}

	pod.Annotations = map[string]string{
		AnnotationTaskOrderKey: AnnotationTaskOrderInitValue,
	}
}

// getChildPod 判断Task管理的Pod是否存在，如果存在直接返回Pod
func (pb *PodBuilder) getChildPod() (*v1.Pod, error) {
	pods := &v1.PodList{}
	err := pb.client.List(context.Background(), pods,
		client.HasLabels{"type", "taskname"}, client.InNamespace(pb.task.Namespace))

	if err != nil {
		klog.Error("getChildPod: ", err)
		return nil, err
	}

	// 遍例找特定Pod
	for _, pod := range pods.Items {
		for _, own := range pod.OwnerReferences {
			if own.UID == pb.task.UID {
				return &pod, err
			}
		}
	}
	return nil, fmt.Errorf("found no task-pod")
}

// setContainer 设置Container
func (pb *PodBuilder) setContainer(index int, step v1alpha1.TaskStep) (v1.Container, error) {
	// 注意：step.Command必须要设置，如果没有设置则通过http远程获取，取不到直接报错
	command := step.Command // 取出原command
	if step.Script == "" {
		klog.Info(step.Name, " use normal mode.....")
		// 如果没有command，远程获取
		if len(command) == 0 {
			ref, err := name.ParseReference(step.Image, name.WeakValidation)
			if err != nil {
				klog.Error("parse container command reference error: ", err)
				return step.Container, err
			}

			// 从缓存获取，如果有就从缓存拿取，没有就远程调用
			var getImage *Image
			if v, ok := common.ImageCache.Get(ref); ok {
				getImage = v.(*Image)
			} else {
				// 解析镜像
				img, err := ParseImage(step.Image)
				if err != nil {
					klog.Error("parse container image error: ", err)
					return step.Container, err
				}
				// 加入缓存
				common.ImageCache.Add(img.Ref, img)
				getImage = img
			}
			// 仅支持 OS=Linux/amd64
			tempOs := "linux/amd64"
			if imgObj, ok := getImage.Command[tempOs]; ok {
				command = imgObj.Command
				// 如果有args，覆盖args
				if len(step.Args) == 0 {
					step.Args = imgObj.Args
				}
			} else {
				return step.Container, fmt.Errorf("error image command")
			}

		}

		// 先取出它原始的 args，把其他需要的数据都放入，再做拼接
		args := step.Args

		step.Container.ImagePullPolicy = v1.PullIfNotPresent //强迫设置拉取策略
		step.Container.Command = []string{"/entrypoint/bin/entrypoint"}
		step.Container.Args = []string{
			"--wait", "/etc/podinfo/order",
			"--waitcontent", strconv.Itoa(index + 1),
			"--out", "stdout", // entrypoint中写上stdout 就会定向到标准输出
			"--command",
		}
		// ex: "sh -c"，进行拼接
		step.Container.Args = append(step.Container.Args, strings.Join(command, " "))
		step.Container.Args = append(step.Container.Args, args...)
		fmt.Println(step.Container.Args)

	} else {
		// 脚本模式
		klog.Info(step.Name, "use script mode.....")
		// 如果script不为空，表示使用script模式，command和args无效

		step.Container.Command = []string{"sh"} // 使用sh命令
		// 使用脚本：1.找到文件夹，2.创建文件并修改权限，3.写入文件 4.解密
		step.Container.Args = []string{"-c", fmt.Sprintf(`scriptfile="/execute/scripts/%s";touch ${scriptfile} && chmod +x ${scriptfile};echo "%s" > ${scriptfile};/entrypoint/bin/entrypoint --wait /etc/podinfo/order --waitcontent %d --out stdout  --encodefile ${scriptfile};`,
			step.Name, common.EncodeScript(step.Script), index+1)}
	}

	// 设置VolumeMounts挂载点
	step.Container.VolumeMounts = []v1.VolumeMount{
		{
			Name:      EntryPointVolume,
			MountPath: "/entrypoint/bin",
		},
		{
			Name:      ExecuteScriptsVolume, //设置 script挂载卷，不管有没有设置
			MountPath: "/execute/scripts",
		},
		{
			Name:      PodInfoVolume,
			MountPath: "/etc/podinfo",
		},
	}
	return step.Container, nil
}

const (
	EntryPointVolume     = "entrypoint-volume"     // 入口程序挂载
	ExecuteScriptsVolume = "execute-inner-scripts" // script属性 存储卷
	PodInfoVolume        = "podinfo"               // 存储Pod信息  用于dowardApi
)

// Build 创建方法
func (pb *PodBuilder) Build(ctx context.Context) error {

	// 1. 先判断pod是否存在
	getPod, err := pb.getChildPod()
	// 代表Pod已被创建
	if err == nil {
		if getPod.Status.Phase == v1.PodRunning {
			// 如果为起始状态，先把annotation 设置为1，让流水线可以执行，并更新
			if getPod.Annotations[AnnotationTaskOrderKey] == AnnotationTaskOrderInitValue {
				getPod.Annotations[AnnotationTaskOrderKey] = "1"
				return pb.client.Update(ctx, getPod)
			} else {
				// 如果是其他状态，则调用forward方法前进
				if err := pb.forward(ctx, getPod); err != nil {
					return err
				}
			}
		}

		fmt.Println("new status: ", getPod.Status.Phase)
		fmt.Println("annotation order:  ", getPod.Annotations[AnnotationTaskOrderKey])

		return nil
	}

	// 2. 创建流程，准备数据
	newPod := &v1.Pod{}
	pb.setPodMeta(newPod)
	pb.setInitContainer(newPod)

	cList := make([]v1.Container, 0)
	for index, step := range pb.task.Spec.Steps {
		// 设置image拉起策略
		getContainer, err := pb.setContainer(index, step)
		if err != nil {
			klog.Error("setContainer err: ", err)
			return err
		}
		cList = append(cList, getContainer)

	}

	newPod.Spec.Containers = cList
	pb.setPodVolumes(newPod)

	// 设置ownerReferences
	newPod.OwnerReferences = append(newPod.OwnerReferences, metav1.OwnerReference{
		APIVersion: pb.task.APIVersion,
		Kind:       pb.task.Kind,
		Name:       pb.task.Name,
		UID:        pb.task.UID,
	})

	return pb.client.Create(ctx, newPod)
}

// forward step前进的方法，容器的流转
func (pb *PodBuilder) forward(ctx context.Context, pod *v1.Pod) error {
	if pod.Status.Phase == v1.PodSucceeded {
		return nil
	}
	// Order值 ==-1  代表 有一个step出错了。不做处理。
	if pod.Annotations[AnnotationTaskOrderKey] == AnnotationExitOrder {
		return nil
	}

	order, err := strconv.Atoi(pod.Annotations[AnnotationTaskOrderKey])
	if err != nil {
		klog.Error("AnnotationTaskOrderKey err: ", err)
		return err
	}
	// 容器长度相等，代表已经到了最后一个
	if order == len(pod.Spec.Containers) {
		return nil
	}

	// 代表当前的容器可能在等待或者正在运行
	containerState := pod.Status.ContainerStatuses[order-1].State

	// 代表目前的容器正在等待或者运行
	if containerState.Terminated == nil {
		return nil
	} else {
		// 代表非正常退出，容器出错
		if containerState.Terminated.ExitCode != 0 {
			//把Order 值改成 -1
			pod.Annotations[AnnotationTaskOrderKey] = AnnotationExitOrder
			return pb.client.Update(ctx, pod)
			//pod.Status.Phase=v1.PodFailed
			//return pb.Client.Status().Update(ctx,pod)
		}
	}

	// 流水线加一
	order++
	pod.Annotations[AnnotationTaskOrderKey] = strconv.Itoa(order)
	return pb.client.Update(ctx, pod)
}
