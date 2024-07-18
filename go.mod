module github.com/rancher/csp-adapter

go 1.20

replace (
	github.com/rancher/rancher/pkg/apis => github.com/rancher/rancher/pkg/apis v0.0.0-20220519154712-0e2fdc8060bc
	github.com/rancher/rancher/pkg/client => github.com/rancher/rancher/pkg/client v0.0.0-20220519154712-0e2fdc8060bc
	k8s.io/api => k8s.io/api v0.27.16
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.27.16
	k8s.io/apimachinery => k8s.io/apimachinery v0.27.16
	k8s.io/apiserver => k8s.io/apiserver v0.27.16
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.27.16
	k8s.io/client-go => github.com/rancher/client-go v1.27.4-rancher1
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.27.16
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.27.16
	k8s.io/code-generator => k8s.io/code-generator v0.27.16
	k8s.io/component-base => k8s.io/component-base v0.27.16
	k8s.io/cri-api => k8s.io/cri-api v0.27.16
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.27.16
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.27.16
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.27.16
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.27.16
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.27.16
	k8s.io/kubectl => k8s.io/kubectl v0.27.16
	k8s.io/kubelet => k8s.io/kubelet v0.27.16
	k8s.io/kubernetes => k8s.io/kubernetes v1.27.4
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.27.16
	k8s.io/metrics => k8s.io/metrics v0.27.16
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.27.16
)

require (
	github.com/aws/aws-sdk-go-v2/config v1.18.25
	github.com/aws/aws-sdk-go-v2/service/licensemanager v1.18.4
	github.com/aws/aws-sdk-go-v2/service/sts v1.19.0
	github.com/google/uuid v1.3.0
	github.com/prometheus/client_model v0.4.0
	github.com/prometheus/common v0.43.0
	github.com/rancher/lasso v0.0.0-20230830164424-d684fdeb6f29
	github.com/rancher/rancher/pkg/apis v0.0.0
	github.com/rancher/wrangler v0.8.11
	github.com/sirupsen/logrus v1.9.0
	github.com/stretchr/testify v1.8.4
	k8s.io/api v0.27.16
	k8s.io/apimachinery v0.27.16
	k8s.io/client-go v12.0.0+incompatible
)

require (
	github.com/aws/aws-sdk-go-v2 v1.18.0 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.13.24 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.13.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.33 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.27 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.34 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.27 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.12.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.14.10 // indirect
	github.com/aws/smithy-go v1.13.5 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/emicklei/go-restful/v3 v3.9.0 // indirect
	github.com/evanphx/json-patch v4.12.0+incompatible // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-openapi/jsonpointer v0.19.6 // indirect
	github.com/go-openapi/jsonreference v0.20.1 // indirect
	github.com/go-openapi/swag v0.22.3 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/gnostic v0.5.7-v3refs // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.15.1 // indirect
	github.com/prometheus/procfs v0.9.0 // indirect
	github.com/rancher/aks-operator v1.0.5 // indirect
	github.com/rancher/eks-operator v1.1.3 // indirect
	github.com/rancher/fleet/pkg/apis v0.0.0-20210918015053-5a141a6b18f0 // indirect
	github.com/rancher/gke-operator v1.1.3 // indirect
	github.com/rancher/norman v0.0.0-20220406153559-82478fb169cb // indirect
	github.com/rancher/rke v1.3.11 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/net v0.23.0 // indirect
	golang.org/x/oauth2 v0.10.0 // indirect
	golang.org/x/sync v0.5.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/term v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/apiextensions-apiserver v0.23.1 // indirect
	k8s.io/apiserver v0.27.16 // indirect
	k8s.io/component-base v0.27.16 // indirect
	k8s.io/klog v1.0.0 // indirect
	k8s.io/klog/v2 v2.90.1 // indirect
	k8s.io/kube-aggregator v0.21.0 // indirect
	k8s.io/kube-openapi v0.0.0-20230501164219-8b0f38b5fd1f // indirect
	k8s.io/utils v0.0.0-20230209194617-a36077c30491 // indirect
	sigs.k8s.io/cli-utils v0.16.0 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)
