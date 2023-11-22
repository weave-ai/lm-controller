package v1alpha1

import (
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

func (in *LanguageModel) ToBytes(scheme *runtime.Scheme) ([]byte, error) {
	return runtime.Encode(
		serializer.NewCodecFactory(scheme).LegacyCodec(
			corev1.SchemeGroupVersion,
			GroupVersion,
			sourcev1.GroupVersion,
		), in)
}

func (in *LanguageModel) FromBytes(scheme *runtime.Scheme, b []byte) error {
	return runtime.DecodeInto(
		serializer.NewCodecFactory(scheme).LegacyCodec(
			corev1.SchemeGroupVersion,
			GroupVersion,
			sourcev1.GroupVersion,
		), b, in)
}
