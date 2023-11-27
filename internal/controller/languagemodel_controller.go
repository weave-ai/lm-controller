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

package controller

import (
	"context"
	"errors"
	"fmt"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/fluxcd/pkg/runtime/predicates"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"knative.dev/pkg/ptr"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	"sigs.k8s.io/cli-utils/pkg/object"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/ratelimiter"

	apiacl "github.com/fluxcd/pkg/apis/acl"
	eventv1 "github.com/fluxcd/pkg/apis/event/v1beta1"
	"github.com/fluxcd/pkg/apis/meta"
	"github.com/fluxcd/pkg/http/fetch"
	"github.com/fluxcd/pkg/runtime/acl"
	runtimeClient "github.com/fluxcd/pkg/runtime/client"
	"github.com/fluxcd/pkg/runtime/conditions"
	runtimeCtrl "github.com/fluxcd/pkg/runtime/controller"
	"github.com/fluxcd/pkg/runtime/jitter"
	"github.com/fluxcd/pkg/runtime/patch"
	"github.com/fluxcd/pkg/ssa"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	sourcev1b2 "github.com/fluxcd/source-controller/api/v1beta2"
	aiv1a1 "github.com/weave-ai/lm-controller/api/v1alpha1"
	"github.com/weave-ai/lm-controller/internal/inventory"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	kuberecorder "k8s.io/client-go/tools/record"
	"sigs.k8s.io/cli-utils/pkg/kstatus/polling"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// +kubebuilder:rbac:groups=ai.contrib.fluxcd.io,resources=languagemodels,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ai.contrib.fluxcd.io,resources=languagemodels/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=ai.contrib.fluxcd.io,resources=languagemodels/finalizers,verbs=get;create;update;patch;delete
// +kubebuilder:rbac:groups=source.toolkit.fluxcd.io,resources=buckets;ocirepositories;gitrepositories,verbs=get;list;watch
// +kubebuilder:rbac:groups=source.toolkit.fluxcd.io,resources=buckets/status;ocirepositories/status;gitrepositories/status,verbs=get
// +kubebuilder:rbac:groups="",resources=configmaps;secrets;serviceaccounts,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// LanguageModelReconciler reconciles a LanguageModel object
type LanguageModelReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	kuberecorder.EventRecorder
	runtimeCtrl.Metrics
	statusManager  patch.WithFieldOwner
	ControllerName string

	StatusPoller          *polling.StatusPoller
	PollingOpts           polling.Options
	DefaultServiceAccount string
	KubeConfigOpts        runtimeClient.KubeConfigOptions
	ConcurrentSSA         int
	NoCrossNamespaceRefs  bool
	requeueDependency     time.Duration
	FailFast              bool
}

// LanguageModelReconcilerOptions contains options for the LanguageModelReconciler.
type LanguageModelReconcilerOptions struct {
	HTTPRetry   int
	RateLimiter ratelimiter.RateLimiter
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the LanguageModel object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
func (r *LanguageModelReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, retErr error) {
	log := ctrl.LoggerFrom(ctx)
	reconcileStart := time.Now()

	obj := &aiv1a1.LanguageModel{}
	if err := r.Get(ctx, req.NamespacedName, obj); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Initialize the runtime patcher with the current version of the object.
	patcher := patch.NewSerialPatcher(obj, r.Client)

	// Finalise the reconciliation and report the results.
	defer func() {
		// Patch finalizers, status and conditions.
		if err := r.finalizeStatus(ctx, obj, patcher); err != nil {
			retErr = kerrors.NewAggregate([]error{retErr, err})
		}

		// Record Prometheus metrics.
		r.Metrics.RecordReadiness(ctx, obj)
		r.Metrics.RecordDuration(ctx, obj, reconcileStart)
		r.Metrics.RecordSuspend(ctx, obj, obj.Spec.Suspend)

		// Log and emit success event.
		if conditions.IsReady(obj) {
			msg := fmt.Sprintf("Reconciliation finished in %s, next run in %s",
				time.Since(reconcileStart).String(),
				obj.Spec.Interval.Duration.String())
			log.Info(msg, "revision", obj.Status.LastAttemptedRevision)
			r.event(obj, obj.Status.LastAppliedRevision, eventv1.EventSeverityInfo, msg,
				map[string]string{
					aiv1a1.GroupVersion.Group + "/" + eventv1.MetaCommitStatusKey: eventv1.MetaCommitStatusUpdateValue,
				})
		}
	}()

	// Prune managed resources if the object is under deletion.
	if !obj.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.finalize(ctx, obj)
	}

	// Add finalizer first if it doesn't exist to avoid the race condition
	// between init and delete.
	// Note: Finalizers in general can only be added when the deletionTimestamp
	// is not set.
	if !controllerutil.ContainsFinalizer(obj, aiv1a1.LanguageModelFinalizer) {
		controllerutil.AddFinalizer(obj, aiv1a1.LanguageModelFinalizer)
		return ctrl.Result{Requeue: true}, nil
	}

	// Skip reconciliation if the object is suspended.
	if obj.Spec.Suspend {
		log.Info("Reconciliation is suspended for this object")
		return ctrl.Result{}, nil
	}

	// Resolve the source reference and requeue the reconciliation if the source is not found.
	artifactSource, err := r.getSource(ctx, obj)
	if err != nil {
		conditions.MarkFalse(obj, meta.ReadyCondition, aiv1a1.ArtifactFailedReason, err.Error())

		if apierrors.IsNotFound(err) {
			retryInterval := jitter.JitteredIntervalDuration(obj.GetRetryInterval())
			msg := fmt.Sprintf("Source '%s' not found", obj.Spec.SourceRef.String())
			log.Info(msg)
			return ctrl.Result{RequeueAfter: retryInterval}, nil
		}

		if acl.IsAccessDenied(err) {
			retryInterval := jitter.JitteredIntervalDuration(obj.GetRetryInterval())
			conditions.MarkFalse(obj, meta.ReadyCondition, apiacl.AccessDeniedReason, err.Error())
			log.Error(err, "Access denied to cross-namespace source")
			r.event(obj, "unknown", eventv1.EventSeverityError, err.Error(), nil)
			return ctrl.Result{RequeueAfter: retryInterval}, nil
		}

		// Retry with backoff on transient errors.
		return ctrl.Result{Requeue: true}, err
	}

	// Requeue the reconciliation if the source artifact is not found.
	if artifactSource.GetArtifact() == nil {
		retryInterval := jitter.JitteredIntervalDuration(obj.GetRetryInterval())
		msg := fmt.Sprintf("Source artifact not found, retrying in %s", retryInterval.String())
		conditions.MarkFalse(obj, meta.ReadyCondition, aiv1a1.ArtifactFailedReason, msg)
		log.Info(msg)
		return ctrl.Result{RequeueAfter: retryInterval}, nil
	}

	// Reconcile the latest revision.
	reconcileErr := r.reconcile(ctx, obj, artifactSource, patcher)

	// Requeue at the specified retry interval if the artifact tarball is not found.
	if errors.Is(reconcileErr, fetch.FileNotFoundError) {
		retryInterval := jitter.JitteredIntervalDuration(obj.GetRetryInterval())
		msg := fmt.Sprintf("Source is not ready, artifact not found, retrying in %s", retryInterval.String())
		conditions.MarkFalse(obj, meta.ReadyCondition, aiv1a1.ArtifactFailedReason, msg)
		log.Info(msg)
		return ctrl.Result{RequeueAfter: retryInterval}, nil
	}

	// Broadcast the reconciliation failure and requeue at the specified retry interval.
	if reconcileErr != nil {
		retryInterval := jitter.JitteredIntervalDuration(obj.GetRetryInterval())
		log.Error(reconcileErr, fmt.Sprintf("Reconciliation failed after %s, next try in %s",
			time.Since(reconcileStart).String(),
			retryInterval.String()),
			"revision",
			artifactSource.GetArtifact().Revision)
		r.event(obj, artifactSource.GetArtifact().Revision, eventv1.EventSeverityError,
			reconcileErr.Error(), nil)
		return ctrl.Result{RequeueAfter: retryInterval}, nil
	}

	// Requeue the reconciliation at the specified interval.
	return ctrl.Result{RequeueAfter: jitter.JitteredIntervalDuration(obj.GetRequeueAfter())}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *LanguageModelReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, opts LanguageModelReconcilerOptions) error {
	const (
		ociRepositoryIndexKey string = ".metadata.ociRepository"
		gitRepositoryIndexKey string = ".metadata.gitRepository"
		bucketIndexKey        string = ".metadata.bucket"
	)

	// Index the LanguageModels by the OCIRepository references they (may) point at.
	if err := mgr.GetCache().IndexField(ctx, &aiv1a1.LanguageModel{}, ociRepositoryIndexKey,
		r.indexBy(sourcev1b2.OCIRepositoryKind)); err != nil {
		return fmt.Errorf("failed setting index fields: %w", err)
	}

	// Index the Kustomizations by the GitRepository references they (may) point at.
	if err := mgr.GetCache().IndexField(ctx, &aiv1a1.LanguageModel{}, gitRepositoryIndexKey,
		r.indexBy(sourcev1.GitRepositoryKind)); err != nil {
		return fmt.Errorf("failed setting index fields: %w", err)
	}

	// Index the Kustomizations by the Bucket references they (may) point at.
	if err := mgr.GetCache().IndexField(ctx, &aiv1a1.LanguageModel{}, bucketIndexKey,
		r.indexBy(sourcev1b2.BucketKind)); err != nil {
		return fmt.Errorf("failed setting index fields: %w", err)
	}

	r.statusManager = patch.WithFieldOwner(fmt.Sprintf("gotk-%s", string(r.ControllerName)))
	return ctrl.NewControllerManagedBy(mgr).
		For(&aiv1a1.LanguageModel{}, builder.WithPredicates(
			predicate.Or(predicate.GenerationChangedPredicate{}, predicates.ReconcileRequestedPredicate{}),
		)).
		Watches(
			&sourcev1b2.OCIRepository{},
			handler.EnqueueRequestsFromMapFunc(r.requestsForRevisionChangeOf(ociRepositoryIndexKey)),
			builder.WithPredicates(SourceRevisionChangePredicate{}),
		).
		Watches(
			&sourcev1.GitRepository{},
			handler.EnqueueRequestsFromMapFunc(r.requestsForRevisionChangeOf(gitRepositoryIndexKey)),
			builder.WithPredicates(SourceRevisionChangePredicate{}),
		).
		Watches(
			&sourcev1b2.Bucket{},
			handler.EnqueueRequestsFromMapFunc(r.requestsForRevisionChangeOf(bucketIndexKey)),
			builder.WithPredicates(SourceRevisionChangePredicate{}),
		).
		WithOptions(controller.Options{
			RateLimiter: opts.RateLimiter,
		}).
		Complete(r)
}

func (r *LanguageModelReconciler) finalizeStatus(ctx context.Context,
	obj *aiv1a1.LanguageModel,
	patcher *patch.SerialPatcher) error {
	// Set the value of the reconciliation request in status.
	if v, ok := meta.ReconcileAnnotationValue(obj.GetAnnotations()); ok {
		obj.Status.LastHandledReconcileAt = v
	}

	// Remove the Reconciling condition and update the observed generation
	// if the reconciliation was successful.
	if conditions.IsTrue(obj, meta.ReadyCondition) {
		conditions.Delete(obj, meta.ReconcilingCondition)
		obj.Status.ObservedGeneration = obj.Generation
	}

	// Set the Reconciling reason to ProgressingWithRetry if the
	// reconciliation has failed.
	if conditions.IsFalse(obj, meta.ReadyCondition) &&
		conditions.Has(obj, meta.ReconcilingCondition) {
		rc := conditions.Get(obj, meta.ReconcilingCondition)
		rc.Reason = meta.ProgressingWithRetryReason
		conditions.Set(obj, rc)
	}

	// Patch finalizers, status and conditions.
	return r.patch(ctx, obj, patcher)
}

func (r *LanguageModelReconciler) patch(ctx context.Context,
	obj *aiv1a1.LanguageModel,
	patcher *patch.SerialPatcher) (retErr error) {

	// Configure the runtime patcher.
	patchOpts := []patch.Option{}
	ownedConditions := []string{
		aiv1a1.HealthyCondition,
		meta.ReadyCondition,
		meta.ReconcilingCondition,
		meta.StalledCondition,
	}
	patchOpts = append(patchOpts,
		patch.WithOwnedConditions{Conditions: ownedConditions},
		patch.WithForceOverwriteConditions{},
		r.statusManager,
	)

	// Patch the object status, conditions and finalizers.
	if err := patcher.Patch(ctx, obj, patchOpts...); err != nil {
		if !obj.GetDeletionTimestamp().IsZero() {
			err = kerrors.FilterOut(err, func(e error) bool { return apierrors.IsNotFound(e) })
		}
		retErr = kerrors.NewAggregate([]error{retErr, err})
		if retErr != nil {
			return retErr
		}
	}

	return nil
}

func (r *LanguageModelReconciler) finalize(ctx context.Context,
	obj *aiv1a1.LanguageModel) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	if obj.Spec.Prune &&
		!obj.Spec.Suspend &&
		obj.Status.Inventory != nil &&
		obj.Status.Inventory.Entries != nil {
		objects, _ := inventory.List(obj.Status.Inventory)
		impersonation := runtimeClient.NewImpersonator(
			r.Client,
			r.StatusPoller,
			r.PollingOpts,
			nil, // obj.Spec.KubeConfig
			r.KubeConfigOpts,
			r.DefaultServiceAccount,
			obj.Spec.ServiceAccountName,
			obj.GetNamespace(),
		)
		if impersonation.CanImpersonate(ctx) {
			kubeClient, _, err := impersonation.GetClient(ctx)
			if err != nil {
				return ctrl.Result{}, err
			}

			resourceManager := ssa.NewResourceManager(kubeClient, nil, ssa.Owner{
				Field: r.ControllerName,
				Group: aiv1a1.GroupVersion.Group,
			})

			opts := ssa.DeleteOptions{
				PropagationPolicy: metav1.DeletePropagationBackground,
				Inclusions:        resourceManager.GetOwnerLabels(obj.Name, obj.Namespace),
				Exclusions: map[string]string{
					fmt.Sprintf("%s/prune", aiv1a1.GroupVersion.Group):     aiv1a1.DisabledValue,
					fmt.Sprintf("%s/reconcile", aiv1a1.GroupVersion.Group): aiv1a1.DisabledValue,
				},
			}

			changeSet, err := resourceManager.DeleteAll(ctx, objects, opts)
			if err != nil {
				r.event(obj, obj.Status.LastAppliedRevision, eventv1.EventSeverityError, "pruning for deleted resource failed", nil)
				// Return the error so we retry the failed garbage collection
				return ctrl.Result{}, err
			}

			if changeSet != nil && len(changeSet.Entries) > 0 {
				r.event(obj, obj.Status.LastAppliedRevision, eventv1.EventSeverityInfo, changeSet.String(), nil)
			}
		} else {
			// when the account to impersonate is gone, log the stale objects and continue with the finalization
			msg := fmt.Sprintf("unable to prune objects: \n%s", ssa.FmtUnstructuredList(objects))
			log.Error(fmt.Errorf("skiping pruning, failed to find account to impersonate"), msg)
			r.event(obj, obj.Status.LastAppliedRevision, eventv1.EventSeverityError, msg, nil)
		}
	}

	// Remove our finalizer from the list and update it
	controllerutil.RemoveFinalizer(obj, aiv1a1.LanguageModelFinalizer)
	// Stop reconciliation as the object is being deleted
	return ctrl.Result{}, nil
}

func (r *LanguageModelReconciler) event(obj *aiv1a1.LanguageModel,
	revision, severity, msg string,
	metadata map[string]string) {
	if metadata == nil {
		metadata = map[string]string{}
	}
	if revision != "" {
		metadata[aiv1a1.GroupVersion.Group+"/revision"] = revision
	}

	reason := severity
	conditions.GetReason(obj, meta.ReadyCondition)
	if r := conditions.GetReason(obj, meta.ReadyCondition); r != "" {
		reason = r
	}

	eventType := "Normal"
	if severity == eventv1.EventSeverityError {
		eventType = "Warning"
	}

	r.EventRecorder.AnnotatedEventf(obj, metadata, eventType, reason, msg)
}

func (r *LanguageModelReconciler) getSource(ctx context.Context,
	obj *aiv1a1.LanguageModel) (sourcev1.Source, error) {
	var src sourcev1.Source
	sourceNamespace := obj.GetNamespace()
	if obj.Spec.SourceRef.Namespace != "" {
		sourceNamespace = obj.Spec.SourceRef.Namespace
	}
	namespacedName := types.NamespacedName{
		Namespace: sourceNamespace,
		Name:      obj.Spec.SourceRef.Name,
	}

	if r.NoCrossNamespaceRefs && sourceNamespace != obj.GetNamespace() {
		return src, acl.AccessDeniedError(
			fmt.Sprintf("can't access '%s/%s', cross-namespace references have been blocked",
				obj.Spec.SourceRef.Kind, namespacedName))
	}

	switch obj.Spec.SourceRef.Kind {
	case sourcev1b2.OCIRepositoryKind:
		var repository sourcev1b2.OCIRepository
		err := r.Client.Get(ctx, namespacedName, &repository)
		if err != nil {
			if apierrors.IsNotFound(err) {
				return src, err
			}
			return src, fmt.Errorf("unable to get source '%s': %w", namespacedName, err)
		}
		src = &repository
	case sourcev1.GitRepositoryKind:
		var repository sourcev1.GitRepository
		err := r.Client.Get(ctx, namespacedName, &repository)
		if err != nil {
			if apierrors.IsNotFound(err) {
				return src, err
			}
			return src, fmt.Errorf("unable to get source '%s': %w", namespacedName, err)
		}
		src = &repository
	case sourcev1b2.BucketKind:
		var bucket sourcev1b2.Bucket
		err := r.Client.Get(ctx, namespacedName, &bucket)
		if err != nil {
			if apierrors.IsNotFound(err) {
				return src, err
			}
			return src, fmt.Errorf("unable to get source '%s': %w", namespacedName, err)
		}
		src = &bucket
	default:
		return src, fmt.Errorf("source `%s` kind '%s' not supported",
			obj.Spec.SourceRef.Name, obj.Spec.SourceRef.Kind)
	}
	return src, nil
}

func (r *LanguageModelReconciler) reconcile(ctx context.Context, obj *aiv1a1.LanguageModel, src sourcev1.Source, patcher *patch.SerialPatcher) error {
	log := ctrl.LoggerFrom(ctx)

	// Update status with the reconciliation progress.
	revision := src.GetArtifact().Revision
	progressingMsg := fmt.Sprintf("Fetching manifests for revision %s with a timeout of %s", revision, obj.GetTimeout().String())
	conditions.MarkUnknown(obj, meta.ReadyCondition, meta.ProgressingReason, "Reconciliation in progress")
	conditions.MarkReconciling(obj, meta.ProgressingReason, progressingMsg)
	if err := r.patch(ctx, obj, patcher); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	log.Info("Reconciling", "revision", revision)

	// Create a snapshot of the current inventory.
	oldInventory := inventory.New()
	if obj.Status.Inventory != nil {
		obj.Status.Inventory.DeepCopyInto(oldInventory)
	}

	// Report progress and set last attempted revision in status.
	obj.Status.LastAttemptedRevision = revision
	progressingMsg = fmt.Sprintf("Building manifests for revision %s with a timeout of %s", revision, obj.GetTimeout().String())
	conditions.MarkReconciling(obj, meta.ProgressingReason, progressingMsg)
	if err := r.patch(ctx, obj, patcher); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	// Configure the Kubernetes client for impersonation.
	impersonation := runtimeClient.NewImpersonator(
		r.Client,
		r.StatusPoller,
		r.PollingOpts,
		nil, // TODO support multi-cluster via obj.Spec.KubeConfig,
		r.KubeConfigOpts,
		r.DefaultServiceAccount,
		obj.Spec.ServiceAccountName,
		obj.GetNamespace(),
	)

	log.Info("Impersonating", "revision", revision)

	// Create the Kubernetes client that runs under impersonation.
	kubeClient, statusPoller, err := impersonation.GetClient(ctx)
	if err != nil {
		conditions.MarkFalse(obj, meta.ReadyCondition, aiv1a1.ReconciliationFailedReason, err.Error())
		return fmt.Errorf("failed to build kube client: %w", err)
	}

	log.Info("Building manifests", "revision", revision)

	objects, err := r.build(obj, src.GetArtifact().URL, src.GetArtifact().Metadata)
	if err != nil {
		conditions.MarkFalse(obj, meta.ReadyCondition, aiv1a1.ReconciliationFailedReason, err.Error())
		return fmt.Errorf("failed to build resources: %w", err)
	}
	log.Info("Objects built", "revision", revision, "objects", len(objects))

	log.Info("Applying manifests", "revision", revision, "objects", len(objects))

	// Create the server-side apply manager.
	resourceManager := ssa.NewResourceManager(kubeClient, statusPoller, ssa.Owner{
		Field: r.ControllerName,
		Group: aiv1a1.GroupVersion.Group,
	})
	resourceManager.SetOwnerLabels(objects, obj.GetName(), obj.GetNamespace())
	resourceManager.SetConcurrency(r.ConcurrentSSA)

	// Update status with the reconciliation progress.
	progressingMsg = fmt.Sprintf("Detecting drift for revision %s with a timeout of %s", revision, obj.GetTimeout().String())
	conditions.MarkReconciling(obj, meta.ProgressingReason, progressingMsg)
	if err := r.patch(ctx, obj, patcher); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	// Validate and apply resources in stages.
	drifted, changeSet, err := r.apply(ctx, resourceManager, obj, revision, objects)
	if err != nil {
		conditions.MarkFalse(obj, meta.ReadyCondition, aiv1a1.ReconciliationFailedReason, err.Error())
		return err
	}

	log.Info("Detecting drift", "revision", revision, "drifted", drifted)
	log.Info("Applied manifests", "revision", revision, "objects", len(objects), "changeSet", changeSet)

	// Create an inventory from the reconciled resources.
	newInventory := inventory.New()
	err = inventory.AddChangeSet(newInventory, changeSet)
	if err != nil {
		conditions.MarkFalse(obj, meta.ReadyCondition, aiv1a1.ReconciliationFailedReason, err.Error())
		return err
	}

	// Set last applied inventory in status.
	obj.Status.Inventory = newInventory

	// Detect stale resources which are subject to garbage collection.
	staleObjects, err := inventory.Diff(oldInventory, newInventory)
	if err != nil {
		conditions.MarkFalse(obj, meta.ReadyCondition, aiv1a1.ReconciliationFailedReason, err.Error())
		return err
	}

	log.Info("Detecting stale resources", "revision", revision, "staleObjects", len(staleObjects))

	// Run garbage collection for stale resources that do not have pruning disabled.
	if _, err := r.prune(ctx, resourceManager, obj, revision, staleObjects); err != nil {
		conditions.MarkFalse(obj, meta.ReadyCondition, aiv1a1.PruneFailedReason, err.Error())
		return err
	}

	// Run the health checks for the last applied resources.
	isNewRevision := !src.GetArtifact().HasRevision(obj.Status.LastAppliedRevision)
	if err := r.checkHealth(ctx,
		resourceManager,
		patcher,
		obj,
		revision,
		isNewRevision,
		drifted,
		changeSet.ToObjMetadataSet()); err != nil {
		conditions.MarkFalse(obj, meta.ReadyCondition, aiv1a1.HealthCheckFailedReason, err.Error())
		return err
	}

	// Set last applied revision.
	obj.Status.LastAppliedRevision = revision

	// Mark the object as ready.
	conditions.MarkTrue(obj,
		meta.ReadyCondition,
		aiv1a1.ReconciliationSucceededReason,
		fmt.Sprintf("Applied revision: %s", revision))

	return nil
}

func roundGigaQuantity(input string, safetyFactor float64) (string, error) {
	quantity, err := resource.ParseQuantity(input)
	if err != nil {
		return "", err
	}

	// Extract the numeric value and round it
	// value := quantity.AsApproximateFloat64()
	value := float64(quantity.ScaledValue(resource.Giga))
	roundedValue := math.Ceil(value * safetyFactor)
	if quantity.Format == resource.BinarySI {
		return fmt.Sprintf("%dGi", int64(roundedValue)), nil
	} else if quantity.Format == resource.DecimalSI {
		return fmt.Sprintf("%dG", int64(roundedValue)), nil
	} else {
		return "", fmt.Errorf("unknown quantity format %s", quantity.Format)
	}
}

func (r *LanguageModelReconciler) build(obj *aiv1a1.LanguageModel, url string, metadata map[string]string) ([]*unstructured.Unstructured, error) {
	usePv := obj.Spec.ModelCacheStrategy == aiv1a1.ModelCacheStrategyPV
	switch obj.Spec.Engine.DeploymentType {
	case aiv1a1.DeploymentTypeDefault:
		return r.buildKubernetes(obj, url, metadata, usePv)
	case aiv1a1.DeploymentTypeKubernetesDeployment:
		return r.buildKubernetes(obj, url, metadata, usePv)
	case aiv1a1.DeploymentTypeKnativeService:
		return r.buildKnative(obj, url, metadata)
		// case aiv1a1.DeploymentTypeKFServing:
		//	return r.buildKFServing(ctx, obj, url)
		// case aiv1a1.DeploymentTypeSeldon:
		//	return r.buildSeldon(ctx, obj, url)
	}

	return nil, fmt.Errorf("unknown deployment type %s", obj.Spec.Engine.DeploymentType)
}

func (r *LanguageModelReconciler) buildKubernetes(obj *aiv1a1.LanguageModel, url string, metadata map[string]string, usePvc bool) ([]*unstructured.Unstructured, error) {
	var (
		pvc               *corev1.PersistentVolumeClaim
		modelVolumeSource corev1.VolumeSource
	)

	deploymentStrategy := appsv1.RollingUpdateDeploymentStrategyType
	if usePvc {
		deploymentStrategy = appsv1.RecreateDeploymentStrategyType
		pvc = &corev1.PersistentVolumeClaim{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "PersistentVolumeClaim",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-pvc-model", obj.Name),
				Namespace: obj.Namespace,
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{
					"ReadWriteOnce",
				},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						StorageResourceName: resource.MustParse("10Gi"), // TODO
					},
				},
				StorageClassName: obj.Spec.Engine.StorageClass,
			},
		}
		modelVolumeSource = corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: pvc.Name,
			},
		}
	} else {
		modelVolumeSource = corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		}
	}

	svc := corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      obj.Name,
			Namespace: obj.Namespace,
		},
		Spec: corev1.ServiceSpec{

			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       8000,
					TargetPort: intstr.FromInt(8000),
				},
			},
			Selector: map[string]string{
				"app": obj.Name,
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	resourceRequirements := obj.Spec.Engine.Resources.DeepCopy()
	if resourceRequirements.Requests == nil {
		resourceRequirements.Requests = corev1.ResourceList{}
	}
	if resourceRequirements.Limits == nil {
		resourceRequirements.Limits = corev1.ResourceList{}
	}
	resourceRequirements.Limits[CPUResourceName] = resourceRequirements.Requests[CPUResourceName]

	// we need to leave some cpu for the web server, so rounding up.
	cpu := resourceRequirements.Requests.Cpu()
	threads := int(math.Ceil(cpu.AsApproximateFloat64())) - 1
	if threads < 1 {
		threads = 1
	}

	if sizeOnDisk, ok := metadata[MetadataSizeOnDisk]; ok {
		s, err := roundGigaQuantity(sizeOnDisk, SizeOnDiskSafetyFactor)
		if err != nil {
			return nil, err
		}

		q, err := resource.ParseQuantity(s)
		if err != nil {
			return nil, err
		}

		resourceRequirements.Requests[EphemeralStorageResourceName] = q
		resourceRequirements.Limits[EphemeralStorageResourceName] = q
	}

	if memoryRequired, ok := metadata[MetadataMemoryRequired]; ok {
		s, err := roundGigaQuantity(memoryRequired, MemorySafetyFactor)
		if err != nil {
			return nil, err
		}

		q, err := resource.ParseQuantity(s)
		if err != nil {
			return nil, err
		}

		resourceRequirements.Requests[MemoryResourceName] = q
		resourceRequirements.Limits[MemoryResourceName] = q
	}

	env := []corev1.EnvVar{}
	env = append(env, corev1.EnvVar{Name: EnvVarModelName, Value: obj.Spec.SourceRef.Name})
	if val, ok := metadata[MetadataModelDescription]; ok {
		env = append(env, corev1.EnvVar{Name: EnvVarModelDescription, Value: val})
	}
	if val, ok := metadata[MetadataContextSize]; ok {
		env = append(env, corev1.EnvVar{Name: EnvVarContextSize, Value: val})
	}
	if val, ok := metadata[MetadataFamily]; ok {
		env = append(env, corev1.EnvVar{Name: EnvVarFamily, Value: val})
	}
	if val, ok := metadata[MetadataFormat]; ok {
		env = append(env, corev1.EnvVar{Name: EnvVarFormat, Value: val})
	}
	if val, ok := metadata[MetadataQuantization]; ok {
		env = append(env, corev1.EnvVar{Name: EnvVarQuantization, Value: val})
	}
	if val, ok := metadata[MetadataPromptTemplate]; ok {
		env = append(env, corev1.EnvVar{Name: EnvVarPromptTemplate, Value: val})
	}
	// TODO implement stop words in weave-ai/model
	// what's the best way to pass this in?
	// if val, ok := metadata[MetadataStopWords]; ok {
	//	env = append(env, corev1.EnvVar{Name: EnvVarStopWords, Value: val})
	// }

	chatFormat := getChatFormatFromModelFamily(metadata[MetadataFamily])

	deploy := appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-deployment", obj.Name),
			Namespace: obj.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: obj.Spec.Engine.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": obj.Name,
				},
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: deploymentStrategy,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": obj.Name,
					},
				},
				Spec: corev1.PodSpec{
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser:  &[]int64{65532}[0],
						RunAsGroup: &[]int64{65532}[0],
						FSGroup:    &[]int64{65532}[0],
					},
					NodeSelector: obj.Spec.Engine.NodeSelector,
					InitContainers: []corev1.Container{
						{
							Name:  "model-downloader",
							Image: aiv1a1.ImageBlobDownloader,
							Env: []corev1.EnvVar{
								{
									Name:  "BLOB_URL",
									Value: url,
								},
								{
									Name:  "DESTINATION",
									Value: "/models/model.gguf",
								},
							},
							Resources: *resourceRequirements,
							VolumeMounts: []corev1.VolumeMount{
								{
									MountPath: "/models",
									Name:      "model-volume",
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "engine",
							Image: obj.Spec.GetEngineImage(),
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8000,
								},
							},
							Args: []string{
								"--host",
								"0.0.0.0",
								"--port",
								"8000",
								"--model",
								"/models/model.gguf",
								"--model_alias",
								obj.Name,
								"--chat_format",
								chatFormat,
								"--n_threads",
								strconv.Itoa(threads),
								"--n_threads_batch",
								strconv.Itoa(threads),
							},
							Env:       env,
							Resources: *resourceRequirements,
							SecurityContext: &corev1.SecurityContext{
								Capabilities: &corev1.Capabilities{
									Drop: []corev1.Capability{
										"NET_RAW",
									},
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									MountPath: "/models",
									Name:      "model-volume",
									ReadOnly:  true,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name:         "model-volume",
							VolumeSource: modelVolumeSource,
						},
					},
				},
			},
		},
	}

	objects := []*unstructured.Unstructured{
		convertToUnstructured(&deploy),
		convertToUnstructured(&svc),
	}
	if usePvc {
		objects = append(objects, convertToUnstructured(pvc))
	}
	return objects, nil
}

func (r *LanguageModelReconciler) buildKnative(obj *aiv1a1.LanguageModel, url string, _ map[string]string) ([]*unstructured.Unstructured, error) {
	service := servingv1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "serving.knative.dev/v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-serve", obj.Name),
			Namespace: obj.Namespace,
		},
		Spec: servingv1.ServiceSpec{
			ConfigurationSpec: servingv1.ConfigurationSpec{
				Template: servingv1.RevisionTemplateSpec{
					Spec: servingv1.RevisionSpec{
						PodSpec: corev1.PodSpec{
							InitContainers: []corev1.Container{
								{
									Name:  "model-downloader",
									Image: aiv1a1.ImageBlobDownloader,
									Env: []corev1.EnvVar{
										{
											Name:  "BLOB_URL",
											Value: url,
										},
										{
											Name:  "DESTINATION",
											Value: "/models/model.gguf",
										},
									},
									VolumeMounts: []corev1.VolumeMount{
										{
											MountPath: "/models",
											Name:      "model-volume",
										},
									},
								},
							},
							Containers: []corev1.Container{
								{
									Name:  "engine",
									Image: obj.Spec.GetEngineImage(),
									Ports: []corev1.ContainerPort{
										{
											ContainerPort: 8000,
										},
									},
									Args: []string{
										"--model",
										"/models/model.gguf",
										"--model_alias",
										obj.Name,
									},
									Resources: obj.Spec.Engine.Resources,
									SecurityContext: &corev1.SecurityContext{
										Capabilities: &corev1.Capabilities{
											Drop: []corev1.Capability{
												"NET_RAW",
											},
										},
									},
									VolumeMounts: []corev1.VolumeMount{
										{
											MountPath: "/models",
											Name:      "model-volume",
											ReadOnly:  true,
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "model-volume",
									VolumeSource: corev1.VolumeSource{
										EmptyDir: &corev1.EmptyDirVolumeSource{},
									},
								},
							},
						},
						// TODO parameterize concurrency
						ContainerConcurrency: ptr.Int64(0),
						// TODO parameterize timeout
						TimeoutSeconds: ptr.Int64(300),
					},
				},
			},
		},
	}
	objects := []*unstructured.Unstructured{
		convertToUnstructured(&service),
	}
	return objects, nil
}

func convertToUnstructured(obj runtime.Object) *unstructured.Unstructured {
	m, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil
	}
	return &unstructured.Unstructured{Object: m}
}

func (r *LanguageModelReconciler) apply(ctx context.Context,
	manager *ssa.ResourceManager,
	obj *aiv1a1.LanguageModel,
	revision string,
	objects []*unstructured.Unstructured) (bool, *ssa.ChangeSet, error) {
	log := ctrl.LoggerFrom(ctx)

	if err := ssa.SetNativeKindsDefaults(objects); err != nil {
		return false, nil, err
	}

	// TODO support common metadata
	// if meta := obj.Spec.CommonMetadata; meta != nil {
	// ssa.SetCommonMetadata(objects, meta.Labels, meta.Annotations)
	// }

	applyOpts := ssa.DefaultApplyOptions()
	applyOpts.Force = obj.Spec.Force
	applyOpts.ExclusionSelector = map[string]string{
		fmt.Sprintf("%s/reconcile", aiv1a1.GroupVersion.Group): aiv1a1.DisabledValue,
		fmt.Sprintf("%s/ssa", aiv1a1.GroupVersion.Group):       aiv1a1.IgnoreValue,
	}
	applyOpts.IfNotPresentSelector = map[string]string{
		fmt.Sprintf("%s/ssa", aiv1a1.GroupVersion.Group): aiv1a1.IfNotPresentValue,
	}
	applyOpts.ForceSelector = map[string]string{
		fmt.Sprintf("%s/force", aiv1a1.GroupVersion.Group): aiv1a1.EnabledValue,
	}

	applyOpts.Cleanup = ssa.ApplyCleanupOptions{
		Annotations: []string{
			// remove the kubectl annotation
			corev1.LastAppliedConfigAnnotation,
			// remove deprecated fluxcd.io annotations
			"ai.contrib.fluxcd.io/checksum",
			"fluxcd.io/sync-checksum",
		},
		Labels: []string{
			// remove deprecated fluxcd.io labels
			"fluxcd.io/sync-gc-mark",
		},
		FieldManagers: []ssa.FieldManager{
			{
				// to undo changes made with 'kubectl apply --server-side --force-conflicts'
				Name:          "kubectl",
				OperationType: metav1.ManagedFieldsOperationApply,
			},
			{
				// to undo changes made with 'kubectl apply'
				Name:          "kubectl",
				OperationType: metav1.ManagedFieldsOperationUpdate,
			},
			{
				// to undo changes made with 'kubectl apply'
				Name:          "before-first-apply",
				OperationType: metav1.ManagedFieldsOperationUpdate,
			},
			{
				// to undo changes made by the controller before SSA
				Name:          r.ControllerName,
				OperationType: metav1.ManagedFieldsOperationUpdate,
			},
		},
		Exclusions: map[string]string{
			fmt.Sprintf("%s/ssa", aiv1a1.GroupVersion.Group): aiv1a1.MergeValue,
		},
	}

	// contains only CRDs and Namespaces
	var defStage []*unstructured.Unstructured

	// contains only Kubernetes Class types e.g.: RuntimeClass, PriorityClass,
	// StorageClass, VolumeSnapshotClass, IngressClass, GatewayClass, ClusterClass, etc
	var classStage []*unstructured.Unstructured

	// contains all objects except for CRDs, Namespaces and Class type objects
	var resStage []*unstructured.Unstructured

	// contains the objects' metadata after apply
	resultSet := ssa.NewChangeSet()

	for _, u := range objects {
		switch {
		case ssa.IsClusterDefinition(u):
			defStage = append(defStage, u)
		case strings.HasSuffix(u.GetKind(), "Class"):
			classStage = append(classStage, u)
		default:
			resStage = append(resStage, u)
		}

	}

	var changeSetLog strings.Builder

	// validate, apply and wait for CRDs and Namespaces to register
	if len(defStage) > 0 {
		changeSet, err := manager.ApplyAll(ctx, defStage, applyOpts)
		if err != nil {
			return false, nil, err
		}

		if changeSet != nil && len(changeSet.Entries) > 0 {
			resultSet.Append(changeSet.Entries)

			log.Info("server-side apply for cluster definitions completed", "output", changeSet.ToMap())
			for _, change := range changeSet.Entries {
				if HasChanged(change.Action) {
					changeSetLog.WriteString(change.String() + "\n")
				}
			}

			if err := manager.WaitForSet(changeSet.ToObjMetadataSet(), ssa.WaitOptions{
				Interval: 2 * time.Second,
				Timeout:  obj.GetTimeout(),
			}); err != nil {
				return false, nil, err
			}
		}
	}

	// validate, apply and wait for Class type objects to register
	if len(classStage) > 0 {
		changeSet, err := manager.ApplyAll(ctx, classStage, applyOpts)
		if err != nil {
			return false, nil, err
		}

		if changeSet != nil && len(changeSet.Entries) > 0 {
			resultSet.Append(changeSet.Entries)

			log.Info("server-side apply for cluster class types completed", "output", changeSet.ToMap())
			for _, change := range changeSet.Entries {
				if HasChanged(change.Action) {
					changeSetLog.WriteString(change.String() + "\n")
				}
			}

			if err := manager.WaitForSet(changeSet.ToObjMetadataSet(), ssa.WaitOptions{
				Interval: 2 * time.Second,
				Timeout:  obj.GetTimeout(),
			}); err != nil {
				return false, nil, err
			}
		}
	}

	// sort by kind, validate and apply all the others objects
	sort.Sort(ssa.SortableUnstructureds(resStage))
	if len(resStage) > 0 {
		changeSet, err := manager.ApplyAll(ctx, resStage, applyOpts)
		if err != nil {
			return false, nil, fmt.Errorf("%w\n%s", err, changeSetLog.String())
		}

		if changeSet != nil && len(changeSet.Entries) > 0 {
			resultSet.Append(changeSet.Entries)

			log.Info("server-side apply completed", "output", changeSet.ToMap(), "revision", revision)
			for _, change := range changeSet.Entries {
				if HasChanged(change.Action) {
					changeSetLog.WriteString(change.String() + "\n")
				}
			}
		}
	}

	// emit event only if the server-side apply resulted in changes
	applyLog := strings.TrimSuffix(changeSetLog.String(), "\n")
	if applyLog != "" {
		r.event(obj, revision, eventv1.EventSeverityInfo, applyLog, nil)
	}

	return applyLog != "", resultSet, nil
}

func (r *LanguageModelReconciler) checkHealth(ctx context.Context,
	manager *ssa.ResourceManager,
	patcher *patch.SerialPatcher,
	obj *aiv1a1.LanguageModel,
	revision string,
	isNewRevision bool,
	drifted bool,
	objects object.ObjMetadataSet) error {

	checkStart := time.Now()
	if len(objects) == 0 {
		conditions.Delete(obj, aiv1a1.HealthyCondition)
		return nil
	}

	// Guard against deadlock (waiting on itself).
	var toCheck []object.ObjMetadata
	for _, o := range objects {
		if o.GroupKind.Kind == aiv1a1.LanguageModelKind &&
			o.Name == obj.GetName() &&
			o.Namespace == obj.GetNamespace() {
			continue
		}
		toCheck = append(toCheck, o)
	}

	// Find the previous health check result.
	wasHealthy := apimeta.IsStatusConditionTrue(obj.Status.Conditions, aiv1a1.HealthyCondition)

	// Update status with the reconciliation progress.
	message := fmt.Sprintf("Running health checks for revision %s with a timeout of %s", revision, obj.GetTimeout().String())
	conditions.MarkReconciling(obj, meta.ProgressingReason, message)
	conditions.MarkUnknown(obj, aiv1a1.HealthyCondition, meta.ProgressingReason, message)
	if err := r.patch(ctx, obj, patcher); err != nil {
		return fmt.Errorf("unable to update the healthy status to progressing: %w", err)
	}

	// Check the health with a default timeout of 30sec shorter than the reconciliation interval.
	if err := manager.WaitForSet(toCheck, ssa.WaitOptions{
		Interval: 5 * time.Second,
		Timeout:  obj.GetTimeout(),
		FailFast: false, // r.FailFast,
	}); err != nil {
		conditions.MarkFalse(obj, meta.ReadyCondition, aiv1a1.HealthCheckFailedReason, err.Error())
		conditions.MarkFalse(obj, aiv1a1.HealthyCondition, aiv1a1.HealthCheckFailedReason, err.Error())
		return fmt.Errorf("health check failed after %s: %w", time.Since(checkStart).String(), err)
	}

	// Emit recovery event if the previous health check failed.
	msg := fmt.Sprintf("Health check passed in %s", time.Since(checkStart).String())
	if !wasHealthy || (isNewRevision && drifted) {
		r.event(obj, revision, eventv1.EventSeverityInfo, msg, nil)
	}

	conditions.MarkTrue(obj, aiv1a1.HealthyCondition, meta.SucceededReason, msg)
	if err := r.patch(ctx, obj, patcher); err != nil {
		return fmt.Errorf("unable to update the healthy status to progressing: %w", err)
	}

	return nil
}

func (r *LanguageModelReconciler) prune(ctx context.Context,
	manager *ssa.ResourceManager,
	obj *aiv1a1.LanguageModel,
	revision string,
	objects []*unstructured.Unstructured) (bool, error) {
	if !obj.Spec.Prune {
		return false, nil
	}

	log := ctrl.LoggerFrom(ctx)

	opts := ssa.DeleteOptions{
		PropagationPolicy: metav1.DeletePropagationBackground,
		Inclusions:        manager.GetOwnerLabels(obj.Name, obj.Namespace),
		Exclusions: map[string]string{
			fmt.Sprintf("%s/prune", aiv1a1.GroupVersion.Group):     aiv1a1.DisabledValue,
			fmt.Sprintf("%s/reconcile", aiv1a1.GroupVersion.Group): aiv1a1.DisabledValue,
		},
	}

	changeSet, err := manager.DeleteAll(ctx, objects, opts)
	if err != nil {
		return false, err
	}

	// emit event only if the prune operation resulted in changes
	if changeSet != nil && len(changeSet.Entries) > 0 {
		log.Info(fmt.Sprintf("garbage collection completed: %s", changeSet.String()))
		r.event(obj, revision, eventv1.EventSeverityInfo, changeSet.String(), nil)
		return true, nil
	}

	return false, nil
}

// HasChanged evaluates the given action and returns true
// if the action type matches a resource mutation or deletion.
func HasChanged(action ssa.Action) bool {
	switch action {
	case ssa.SkippedAction:
		return false
	case ssa.UnchangedAction:
		return false
	default:
		return true
	}
}
