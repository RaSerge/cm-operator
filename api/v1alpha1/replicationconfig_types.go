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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ReplicationConfigSpec defines the desired state of ReplicationConfig
type ReplicationConfigSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// TargetName is name of taregt config
	TargetName string `json:"targetname"`

	// TargetNamespace are target config namespace
	TargetNamespace string `json:"targetnamespace"`

	// Data contains the configuration data.
	// Each key must consist of alphanumeric characters, '-', '_' or '.'.
	// Values with non-UTF-8 byte sequences must use the BinaryData field.
	// The keys stored in Data must not overlap with the keys in
	// the BinaryData field.
	// +optional
	Data map[string]string `json:"data,omitempty"`

	// BinaryData contains the binary data.
	// Each key must consist of alphanumeric characters, '-', '_' or '.'.
	// BinaryData can contain byte sequences that are not in the UTF-8 range.
	// The keys stored in BinaryData must not overlap with the keys in
	// the Data field.
	// +optional
	BinaryData map[string][]byte `json:"binaryData,omitempty"`
}

// ReplicationConfigStatus defines the observed state of ReplicationConfig
type ReplicationConfigStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// ReplicatedKeys are the keys replicated to target config
	ReplicatedKeys []string `json:"replicatedkeys"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ReplicationConfig is the Schema for the replicationconfigs API
type ReplicationConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ReplicationConfigSpec   `json:"spec,omitempty"`
	Status ReplicationConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ReplicationConfigList contains a list of ReplicationConfig
type ReplicationConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ReplicationConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ReplicationConfig{}, &ReplicationConfigList{})
}
