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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	smarthomev1alpha1 "github.com/loodse/godays-2020-k8s-workshop/smart-home/api/v1alpha1"
	"github.com/loodse/godays-2020-k8s-workshop/smart-home/pkg/smarthome"
)

// ShutterReconciler reconciles a Shutter object
type ShutterReconciler struct {
	client.Client
	Log             logr.Logger
	SmartHomeClient *smarthome.Client
}

// +kubebuilder:rbac:groups=smarthome.loodse.io,resources=shutters,verbs=get;list;watch;create;update
// +kubebuilder:rbac:groups=smarthome.loodse.io,resources=shutters/status,verbs=get;update;patch

func (r *ShutterReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	var (
		ctx    = context.Background()
		result ctrl.Result
		_      = r.Log.WithValues("shutter", req.NamespacedName)
	)

	// Load Shutter instance from cache.
	shutter := &smarthomev1alpha1.Shutter{}
	if err := r.Get(ctx, req.NamespacedName, shutter); err != nil {
		return result, client.IgnoreNotFound(err)
	}

	// Just update the Shutter - it will not move when it's already in position
	// If you have a LOT of shutters and want to save network bandwith,
	// you can also check the state of the shutter first.
	if err := r.SmartHomeClient.Shutters().Set(ctx, req.NamespacedName.String(), shutter.Spec.ClosedPercentage); err != nil {
		return result, fmt.Errorf("updating shutter: %v", err)
	}

	state, err := r.SmartHomeClient.Shutters().Get(ctx, req.NamespacedName.String())
	if err != nil {
		return result, fmt.Errorf("checking shutter state: %v", err)
	}

	// Update the Status of the shutter, to tell the rest of the system what is going on.
	shutter.Status.ObservedGeneration = shutter.Generation
	shutter.Status.ClosedPercentage = state.Current
	if state.Moving {
		shutter.Status.Phase = smarthomev1alpha1.ShutterMoving
	} else {
		shutter.Status.Phase = smarthomev1alpha1.ShutterIdle
	}
	if err := r.Client.Status().Update(ctx, shutter); err != nil {
		return result, fmt.Errorf("updating shutter status: %v", err)
	}

	// When the Shutter is not at target position, requeue this entry to check again.
	if state.Current != state.Target {
		result.Requeue = true
		return result, nil
	}

	return result, nil
}

func (r *ShutterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&smarthomev1alpha1.Shutter{}).
		Complete(r)
}
