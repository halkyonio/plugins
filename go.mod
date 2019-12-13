module halkyon.io/plugins

go 1.13

require (
	github.com/hashicorp/go-plugin v1.0.2-0.20191004171845-809113480b55
	halkyon.io/api v1.0.0-beta.7
	halkyon.io/operator-framework v0.0.0-20191212091852-c4cae77ce280
	k8s.io/apimachinery v0.0.0-20190817020851-f2f3a405f61d
)

replace k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d // kubernetes-1.14.5
