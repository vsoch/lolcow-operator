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
	api "vsoch/lolcow-operator/api/lolcow/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	logctrl "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// EnsureService ensures the service is running
func (r *LolcowReconciler) EnsureService(ctx context.Context, request reconcile.Request, instance *api.Lolcow, existing *corev1.Service) error {

	log := logctrl.FromContext(ctx).WithValues("LolcowEnsureService", request.NamespacedName)
	err := r.Client.Get(ctx, types.NamespacedName{Name: request.Name, Namespace: request.Namespace}, existing)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Nothing to do, no service found.")
			return nil
		}
		log.Error(err, "❌ Failed to get Service")
		return err
	}
	log.Info("☠️ Service exists: delete it. ☠️")
	r.Client.Delete(ctx, existing)
	return nil
}

// createService creates a backend service
func (r *LolcowReconciler) createService(instance *api.Lolcow) *corev1.Service {

	labels := labels(instance, "backend")

	// We shouldn't need this, as the port comes from the manifest
	if instance.Spec.Port == 0 {
		instance.Spec.Port = 30685
	}
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "backend-service",
			Namespace: instance.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Protocol:   corev1.ProtocolTCP,
					Port:       80,
					TargetPort: intstr.FromInt(8080),
					NodePort:   instance.Spec.Port,
				},
			},
			Type: corev1.ServiceTypeLoadBalancer,
		},
	}

	return service
}
