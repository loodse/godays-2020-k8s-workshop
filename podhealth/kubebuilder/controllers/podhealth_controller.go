/*

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

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	trainingv1alpha1 "github.com/loodse/operator-workshop/podhealth/kubebuilder/api/v1alpha1"
)

// PodHealthReconciler reconciles a PodHealth object
type PodHealthReconciler struct {
	client.Client
	Log logr.Logger
}

// +kubebuilder:rbac:groups=training.loodse.io,resources=podhealths,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods,verbs=list;watch
// +kubebuilder:rbac:groups=training.loodse.io,resources=podhealths/status,verbs=get;update;patch

func (r *PodHealthReconciler) Reconcile(req ctrl.Request) (result ctrl.Result, err error) {
	ctx := context.Background()
	log := r.Log.WithValues("podhealth", req.NamespacedName)

	// Get current State
	podHealth := &trainingv1alpha1.PodHealth{}
	if err = r.Get(ctx, req.NamespacedName, podHealth); err != nil {
		return result, client.IgnoreNotFound(err)
	}

	// List Pods
	podSelector, err := metav1.LabelSelectorAsSelector(&podHealth.Spec.PodSelector)
	if err != nil {
		log.Error(err, "invalid podSelector")
		// don't return an error here, because we don't want to retry
		return result, nil
	}
	podList := &corev1.PodList{}
	if err = r.List(ctx, podList,
		client.InNamespace(podHealth.Namespace),
		client.MatchingLabelsSelector{Selector: podSelector}); err != nil {
		return result, fmt.Errorf("listing pods: %v", err)
	}

	// Count ready/unready
	var (
		ready   int
		unready int
	)
	for _, pod := range podList.Items {
		if isReady(&pod) {
			ready++
			continue
		}
		unready++
	}

	// Update PodHealth Status
	podHealth.Status.Total = len(podList.Items)
	podHealth.Status.Ready = ready
	podHealth.Status.Unready = unready
	podHealth.Status.LastChecked = metav1.Now()
	if err = r.Status().Update(ctx, podHealth); err != nil {
		return result, fmt.Errorf("updating PodHealth Status: %v", err)
	}

	return
}

func isReady(pod *corev1.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady &&
			condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func (r *PodHealthReconciler) SetupWithManager(mgr ctrl.Manager) error {
	enqueueAllPodHealthsInNamespace := &handler.EnqueueRequestsFromMapFunc{
		ToRequests: handler.ToRequestsFunc(func(obj handler.MapObject) (requests []reconcile.Request) {
			ctx := context.Background()

			// List all PodHealth objects in the same Namespace
			podHealthList := &trainingv1alpha1.PodHealthList{}
			if err := mgr.GetClient().List(ctx, podHealthList, client.InNamespace(obj.Meta.GetNamespace())); err != nil {
				utilruntime.HandleError(err)
				return requests
			}

			// Add all PodHealth objects to the workqueue
			for _, podHealth := range podHealthList.Items {
				requests = append(requests, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      podHealth.Name,
						Namespace: podHealth.Namespace,
					},
				})
			}

			return requests
		}),
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&trainingv1alpha1.PodHealth{}).
		Watches(&source.Kind{Type: &corev1.Pod{}}, enqueueAllPodHealthsInNamespace).
		Complete(r)
}
