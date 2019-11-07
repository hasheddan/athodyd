/*
Copyright 2019 The Athodyd Authors.

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

package example

import (
	"log"

	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// C just is
type C struct{}

// SetupWithManager just is
func (c *C) SetupWithManager(m manager.Manager) error {
	r := &R{}
	return ctrl.NewControllerManagedBy(m).Named("test").For(&corev1.Service{}).Complete(r)
}

// R just is
type R struct{}

// Reconcile likes to
func (r *R) Reconcile(o reconcile.Request) (reconcile.Result, error) {
	// Your business logic to implement the API by creating, updating, deleting objects goes here.
	log.Print("here")
	return reconcile.Result{}, nil
}
