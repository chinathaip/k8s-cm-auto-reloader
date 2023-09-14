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

package controller

import (
	"context"
	"log"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

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

	log.Println("reloader status", reloader.Status.Type)
	if reloader.Status.Type == "" {
		reloader.Status.Type = statusCreated
		if err := r.Client.Status().Update(ctx, reloader); err != nil {
			log.Println(err, "unable to update Reloader")
			return ctrl.Result{}, err
		}
		log.Println("Reloader status updated to", reloader.Status.Type)
	}

	if reloader.Status.Type == statusReloaded {
		log.Println("Reloader already reloaded")

		cm := &corev1.ConfigMap{}
		if err := r.Get(ctx, types.NamespacedName{
			Name:      reloader.Spec.ConfigMap,
			Namespace: req.Namespace,
		}, cm); err != nil {
			log.Println(err, "could not find the specified configmap")
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}

		for k, v := range cm.Data {
			if reloader.Spec.Data[k] != v {
				log.Println("Configmap data has been changed")
				reloader.Status.Type = statusReloadRequired
				if err := r.Status().Update(ctx, reloader); err != nil {
					log.Println(err, "unable to update Reloader status")
					return ctrl.Result{}, err
				}
				return ctrl.Result{Requeue: true}, nil
			}
		}

		return ctrl.Result{RequeueAfter: time.Second * 10}, nil
	}

	if reloader.Status.Type == statusCreated {
		cm := &corev1.ConfigMap{}
		if err := r.Get(ctx, types.NamespacedName{
			Name:      reloader.Spec.ConfigMap,
			Namespace: req.Namespace,
		}, cm); err != nil {
			log.Println(err, "could not find the specified configmap")
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}

		reloader.Spec.Data = cm.Data
		if err := r.Update(ctx, reloader); err != nil {
			log.Println(err, "unable to update Reloader")
			return ctrl.Result{}, err
		}

		reloader.Status.Type = statusStored
		if err := r.Status().Update(ctx, reloader); err != nil {
			log.Println(err, "unable to update Reloader status")
			return ctrl.Result{}, err
		}

		log.Println("Reloader status updated to", reloader.Status.Type)
	}

	if reloader.Status.Type == statusStored {
		cm := &corev1.ConfigMap{}
		if err := r.Get(ctx, types.NamespacedName{
			Name:      reloader.Spec.ConfigMap,
			Namespace: req.Namespace,
		}, cm); err != nil {
			log.Println(err, "could not find the specified configmap")
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}

		if cm.Data == nil {
			log.Println("Configmap data is nil")
			return ctrl.Result{}, nil
		}

		for k, v := range cm.Data {
			if reloader.Spec.Data[k] != v {
				log.Println("Configmap data has been changed")
				reloader.Status.Type = statusReloadRequired
				if err := r.Status().Update(ctx, reloader); err != nil {
					log.Println(err, "unable to update Reloader status")
					return ctrl.Result{}, err
				}
				return ctrl.Result{Requeue: true}, nil
			}
		}
		log.Println("Reloader status updated to", reloader.Status.Type)
	}

	if reloader.Status.Type == statusReloadRequired {
		log.Println("In reload required")
		deploymentList := &appsv1.DeploymentList{}
		if err := r.Client.List(ctx, deploymentList, client.InNamespace(req.Namespace)); err != nil {
			log.Println(err, "unable to list deployments")
			return ctrl.Result{}, err
		}

		cm := &corev1.ConfigMap{}
		if err := r.Get(ctx, types.NamespacedName{
			Name:      reloader.Spec.ConfigMap,
			Namespace: req.Namespace,
		}, cm); err != nil {
			log.Println(err, "could not find the specified configmap")
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}

		reloader.Spec.Data = cm.Data
		if err := r.Update(ctx, reloader); err != nil {
			log.Println(err, "unable to update Reloader")
			return ctrl.Result{}, err
		}

		for _, deployment := range deploymentList.Items {
			use := useConfigMap(deployment.Spec.Template.Spec.Containers, reloader.Spec.ConfigMap)
			if use {
				if *deployment.Spec.Replicas != 0 {
					*deployment.Spec.Replicas = 0
					if err := r.Update(ctx, &deployment); err != nil {
						log.Println(err, "unable to update deployment", deployment.Name)
						return ctrl.Result{}, err
					}
					log.Println("set replicas to 0: ", deployment.Name)

				}

				if deployment.Status.Replicas == 0 {
					*deployment.Spec.Replicas = 3 //hardcoded
					if err := r.Update(ctx, &deployment); err != nil {
						log.Println(err, "unable to update deployment", deployment.Name)
						return ctrl.Result{}, err
					}
					log.Println("set replicas to 3: ", deployment.Name)
					reloader.Status.Type = statusReloaded
					if err := r.Status().Update(ctx, reloader); err != nil {
						log.Println(err, "unable to update Reloader status")
						return ctrl.Result{}, err
					}
				}
			} else {
				log.Println("Deployment", deployment.Name, "is not using configmap")
			}
		}
		log.Println("Reloader status updated to", reloader.Status.Type)
	}

	return ctrl.Result{Requeue: true}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ReloaderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&khingv1.Reloader{}).
		Complete(r)
}

func useConfigMap(containers []corev1.Container, ref string) bool {
	for _, c := range containers {
		for _, e := range c.Env {
			if e.ValueFrom.ConfigMapKeyRef.Name == ref {
				return true
			}
		}
	}
	return false
}
