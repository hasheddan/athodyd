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
	storagev1alpha3 "github.com/crossplaneio/stack-gcp/apis/storage/v1alpha3"
	"github.com/crossplaneio/stack-gcp/apis/v1alpha3"
	"github.com/hasheddan/athodyd"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// TestThis tests this
func TestThis(t *testing.T) {
	name := "MyExampleJob"
	description := "An example job for testing athodyd"

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

				return c.Create(context.TODO(), n)
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
						Name: "gcp-provider",
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

				return c.Create(context.TODO(), p)
			},
			Janitor: func(client.Client) error {
				return nil
			},
			Persist: true,
		},
		{
			Name:        "TestGCPBucketClassCreation",
			Description: "This test checks to see if a GCP BucketClass can be created successfully.",
			Executor: func(c client.Client) error {
				s := &storagev1alpha3.BucketClass{
					ObjectMeta: metav1.ObjectMeta{
						Name: "gcp-bucket",
					},
					SpecTemplate: storagev1alpha3.BucketClassSpecTemplate{
						ClassSpecTemplate: runtimev1alpha1.ClassSpecTemplate{
							WriteConnectionSecretsToNamespace: "cool-namespace",
							ProviderReference: &corev1.ObjectReference{
								Name: "gcp-provider",
							},
						},
						BucketParameters: storagev1alpha3.BucketParameters{
							BucketSpecAttrs: storagev1alpha3.BucketSpecAttrs{
								StorageClass: "REGIONAL",
							},
						},
					},
				}

				return c.Create(context.TODO(), s)
			},
			Janitor: func(client.Client) error {
				return nil
			},
			Persist: false,
		},
		{
			Name:        "TestCreateAnotherNamespace",
			Description: "This test creates a different namespace.",
			Executor: func(c client.Client) error {
				n := &corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "keen-namespace",
					},
				}

				return c.Create(context.TODO(), n)
			},
			Janitor: func(client.Client) error {
				return nil
			},
			Persist: true,
		},
	}

	cfg, err := config.GetConfig()
	if err != nil {
		t.Fatal(err)
	}

	cleaner := func(c client.Client) error {
		if err := c.DeleteAllOf(context.TODO(), &v1alpha3.Provider{}); err != nil {
			return err
		}

		if err := c.DeleteAllOf(context.TODO(), &storagev1alpha3.BucketClass{}); err != nil {
			return err
		}

		return nil
	}

	job := athodyd.NewJob(name, description, tests, t,
		athodyd.WithCluster(cfg),
		athodyd.WithCRDDirectoryPaths([]string{"./crds"}),
		athodyd.WithSetupWithManager(controllerSetupWithManager),
		athodyd.WithAddToScheme(addToScheme),
		athodyd.WithCleaner(cleaner),
	)

	if err := job.Run(); err != nil {
		t.Fatal(err)
	}
}
