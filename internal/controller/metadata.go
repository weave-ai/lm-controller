package controller

const (
	CPUResourceName              = "cpu"
	MemoryResourceName           = "memory"
	StorageResourceName          = "storage"
	EphemeralStorageResourceName = "ephemeral-storage"
)

const (
	MetadataContextSize      = "ai.contrib.fluxcd.io/context-size"
	MetadataFamily           = "ai.contrib.fluxcd.io/family"
	MetadataFormat           = "ai.contrib.fluxcd.io/format"
	MetadataMemoryRequired   = "ai.contrib.fluxcd.io/memory-required"
	MetadataModelDescription = "ai.contrib.fluxcd.io/model-description"
	MetadataPromptTemplate   = "ai.contrib.fluxcd.io/prompt-template"
	MetadataQuantization     = "ai.contrib.fluxcd.io/quantization"
	MetadataSizeOnDisk       = "ai.contrib.fluxcd.io/size-on-disk"
	MetadataStopWords        = "ai.contrib.fluxcd.io/stop-words"
)

const (
	SizeOnDiskSafetyFactor = 1.20
	MemorySafetyFactor     = 1.00
)

const (
	EnvVarContextSize      = "LM_CTRL_CONTEXT_SIZE"
	EnvVarFamily           = "LM_CTRL_FAMILY"
	EnvVarFormat           = "LM_CTRL_FORMAT"
	EnvVarModelDescription = "LM_CTRL_MODEL_DESCRIPTION"
	EnvVarModelName        = "LM_CTRL_MODEL_NAME"
	EnvVarPromptTemplate   = "LM_CTRL_PROMPT_TEMPLATE"
	EnvVarQuantization     = "LM_CTRL_QUANTIZATION"
	EnvVarStopWords        = "LM_CTRL_STOP_WORDS"
)
