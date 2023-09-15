package reloader

import (
	"context"
	"log"
	"time"

	khingv1 "github.com/chinathaip/k8s-cm-auto-reloader/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Status interface {
	Reconcile(context.Context, ctrl.Request) (ctrl.Result, error)
}

type StatusCreated struct {
	c client.Client
}

func (s StatusCreated) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log.Println("in Created")
	reloader := &khingv1.Reloader{}
	if err := s.c.Get(ctx, req.NamespacedName, reloader); err != nil {
		log.Println(err, "unable to fetch Reloader")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	cm := &corev1.ConfigMap{}
	if err := s.c.Get(ctx, types.NamespacedName{
		Name:      reloader.Spec.ConfigMap,
		Namespace: req.Namespace,
	}, cm); err != nil {
		log.Println(err, "could not find the specified configmap")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	reloader.Spec.Data = cm.Data
	if err := s.c.Update(ctx, reloader); err != nil {
		log.Println(err, "unable to update Reloader")
		return ctrl.Result{}, err
	}

	reloader.Status.Type = statusStored
	if err := s.c.Status().Update(ctx, reloader); err != nil {
		log.Println(err, "unable to update Reloader status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: time.Second * 10}, nil
}

type StatusStored struct {
	c client.Client
}

func (s StatusStored) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log.Println("in Stored")
	reloader := &khingv1.Reloader{}
	if err := s.c.Get(ctx, req.NamespacedName, reloader); err != nil {
		log.Println(err, "unable to fetch Reloader")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	cm := &corev1.ConfigMap{}
	if err := s.c.Get(ctx, types.NamespacedName{
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
		log.Println("Checking configmap data")
		if reloader.Spec.Data[k] != v {
			log.Println("Configmap data has been changed")
			reloader.Status.Type = statusReloadRequired
			if err := s.c.Status().Update(ctx, reloader); err != nil {
				log.Println(err, "unable to update Reloader status")
				return ctrl.Result{}, err
			}
			return StatusReloadRequired{s.c}.Reconcile(ctx, req)
		}
	}

	return ctrl.Result{RequeueAfter: time.Second * 10}, nil
}

type StatusReloadRequired struct {
	c client.Client
}

func (s StatusReloadRequired) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log.Println("in ReloadRequired")
	reloader := &khingv1.Reloader{}
	if err := s.c.Get(ctx, req.NamespacedName, reloader); err != nil {
		log.Println(err, "unable to fetch Reloader")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	deploymentList := &appsv1.DeploymentList{}
	if err := s.c.List(ctx, deploymentList, client.InNamespace(req.Namespace)); err != nil {
		log.Println(err, "unable to list deployments")
		return ctrl.Result{}, err
	}

	cm := &corev1.ConfigMap{}
	if err := s.c.Get(ctx, types.NamespacedName{
		Name:      reloader.Spec.ConfigMap,
		Namespace: req.Namespace,
	}, cm); err != nil {
		log.Println(err, "could not find the specified configmap")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	for _, deployment := range deploymentList.Items {
		if useConfigMap(deployment.Spec.Template.Spec.Containers, reloader.Spec.ConfigMap) {
			if *deployment.Spec.Replicas != 0 {
				*deployment.Spec.Replicas = 0
				if err := s.c.Update(ctx, &deployment); err != nil {
					log.Println(err, "unable to update deployment", deployment.Name)
					return ctrl.Result{}, err
				}
				log.Println("set replicas to 0: ", deployment.Name)
				continue
			}

			if deployment.Status.Replicas == 0 {
				*deployment.Spec.Replicas = 3 //hardcoded
				if err := s.c.Update(ctx, &deployment); err != nil {
					log.Println(err, "unable to update deployment", deployment.Name)
					return ctrl.Result{}, err
				}
				log.Println("set replicas to 3: ", deployment.Name)
				reloader.Status.Type = statusCreated
				if err := s.c.Status().Update(ctx, reloader); err != nil {
					log.Println(err, "unable to update Reloader status")
					return ctrl.Result{}, err
				}
			}
		}
	}
	return ctrl.Result{RequeueAfter: time.Second * 10}, nil
}
