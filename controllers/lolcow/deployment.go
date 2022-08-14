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
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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

// ensureDeployment ensures Deployment resource presence in given namespace.
func (r *LolcowReconciler) ensureDeployment(request reconcile.Request, instance *api.Lolcow, dep *appsv1.Deployment) (*reconcile.Result, error) {

	// See if deployment already exists and create if it doesn't
	found := &appsv1.Deployment{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{
		Name:      dep.Name,
		Namespace: instance.Namespace,
	}, found)
	if err != nil && errors.IsNotFound(err) {

		// Create the deployment
		err = r.Client.Create(context.TODO(), dep)

		if err != nil {
			// Deployment failed
			return &reconcile.Result{}, err
		} else {
			// Deployment was successful
			return nil, nil
		}
	} else if err != nil {
		// Error that isn't due to the deployment not existing
		return &reconcile.Result{}, err
	}

	return nil, nil
}

// backendDeployment is a code for Creating Deployment
func (r *LolcowReconciler) backendDeployment(v *api.Lolcow) *appsv1.Deployment {

	labels := labels(v, "backend")
	size := int32(1)
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "lolcow-pod",
			Namespace: v.Namespace,
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
						Image:           "paulbouwer/hello-kubernetes:1.10",
						ImagePullPolicy: corev1.PullAlways,
						Name:            "lolcow-pod",
						Ports: []corev1.ContainerPort{{
							ContainerPort: 8080,
							Name:          "lolcow",
						}},
					}},
				},
			},
		},
	}

	controllerutil.SetControllerReference(v, dep, r.Scheme)
	return dep
}
