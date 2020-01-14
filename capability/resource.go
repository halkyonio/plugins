package capability

import (
	"fmt"
	halkyon "halkyon.io/api/capability/v1beta1"
	"halkyon.io/api/v1beta1"
	framework "halkyon.io/operator-framework"
)

type PluginResource interface {
	GetSupportedCategory() halkyon.CapabilityCategory
	GetSupportedTypes() []TypeInfo
	GetDependentResourcesWith(owner v1beta1.HalkyonResource) []framework.DependentResource
}

type SimplePluginResourceStem struct {
	ct []TypeInfo
	cc halkyon.CapabilityCategory
}

func NewSimplePluginResourceStem(cat halkyon.CapabilityCategory, typ TypeInfo) SimplePluginResourceStem {
	return SimplePluginResourceStem{cc: cat, ct: []TypeInfo{typ}}
}
func (p SimplePluginResourceStem) GetSupportedCategory() halkyon.CapabilityCategory {
	return p.cc
}

func (p SimplePluginResourceStem) GetSupportedTypes() []TypeInfo {
	return p.ct
}

type AggregatePluginResource struct {
	category        halkyon.CapabilityCategory
	pluginResources map[halkyon.CapabilityType]PluginResource
}

func NewAggregatePluginResource(resources ...PluginResource) (PluginResource, error) {
	apr := AggregatePluginResource{
		pluginResources: make(map[halkyon.CapabilityType]PluginResource, len(resources)),
	}
	for _, resource := range resources {
		category := categoryKey(resource.GetSupportedCategory())
		if len(apr.category) == 0 {
			apr.category = category
		}
		if !apr.category.Equals(category) {
			return nil, fmt.Errorf("can only aggregate PluginResources providing the same category, got %v and %v", apr.category, category)
		}
		for _, typeInfo := range resource.GetSupportedTypes() {
			apr.pluginResources[typeKey(typeInfo.Type)] = resource
		}
	}
	return apr, nil
}

func (a AggregatePluginResource) GetSupportedCategory() halkyon.CapabilityCategory {
	return a.category
}

func (a AggregatePluginResource) GetSupportedTypes() []TypeInfo {
	types := make([]TypeInfo, 0, len(a.pluginResources))
	for _, resource := range a.pluginResources {
		types = append(types, resource.GetSupportedTypes()...)
	}
	return types
}

func (a AggregatePluginResource) GetDependentResourcesWith(owner v1beta1.HalkyonResource) []framework.DependentResource {
	capType := typeKey(owner.(*halkyon.Capability).Spec.Type)
	return a.pluginResources[capType].GetDependentResourcesWith(owner)
}
