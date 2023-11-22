module github.com/weave-ai/lm-controller/api

go 1.20

require (
	github.com/fluxcd/source-controller/api v1.1.2
	k8s.io/api v0.28.2
	// k8s.io/apiextensions-apiserver v0.27.4
	k8s.io/apimachinery v0.28.2
	sigs.k8s.io/controller-runtime v0.16.2
)

replace sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.15.1

// Pin k8s apis to v0.27.3
replace (
	k8s.io/api => k8s.io/api v0.27.4
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.27.4
	k8s.io/apimachinery => k8s.io/apimachinery v0.27.4
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.27.4
	k8s.io/client-go => k8s.io/client-go v0.27.4
	k8s.io/component-base => k8s.io/component-base v0.27.4
	k8s.io/kubectl => k8s.io/kubectl v0.27.4
)

// Fix CVE-2022-28948
replace gopkg.in/yaml.v3 => gopkg.in/yaml.v3 v3.0.1

require github.com/fluxcd/pkg/apis/meta v1.1.2

require (
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	golang.org/x/net v0.17.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/klog/v2 v2.100.1 // indirect
	k8s.io/utils v0.0.0-20230726121419-3b25d923346b // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.3.0 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)
