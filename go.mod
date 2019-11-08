module halkyon.io/plugins

go 1.13

require (
	halkyon.io/api v1.0.0-beta.7
	halkyon.io/operator-framework v0.0.0-20191108175501-3d0a053bc383
)

replace (
	k8s.io/api => k8s.io/api v0.0.0-20190805182251-6c9aa3caf3d6 // kubernetes-1.14.5
	k8s.io/client-go => k8s.io/client-go v11.0.1-0.20190805182715-88a2adca7e76+incompatible
)
