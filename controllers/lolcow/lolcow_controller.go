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

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"vsoch/lolcow-operator/pkg/lolcow"

	api "vsoch/lolcow-operator/api/lolcow/v1alpha1"
)

// LolcowReconciler reconciles a Lolcow object
type LolcowReconciler struct {
	Client client.Client
	Scheme *runtime.Scheme

	// An added "greeter" to hold the greeting
	Greeter *lolcow.Greeter
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

	// Prepare a logger to communicate to the developer user
	// Note that we could attach a named logger to the reconciler object,
	// and that might be a better approach for organization or state
	// https://github.com/kubernetes-sigs/kueue/blob/main/pkg/controller/core/queue_controller.go#L50
	log := log.FromContext(ctx).WithValues("Lolcow", req.NamespacedName)

	// Do we have a current lolcow instance deployed?

	if err := r.Client.Get(ctx, req.NamespacedName, &instance); err != nil {

		// Check if this Service already exists
		result, err := r.ensureService(req, &instance, r.backendService(&instance))
		if result != nil {
			log.Error(err, "Service Not ready")
			return *result, err
		}

		// ctrl.Result is a reconcile.Result
		// IgnoreNotFound will return nil in the case of not found (don't re-queue, nothing to do)
		// and the error otherwise (so we DO want to re-queue)
		// https://github.com/kubernetes-sigs/controller-runtime/blob/master/pkg/client/interfaces.go#L140
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Check if this Deployment already exists (is found)
	var deployment appsv1.Deployment
	err := r.Client.Get(ctx, types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, &deployment)
	var result *ctrl.Result
	result, err = r.ensureDeployment(req, &instance, r.backendDeployment(&instance))
	if result != nil {
		log.Error(err, "Deployment Not ready")
		return *result, err
	}

	// Check if this Service already exists
	result, err = r.ensureService(req, &instance, r.backendService(&instance))
	if result != nil {
		log.Error(err, "Service Not ready")
		return *result, err
	}

	// If we get here, we have an lolcow object!
	// The only thing we care about is the greeting! Did it change?
	log.Info("Spec greeting: " + instance.Spec.Greeting)
	log.Info("Request greeting: " + r.Greeter.Greeting)
	if instance.Spec.Greeting != r.Greeter.Greeting {

		// Change state of our reconciler object
		r.Greeter.Greeting = instance.Spec.Greeting
		err := r.Client.Status().Update(ctx, &instance)
		if err != nil {
			log.Error(err, "Error updating instance.")
		}

		// Update the deployment cmd provided
		updatedDeployment := deployment.DeepCopy()
		updatedDeployment.Spec.Template.Spec = r.NewContainer(instance.Spec.Greeting)

		// Patch the deploymemt
		if err := r.Client.Patch(ctx, updatedDeployment, client.StrategicMergeFrom(&deployment)); err != nil {
			log.Error(err, "Unable to patch Deployment")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *LolcowReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&api.Lolcow{}).
		Complete(r)
}
