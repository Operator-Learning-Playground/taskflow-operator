package controller

import (
	"context"
	"github.com/myoperator/cicdoperator/pkg/apis/task/v1alpha1"
	clientset "github.com/myoperator/cicdoperator/pkg/client/clientset/versioned"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type TaskController struct {
	client.Client
	*clientset.Clientset
	Event record.EventRecorder

}

func NewTaskController(event record.EventRecorder, clientset *clientset.Clientset) *TaskController {
	return &TaskController{
		Event: event,
		Clientset: clientset,
	}
}

func (r *TaskController) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {

	t := &v1alpha1.Task{}
	err := r.Client.Get(ctx, req.NamespacedName, t)
	if err != nil {
		return reconcile.Result{}, err
	}

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
