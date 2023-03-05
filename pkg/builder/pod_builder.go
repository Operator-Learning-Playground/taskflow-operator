package builder

import (
	"context"
	"github.com/myoperator/cicdoperator/pkg/apis/task/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type PodBuilder struct {
	task *v1alpha1.Task
	client client.Client
}

func NewPodBuilder(task *v1alpha1.Task, client client.Client) *PodBuilder {
	return &PodBuilder{task: task, client: client}
}

func (pb *PodBuilder) Build(ctx context.Context) error {

	newPod := &v1.Pod{}
	newPod.Name = "task-pod-" + pb.task.Name
	newPod.Namespace = pb.task.Namespace
	// 设置永不重启
	newPod.Spec.RestartPolicy = v1.RestartPolicyNever
	cList := make([]v1.Container, 0)
	for _, step := range pb.task.Spec.Steps {
		// 设置image拉起策略
		step.Container.ImagePullPolicy = v1.PullIfNotPresent
		cList = append(cList, step.Container)
	}

	newPod.Spec.Containers = cList
	// 设置ownerReferences
	newPod.OwnerReferences = append(newPod.OwnerReferences, metav1.OwnerReference{
		APIVersion: pb.task.APIVersion,
		Kind: pb.task.Kind,
		Name: pb.task.Name,
		UID: pb.task.UID,
	})

	return pb.client.Create(ctx, newPod)

}

