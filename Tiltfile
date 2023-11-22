docker_build('ghcr.io/llm-gitops/lm-controller', '.')

yaml = kustomize('./config/manager')
k8s_yaml(yaml)
