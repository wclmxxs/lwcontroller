/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sort"
	"time"

	webappv1 "github.com/wclmxxs/lwcontroller/api/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// LwcontrollerReconciler reconciles a Lwcontroller object
type Lw struct {
	client        client.Client
	cache         cache.Cache
	Scheme        *runtime.Scheme
	nodeName      string
	checkInterval int64
	FreeMemory    uint64
	TotalMemory   uint64
	lessGpuNumber int16
}

func New(interval int64, client client.Client, cache cache.Cache) *Lw {
	return &Lw{
		lessGpuNumber: 0,
		checkInterval: interval,
		client:        client,
		cache:         cache,
	}
}

func Start(lw *Lw) {
	nodeName := os.Getenv("NODE_NAME")
	lwcontroller := webappv1.Lwcontroller{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "lw",
			Name:      nodeName,
		},
		Spec: webappv1.LwcontrollerSpec{
			CheckInterval: lw.checkInterval,
		},
	}
	lw.nodeName = nodeName
	err := lw.client.Create(context.Background(), &lwcontroller)
	if err != nil {
		fmt.Printf("%s - [Lw] Create-Error: %s\n", time.Now(), err)
	}
	lw.Start(nodeName)
}

func (lw *Lw) Start(nodeName string) {
	duration := time.Duration(lw.checkInterval) * time.Millisecond
	ticker := time.NewTicker(duration)
	for {
		<-ticker.C
		gpuList := lw.checkGPU()
		oldLwcontroller := webappv1.Lwcontroller{}
		key := client.ObjectKey{
			Namespace: "lw",
			Name:      nodeName,
		}
		/*types.NamespacedName{
			Name: lw.nodeName,
		}*/
		err := lw.client.Get(context.Background(), key, &oldLwcontroller)
		if err != nil {
			fmt.Printf("%s - [Lw] Get--Error: %s\n", time.Now(), err)
			continue
		}
		if lw.toUpdate(oldLwcontroller.Status, *gpuList) {
			lwcontroller := oldLwcontroller.DeepCopy()
			lwcontroller.Status = webappv1.LwcontrollerStatus{
				GpuList:     *gpuList,
				TotalMemory: lw.TotalMemory,
				FreeMemory:  lw.FreeMemory,
				GpuNumber:   lw.lessGpuNumber,
				UpdateTime: &metav1.Time{
					Time: time.Now(),
				},
			}
			err := lw.client.Update(context.Background(), lwcontroller)
			if err != nil {
				fmt.Printf("%s - [Lw] Update-Error: %s\n", time.Now(), err)
			}
		}
	}
}

func (lw *Lw) checkGPU() *webappv1.GpuList {
	gpuList := make(webappv1.GpuList, 0)
	gpuNum := lw.calculateGPU()
	for i := int16(0); i < gpuNum; i++ {
		gpuList = append(gpuList, webappv1.Gpu{
			ID:          i,
			Health:      "Healthy",
			Power:       100,
			TotalMemory: 1000,
			Clock:       1000,
			FreeMemory:  1000,
			Core:        1,
			Bandwidth:   30,
		})
	}
	lw.lessGpuNumber = gpuNum
	sort.Sort(gpuList)
	total := uint64(0)
	free := uint64(0)
	for _, gpu := range gpuList {
		total += gpu.TotalMemory
		free += gpu.FreeMemory
	}
	lw.FreeMemory = free
	lw.TotalMemory = total
	return &gpuList
}

func (lw *Lw) calculateGPU() int16 {
	return 3
}

func (lw *Lw) toUpdate(status webappv1.LwcontrollerStatus, gpuList webappv1.GpuList) bool {
	if status.UpdateTime == nil || status.TotalMemory != lw.TotalMemory {
		return true
	}
	if status.FreeMemory != lw.FreeMemory || status.GpuNumber != lw.lessGpuNumber {
		return true
	}
	if !reflect.DeepEqual(status.GpuList, gpuList) {
		return true
	}
	return false
}
