/*

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

// +kubebuilder:validation:Optional

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AlicloudMachineTemplateSpec defines the desired state of AlicloudMachineTemplate
type AlicloudMachineTemplateSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Template AlicloudMachineTemplateResource `json:"template"`
}

type AlicloudMachineTemplateResource struct {
	// Spec is the specification of the desired behavior of the machine.
	Spec AlicloudMachineSpec `json:"spec"`
}

// +kubebuilder:object:root=true

// AlicloudMachineTemplate is the Schema for the alicloudmachinetemplates API
type AlicloudMachineTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec AlicloudMachineTemplateSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// AlicloudMachineTemplateList contains a list of AlicloudMachineTemplate
type AlicloudMachineTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AlicloudMachineTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AlicloudMachineTemplate{}, &AlicloudMachineTemplateList{})
}
