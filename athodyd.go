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
	"context"
	"io/ioutil"
	"testing"
	"time"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	syncPeriod = "30s"
)

type setupWithManagerFunc func(manager.Manager) error

type addToSchemeFunc func(*runtime.Scheme) error

type cleanFn func(client.Client) error

type executorFn func(client.Client) error

type janitorFn func(client.Client) error

// Test is something
type Test struct {
	Name        string
	Description string
	Executor    executorFn
	Janitor     janitorFn
	Persist     bool
}

// Config does this
type Config struct {
	timeout *time.Duration
}

// Job does this
type Job struct {
	name             string
	description      string
	mcfg             *Config
	runner           *testing.T
	setupWithManager setupWithManagerFunc
	addToScheme      addToSchemeFunc
	clean            cleanFn
	tests            []Test
}

// NewJob does this
func NewJob(name string, description string, tests []Test, timeout string, runner *testing.T, swm setupWithManagerFunc, ats addToSchemeFunc) (*Job, error) {
	t, err := time.ParseDuration(timeout)
	if err != nil {
		return nil, err
	}

	jc := func(c client.Client) error {
		// return errors.New("you probably have some external resources to clean up")
		return nil
	}

	return &Job{
		name:             name,
		runner:           runner,
		tests:            tests,
		setupWithManager: swm,
		addToScheme:      ats,
		clean:            jc,
		mcfg: &Config{
			timeout: &t,
		},
	}, nil
}

// Run does this
func (j *Job) Run() error {
	j.runner.Helper()

	j.runner.Log("Starting up control plane...")
	e := &envtest.Environment{}
	cfg, err := e.Start()
	if err != nil {
		return err
	}

	mgr, err := manager.New(cfg, manager.Options{SyncPeriod: j.mcfg.timeout})
	if err != nil {
		return err
	}

	if err := j.addToScheme(mgr.GetScheme()); err != nil {
		return err
	}

	if err := j.setupWithManager(mgr); err != nil {
		return err
	}

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

	data, err := ioutil.ReadFile("./crds/provider.yaml")
	if err != nil {
		panic(err)
	}

	crd := &apiextensions.CustomResourceDefinition{}

	if err := convertV1Beta1ToInternal(data, crd); err != nil {
		return err
	}

	if err := client.Create(context.TODO(), crd); err != nil {
		return err
	}

	for _, tt := range j.tests {
		successful := j.runner.Run(tt.Name, func(t *testing.T) {
			t.Helper()
			if err := tt.Executor(client); err != nil {
				t.Fatalf("%s executor failed with error: %s", tt.Name, err)
				err := tt.Janitor(client)
				if err != nil {
					t.Fatalf("%s janitor failed with error: %s", tt.Name, err)
				}
			}
		})

		if !successful && !tt.Persist {
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
