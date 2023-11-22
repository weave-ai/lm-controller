package controller

import (
	. "github.com/onsi/gomega"
	aiv1a1 "github.com/weave-ai/lm-controller/api/v1alpha1"

	"testing"
)

func Test_00200_GenerateConfigMap(t *testing.T) {
	g := NewGomegaWithT(t)

	By("creating a LanguageModel from bytes")
	spec := []byte(fixtureO1)
	lm := &aiv1a1.LanguageModel{}
	err := lm.FromBytes(testScheme, spec)
	g.Expect(err).NotTo(HaveOccurred())

	By("generating a ConfigMap from the LanguageModel")
	cm, err := GenerateConfigMap(lm)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cm.Name).To(Equal(lm.Name + LanguageModelConfigMapSuffix))
	g.Expect(cm.OwnerReferences[0].APIVersion).To(Equal("ai.contrib.fluxcd.io/v1alpha1"))
	g.Expect(cm.OwnerReferences[0].Name).To(Equal("llama-2-7b-chat"))
	g.Expect(cm.OwnerReferences[0].Kind).To(Equal("LanguageModel"))
}
