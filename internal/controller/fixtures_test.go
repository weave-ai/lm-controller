package controller

const fixtureO1 = `
apiVersion: ai.contrib.fluxcd.io/v1alpha1
kind: LanguageModel
metadata:
  name: llama-2-7b-chat
  namespace: default
spec:
  engine: 
    engineType: default
    deploymentType: default
  sourceRef:
    kind: OCIRepository
    name: llama-2-7b-chat
`

const fixtureO2 = `
apiVersion: ai.contrib.fluxcd.io/v1alpha1
kind: LanguageModel
metadata:
  name: llama-2-7b-chat
  namespace: default
spec:
  engine: 
    engineType: default
    deploymentType: default
  sourceRef:
    kind: OCIRepository
    name: llama-2-7b-chat
`
