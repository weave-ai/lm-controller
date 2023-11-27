docker_build('ghcr.io/weave-ai/lm-controller', '.')

yaml = kustomize('./config/manager')
k8s_yaml(yaml)
