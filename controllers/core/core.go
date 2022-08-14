/*
Copyright 2022 The Kubernetes Authors.

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

package core

import (
	ctrl "sigs.k8s.io/controller-runtime"
)

const updateChBuffer = 10

// SetupControllers sets up all controllers.
func SetupControllers(mgr ctrl.Manager) (string, error) {

	// TODO: we can setup other controllers here
	// For now this does nothing :)
	//recon := NewLolcowReconciler(mgr.GetClient(), mgr.GetScheme())
	//if err := recon.SetupWithManager(mgr); err != nil {
	//	return "Lolcow", err
	//}
	return "", nil
}
