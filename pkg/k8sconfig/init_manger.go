package k8sconfig

import (
	taskv1alpha1 "github.com/myoperator/cicdoperator/pkg/apis/task/v1alpha1"
	"github.com/myoperator/cicdoperator/pkg/controller"
	corev1 "k8s.io/api/core/v1"
	"log"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"os"

	"github.com/myoperator/cicdoperator/pkg/client/clientset/versioned"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

//InitManager 初始化控制器管理器
func InitManager() {

	logf.SetLogger(zap.New())
	// 1. 初始化管理器
	mgr, err := manager.New(K8sRestConfig(),
		manager.Options{
			Logger: logf.Log.WithName("cicd_task"),
		})
	if err != nil {
		log.Fatal("创建管理器失败:", err.Error())
	}

	// 2. 将crd自定义的资源对象加入Schema
	err = taskv1alpha1.SchemeBuilder.AddToScheme(mgr.GetScheme())
	if err != nil {
		mgr.GetLogger().Error(err, "unable add schema")
		os.Exit(1)
	}

	// 3. 初始化使用code-generator
	taskClient := versioned.NewForConfigOrDie(K8sRestConfig())

	// 4. 初始化 controller
	taskController := controller.NewTaskController(
		mgr.GetEventRecorderFor("cicd_task"),
		taskClient,
	)
	// 5. 定义controller的管理对象等工作
	if err = builder.ControllerManagedBy(mgr).
		For(&taskv1alpha1.Task{}).
		Watches(&source.Kind{Type: &corev1.Pod{}},
			handler.Funcs{
				UpdateFunc: taskController.OnUpdate,
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
