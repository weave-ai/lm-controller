module github.com/weave-ai/lm-controller

go 1.20

replace github.com/weave-ai/lm-controller/api => ./api

// Replace digest lib to master to gather access to BLAKE3.
// xref: https://github.com/opencontainers/go-digest/pull/66
replace github.com/opencontainers/go-digest => github.com/opencontainers/go-digest v1.0.1-0.20220411205349-bde1400a84be

require (
	github.com/fluxcd/pkg/apis/acl v0.1.0
	github.com/fluxcd/pkg/apis/event v0.5.2
	github.com/fluxcd/pkg/apis/meta v1.1.2
	github.com/fluxcd/pkg/http/fetch v0.6.0
	github.com/fluxcd/pkg/runtime v0.42.0
	github.com/fluxcd/pkg/ssa v0.32.1
	github.com/fluxcd/source-controller/api v1.1.2
	github.com/onsi/gomega v1.28.0
	github.com/spf13/pflag v1.0.5
	github.com/weave-ai/lm-controller/api v0.0.0-00010101000000-000000000000
	k8s.io/api v0.28.2
	k8s.io/apimachinery v0.28.2
	k8s.io/client-go v1.5.2
	k8s.io/utils v0.0.0-20230726121419-3b25d923346b
	knative.dev/pkg v0.0.0-20231023151236-29775d7c9e5c
	knative.dev/serving v0.39.0
	sigs.k8s.io/cli-utils v0.35.0
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

// Pin kustomize to v5.0.3
replace (
	sigs.k8s.io/kustomize/api => sigs.k8s.io/kustomize/api v0.13.4
	sigs.k8s.io/kustomize/kyaml => sigs.k8s.io/kustomize/kyaml v0.14.2
)

// Fix CVE-2022-28948
replace gopkg.in/yaml.v3 => gopkg.in/yaml.v3 v3.0.1

require (
	github.com/AdaLogics/go-fuzz-headers v0.0.0-20230811130428-ced1acdcaa24 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20230124172434-306776ec8161 // indirect
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blendle/zapdriver v1.3.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/chai2010/gettext-go v1.0.2 // indirect
	github.com/cyphar/filepath-securejoin v0.2.4 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/emicklei/go-restful/v3 v3.11.0 // indirect
	github.com/evanphx/json-patch v5.7.0+incompatible // indirect
	github.com/evanphx/json-patch/v5 v5.7.0 // indirect
	github.com/exponent-io/jsonpath v0.0.0-20210407135951-1de76d718b3f // indirect
	github.com/fatih/color v1.15.0 // indirect
	github.com/fluxcd/pkg/tar v0.3.0 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-errors/errors v1.5.1 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-logr/zapr v1.2.4 // indirect
	github.com/go-openapi/jsonpointer v0.20.0 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.22.4 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/btree v1.1.2 // indirect
	github.com/google/gnostic v0.7.0 // indirect
	github.com/google/gnostic-models v0.6.9-0.20230804172637-c7be7c783f49 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/go-containerregistry v0.13.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/google/uuid v1.3.1 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-hclog v1.2.1 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.4 // indirect
	github.com/imdario/mergo v0.3.16 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/cpuid/v2 v2.2.5 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/moby/spdystream v0.2.0 // indirect
	github.com/moby/term v0.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/go-digest/blake3 v0.0.0-20230920164345-5d0a5887d130 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.17.0 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.44.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/spf13/cobra v1.7.0 // indirect
	github.com/xlab/treeprint v1.2.0 // indirect
	github.com/zeebo/blake3 v0.2.3 // indirect
	go.starlark.net v0.0.0-20231013162135-47c85baa7a64 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.26.0 // indirect
	golang.org/x/net v0.17.0 // indirect
	golang.org/x/oauth2 v0.13.0 // indirect
	golang.org/x/sync v0.4.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
	golang.org/x/term v0.13.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	gomodules.xyz/jsonpatch/v2 v2.4.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/apiextensions-apiserver v0.28.2 // indirect
	k8s.io/cli-runtime v0.28.2 // indirect
	k8s.io/component-base v0.28.2 // indirect
	k8s.io/klog/v2 v2.100.1 // indirect
	k8s.io/kube-openapi v0.0.0-20231010175941-2dd684a91f00 // indirect
	k8s.io/kubectl v0.28.2 // indirect
	knative.dev/networking v0.0.0-20231017124814-2a7676e912b7 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/kustomize/api v0.14.0 // indirect
	sigs.k8s.io/kustomize/kyaml v0.14.3 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.3.0 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)
