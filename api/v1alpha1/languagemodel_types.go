/*
Copyright 2023.

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
	"github.com/fluxcd/pkg/apis/meta"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

const (
	LanguageModelKind      = "LanguageModel"
	LanguageModelFinalizer = "finalizer.lm-controller.ai.contrib.fluxcd.io"
	EnabledValue           = "enabled"
	DisabledValue          = "disabled"
	MergeValue             = "Merge"
	IfNotPresentValue      = "IfNotPresent"
	IgnoreValue            = "Ignore"
)

// LanguageModelSpec defines the desired state of LanguageModel
type LanguageModelSpec struct {
	// SourceRef is the reference to the source of the model
	// +required
	SourceRef CrossNamespaceSourceReference `json:"sourceRef,omitempty"`

	// Interval at which to check for new versions of the model, and update the model deployment.
	// This interval is approximate and may be subject to jitter to ensure
	// efficient use of resources.
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Pattern="^([0-9]+(\\.[0-9]+)?(ms|s|m|h))+$"
	// +required
	Interval metav1.Duration `json:"interval"`

	// RetryInterval is the interval at which to retry a failed model deployment.
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Pattern="^([0-9]+(\\.[0-9]+)?(ms|s|m|h))+$"
	// +optional
	RetryInterval metav1.Duration `json:"retryInterval"`

	// Engine is the engine to use for the model.
	//+kubebuilder:default:={"engineType":"Default","deploymentType":"Default"}
	//+required
	Engine EngineSpec `json:"engine,omitempty"`

	// Suspend is the object suspend flag.
	// Default to false.
	// +optional
	Suspend bool `json:"suspend,omitempty"`

	// Prune is the flag to prune old revisions.
	// Default to false.
	// +optional
	Prune bool `json:"prune,omitempty"`

	// Force is the flag to force the reconciliation.
	// +optional
	Force bool `json:"force,omitempty"`

	// ModelPullPolicy is the policy used for pulling model from the Source Controller.
	//+kubebuilder:validation:Enum=Always;IfNotPresent;Never
	//+kubebuilder:default=IfNotPresent
	//+optional
	ModelPullPolicy string `json:"modelPullPolicy,omitempty"`

	// ModelCacheStrategy is the strategy for caching models. Default to None.
	// Set to PV if you want to use PersistantVolume to cache the model.
	//+kubebuilder:validation:Enum=None;PV
	//+kubebuilder:default=None
	//+optional
	ModelCacheStrategy string `json:"modelCacheStrategy,omitempty"`

	// Parameters is the parameters to use for the model.
	// Parameters *apiextensionsv1.JSON `json:"parameters,omitempty"`

	// SystemPrompt is the system prompt to use for the model.
	// SystemPrompt string `json:"systemPrompt,omitempty"`

	// PromptTemplates is the prompt templates to use for the model.
	// PromptTemplates []PromptTemplate `json:"promptTemplates,omitempty"`

	// ServiceAccountName is the name of the service account to use for deploying the model server.
	//+kubebuilder:default=default
	//+optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// Timeout for validation, and apply operations.
	// Defaults to 'Interval' duration.
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Pattern="^([0-9]+(\\.[0-9]+)?(ms|s|m|h))+$"
	// +optional
	Timeout *metav1.Duration `json:"timeout,omitempty"`
}

// LanguageModelStatus defines the observed state of LanguageModel
type LanguageModelStatus struct {
	meta.ReconcileRequestStatus `json:",inline"`

	// ObservedGeneration is the last reconciled generation.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// The last successfully applied revision.
	// Equals the Revision of the applied Artifact from the referenced Source.
	// +optional
	LastAppliedRevision string `json:"lastAppliedRevision,omitempty"`

	// LastAttemptedRevision is the revision of the last reconciliation attempt.
	// +optional
	LastAttemptedRevision string `json:"lastAttemptedRevision,omitempty"`

	// Inventory contains the list of Kubernetes resource object references that
	// have been successfully applied.
	// +optional
	Inventory *ResourceInventory `json:"inventory,omitempty"`
}

// +genclient
// +genclient:Namespaced
// +kubebuilder:storageversion
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=lm
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description=""
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status",description=""
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].message",description=""

// LanguageModel is the Schema for the language models API
type LanguageModel struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec LanguageModelSpec `json:"spec,omitempty"`

	// +kubebuilder:default:={"observedGeneration":-1}
	Status LanguageModelStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// LanguageModelList contains a list of LanguageModel
type LanguageModelList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LanguageModel `json:"items"`
}

// PromptTemplate is the prompt template to use for the model
type PromptTemplate struct {
	// Name is the name of the prompt template
	Name string `json:"name,omitempty"`

	// Template is the template of the prompt template
	Template string `json:"template,omitempty"`
}

type EngineSpec struct {
	// Name is the name of the engine.
	//+kubebuilder:validation:Enum=Default;LocalAI;LLamaCppPython
	//+kubebuilder:default=Default
	EngineType string `json:"engineType,omitempty"`

	//+kubebuilder:validation:Enum=Default;KubernetesDeployment;KnativeService
	//+kubebuilder:default=Default
	DeploymentType string `json:"deploymentType,omitempty"`

	// Replicas is the replicas of the engine.
	//+kubebuilder:default=1
	Replicas *int32 `json:"replicas,omitempty"`

	// Image is the image of the engine.
	// +optional
	Image string `json:"image,omitempty"`

	// NodeSelector is a selector which must be true for the pod to fit on a node.
	// Selector which must match a node's labels for the pod to be scheduled on that node.
	// More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/
	// +optional
	// +mapType=atomic
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Resources is the resources of the engine.
	//+kubebuilder:default={"limits":{"cpu":"1000m","ephemeral-storage":"1Gi","memory":"1Gi"},"requests":{"cpu":"1000m","ephemeral-storage":"1Gi", "memory":"1Gi"}}
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// StorageClass is the storage class of the engine.
	//+kubebuilder:default=standard
	StorageClass *string `json:"storageClass,omitempty"`

	// TODO Security SecuritySpec `json:"security,omitempty"`

	// ImagePullPolicy is the image pull policy of the engine.
	// +kubebuilder:validation:Enum=Always;IfNotPresent;Never
	// +kubebuilder:default=IfNotPresent
	// ImagePullPolicy string `json:"imagePullPolicy,omitempty"`

	// ImagePullSecrets is the image pull secrets of the engine.
	// +kubebuilder:default=weave-ai-registry
	// ImagePullSecrets []string `json:"imagePullSecrets,omitempty"`
}

func (in *LanguageModelSpec) GetEngine() EngineSpec {
	return in.Engine
}

func (in *LanguageModelSpec) GetEngineImage() string {
	image := in.Engine.Image
	if image == "" {
		image = ImageEngineLlamaCppPython
	}
	return image
}

func (in *LanguageModelSpec) GetModelPullPolicy() string {
	if in.ModelPullPolicy == "" {
		return "IfNotPresent"
	}
	return in.ModelPullPolicy
}

func (in *LanguageModelSpec) GetModelCacheStrategy() string {
	if in.ModelCacheStrategy == "" {
		return "None"
	}
	return in.ModelCacheStrategy
}

// GetConditions returns the status conditions of the object.
func (in *LanguageModel) GetConditions() []metav1.Condition {
	return in.Status.Conditions
}

// SetConditions sets the status conditions on the object.
func (in *LanguageModel) SetConditions(conditions []metav1.Condition) {
	in.Status.Conditions = conditions
}

func (in *LanguageModel) GetRetryInterval() time.Duration {
	// check if the retry interval is set
	if in.Spec.RetryInterval.Duration > 0 {
		return in.Spec.RetryInterval.Duration
	}
	return in.GetRequeueAfter()
}

func (in *LanguageModel) GetRequeueAfter() time.Duration {
	return in.Spec.Interval.Duration
}

// GetTimeout returns the timeout with default.
func (in *LanguageModel) GetTimeout() time.Duration {
	duration := in.Spec.Interval.Duration - 30*time.Second
	if in.Spec.Timeout != nil {
		duration = in.Spec.Timeout.Duration
	}
	if duration < 30*time.Second {
		return 30 * time.Second
	}
	return duration
}

func init() {
	SchemeBuilder.Register(&LanguageModel{}, &LanguageModelList{})
}
