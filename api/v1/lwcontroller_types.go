package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type LwcontrollerSpec struct {
	CheckInterval int64 `json:"updateInterval,omitempty"`
}

type LwcontrollerStatus struct {
	GpuList     GpuList      `json:"gpuList,omitempty"`
	GpuNumber   int16        `json:"cardNumber,omitempty"`
	UpdateTime  *metav1.Time `json:"updateTime,omitempty"`
	TotalMemory uint64       `json:"totalMemory,omitempty"`
	FreeMemory  uint64       `json:"freeMemory,omitempty"`
}

type LwcontrollerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Lwcontroller `json:"items"`
}

type Lwcontroller struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              LwcontrollerSpec   `json:"spec,omitempty"`
	Status            LwcontrollerStatus `json:"status,omitempty"`
}

type GpuList []Gpu

func (in GpuList) Len() int {
	return len(in)
}

//used in sort
func (in GpuList) Less(i, j int) bool {
	return in[i].ID < in[j].ID
}

func (in GpuList) Swap(i, j int) {
	in[i], in[j] = in[j], in[i]
}

type Gpu struct {
	ID          int16  `json:"id"`
	Health      string `json:"health,omitempty"`
	Power       uint   `json:"power,omitempty"`
	TotalMemory uint64 `json:"totalMemory,omitempty"`
	Clock       uint   `json:"clock,omitempty"`
	FreeMemory  uint64 `json:"freeMemory,omitempty"`
	Core        uint   `json:"core,omitempty"`
	Bandwidth   uint   `json:"bandwidth,omitempty"`
}

func init() {
	SchemeBuilder.Register(&Lwcontroller{}, &LwcontrollerList{})
}
