package capability

import (
	"fmt"
	halkyon "halkyon.io/api/capability/v1beta1"
	"halkyon.io/api/v1beta1"
	framework "halkyon.io/operator-framework"
)

type PluginResource interface {
	GetSupportedCategory() halkyon.CapabilityCategory
	GetSupportedTypes() []halkyon.CapabilityType
	GetVersionsFor(capabilityType halkyon.CapabilityType) []string
	GetDependentResourcesWith(owner v1beta1.HalkyonResource) []framework.DependentResource
}

type SimplePluginResourceStem struct {
	ct []halkyon.CapabilityType
	cc halkyon.CapabilityCategory
}

func NewSimplePluginResourceStem(cat halkyon.CapabilityCategory, typ halkyon.CapabilityType) SimplePluginResourceStem {
	return SimplePluginResourceStem{cc: cat, ct: []halkyon.CapabilityType{typ}}
}
func (p SimplePluginResourceStem) GetSupportedCategory() halkyon.CapabilityCategory {
	return p.cc
}

func (p SimplePluginResourceStem) GetSupportedTypes() []halkyon.CapabilityType {
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
		for _, capabilityType := range resource.GetSupportedTypes() {
			apr.pluginResources[typeKey(capabilityType)] = resource
		}
	}
	return apr, nil
}

func (a AggregatePluginResource) GetSupportedCategory() halkyon.CapabilityCategory {
	return a.category
}

func (a AggregatePluginResource) GetSupportedTypes() []halkyon.CapabilityType {
	types := make([]halkyon.CapabilityType, 0, len(a.pluginResources))
	for capabilityType := range a.pluginResources {
		types = append(types, capabilityType)
	}
	return types
}

func (a AggregatePluginResource) GetDependentResourcesWith(owner v1beta1.HalkyonResource) []framework.DependentResource {
	capType := typeKey(owner.(*halkyon.Capability).Spec.Type)
	return a.pluginResources[capType].GetDependentResourcesWith(owner)
}

func (a AggregatePluginResource) GetVersionsFor(capabilityType halkyon.CapabilityType) []string {
	capType := typeKey(capabilityType)
	if resource, ok := a.pluginResources[capType]; ok {
		return resource.GetVersionsFor(capType)
	}
	return []string{}
}
