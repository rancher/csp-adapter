module github.com/rancher/csp-adapter

go 1.22

replace (
	github.com/rancher/rancher/pkg/apis => github.com/rancher/rancher/pkg/apis v0.0.0-20220519154712-0e2fdc8060bc
	github.com/rancher/rancher/pkg/client => github.com/rancher/rancher/pkg/client v0.0.0-20220519154712-0e2fdc8060bc
	k8s.io/api => k8s.io/api v0.29.3
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.29.3
	k8s.io/apimachinery => k8s.io/apimachinery v0.29.3
	k8s.io/apiserver => k8s.io/apiserver v0.29.3
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.29.3
	k8s.io/client-go => github.com/rancher/client-go v1.29.3-rancher1
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.29.3
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.29.3
	k8s.io/code-generator => k8s.io/code-generator v0.29.3
	k8s.io/component-base => k8s.io/component-base v0.29.3
	k8s.io/cri-api => k8s.io/cri-api v0.29.3
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.29.3
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.29.3
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.29.3
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.29.3
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.29.3
	k8s.io/kubectl => k8s.io/kubectl v0.29.3
	k8s.io/kubelet => k8s.io/kubelet v0.29.3
	k8s.io/kubernetes => k8s.io/kubernetes v1.29.3
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.29.3
	k8s.io/metrics => k8s.io/metrics v0.29.3
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.29.3
)

require (
	github.com/aws/aws-sdk-go-v2/config v1.27.11
	github.com/aws/aws-sdk-go-v2/service/licensemanager v1.18.4
	github.com/aws/aws-sdk-go-v2/service/sts v1.28.6
	github.com/google/uuid v1.6.0
        github.com/prometheus/client_model v0.4.1-0.20230718164431-9a2bf3000d16 // indirect
	github.com/prometheus/common v0.44.0
        github.com/rancher/lasso v0.0.0-20240424194130-d87ec407d941 // indirect
        github.com/rancher/rancher/pkg/apis v0.0.0-20240425061024-5ce684a45887
        github.com/rancher/wrangler/v2 v2.2.0-rc6
	7ithub.com/sirupsen/logrus v1.9.3
	github.com/stretchr/testify v1.9.0
	k8s.io/api v0.28.8
	k8s.io/apimachinery v0.28.8
	k8s.io/client-go v12.0.0+incompatible
)

require (
	github.com/aws/aws-sdk-go-v2 v1.26.1 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.7 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.5 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.5 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.11.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.20.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.23.4 // indirect
	github.com/aws/smithy-go v1.20.2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
        github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
        github.com/emicklei/go-restful/v3 v3.11.0 // indirect
	github.com/evanphx/json-patch v5.7.0+incompatible // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-openapi/jsonpointer v0.19.6 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.22.10 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/gnostic-models v0.6.8 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/imdario/mergo v0.3.16 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.17.0 // indirect
	github.com/prometheus/procfs v0.11.1 // indirect
        github.com/rancher/aks-operator v1.9.0-rc.1
        github.com/rancher/eks-operator v1.9.0-rc.1
        github.com/rancher/fleet/pkg/apis v0.0.0-20231017140638-93432f288e79
        github.com/rancher/gke-operator v1.9.0-rc.1
        github.com/rancher/norman v0.0.0-20240503193601-9f5f6586bb5b
        github.com/rancher/rke 1.6.0-rc4
        github.com/rancher/wrangler/v2 v2.2.0-rc6
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/net v0.25.0 // indirect
	golang.org/x/oauth2 v0.20.0 // indirect
	golang.org/x/sync v0.7.0 // indirect
	golang.org/x/sys v0.20.0 // indirect
	golang.org/x/term v0.20.0 // indirect
	golang.org/x/text v0.15.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/apiextensions-apiserver v0.29.4 // indirect
	k8s.io/apiserver v0.29.4 // indirect
	k8s.io/component-base v0.28.8 // indirect
	k8s.io/klog v1.0.0 // indirect
	k8s.io/klog/v2 v2.120.1 // indirect
	k8s.io/kube-aggregator v0.29.3 // indirect
	k8s.io/kube-openapi v0.0.0-20230717233707-2695361300d9 // indirect
        k8s.io/utils v0.0.0-20240102154912-e7106e64919e
	sigs.k8s.io/cli-utils v0.35.0 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.4.1 // indirect
	sigs.k8s.io/yaml v1.4.0 // indirect
)
