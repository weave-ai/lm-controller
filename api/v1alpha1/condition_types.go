package v1alpha1

const (
	// HealthyCondition represents the last recorded
	// health assessment result.
	HealthyCondition string = "Healthy"

	// PruneFailedReason represents the fact that the
	// pruning of the Kustomization failed.
	PruneFailedReason string = "PruneFailed"

	// ArtifactFailedReason represents the fact that the
	// source artifact download failed.
	ArtifactFailedReason string = "ArtifactFailed"

	// BuildFailedReason represents the fact that the
	// kustomize build failed.
	BuildFailedReason string = "BuildFailed"

	// HealthCheckFailedReason represents the fact that
	// one of the health checks failed.
	HealthCheckFailedReason string = "HealthCheckFailed"

	// DependencyNotReadyReason represents the fact that
	// one of the dependencies is not ready.
	DependencyNotReadyReason string = "DependencyNotReady"

	// ReconciliationSucceededReason represents the fact that
	// the reconciliation succeeded.
	ReconciliationSucceededReason string = "ReconciliationSucceeded"

	// ReconciliationFailedReason represents the fact that
	// the reconciliation failed.
	ReconciliationFailedReason string = "ReconciliationFailed"
)
