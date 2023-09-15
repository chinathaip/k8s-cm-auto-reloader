/*
Copyright 2023.

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

package reloader

import (
	"context"
	"log"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	khingv1 "github.com/chinathaip/k8s-cm-auto-reloader/api/v1"
)

const (
	statusCreated        = "created"
	statusStored         = "stored"
	statusReloadRequired = "reloadRequired"
	statusReloaded       = "reloaded"
)

// ReloaderReconciler reconciles a Reloader object
type ReloaderReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups="apps",resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=khing.khing.k8s,resources=reloaders;configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=khing.khing.k8s,resources=reloaders/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=khing.khing.k8s,resources=reloaders/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Reloader object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.0/pkg/reconcile
func (r *ReloaderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reloader := &khingv1.Reloader{}
	if err := r.Client.Get(ctx, req.NamespacedName, reloader); err != nil {
		log.Println(err, "unable to fetch Reloader")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	switch reloader.Status.Type {
	case statusStored:
		return StatusStored{r.Client}.Reconcile(ctx, req)
	case statusReloadRequired:
		return StatusReloadRequired{r.Client}.Reconcile(ctx, req)
	default:
		return StatusCreated{r.Client}.Reconcile(ctx, req)
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *ReloaderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&khingv1.Reloader{}).
		Complete(r)
}

func NewReconciler(c client.Client, s *runtime.Scheme) *ReloaderReconciler {
	return &ReloaderReconciler{Client: c, Scheme: s}
}
