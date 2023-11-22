package controller

import (
	aiv1a1 "github.com/weave-ai/lm-controller/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const LanguageModelConfigMapSuffix = "-lmcm"

func GenerateConfigMap(lm *aiv1a1.LanguageModel) (*corev1.ConfigMap, error) {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      lm.Name + LanguageModelConfigMapSuffix,
			Namespace: lm.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: lm.APIVersion,
					Kind:       lm.Kind,
					Name:       lm.Name,
				},
			},
		},
	}

	return cm, nil
}

func generateConfigMapDataForLocalAI(lm *aiv1a1.LanguageModel) (map[string]string, error) {
	return nil, nil
}
