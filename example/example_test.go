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
	"context"
	"testing"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/stack-gcp/apis/v1alpha3"
	"github.com/hasheddan/athodyd"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// TestThis tests this
func TestThis(t *testing.T) {
	tests := []athodyd.Test{
		{
			Name:        "TestCreateNamespaceSuccessful",
			Description: "This test checks to see if a namespace is created successfully.",
			Executor: func(c client.Client) error {
				n := &corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "cool-namespace",
					},
				}
				if err := c.Create(context.TODO(), n); err != nil {
					return err
				}

				return nil
			},
			Janitor: func(client.Client) error {
				return nil
			},
		},
		{
			Name:        "TestGCPProvider",
			Description: "This test checks to see if a GCP Provider can be created successfully.",
			Executor: func(c client.Client) error {
				p := &v1alpha3.Provider{
					ObjectMeta: metav1.ObjectMeta{
						Name: "yooooo",
					},
					Spec: v1alpha3.ProviderSpec{
						Secret: runtimev1alpha1.SecretKeySelector{
							Key: "credentials.json",
							SecretReference: runtimev1alpha1.SecretReference{
								Name:      "a-cool-secret",
								Namespace: "a-secret-namespace",
							},
						},
						ProjectID: "some-project",
					},
				}
				if err := c.Create(context.TODO(), p); err != nil {
					return err
				}

				return nil
			},
			Janitor: func(client.Client) error {
				return nil
			},
			Persist: true,
		},
		// Uncomment to view failure case
		// {
		// 	Name:        "TestCreateAnotherNamespace",
		// 	Description: "This test creates the same namespace as before.",
		// 	Executor: func(c client.Client) error {
		// 		n := &corev1.Namespace{
		// 			ObjectMeta: metav1.ObjectMeta{
		// 				Name: "cool-namespace",
		// 			},
		// 		}
		// 		if err := c.Create(context.TODO(), n); err != nil {
		// 			return err
		// 		}

		// 		return nil
		// 	},
		// 	Janitor: func(client.Client) error {
		// 		return nil
		// 	},
		// 	Persist: true,
		// },
		{
			Name:        "TestCreateYetAnotherNamespace",
			Description: "This test creates a different namespace.",
			Executor: func(c client.Client) error {
				n := &corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "keen-namespace",
					},
				}
				if err := c.Create(context.TODO(), n); err != nil {
					return err
				}

				return nil
			},
			Janitor: func(client.Client) error {
				return nil
			},
		},
	}

	job, err := athodyd.NewJob("MyExampleJob", "An example job for testing athodyd.", "./crds", tests, "30s", t, controllerSetupWithManager, addToScheme)
	if err != nil {
		t.Fatal(err)
	}

	if err := job.Run(); err != nil {
		t.Fatal(err)
	}
}
