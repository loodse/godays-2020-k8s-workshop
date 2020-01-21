package podhealth

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	trainingv1alpha1 "github.com/loodse/operator-workshop/podhealth/operatorsdk/pkg/apis/training/v1alpha1"
)

var log = logf.Log.WithName("controller_podhealth")

// Add creates a new PodHealth Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcilePodHealth{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("podhealth-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource PodHealth
	err = c.Watch(&source.Kind{Type: &trainingv1alpha1.PodHealth{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

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

	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, enqueueAllPodHealthsInNamespace)
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcilePodHealth implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcilePodHealth{}

// ReconcilePodHealth reconciles a PodHealth object
type ReconcilePodHealth struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a PodHealth object and makes changes based on the state read
// and what is in the PodHealth.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcilePodHealth) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling PodHealth")

	// Fetch the PodHealth instance
	instance := &trainingv1alpha1.PodHealth{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// List Pods
	podSelector, err := metav1.LabelSelectorAsSelector(&instance.Spec.PodSelector)
	if err != nil {
		reqLogger.Error(err, "invalid podSelector")
		// don't return an error here, because we don't want to retry
		return reconcile.Result{}, nil
	}
	ctx := context.Background()
	podList := &corev1.PodList{}
	if err = r.client.List(ctx, podList,
		client.InNamespace(instance.Namespace),
		client.MatchingLabelsSelector{Selector: podSelector}); err != nil {
		return reconcile.Result{}, fmt.Errorf("listing pods: %v", err)
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
	instance.Status.Total = len(podList.Items)
	instance.Status.Ready = ready
	instance.Status.Unready = unready
	instance.Status.LastChecked = metav1.Now()
	if err = r.client.Status().Update(ctx, instance); err != nil {
		return reconcile.Result{}, fmt.Errorf("updating PodHealth Status: %v", err)
	}
	return reconcile.Result{}, nil
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
