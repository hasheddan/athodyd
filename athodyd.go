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

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const syncPeriod = "30s"

type setupWithManagerFunc func(manager.Manager) error

type addToSchemeFunc func(*runtime.Scheme) error

type cleanFn func(client.Client) error

type executorFn func(client.Client) error

type janitorFn func(client.Client) error

// A Test is a logical operation in a cluster environment
type Test struct {
	Name        string
	Description string
	Executor    executorFn
	Janitor     janitorFn
	Persist     bool
}

// Config is a set of configuration values for a Job
type Config struct {
	SyncPeriod *time.Duration
}

// A Job is a set of tests that are executed sequentially in the same cluster environment
type Job struct {
	name             string
	description      string
	crdPath          string
	mcfg             *Config
	runner           *testing.T
	setupWithManager setupWithManagerFunc
	addToScheme      addToSchemeFunc
	clean            cleanFn
	tests            []Test
}

// NewJob creates a new Job with provided config
func NewJob(name string, description string, crdPath string, tests []Test, SyncPeriod string, runner *testing.T, swm setupWithManagerFunc, ats addToSchemeFunc) (*Job, error) {
	t, err := time.ParseDuration(SyncPeriod)
	if err != nil {
		return nil, err
	}

	jc := func(c client.Client) error {
		// return errors.New("you probably have some external resources to clean up")
		return nil
	}

	return &Job{
		name:             name,
		description:      description,
		crdPath:          crdPath,
		runner:           runner,
		tests:            tests,
		setupWithManager: swm,
		addToScheme:      ats,
		clean:            jc,
		mcfg: &Config{
			SyncPeriod: &t,
		},
	}, nil
}

// Run executes provided Tests as subtests in the Job environment
func (j *Job) Run() error {
	j.runner.Helper()

	j.runner.Log("Adding CRDs...")
	e := &envtest.Environment{
		CRDDirectoryPaths: []string{j.crdPath},
	}

	j.runner.Log("Starting up control plane...")
	cfg, err := e.Start()
	if err != nil {
		return err
	}

	j.runner.Log("Creating manager...")
	mgr, err := manager.New(cfg, manager.Options{SyncPeriod: j.mcfg.SyncPeriod})
	if err != nil {
		return err
	}

	j.runner.Log("Adding API types to scheme...")
	if err := j.addToScheme(mgr.GetScheme()); err != nil {
		return err
	}

	j.runner.Log("Registering controllers with manager...")
	if err := j.setupWithManager(mgr); err != nil {
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

	for _, tt := range j.tests {
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
			j.runner.Logf("%s specified exit on failure", tt.Name)
			break
		}

		if !successful && tt.Persist {
			j.runner.Logf("%s specified persistence on failure, continuing with additional tests", tt.Name)
		}
	}

	j.runner.Logf("running clean up for job %s", j.name)
	if err := j.clean(client); err != nil {
		j.runner.Fatal(err)
	}

	j.runner.Logf("successful clean up for job %s", j.name)

	close(ch)

	return e.Stop()
}
