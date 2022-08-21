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

// TODO vsoch I like this design better
// https://github.com/GoogleCloudPlatform/airflow-operator/blob/master/pkg/controller/controller.go

package controllers

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	logctrl "sigs.k8s.io/controller-runtime/pkg/log"
	api "vsoch/lolcow-operator/api/lolcow/v1alpha1"
	"vsoch/lolcow-operator/pkg/lolcow"
)

// LolcowReconciler reconciles a Lolcow object
type LolcowReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	// An added "greeter" to hold the greeting
	Greeter *lolcow.Greeter
}

// GenericResource have both meta and runtime interfaces
type LolcowResources interface {
	metav1.Object
	runtime.Object
}

// NewLolcowReconciler returns the Lolcow Reconciler to the core controller
func NewLolcowReconciler(client client.Client, scheme *runtime.Scheme, greeter *lolcow.Greeter) *LolcowReconciler {

	// TODO Could this have options (ops) instead of hard coding the specific params (e.g., Greeting)
	// https://github.com/kubernetes-sigs/kueue/blob/47ec7d6033ae7527b5495ed432ae4390fc052523/pkg/controller/workload/job/job_controller.go#L78
	return &LolcowReconciler{
		Client:  client,
		Scheme:  scheme,
		Greeter: greeter,
	}
}

//+kubebuilder:rbac:groups=my.domain,resources=lolcows,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=my.domain,resources=lolcows/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=my.domain,resources=lolcows/finalizers,verbs=update
//+kubebuilder:rbac:groups=my.domain,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=my.domain,resources=pods,verbs=get;list;watch;create;
//+kubebuilder:rbac:groups=my.domain,resources=services,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// 1. Compare the state specified by the Lolcow instance against actual cluster state
// 2. Does it not exist? Oh well, nothing to do.
// 3. Is it different? Perform operations to reflect new user preferences
// 4. Is it the same? No changes needed.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *LolcowReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	// The Lolcow instance!
	var instance api.Lolcow

	// Prepare a logger to communicate to the developer user
	// Note that we could attach a named logger to the reconciler object,
	// and that might be a better approach for organization or state
	// https://github.com/kubernetes-sigs/kueue/blob/main/pkg/controller/core/queue_controller.go#L50
	log := logctrl.FromContext(ctx).WithValues("Lolcow", req.NamespacedName)

	// Keep developed informed what is going on.
	log.Info("⚡️ Event received! ⚡️")
	log.Info("Request: ", "req", req)

	err := r.Get(ctx, req.NamespacedName, &instance)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Lolcow resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		log.Info("Failed to get Lolcow resource. Re-running reconcile.")
		return ctrl.Result{}, err
	}
	log.Info("🥑️ Found instance 🥑️", instance.Spec.Greeting, instance.Spec.Port)

	// Do we have a current lolcow instance deployed?
	existingD := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, existingD)
	if err != nil {

		// Case 1: not found yet, check if deployment needs deletion
		if errors.IsNotFound(err) {
			dep := r.createDeployment(&instance)
			log.Info("✨ Creating a new Deployment ✨", "Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			err = r.Create(ctx, dep)
			if err != nil {
				log.Error(err, "❌ Failed to create new Deployment", "Namespace", dep.Namespace, "Name", dep.Name)
				return ctrl.Result{}, err
			}
			// Deployment created successfully - return and requeue
			return ctrl.Result{Requeue: true}, nil
		} else if err != nil {
			log.Error(err, "Failed to get Deployment")
			return ctrl.Result{}, err
		}
	}

	// If we get down here, we have a current instance deployed (not freshly created)
	greeting := instance.Spec.Greeting

	// The third / last entry in the command is the greeting
	existingGreeting := existingD.Spec.Template.Spec.Containers[0].Command[2]
	log.Info("Requested vs existing greeting : ", greeting, existingGreeting)

	// Do we need to update the greeting?
	if existingGreeting != greeting {

		log.Info("👋️ New Greeting! 👋️: ", greeting, existingGreeting)

		// /bin/bash /entrypoint.sh <greeting>
		existingD.Spec.Template.Spec.Containers[0].Command[2] = greeting
		err = r.Update(ctx, existingD)
		if err != nil {
			log.Error(err, "Failed to update Deployment", "Namespace", existingD.Namespace, "Name", existingD.Name)
			return ctrl.Result{}, err
		}
		// Deployment updated - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else {
		log.Info("👋️ No Change to Greeting! 👋️: ", greeting, existingGreeting)
	}

	// Do we have a current service deployed?
	existingS := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, existingS)
	if err != nil {

		// Case 1: not found yet, check if deployment needs deletion
		if errors.IsNotFound(err) {
			service := r.createService(&instance)
			log.Info("✨ Creating a new Service ✨", "Namespace", service.Namespace, "Name", service.Name)
			err = r.Create(ctx, service)
			if err != nil {
				log.Error(err, "❌ Failed to create new Service", "Namespace", service.Namespace, "Name", service.Name)
				return ctrl.Result{}, err
			}
			// Service created successfully - return and requeue
			return ctrl.Result{Requeue: true}, nil
		} else if err != nil {
			log.Error(err, "Failed to get Service")
			return ctrl.Result{}, err
		}

		// We found the service, but did the port change (and needs to be updated?)
	} else {

		var port int32 = instance.Spec.Port
		existingPort := existingS.Spec.Ports[0].NodePort

		// Do we need to update the port
		if port != existingPort {
			log.Info("🔁 New Port! 🔁")
			existingS.Spec.Ports[0].NodePort = port
			err = r.Update(ctx, existingS)
			if err != nil {
				log.Error(err, "Failed to update Service", "Namespace", existingS.Namespace, "Name", existingS.Name)
				return ctrl.Result{}, err
			}
			// Service updated - return and requeue
			return ctrl.Result{Requeue: true}, nil
		} else {
			log.Info("🔁 No Change to Port! 🔁")
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *LolcowReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&api.Lolcow{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		// Defaults to 1, putting here so we know it exists!
		WithOptions(controller.Options{MaxConcurrentReconciles: 1}).
		Complete(r)
}
