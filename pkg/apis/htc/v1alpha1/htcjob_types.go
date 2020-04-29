package v1alpha1

import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    //batchv1 "k8s.io/api/batch/v1"
    //corev1 "k8s.io/api/core/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type ScriptSpec struct {
    Image string `json:"image"`
    Command string `json:"command"`
    Source string `json:"source"`
}

// HTCJobSpec defines the desired state of HTCJob
type HTCJobSpec struct {
    Name string `json:"name"`
    // +optional
    Queue int `json:"queue"`
    Script ScriptSpec `json:"script"`
    // +optional
    HTCopts string `json:"htcopts"`
}

type HTCJobStatus struct {
    UniqId int `json:"uniqid"`
    Active int `json:"active"`
    Succeeded int `json:"succeeded"`
    Failed int `json:"failed"`
    // +optional
    JobId []string `json:"jobid"`
}

// HTCJobStatus defines the observed state of HTCJob
////type HTCJobStatus struct {
////    Status batchv1.JobStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
////}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HTCJob is the Schema for the htcjobs API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=htcjobs,scope=Namespaced
type HTCJob struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

////    Spec   HTCJobSpec   `json:"spec,omitempty"`
////    Status HTCJobStatus `json:"status,omitempty"`
    Spec HTCJobSpec `json:"spec,omitempty"`
    Status HTCJobStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HTCJobList contains a list of HTCJob
type HTCJobList struct {
    metav1.TypeMeta `json:",inline"`
    metav1.ListMeta `json:"metadata,omitempty"`
    Items           []HTCJob `json:"items"`
}

func init() {
    SchemeBuilder.Register(&HTCJob{}, &HTCJobList{})
}
