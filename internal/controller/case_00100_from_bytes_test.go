package controller

import (
	"testing"

	. "github.com/onsi/gomega"
	aiv1a1 "github.com/weave-ai/lm-controller/api/v1alpha1"
)

func Test_00100_FromBytesDefaults(t *testing.T) {
	g := NewGomegaWithT(t)

	By("creating a LanguageModel from bytes")

	spec := []byte(fixtureO1)
	lm := &aiv1a1.LanguageModel{}
	err := lm.FromBytes(testScheme, spec)
	g.Expect(err).NotTo(HaveOccurred())

	By("checking the LanguageModel's spec")
	g.Expect(lm.Spec.GetEngine().EngineType).To(Equal("default"))
	g.Expect(lm.Spec.GetEngine().DeploymentType).To(Equal("default"))
	g.Expect(lm.Spec.GetModelPullPolicy()).To(Equal("IfNotPresent"))
	g.Expect(lm.Spec.GetModelCacheStrategy()).To(Equal("None"))
	g.Expect(lm.Spec.SourceRef.Kind).To(Equal("OCIRepository"))
	g.Expect(lm.Spec.SourceRef.Name).To(Equal("llama-2-7b-chat"))
}

func Test_00100_FromBytes(t *testing.T) {
	g := NewGomegaWithT(t)

	By("creating a LanguageModel from bytes")

	spec := []byte(fixtureO2)
	lm := &aiv1a1.LanguageModel{}
	err := lm.FromBytes(testScheme, spec)
	g.Expect(err).NotTo(HaveOccurred())

	By("checking the LanguageModel's spec")
	g.Expect(lm.Spec.GetEngine().EngineType).To(Equal("default"))
	g.Expect(lm.Spec.GetEngine().DeploymentType).To(Equal("default"))
	g.Expect(lm.Spec.SourceRef.Kind).To(Equal("OCIRepository"))
	g.Expect(lm.Spec.SourceRef.Name).To(Equal("llama-2-7b-chat"))
	/*{
		params, err := lm.GetParameters()
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(params).To(HaveLen(4))
		g.Expect(params["temperature"]).To(Equal(0.9))
		g.Expect(params["max_tokens"]).To(Equal(64.0))
		g.Expect(params["top_p"]).To(Equal(1.0))
		g.Expect(params["top_k"]).To(Equal(40.0))
	}*/
	//g.Expect(lm.Spec.SystemPrompt).To(Equal("You are Weave AI, a human-like AI assistant that helps you with your daily tasks.\n"))
	//g.Expect(lm.Spec.PromptTemplates).To(HaveLen(1))
	//g.Expect(lm.Spec.PromptTemplates[0].Name).To(Equal("default"))
	//g.Expect(lm.Spec.PromptTemplates[0].Template).To(Equal("{{ .Input }}\n"))
}
