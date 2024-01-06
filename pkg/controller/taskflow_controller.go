package controller

import (
	"context"
	"github.com/myoperator/taskflowoperator/pkg/apis/task/v1alpha1"
	"github.com/myoperator/taskflowoperator/pkg/image"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type TaskController struct {
	client       client.Client
	event        record.EventRecorder
	imageManager *builder.ImageManager
}

func NewTaskController(event record.EventRecorder, client client.Client, imageManager *builder.ImageManager) *TaskController {
	return &TaskController{
		event:        event,
		client:       client,
		imageManager: imageManager,
	}
}

func (r *TaskController) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {

	t := &v1alpha1.Task{}
	err := r.client.Get(ctx, req.NamespacedName, t)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			klog.Error("get task error: ", err)
			return reconcile.Result{}, err
		}
		// 如果未找到的错误，不再进入调协
		return reconcile.Result{}, nil
	}

	// 调协 task flow
	err = r.deployTaskFlow(ctx, t)
	if err != nil {
		klog.Error("deploy task flow err: ", err)
		return reconcile.Result{}, err
	}
	klog.Info("successful reconcile")

	return reconcile.Result{}, nil

}

func (r *TaskController) OnUpdatePodHandler(event event.UpdateEvent, limitingInterface workqueue.RateLimitingInterface) {
	for _, ref := range event.ObjectNew.GetOwnerReferences() {
		if ref.Kind == v1alpha1.TaskKind && ref.APIVersion == v1alpha1.TaskApiVersion {
			// 重新放入Reconcile调协方法
			limitingInterface.Add(reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name: ref.Name, Namespace: event.ObjectNew.GetNamespace(),
				},
			})
		}
	}
}

func (r *TaskController) OnDeletePodHandler(event event.DeleteEvent, limitingInterface workqueue.RateLimitingInterface) {
	for _, ref := range event.Object.GetOwnerReferences() {
		if ref.Kind == v1alpha1.TaskKind && ref.APIVersion == v1alpha1.TaskApiVersion {
			// 重新入列，这样删除pod后，就会进入调和loop，发现ownerReference还在，会立即创建出新的pod。
			klog.Info("delete pod: ", event.Object.GetName(), event.Object.GetObjectKind())
			limitingInterface.Add(reconcile.Request{
				NamespacedName: types.NamespacedName{Name: ref.Name,
					Namespace: event.Object.GetNamespace()}})
		}
	}
}
