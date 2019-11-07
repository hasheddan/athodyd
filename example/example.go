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
	crossplaneapis "github.com/crossplaneio/crossplane/apis"
	"github.com/crossplaneio/stack-gcp/apis"
	"github.com/crossplaneio/stack-gcp/pkg/controller"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// Adds GCP controllers to manager
// Copied from https://github.com/crossplaneio/stack-gcp/blob/master/cmd/stack/main.go
func controllerSetupWithManager(mgr manager.Manager) error {
	if err := (&controller.Controllers{}).SetupWithManager(mgr); err != nil {
		return err
	}

	return nil
}

// Adds GCP and Crossplane API types to scheme
// Copied from https://github.com/crossplaneio/stack-gcp/blob/master/cmd/stack/main.go
func addToScheme(scheme *runtime.Scheme) error {
	if err := apis.AddToScheme(scheme); err != nil {
		return err
	}

	if err := crossplaneapis.AddToScheme(scheme); err != nil {
		return err
	}

	return nil
}
