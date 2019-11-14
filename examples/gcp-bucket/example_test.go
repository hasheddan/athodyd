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
	"time"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	storagev1alpha1 "github.com/crossplaneio/crossplane/apis/storage/v1alpha1"
	storagev1alpha3 "github.com/crossplaneio/stack-gcp/apis/storage/v1alpha3"
	"github.com/crossplaneio/stack-gcp/apis/v1alpha3"
	"github.com/hasheddan/athodyd"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// TestThis tests this
func TestThis(t *testing.T) {
	cases := map[string]struct {
		reason string
		test   func(c client.Client) error
	}{
		"CreateProvider": {
			reason: "A GCP Provider should be created without error.",
			test: func(c client.Client) error {
				p := &v1alpha3.Provider{
					ObjectMeta: metav1.ObjectMeta{
						Name: "gcp-provider",
					},
					Spec: v1alpha3.ProviderSpec{
						Secret: runtimev1alpha1.SecretKeySelector{
							Key: "credentials.json",
							SecretReference: runtimev1alpha1.SecretReference{
								Name:      "example-provider-gcp",
								Namespace: "crossplane-system",
							},
						},
						ProjectID: "crossplane-playground",
					},
				}

				defer func() {
					if err := c.Delete(context.Background(), p); err != nil {
						t.Error(err)
					}
				}()

				return c.Create(context.Background(), p)
			},
		},
		"DynamicallyProvisionGCPBucket": {
			reason: "A GCP Bucket should be provisioned without error.",
			test: func(c client.Client) error {
				n := &corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "cool-namespace",
					},
				}

				p := &v1alpha3.Provider{
					ObjectMeta: metav1.ObjectMeta{
						Name: "gcp-provider",
					},
					Spec: v1alpha3.ProviderSpec{
						Secret: runtimev1alpha1.SecretKeySelector{
							Key: "credentials.json",
							SecretReference: runtimev1alpha1.SecretReference{
								Name:      "example-provider-gcp",
								Namespace: "crossplane-system",
							},
						},
						ProjectID: "crossplane-playground",
					},
				}

				s := &storagev1alpha3.BucketClass{
					ObjectMeta: metav1.ObjectMeta{
						Name: "gcp-bucket",
						Labels: map[string]string{
							"provider": "gcp",
						},
					},
					SpecTemplate: storagev1alpha3.BucketClassSpecTemplate{
						ClassSpecTemplate: runtimev1alpha1.ClassSpecTemplate{
							WriteConnectionSecretsToNamespace: "cool-namespace",
							ProviderReference: &corev1.ObjectReference{
								Name: "gcp-provider",
							},
							ReclaimPolicy: runtimev1alpha1.ReclaimDelete,
						},
						BucketParameters: storagev1alpha3.BucketParameters{
							BucketSpecAttrs: storagev1alpha3.BucketSpecAttrs{
								StorageClass: "MULTI_REGIONAL",
								Location:     "US",
							},
						},
					},
				}

				cl := &storagev1alpha1.Bucket{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "gcp-claim",
						Namespace: "cool-namespace",
					},
					Spec: storagev1alpha1.BucketSpec{
						ResourceClaimSpec: runtimev1alpha1.ResourceClaimSpec{
							ClassSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"provider": "gcp",
								},
							},
						},
					},
				}

				defer func() {
					if err := c.Delete(context.Background(), cl); err != nil {
						t.Error(err)
					}

					if err := c.Delete(context.Background(), s); err != nil {
						t.Error(err)
					}

					if err := c.Delete(context.Background(), p); err != nil {
						t.Error(err)
					}

					if err := c.Delete(context.Background(), n); err != nil {
						t.Error(err)
					}
				}()

				if err := c.Create(context.Background(), n); err != nil {
					return err
				}

				if err := c.Create(context.Background(), p); err != nil {
					return err
				}

				if err := c.Create(context.Background(), s); err != nil {
					return err
				}

				if err := c.Create(context.Background(), cl); err != nil {
					return err
				}

				d, err := time.ParseDuration("20s")
				if err != nil {
					return err
				}

				dt, err := time.ParseDuration("500s")
				if err != nil {
					return err
				}

				return wait.PollImmediate(d, dt, func() (bool, error) {
					b := &storagev1alpha1.Bucket{}
					if err := c.Get(context.TODO(), types.NamespacedName{Name: "gcp-claim", Namespace: "cool-namespace"}, b); err != nil {
						return false, err
					}
					if b.Status.BindingStatus.Phase == runtimev1alpha1.BindingPhaseBound {
						return true, nil
					}

					return false, nil
				})
			},
		},
	}

	cfg, err := config.GetConfig()
	if err != nil {
		t.Fatal(err)
	}

	a, err := athodyd.New(cfg, athodyd.WithCRDDirectoryPaths([]string{"../crds"}))
	if err != nil {
		t.Fatal(err)
	}

	addToScheme(a.GetScheme())
	controllerSetupWithManager(a)

	a.Run()

	defer func() {
		if err := a.Cleanup(); err != nil {
			t.Fatal(err)
		}
	}()

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := tc.test(a.GetClient())
			if err != nil {
				t.Error(err)
			}
		})
	}
}
