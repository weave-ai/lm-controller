package inventory

import (
	"sort"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/cli-utils/pkg/object"

	"github.com/fluxcd/pkg/apis/meta"
	"github.com/fluxcd/pkg/ssa"

	aiv1a1 "github.com/weave-ai/lm-controller/api/v1alpha1"
)

func New() *aiv1a1.ResourceInventory {
	return &aiv1a1.ResourceInventory{
		Entries: []aiv1a1.ResourceRef{},
	}
}

// AddChangeSet extracts the metadata from the given objects and adds it to the inventory.
func AddChangeSet(inv *aiv1a1.ResourceInventory, set *ssa.ChangeSet) error {
	if set == nil {
		return nil
	}

	for _, entry := range set.Entries {
		inv.Entries = append(inv.Entries, aiv1a1.ResourceRef{
			ID:      entry.ObjMetadata.String(),
			Version: entry.GroupVersion,
		})
	}

	return nil
}

// List returns the inventory entries as unstructured.Unstructured objects.
func List(inv *aiv1a1.ResourceInventory) ([]*unstructured.Unstructured, error) {
	objects := make([]*unstructured.Unstructured, 0)

	if inv.Entries == nil {
		return objects, nil
	}

	for _, entry := range inv.Entries {
		objMetadata, err := object.ParseObjMetadata(entry.ID)
		if err != nil {
			return nil, err
		}

		u := &unstructured.Unstructured{}
		u.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   objMetadata.GroupKind.Group,
			Kind:    objMetadata.GroupKind.Kind,
			Version: entry.Version,
		})
		u.SetName(objMetadata.Name)
		u.SetNamespace(objMetadata.Namespace)
		objects = append(objects, u)
	}

	sort.Sort(ssa.SortableUnstructureds(objects))
	return objects, nil
}

// ListMetadata returns the inventory entries as object.ObjMetadata objects.
func ListMetadata(inv *aiv1a1.ResourceInventory) (object.ObjMetadataSet, error) {
	var metas []object.ObjMetadata
	for _, e := range inv.Entries {
		m, err := object.ParseObjMetadata(e.ID)
		if err != nil {
			return metas, err
		}
		metas = append(metas, m)
	}

	return metas, nil
}

// Diff returns the slice of objects that do not exist in the target inventory.
func Diff(inv *aiv1a1.ResourceInventory, target *aiv1a1.ResourceInventory) ([]*unstructured.Unstructured, error) {
	versionOf := func(i *aiv1a1.ResourceInventory, objMetadata object.ObjMetadata) string {
		for _, entry := range i.Entries {
			if entry.ID == objMetadata.String() {
				return entry.Version
			}
		}
		return ""
	}

	objects := make([]*unstructured.Unstructured, 0)
	aList, err := ListMetadata(inv)
	if err != nil {
		return nil, err
	}

	bList, err := ListMetadata(target)
	if err != nil {
		return nil, err
	}

	list := aList.Diff(bList)
	if len(list) == 0 {
		return objects, nil
	}

	for _, metadata := range list {
		u := &unstructured.Unstructured{}
		u.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   metadata.GroupKind.Group,
			Kind:    metadata.GroupKind.Kind,
			Version: versionOf(inv, metadata),
		})
		u.SetName(metadata.Name)
		u.SetNamespace(metadata.Namespace)
		objects = append(objects, u)
	}

	sort.Sort(ssa.SortableUnstructureds(objects))
	return objects, nil
}

// ReferenceToObjMetadataSet transforms a NamespacedObjectKindReference to an ObjMetadataSet.
func ReferenceToObjMetadataSet(cr []meta.NamespacedObjectKindReference) (object.ObjMetadataSet, error) {
	var objects []object.ObjMetadata

	for _, c := range cr {
		// For backwards compatibility with Kustomization v1beta1
		if c.APIVersion == "" {
			c.APIVersion = "apps/v1"
		}

		gv, err := schema.ParseGroupVersion(c.APIVersion)
		if err != nil {
			return objects, err
		}

		u := &unstructured.Unstructured{}
		u.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   gv.Group,
			Kind:    c.Kind,
			Version: gv.Version,
		})
		u.SetName(c.Name)
		if c.Namespace != "" {
			u.SetNamespace(c.Namespace)
		}

		objects = append(objects, object.UnstructuredToObjMetadata(u))

	}

	return objects, nil
}
