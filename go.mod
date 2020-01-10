module halkyon.io/plugins

go 1.13

require (
	github.com/go-logr/logr v0.1.0
	github.com/hashicorp/go-plugin v1.0.2-0.20191004171845-809113480b55
	halkyon.io/api v1.0.0-rc.1
	halkyon.io/operator-framework v1.0.0-beta.2
	k8s.io/apimachinery v0.17.0
	k8s.io/utils v0.0.0-20190712204705-3dccf664f023 // indirect
)

replace k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d // kubernetes-1.14.5
