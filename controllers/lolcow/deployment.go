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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	logctrl "sigs.k8s.io/controller-runtime/pkg/log"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	api "vsoch/lolcow-operator/api/lolcow/v1alpha1"
)

// labels fetches and sets labels
func labels(v *api.Lolcow, tier string) map[string]string {
	return map[string]string{
		"app":             "lolcow",
		"visitorssite_cr": v.Name,
		"tier":            tier,
	}
}

// EnsureDeployment ensures Deployment resource presence in given namespace.
func (r *LolcowReconciler) EnsureDeployment(ctx context.Context, request reconcile.Request, instance *api.Lolcow, existing *appsv1.Deployment) error {
	log := logctrl.FromContext(ctx).WithValues("LolcowEnsureDeployment", request.NamespacedName)
	log.Info("Lolcow deployment not found, checking if a deployment must be deleted.")
	err := r.Client.Get(ctx, types.NamespacedName{Name: request.Name, Namespace: request.Namespace}, existing)

	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Nothing to do, no deployment found.")
			return nil
		}
		log.Error(err, "❌ Failed to get Deployment")
		return err
	}
	log.Info("☠️ Deployment exists: delete it. ☠️")
	r.Client.Delete(ctx, existing)
	return nil
}

// Create a Deployment for the Nginx server.
func (r *LolcowReconciler) createDeployment(instance *api.Lolcow) *appsv1.Deployment {
	size := int32(1)
	labels := labels(instance, "backend")
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "lolcow-pod",
			Namespace: instance.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &size,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:           "ghcr.io/vsoch/lolcow-operator:latest",
						ImagePullPolicy: corev1.PullAlways,
						Name:            "lolcow-pod",
						Args:            []string{"/bin/bash", "/entrypoint.sh", r.Greeter.Greeting},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 8080,
							Name:          "lolcow",
						}},
					}},
				},
			},
		},
	}
	return deployment
}
