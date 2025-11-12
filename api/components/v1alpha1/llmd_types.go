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

package v1alpha1

import (
	operatorv1 "github.com/openshift/api/operator/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/opendatahub-io/opendatahub-operator/v2/api/common"
)

const (
	LlmdComponentName = "llm-d"
	// value should match what's set in the XValidation below
	LlmdInstanceName = "default-llm-d"
	LlmdKind         = "Llmd"
)

// Check that the component implements common.PlatformObject.
var _ common.PlatformObject = (*Llmd)(nil)

// LlmdCommonSpec spec defines the shared desired state of Llmd
type LlmdCommonSpec struct {
	// ModelService configuration for llm-d modelservice chart
	ModelService LlmdModelServiceSpec `json:"modelService,omitempty"`

	// Infra configuration for llm-d infra chart
	Infra LlmdInfraSpec `json:"infra,omitempty"`

	// GatewayAPI configuration for Gateway API Inference Extension
	GatewayAPI LlmdGatewayAPISpec `json:"gatewayAPI,omitempty"`
}

// LlmdModelServiceSpec defines configuration for the llm-d modelservice Helm chart
type LlmdModelServiceSpec struct {
	// Enable or disable modelservice deployment
	// +kubebuilder:default=true
	Enabled bool `json:"enabled,omitempty"`

	// Chart version to use (default: v0.2.11)
	// +kubebuilder:default="0.2.11"
	Version string `json:"version,omitempty"`

	// Additional Helm values to override defaults
	// +optional
	Values map[string]string `json:"values,omitempty"`
}

// LlmdInfraSpec defines configuration for the llm-d infra Helm chart
type LlmdInfraSpec struct {
	// Enable or disable infra deployment
	// +kubebuilder:default=true
	Enabled bool `json:"enabled,omitempty"`

	// Chart version to use (default: v1.3.3)
	// +kubebuilder:default="1.3.3"
	Version string `json:"version,omitempty"`

	// Additional Helm values to override defaults
	// +optional
	Values map[string]string `json:"values,omitempty"`
}

// LlmdGatewayAPISpec defines configuration for the Gateway API Inference Extension Helm chart
type LlmdGatewayAPISpec struct {
	// Enable or disable Gateway API Inference Extension deployment
	// +kubebuilder:default=true
	Enabled bool `json:"enabled,omitempty"`

	// Chart version to use (default: v1.0.1)
	// +kubebuilder:default="1.0.1"
	Version string `json:"version,omitempty"`

	// Additional Helm values to override defaults
	// +optional
	Values map[string]string `json:"values,omitempty"`
}

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// LlmdSpec defines the desired state of Llmd
type LlmdSpec struct {
	// llmd spec exposed to DSC api
	LlmdCommonSpec `json:",inline"`
	// llmd spec exposed only to internal api
}

// LlmdCommonStatus defines the shared observed state of Llmd
type LlmdCommonStatus struct {
	common.ComponentReleaseStatus `json:",inline"`
}

// LlmdStatus defines the observed state of Llmd
type LlmdStatus struct {
	common.Status      `json:",inline"`
	LlmdCommonStatus   `json:",inline"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:validation:XValidation:rule="self.metadata.name == 'default-llm-d'",message="Llmd name must be default-llm-d"
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`,description="Ready"
// +kubebuilder:printcolumn:name="Reason",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].reason`,description="Reason"

// Llmd is the Schema for the llmd API
type Llmd struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LlmdSpec   `json:"spec,omitempty"`
	Status LlmdStatus `json:"status,omitempty"`
}

func (c *Llmd) GetStatus() *common.Status {
	return &c.Status.Status
}

// +kubebuilder:object:root=true

// LlmdList contains a list of Llmd
type LlmdList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Llmd `json:"items"`
}

func init() { //nolint:gochecknoinits
	SchemeBuilder.Register(&Llmd{}, &LlmdList{})
}

// DSCLlmd contains configuration for llm-d component in DSC
type DSCLlmd struct {
	// ManagementState indicates the component's management state
	// +kubebuilder:validation:Enum=Managed;Removed
	ManagementState operatorv1.ManagementState `json:"managementState,omitempty"`

	LlmdCommonSpec `json:",inline"`
}

// DSCLlmdStatus contains status information for llm-d component in DSC
type DSCLlmdStatus struct {
	ManagementState operatorv1.ManagementState `json:"managementState,omitempty"`
	*LlmdCommonStatus `json:",inline"`
}
