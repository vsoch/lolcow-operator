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
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logctrl "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
	api "vsoch/lolcow-operator/api/lolcow/v1alpha1"
	"vsoch/lolcow-operator/pkg/lolcow"
)

// LolcowReconciler reconciles a Lolcow object
type LolcowReconciler struct {
	Client client.Client
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

	// TODO this should have opts?
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

	// Prepare to look for existing service and pod (lolcow)
	existingLolcow := &appsv1.Deployment{}
	existingService := &corev1.Service{}

	// Prepare a logger to communicate to the developer user
	// Note that we could attach a named logger to the reconciler object,
	// and that might be a better approach for organization or state
	// https://github.com/kubernetes-sigs/kueue/blob/main/pkg/controller/core/queue_controller.go#L50
	log := logctrl.FromContext(ctx).WithValues("Lolcow", req.NamespacedName)

	// Keep developed informed what is going on.
	log.Info("⚡️ Event received! ⚡️")
	log.Info("Request: ", "req", req)

	// Do we have a current lolcow instance deployed?
	err := r.Client.Get(ctx, req.NamespacedName, &instance)
	if err != nil {

		// Case 1: not found yet, check if deployment needs deletion
		if errors.IsNotFound(err) {

			// Ensure deployment and service
			err := r.EnsureDeployment(ctx, req, &instance, existingLolcow)
			if err != nil {
				return ctrl.Result{}, err
			}
			err = r.EnsureService(ctx, req, &instance, existingService)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {

		log.Info("ℹ️  State ℹ️", "Lolcow.Name", instance.Name, " Lolcow.Namespace", instance.Namespace, "Lolcow.Spec.Port", instance.Spec.Port, "Lolcow.Spec.Greeting", instance.Spec.Greeting)

		// Check if the deployment already exists, if not: create a new one.
		err = r.Client.Get(ctx, types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, existingLolcow)
		if err != nil && errors.IsNotFound(err) {
			// Define a new deployment
			newLolcow := r.createDeployment(&instance)
			log.Info("✨ Creating a new Deployment", "Namespace", newLolcow.Namespace, "Name", newLolcow.Name)

			err = r.Client.Create(ctx, newLolcow)
			if err != nil {
				log.Error(err, "❌ Failed to create new Deployment", "Namespace", newLolcow.Namespace, "Name", newLolcow.Name)
				return ctrl.Result{}, err
			}

			//	 	} else if err != nil {
			//			err = r.Update(ctx, existingLolcow)
			//			if err != nil {
			//			log.Error(err, "❌ Failed to update Deployment", "Namespace", existingLolcow.Namespace, "Name", existing.Name)
			//			return ctrl.Result{}, err

		} else if err != nil {
			log.Error(err, "Failed to get Deployment")
			return ctrl.Result{}, err
		}

		// Check if the service already exists, if not: create a new one
		err = r.Client.Get(ctx, types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, existingService)
		if err != nil && errors.IsNotFound(err) {
			// Create the Service
			newService := r.createService(&instance)
			log.Info("✨ Creating a new Service", "Namespace", instance.Namespace, "Name", instance.Name, "Port", newService.Spec.Ports[0].Port)
			err = r.Client.Create(ctx, newService)
			if err != nil {
				log.Error(err, "❌ Failed to create new Service", "Namespace", newService.Namespace, "Name", newService.Name)
				return ctrl.Result{}, err
			}
		} else if err == nil {
			// Service exists, check if the port has to be updated.
			var port int32 = instance.Spec.Port
			if existingService.Spec.Ports[0].Port != port {
				log.Info("🔁 Port number changes, update the service! 🔁")
				existingService.Spec.Ports[0].Port = port
				err = r.Client.Update(ctx, existingService)
				if err != nil {
					log.Error(err, "❌ Failed to update Service", "Namespace", existingService.Namespace, "Name", existingService.Name)
					return ctrl.Result{}, err
				}
			}
		} else if err != nil {
			log.Error(err, "Failed to get Service")
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *LolcowReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&api.Lolcow{}).
		Watches(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForObject{}, builder.WithPredicates(predicate.Funcs{
			// Check only delete events for a service
			UpdateFunc: func(e event.UpdateEvent) bool {
				return false
			},
			CreateFunc: func(e event.CreateEvent) bool {
				return false
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				return true
			},
		})).
		Complete(r)
}
