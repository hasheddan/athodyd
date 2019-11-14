module github.com/hasheddan/athodyd

go 1.12

replace (
	k8s.io/api => k8s.io/api v0.0.0-20190918195907-bd6ac527cfd2
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190918201827-3de75813f604
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190817020851-f2f3a405f61d
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20190918200908-1e17798da8c1
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190918200256-06eb1244587a
)

require (
	github.com/crossplaneio/crossplane v0.4.0
	github.com/crossplaneio/crossplane-runtime v0.2.1
	github.com/crossplaneio/stack-gcp v0.2.1
	github.com/google/go-cmp v0.3.1 // indirect
	github.com/pkg/errors v0.8.1
	github.com/stretchr/testify v1.4.0 // indirect
	golang.org/x/net v0.0.0-20190827160401-ba9fcec4b297 // indirect
	gopkg.in/yaml.v2 v2.2.4 // indirect
	k8s.io/api v0.0.0-20190918195907-bd6ac527cfd2
	k8s.io/apiextensions-apiserver v0.0.0-20190918201827-3de75813f604
	k8s.io/apimachinery v0.0.0-20190913080033-27d36303b655
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	sigs.k8s.io/controller-runtime v0.3.0
	sigs.k8s.io/yaml v1.1.0
)
