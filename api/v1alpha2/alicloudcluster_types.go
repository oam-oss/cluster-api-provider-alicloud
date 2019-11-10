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
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha2"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AlicloudClusterSpec defines the desired state of AlicloudCluster
type AlicloudClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Network  NetworkSpec `json:"network,omitempty"`
	ZoneId   string      `json:"zoneId,omitempty"`
	RegionId string      `json:"regionId,omitempty"`
}

// AlicloudClusterStatus defines the observed state of AlicloudCluster
type AlicloudClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Ready   bool    `json:"ready,omitempty"`
	Network Network `json:"network,omitempty"`
	// +optional
	ApiEndpoints []clusterv1.APIEndpoint `json:"apiEndpoints,omitempty"`
	Reason       string                  `json:"reason,omitempty"`
	Message      string                  `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// AlicloudCluster is the Schema for the alicloudclusters API
type AlicloudCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AlicloudClusterSpec   `json:"spec,omitempty"`
	Status AlicloudClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AlicloudClusterList contains a list of AlicloudCluster
type AlicloudClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AlicloudCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AlicloudCluster{}, &AlicloudClusterList{})
}
