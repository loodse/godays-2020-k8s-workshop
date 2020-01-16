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

package main

import (
	"flag"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"

	smarthomev1alpha1 "github.com/loodse/godays-2020-k8s-workshop/smart-home/api/v1alpha1"
	"github.com/loodse/godays-2020-k8s-workshop/smart-home/controllers"
	"github.com/loodse/godays-2020-k8s-workshop/smart-home/pkg/smarthome"
	"github.com/loodse/godays-2020-k8s-workshop/smart-home/pkg/ui"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = smarthomev1alpha1.AddToScheme(scheme)
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	smartHomeClient := smarthome.NewClient()
	u := ui.NewUI(smartHomeClient, 1*time.Second)

	ctrl.SetLogger(u.Logger())

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		LeaderElection:     enableLeaderElection,
		Port:               9443,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.ShutterReconciler{
		Client:          mgr.GetClient(),
		Log:             ctrl.Log.WithName("controllers").WithName("Shutter"),
		SmartHomeClient: smartHomeClient,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Shutter")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	go u.Run()

	setupLog.Info("starting manager")
	if err := mgr.Start(u.CloseCh()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
