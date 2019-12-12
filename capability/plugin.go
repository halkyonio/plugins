package capability

import (
	halkyon "halkyon.io/api/capability/v1beta1"
	framework "halkyon.io/operator-framework"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Plugin interface {
	framework.DependentResource
	GetCategory() halkyon.CapabilityCategory
	GetType() halkyon.CapabilityType
	GetWatchedResourcesTypes() []schema.GroupVersionKind
	ReadyFor(owner framework.Resource) framework.DependentResource
}

type PluginRequest struct {
	Owner framework.Resource
	Arg   runtime.Object
}

type IsReadyResponse struct {
	Ready   bool
	Message string
}

type UpdateResponse struct {
	NeedsUpdate bool
	Error       error
	Updated     runtime.Object
}

type BuildResponse struct {
	Built runtime.Object
}

type PluginResource interface {
	framework.DependentResource
	SetOwner(owner framework.Resource)
}
