module halkyon.io/plugins

go 1.13

require (
	github.com/hashicorp/go-plugin v1.0.2-0.20191004171845-809113480b55
	halkyon.io/api v1.0.0-beta.8.0.20191219201020-14dfa325eab8
	halkyon.io/operator-framework v0.0.0-20191220094645-78d6d8666a50
	k8s.io/apimachinery v0.17.0
	k8s.io/utils v0.0.0-20190712204705-3dccf664f023 // indirect
)

replace k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d // kubernetes-1.14.5
