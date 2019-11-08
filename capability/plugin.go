package capability

import (
	halkyon "halkyon.io/api/capability/v1beta1"
	framework "halkyon.io/operator-framework"
)

type Plugin interface {
	GetDependentResources() []framework.DependentResource
	GetCategory() halkyon.CapabilityCategory
	GetType() halkyon.CapabilityType
}
