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

// AlicloudMachineSpec defines the desired state of AlicloudMachine
type AlicloudMachineSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ProviderID              string `json:"providerID,omitempty"`
	InternetChargeType      string `json:"internetChargeType,omitempty"`
	InternetMaxBandwidthIn  string `json:"internetMaxBandwidthIn"`
	InternetMaxBandwidthOut string `json:"internetMaxBandwidthOut"`
	SSHKeyPair              string `json:"sshKeyPair"`
	ImageId                 string `json:"imageId"`
	CapacityReservationId   string `json:"capacityReservationId,omitempty"`
	SystemDiskCategory      string `json:"systemDiskCategory,omitempty"`
	InstanceType            string `json:"instanceType"`
	SystemDiskSize          string `json:"systemDiskSize"`
}

// AlicloudMachineStatus defines the observed state of AlicloudMachine
type AlicloudMachineStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Ready bool `json:"ready"`

	Addresses clusterv1.MachineAddresses `json:"addresses,omitempty"`

	Phase string `json:"phase"`

	ErrorReason string `json:"errorReason,omitempty"`

	ErrorMessage string `json:"errorMessage,omitempty"`

	// +optional
	Instance *Instance `json:"instance,omitempty"`

	ID string `json:"id,omitempty"`
}

type Instance struct {
	ImageId                 string `json:"ImageId" xml:"ImageId"`
	InstanceType            string `json:"InstanceType" xml:"InstanceType"`
	OsType                  string `json:"OsType" xml:"OsType"`
	DeviceAvailable         bool   `json:"DeviceAvailable" xml:"DeviceAvailable"`
	InstanceNetworkType     string `json:"InstanceNetworkType" xml:"InstanceNetworkType"`
	LocalStorageAmount      int    `json:"LocalStorageAmount" xml:"LocalStorageAmount"`
	NetworkType             string `json:"NetworkType" xml:"NetworkType"`
	IsSpot                  bool   `json:"IsSpot" xml:"IsSpot"`
	InstanceChargeType      string `json:"InstanceChargeType" xml:"InstanceChargeType"`
	InstanceName            string `json:"InstanceName" xml:"InstanceName"`
	StartTime               string `json:"StartTime" xml:"StartTime"`
	ZoneId                  string `json:"ZoneId" xml:"ZoneId"`
	InternetChargeType      string `json:"InternetChargeType" xml:"InternetChargeType"`
	InternetMaxBandwidthIn  int    `json:"InternetMaxBandwidthIn" xml:"InternetMaxBandwidthIn"`
	HostName                string `json:"HostName" xml:"HostName"`
	Status                  string `json:"Status" xml:"Status"`
	CPU                     int    `json:"CPU" xml:"CPU"`
	Cpu                     int    `json:"Cpu" xml:"Cpu"`
	OSName                  string `json:"OSName" xml:"OSName"`
	OSNameEn                string `json:"OSNameEn" xml:"OSNameEn"`
	SerialNumber            string `json:"SerialNumber" xml:"SerialNumber"`
	RegionId                string `json:"RegionId" xml:"RegionId"`
	InternetMaxBandwidthOut int    `json:"InternetMaxBandwidthOut" xml:"InternetMaxBandwidthOut"`
	InstanceTypeFamily      string `json:"InstanceTypeFamily" xml:"InstanceTypeFamily"`
	InstanceId              string `json:"InstanceId" xml:"InstanceId"`
	Description             string `json:"Description" xml:"Description"`
	ExpiredTime             string `json:"ExpiredTime" xml:"ExpiredTime"`
	OSType                  string `json:"OSType" xml:"OSType"`
	Memory                  int    `json:"Memory" xml:"Memory"`
	CreationTime            string `json:"CreationTime" xml:"CreationTime"`
	KeyPairName             string `json:"KeyPairName" xml:"KeyPairName"`
	LocalStorageCapacity    int64  `json:"LocalStorageCapacity" xml:"LocalStorageCapacity"`
	VlanId                  string `json:"VlanId" xml:"VlanId"`
	StoppedMode             string `json:"StoppedMode" xml:"StoppedMode"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// AlicloudMachine is the Schema for the alicloudmachines API
type AlicloudMachine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AlicloudMachineSpec   `json:"spec,omitempty"`
	Status AlicloudMachineStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AlicloudMachineList contains a list of AlicloudMachine
type AlicloudMachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AlicloudMachine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AlicloudMachine{}, &AlicloudMachineList{})
}
