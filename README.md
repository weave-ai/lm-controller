# Weave AI LM-Controller

Weave AI LM-Controller is a Flux controller that manages the lifecycle of 
Large Language Models (LLMs) on Kubernetes.

## Getting Started

### Prerequisites

- Kubernetes v1.27+
- Flux v2.1.0+

### Installation

LM-Controller should be installed as part of the Weave AI Controllers.
For development and testing, the standalone LM-Controller can be installed using the following commands:

```shell
flux install

VERSION=v0.2.1
kubectl create ns weave-ai
kubectl apply -f  https://github.com/weave-ai/lm-controller/releases/download/${VERSION}/lm-controller.crds.yaml
kubectl apply -f  https://github.com/weave-ai/lm-controller/releases/download/${VERSION}/lm-controller.rbac.yaml
kubectl apply -f  https://github.com/weave-ai/lm-controller/releases/download/${VERSION}/lm-controller.deployment.yaml
```