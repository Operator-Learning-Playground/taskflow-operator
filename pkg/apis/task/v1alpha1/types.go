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
}

type TaskSpec struct {
	Steps []TaskStep `json:"steps,omitempty"`
}

type TaskStep struct {
	corev1.Container 			`json:",inline"` // 容器对象
	JobStatus        bool		`json:"job_status",omitempty", default:"false"`
	Script           string     `json:"script,omitempty"` // 支持脚本命令
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TaskList ..
type TaskList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Task `json:"items"`
}
