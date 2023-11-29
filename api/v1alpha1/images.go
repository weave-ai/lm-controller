package v1alpha1

const (
	ImageBlobDownloader       = "ghcr.io/llm-gitops/blob-downloader/blob-downloader:v1@sha256:74b65fd0095a66ce5e5a04134cb2db08c62fc7d08cbb4e80f6ea31df4f680742"
	ImageEngineLlamaCppPython = "ghcr.io/weave-ai/serve/llama-cpp-python:v0.2.20-avx2"
	// Variants are
	// - CPU noavx
	// - CPU avx
	// - CPU avx2
	// - CPU avx512
	// - GPU TODO
)
