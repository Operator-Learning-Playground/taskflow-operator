package controller

import (
	"context"
	"github.com/myoperator/cicdoperator/pkg/apis/task/v1alpha1"
	"github.com/myoperator/cicdoperator/pkg/builder"
	clientset "github.com/myoperator/cicdoperator/pkg/client/clientset/versioned"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type TaskController struct {
	client.Client
	*clientset.Clientset
	Event record.EventRecorder
}

func NewTaskController(event record.EventRecorder, clientset *clientset.Clientset) *TaskController {
	return &TaskController{
		Event:     event,
		Clientset: clientset,
	}
}

func (r *TaskController) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {

	t := &v1alpha1.Task{}
	err := r.Client.Get(ctx, req.NamespacedName, t)
	if err != nil {
		klog.Error("get task error: ", err)
		return reconcile.Result{}, err
	}

	// 处理调协pod
	podBuilder := builder.NewPodBuilder(t, r.Client)
	err = podBuilder.Build(ctx)
	if err != nil {
		klog.Error("create pod error: ", err)
		return reconcile.Result{}, err
	}

	// 法二：可以使用生成的client端
	//task, err := r.ApiV1alpha1().Tasks(req.Namespace).Get(ctx, req.Name, metav1.GetOptions{})
	//if err!=nil{
	//	return reconcile.Result{}, err
	//}

	klog.Info("task: ", t)

	return reconcile.Result{}, nil

}

// InjectClient 框架要注入clientSet
func (r *TaskController) InjectClient(c client.Client) error {
	r.Client = c
	return nil
}

func (r *TaskController) OnUpdatePodHandler(event event.UpdateEvent, limitingInterface workqueue.RateLimitingInterface) {
	for _, ref := range event.ObjectNew.GetOwnerReferences() {
		if ref.Kind == v1alpha1.TaskKind && ref.APIVersion == v1alpha1.TaskApiVersion {
			// 重新放入Reconcile调协方法
			limitingInterface.Add(reconcile.Request{
				types.NamespacedName{
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
