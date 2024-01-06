package main

import (
	taskv1alpha1 "github.com/myoperator/taskflowoperator/pkg/apis/task/v1alpha1"
	"github.com/myoperator/taskflowoperator/pkg/controller"
	builder2 "github.com/myoperator/taskflowoperator/pkg/image"
	"github.com/myoperator/taskflowoperator/pkg/k8sconfig"
	corev1 "k8s.io/api/core/v1"
	_ "k8s.io/code-generator"
	"log"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func main() {
	logf.SetLogger(zap.New())
	// 1. 初始化管理器
	mgr, err := manager.New(k8sconfig.K8sRestConfig(),
		manager.Options{
			Logger: logf.Log.WithName("task-flow operator"),
		})
	if err != nil {
		log.Fatal("init manager error:", err.Error())
	}

	// 2. 将crd自定义的资源对象加入Schema
	err = taskv1alpha1.SchemeBuilder.AddToScheme(mgr.GetScheme())
	if err != nil {
		mgr.GetLogger().Error(err, "unable add schema")
		os.Exit(1)
	}

	// 3. 初始化使用 code-generator 生成器的 client 示例
	//taskClient := versioned.NewForConfigOrDie(K8sRestConfig())

	// 4. 初始化 controller
	taskController := controller.NewTaskController(
		mgr.GetEventRecorderFor("task-flow operator"),
		mgr.GetClient(),
		builder2.NewImageManager(100),
	)

	// 5. 定义controller的管理对象等工作
	if err = builder.ControllerManagedBy(mgr).
		For(&taskv1alpha1.Task{}).
		Watches(&source.Kind{Type: &corev1.Pod{}},
			handler.Funcs{
				UpdateFunc: taskController.OnUpdatePodHandler,
				DeleteFunc: taskController.OnDeletePodHandler,
			},
		).
		Complete(taskController); err != nil {
		mgr.GetLogger().Error(err, "unable to create manager")
		os.Exit(1)
	}

	// 6. 启动管理器
	if err = mgr.Start(signals.SetupSignalHandler()); err != nil {
		mgr.GetLogger().Error(err, "unable to start manager")
	}
}
