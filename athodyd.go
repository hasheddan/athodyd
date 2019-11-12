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

package athodyd

import (
	"testing"
	"time"

	// Allow auth to cloud providers
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/client-go/rest"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const syncPeriod = "30s"

var useExisting = true

// SetupWithManagerFunc is a function that adds controller to a manager
type SetupWithManagerFunc func(manager.Manager) error

// AddToSchemeFunc is a function that adds api types to a scheme
type AddToSchemeFunc func(*runtime.Scheme) error

// OperationFn is a function that uses a Kubernetes client to perform and operation
type OperationFn func(client.Client) error

// A Test is a logical operation in a cluster environment
type Test struct {
	Name        string
	Description string
	Executor    OperationFn
	Janitor     OperationFn
	Persist     bool
}

// JobConfig is a set of configuration values for a Job
type JobConfig struct {
	CRDDirectoryPaths []string
	Cluster           *rest.Config
	Builder           OperationFn
	Cleaner           OperationFn
	SetupWithManager  SetupWithManagerFunc
	AddToScheme       AddToSchemeFunc
	SyncPeriod        *time.Duration
}

// NewBuilder returns a new no-op Builder
func NewBuilder() OperationFn {
	return func(client.Client) error {
		return nil
	}
}

// NewCleaner returns a new no-op Cleaner
func NewCleaner() OperationFn {
	return func(client.Client) error {
		return nil
	}
}

// NewSetupWithManager returns a new no-op controller setup function
func NewSetupWithManager() SetupWithManagerFunc {
	return func(manager.Manager) error {
		return nil
	}
}

// NewAddToScheme returns a new no-op scheme adder
func NewAddToScheme() AddToSchemeFunc {
	return func(*runtime.Scheme) error {
		return nil
	}
}

// A JobOption configures a Job
type JobOption func(*Job)

// WithBuilder sets a custom builder function for a Job
func WithBuilder(builder OperationFn) JobOption {
	return func(j *Job) {
		j.cfg.Builder = builder
	}
}

// WithCleaner sets a custom cleaner function for a Job
func WithCleaner(cleaner OperationFn) JobOption {
	return func(j *Job) {
		j.cfg.Cleaner = cleaner
	}
}

// WithCluster sets a custom cluster for a Job
func WithCluster(cluster *rest.Config) JobOption {
	return func(j *Job) {
		j.cfg.Cluster = cluster
	}
}

// WithCRDDirectoryPaths sets custom CRD locations for a Job
func WithCRDDirectoryPaths(crds []string) JobOption {
	return func(j *Job) {
		j.cfg.CRDDirectoryPaths = crds
	}
}

// WithSetupWithManager sets custom CRD locations for a Job
func WithSetupWithManager(s SetupWithManagerFunc) JobOption {
	return func(j *Job) {
		j.cfg.SetupWithManager = s
	}
}

// WithAddToScheme sets custom CRD locations for a Job
func WithAddToScheme(a AddToSchemeFunc) JobOption {
	return func(j *Job) {
		j.cfg.AddToScheme = a
	}
}

func defaultConfig() *JobConfig {
	t, err := time.ParseDuration(syncPeriod)
	if err != nil {
		panic(err)
	}

	return &JobConfig{
		CRDDirectoryPaths: []string{},
		Cluster:           nil,
		Builder:           NewBuilder(),
		Cleaner:           NewCleaner(),
		SetupWithManager:  NewSetupWithManager(),
		AddToScheme:       NewAddToScheme(),
		SyncPeriod:        &t,
	}
}

// A Job is a set of tests that are executed sequentially in the same cluster environment
type Job struct {
	Name        string
	Description string
	Tests       []Test
	cfg         *JobConfig
	runner      *testing.T
}

// NewJob creates a new Job with provided config
func NewJob(name string, description string, tests []Test, runner *testing.T, o ...JobOption) *Job {
	j := &Job{
		Name:        name,
		Description: description,
		Tests:       tests,
		cfg:         defaultConfig(),
		runner:      runner,
	}

	for _, jo := range o {
		jo(j)
	}

	return j
}

// Run executes provided Tests as subtests in the Job environment
func (j *Job) Run() error {
	j.runner.Helper()

	j.runner.Log("Adding CRDs...")
	e := &envtest.Environment{
		CRDDirectoryPaths:  j.cfg.CRDDirectoryPaths,
		Config:             j.cfg.Cluster,
		UseExistingCluster: &useExisting,
	}

	j.runner.Log("Connecting to cluster...")
	cfg, err := e.Start()
	if err != nil {
		return err
	}

	j.runner.Log("Creating manager...")
	mgr, err := manager.New(cfg, manager.Options{SyncPeriod: j.cfg.SyncPeriod})
	if err != nil {
		return err
	}

	j.runner.Log("Adding API types to scheme...")
	if err := j.cfg.AddToScheme(mgr.GetScheme()); err != nil {
		return err
	}

	j.runner.Log("Registering controllers with manager...")
	if err := j.cfg.SetupWithManager(mgr); err != nil {
		return err
	}

	j.runner.Log("Starting up manager...")
	ch := make(chan struct{})
	go func() error {
		if err := mgr.Start(ch); err != nil {
			return err
		}
		return nil
	}()

	client, err := client.New(cfg, client.Options{})
	if err != nil {
		return err
	}

	for _, tt := range j.Tests {
		successful := j.runner.Run(tt.Name, func(t *testing.T) {
			t.Helper()
			if err := tt.Executor(client); err != nil {
				t.Fatalf("%s executor failed with error: %s", tt.Name, err)
			}
		})

		if successful {
			j.runner.Logf("%s completed successfully", tt.Name)
		}

		if !successful && !tt.Persist {
			if err := tt.Janitor(client); err != nil {
				j.runner.Fatalf("%s janitor failed with error: %s", tt.Name, err)
			}
			j.runner.Logf("%s specified exit on failure, foregoing additional tests", tt.Name)
			break
		}

		if !successful && tt.Persist {
			j.runner.Logf("%s specified persistence on failure, continuing with additional tests", tt.Name)
		}
	}

	j.runner.Logf("running clean up for job %s", j.Name)
	if err := j.cfg.Cleaner(client); err != nil {
		j.runner.Fatal(err)
	}

	j.runner.Logf("successful clean up for job %s", j.Name)

	close(ch)

	return e.Stop()
}
