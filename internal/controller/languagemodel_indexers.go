package controller

import (
	"context"
	"fmt"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	aiv1a1 "github.com/weave-ai/lm-controller/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *LanguageModelReconciler) requestsForRevisionChangeOf(indexKey string) handler.MapFunc {
	return func(ctx context.Context, obj client.Object) []reconcile.Request {
		log := ctrl.LoggerFrom(ctx)
		repo, ok := obj.(interface {
			GetArtifact() *sourcev1.Artifact
		})
		if !ok {
			log.Error(fmt.Errorf("expected an object conformed with GetArtifact() method, but got a %T", obj),
				"failed to get reconcile requests for revision change")
			return nil
		}
		// If we do not have an artifact, we have no requests to make
		if repo.GetArtifact() == nil {
			return nil
		}

		var list aiv1a1.LanguageModelList
		if err := r.List(ctx, &list, client.MatchingFields{
			indexKey: client.ObjectKeyFromObject(obj).String(),
		}); err != nil {
			log.Error(err, "failed to list objects for revision change")
			return nil
		}
		reqs := make([]reconcile.Request, len(list.Items))
		for i := range list.Items {
			reqs[i].NamespacedName.Name = list.Items[i].Name
			reqs[i].NamespacedName.Namespace = list.Items[i].Namespace
		}
		return reqs
	}
}

func (r *LanguageModelReconciler) indexBy(kind string) func(o client.Object) []string {
	return func(o client.Object) []string {
		lm, ok := o.(*aiv1a1.LanguageModel)
		if !ok {
			panic(fmt.Sprintf("Expected a Kustomization, got %T", o))
		}

		if lm.Spec.SourceRef.Kind == kind {
			namespace := lm.GetNamespace()
			if lm.Spec.SourceRef.Namespace != "" {
				namespace = lm.Spec.SourceRef.Namespace
			}
			return []string{fmt.Sprintf("%s/%s", namespace, lm.Spec.SourceRef.Name)}
		}

		return nil
	}
}
