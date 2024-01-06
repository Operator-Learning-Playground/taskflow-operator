package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Task 任务
type Task struct {
	metav1.TypeMeta `json:",inline"`

	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec TaskSpec `json:"spec,omitempty"`
	// TODO: 加入 TaskStatus 用于查看每个工作流的状态 Running Complete Failed ...
	// TODO: 记录运行时间，需要在 status加入一个时间值字段 开始时记录时间  结束或错误退出时记录时间
	// 在调协中修改 status
	Status TaskStatus `json:"status,omitempty"`
}

type TaskSpec struct {
	Steps []TaskStep `json:"steps,omitempty"`
}

type TaskStep struct {
	corev1.Container `json:",inline"` // 容器对象
	// 支持脚本命令
	Script string `json:"script,omitempty"`
}

type TaskStatus struct {
	// Status 任务状态
	Status string `json:"status"`
	// StartAt 任务起始时间
	StartAt string `json:"startAt"`
	// Duration 任务执行多久
	Duration string `json:"duration"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TaskList ..
type TaskList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Task `json:"items"`
}
