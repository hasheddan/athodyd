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
	"fmt"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	databasev1beta1 "github.com/crossplaneio/stack-gcp/apis/database/v1beta1"
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
	dbVersion := "MYSQL_5_7"
	ddt := "PD_SSD"
	dds := int64(10)

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
								Name:      "example-provider-gcp",
								Namespace: "crossplane-system",
							},
						},
						ProjectID: "crossplane-playground",
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
			Name:        "TestCloudSQLProvisioning",
			Description: "This test checks to see if a GCP CloudSQL instance can be created successfully.",
			Executor: func(c client.Client) error {
				s := &databasev1beta1.CloudSQLInstance{
					ObjectMeta: metav1.ObjectMeta{
						Name: "gcp-cloudsql",
					},
					Spec: databasev1beta1.CloudSQLInstanceSpec{
						ResourceSpec: runtimev1alpha1.ResourceSpec{
							ProviderReference: &corev1.ObjectReference{
								Name: "gcp-provider",
							},
							ReclaimPolicy: runtimev1alpha1.ReclaimDelete,
						},
						ForProvider: databasev1beta1.CloudSQLInstanceParameters{
							Region:          "us-central1",
							DatabaseVersion: &dbVersion,
							Settings: databasev1beta1.Settings{
								Tier:           "db-n1-standard-1",
								DataDiskType:   &ddt,
								DataDiskSizeGb: &dds,
							},
						},
					},
				}

				if err := c.Create(context.TODO(), s); err != nil {
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

				wait.PollImmediate(d, dt, func() (bool, error) {
					fmt.Println("CHECK")
					g := &databasev1beta1.CloudSQLInstance{}
					if err := c.Get(context.TODO(), types.NamespacedName{Name: "gcp-cloudsql"}, g); err != nil {
						return false, err
					}
					if g.Status.AtProvider.State == databasev1beta1.StateRunnable {
						return true, nil
					}
					return false, nil
				})
				return nil
			},
			Janitor: func(client.Client) error {
				// TODO: implement janitor
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

	job := athodyd.NewJob(name, description, tests, t,
		athodyd.WithCluster(cfg),
		athodyd.WithCRDDirectoryPaths([]string{"../crds"}),
		athodyd.WithSetupWithManager(controllerSetupWithManager),
		athodyd.WithAddToScheme(addToScheme),
		// TODO: add cleaner
	)

	if err := job.Run(); err != nil {
		t.Fatal(err)
	}
}
