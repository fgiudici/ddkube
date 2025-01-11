/*
Copyright 2025.

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

package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DDNSServiceData defines endpoint and authentication data for DDNS services.
type DDNSServiceData struct {
	// AuthSecretRef a reference to the secret containing DDNS endpoint auth credentials.
	// +required
	AuthSecretRef *corev1.ObjectReference `json:"authSecretRef"`
	// Endpoint API endpoint of the DDNS service provider.
	// +required
	Endpoint string `json:"endpoint"`
}

// HostnameSpec defines the desired state of Hostname
type HostnameSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Hostname fully qualified domain name to update.
	// +required
	Hostname string `json:"hostname"`
	// Addresss IP address to be assigned to the FQDN hostname: if empty the current public ip address will
	// be retrieved and assigned.
	// +optional
	Address string `json:"address"`
	// CheckIntervalMinutes interval to wait before checking if the ip address of the hostname should be updated.
	// +default:value=0
	CheckIntervalMinutes *int32 `json:"checkIntervalMinutes"`
	// DDNSService DDNS service data to authenticate and update the hostname address.
	// +required
	DDNSService *DDNSServiceData `json:"ddnsService"`
}

type LastUpdateData struct {
	// ScheduledAt tracks the last time the hostname update was attempted.
	// +optional
	ScheduledAt *metav1.Time `json:"scheduledAt"`
	// Failed records if the last update failed.
	// +optional
	Failed bool `json:"failed"`
	// Hostname records the last FQDN sent to the DDNS service API.
	// +optional
	Hostname string `json:"hostname"`
	// Address records the last address sent to the DDNS service API.
	// +optional
	Address string `json:"address"`
}

// HostnameStatus defines the observed state of Hostname
type HostnameStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Conditions describe the state of the hostname object.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
	// LastUpdate reports main data about the last update.
	// +optional
	LastUpdate *LastUpdateData `json:"lastUpdate,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Hostname is the Schema for the hostnames API
type Hostname struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HostnameSpec   `json:"spec,omitempty"`
	Status HostnameStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HostnameList contains a list of Hostname
type HostnameList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Hostname `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Hostname{}, &HostnameList{})
}
