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
	"errors"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/validation"
	"sigs.k8s.io/yaml"
)

// func convertV1ToInternal(data []byte, internal *apiextensions.CustomResourceDefinition) error {
// 	_ = kubernetes.Interface
// 	crd := &v1.CustomResourceDefinition{}
// 	if err := yaml.Unmarshal(data, crd); err != nil {
// 		return err
// 	}
// 	v1.SetDefaults_CustomResourceDefinition(crd)
// 	if err := v1.Convert_v1_CustomResourceDefinition_To_apiextensions_CustomResourceDefinition(crd, internal, nil); err != nil {
// 		return err
// 	}
// 	errList := validation.ValidateCustomResourceDefinition(internal, v1.SchemeGroupVersion)
// 	if len(errList) > 0 {
// 		return errors.New(errList.ToAggregate().Error())
// 	}

// 	return nil
// }

func convertV1Beta1ToInternal(data []byte, internal *apiextensions.CustomResourceDefinition) error {
	crd := &v1beta1.CustomResourceDefinition{}
	if err := yaml.Unmarshal(data, crd); err != nil {
		return err
	}
	v1beta1.SetObjectDefaults_CustomResourceDefinition(crd)
	if err := v1beta1.Convert_v1beta1_CustomResourceDefinition_To_apiextensions_CustomResourceDefinition(crd, internal, nil); err != nil {
		return err
	}
	errList := validation.ValidateCustomResourceDefinition(internal)
	if len(errList) > 0 {
		return errors.New(errList.ToAggregate().Error())
	}

	return nil
}
