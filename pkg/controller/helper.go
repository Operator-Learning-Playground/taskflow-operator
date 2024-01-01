package controller

import (
	"context"
	"fmt"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/myoperator/cicdoperator/pkg/apis/task/v1alpha1"
	"github.com/myoperator/cicdoperator/pkg/common"
	"github.com/myoperator/cicdoperator/pkg/image"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
	"strings"
)

// setInitContainer 设置 InitContainer
func (r *TaskController) setInitContainer(pod *v1.Pod) {
	// entrypoint 容器
	// 用意：把通用入口镜像挂载到 Pod 中，使所有容器都共享
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

// setPodVolumes 设置挂载 volumes
// 设置 pod 数据卷: 包含 downwardAPI 和 emptyDir
func (r *TaskController) setPodVolumes(pod *v1.Pod) {
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
	EntryPointImage              = "docker.io/taskflow/entrypoint:v1.0"
	TaskPodPrefix                = "task-pod-"
	AnnotationTaskOrderKey       = "taskorder"
	AnnotationTaskOrderInitValue = "0"
	AnnotationExitOrder          = "-1" // 退出 step 用的标识
)

// setPodMeta 设置 Pod 信息
func (r *TaskController) setPodMeta(task *v1alpha1.Task, pod *v1.Pod) {
	pod.Namespace = task.Namespace
	//pod.Name = TaskPodPrefix + pb.task.Name 	// pod名称
	// 如果要尾部增加随机字符串的使用方法，在k8s内部层面来保证唯一
	pod.GenerateName = TaskPodPrefix + task.Name + "-"
	// 设置永不重启
	pod.Spec.RestartPolicy = v1.RestartPolicyNever

	// 设置 labels 用于查找 Pod 使用
	pod.Labels = map[string]string{
		"type":     "taskpod",
		"taskname": task.Name,
	}

	// 设置初始 annotation key = "0"
	pod.Annotations = map[string]string{
		AnnotationTaskOrderKey: AnnotationTaskOrderInitValue,
	}
}

// getChildPod 判断 Task 管理的 Pod 是否存在，如果存在直接返回 Pod
func (r *TaskController) getChildPod(task *v1alpha1.Task) (*v1.Pod, error) {
	pods := &v1.PodList{}
	err := r.client.List(context.Background(), pods,
		client.HasLabels{"type", "taskname"}, client.InNamespace(task.Namespace))

	if err != nil {
		klog.Error("found task flow pod error: ", err)
		return nil, err
	}

	// 遍例找特定 Pod ，藉由 UID 查找
	for _, pod := range pods.Items {
		for _, own := range pod.OwnerReferences {
			if own.UID == task.UID {
				return &pod, err
			}
		}
	}
	return nil, fmt.Errorf("found no task-pod")
}

// setContainer 设置 Container
func (r *TaskController) setContainer(index int, step v1alpha1.TaskStep) (v1.Container, error) {
	// 注意：step.Command必须要设置，如果没有设置则通过http远程获取，取不到直接报错
	// 取出原 command
	command := step.Command
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
			var getImage *builder.Image
			if v, ok := r.imageManager.ImageCache.Get(ref); ok {
				getImage = v.(*builder.Image)
			} else {
				// 解析镜像
				img, err := r.imageManager.ParseImage(step.Image)
				if err != nil {
					klog.Error("parse container image error: ", err)
					return step.Container, err
				}
				// 加入缓存
				r.imageManager.ImageCache.Add(img.Ref, img)
				getImage = img
			}

			// TODO: 仅支持 os = Linux/amd64
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
		// 强迫设置拉取策略
		step.Container.ImagePullPolicy = v1.PullIfNotPresent
		step.Container.Command = []string{"/entrypoint/bin/entrypoint"}
		step.Container.Args = []string{
			"--wait", "/etc/podinfo/order",
			"--waitcontent", strconv.Itoa(index + 1),
			// entrypoint 中写上 stdout 就会定向到标准输出
			"--out", "stdout",
			"--command",
		}
		// ex: "sh -c"，进行拼接
		step.Container.Args = append(step.Container.Args, strings.Join(command, " "))
		step.Container.Args = append(step.Container.Args, args...)
		klog.Info(step.Container.Args)

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
	PodInfoVolume        = "podinfo"               // 存储Pod信息用于 downwardAPI
)

// deployTaskFlow 调协 task flow
func (r *TaskController) deployTaskFlow(ctx context.Context, task *v1alpha1.Task) error {

	// 判断 Pod 是否存在
	getPod, err := r.getChildPod(task)
	// 代表 Pod 已被创建，代表 taskFlow 正在执行
	if err == nil {
		if getPod.Status.Phase == v1.PodRunning {
			// 如果为起始状态，先把 annotation 设置为1，让流水线可以执行，并更新
			if getPod.Annotations[AnnotationTaskOrderKey] == AnnotationTaskOrderInitValue {
				getPod.Annotations[AnnotationTaskOrderKey] = "1"
				r.event.Event(task, v1.EventTypeNormal, "TaskFlow Start", "TaskFlow start to run step 1 container")
				return r.client.Update(ctx, getPod)
			} else {
				// 如果是其他状态，则调用forward方法前进
				if err := r.forward(ctx, getPod, task); err != nil {
					return err
				}
			}
		}
		klog.Info("pod status: ", getPod.Status.Phase)
		klog.Info("annotation order:  ", getPod.Annotations[AnnotationTaskOrderKey])

		return nil
	}

	r.event.Event(task, v1.EventTypeNormal, "TaskFlow Create", "TaskFlow start to create pod")
	// 2. 创建流程
	newPod := &v1.Pod{}
	r.setPodMeta(task, newPod)
	r.setInitContainer(newPod)

	cList := make([]v1.Container, 0)
	for index, step := range task.Spec.Steps {
		// 完成容器设置
		getContainer, err := r.setContainer(index, step)
		if err != nil {
			klog.Error("setContainer err: ", err)
			return err
		}
		cList = append(cList, getContainer)
	}

	newPod.Spec.Containers = cList
	// 设置 volumes
	r.setPodVolumes(newPod)

	// 设置 ownerReferences
	newPod.OwnerReferences = append(newPod.OwnerReferences, metav1.OwnerReference{
		APIVersion: task.APIVersion,
		Kind:       task.Kind,
		Name:       task.Name,
		UID:        task.UID,
	})

	return r.client.Create(ctx, newPod)
}

// forward taskFlow 流转逻辑
// 主要使用 AnnotationTaskOrderKey 标示与 Pod, Container 状态进行描述
func (r *TaskController) forward(ctx context.Context, pod *v1.Pod, task *v1alpha1.Task) error {
	// Pod 状态 Succeeded 代表整个 Pod 执行完毕
	if pod.Status.Phase == v1.PodSucceeded {
		r.event.Eventf(task, v1.EventTypeNormal, "TaskFlow Completed", "TaskFlow completed successfully")
		return nil
	}

	// AnnotationTaskOrderKey = "-1" 表示流程有错误，直接退出
	if pod.Annotations[AnnotationTaskOrderKey] == AnnotationExitOrder {
		r.event.Eventf(task, v1.EventTypeWarning, "TaskFlow Failed", "TaskFlow exit")
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
			// AnnotationTaskOrderKey 设置为 -1
			r.event.Eventf(task, v1.EventTypeWarning, "TaskFlow Failed", "TaskFlow failed to run step %v container", order)
			pod.Annotations[AnnotationTaskOrderKey] = AnnotationExitOrder
			return r.client.Update(ctx, pod)
			//pod.Status.Phase=v1.PodFailed
			//return pb.Client.Status().Update(ctx,pod)
		}
	}

	// AnnotationTaskOrderKey++
	order++
	pod.Annotations[AnnotationTaskOrderKey] = strconv.Itoa(order)
	r.event.Eventf(task, v1.EventTypeNormal, "TaskFlow Running", "TaskFlow continuously run step %v container", order)

	return r.client.Update(ctx, pod)
}
