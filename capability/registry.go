package capability

import (
	"fmt"
	halkyon "halkyon.io/api/capability/v1beta1"
	"strings"
)

type typeRegistry map[halkyon.CapabilityType]Plugin
type pluginsRegistry map[halkyon.CapabilityCategory]typeRegistry

var plugins pluginsRegistry

func GetPluginFor(category halkyon.CapabilityCategory, capabilityType halkyon.CapabilityType) (Plugin, error) {
	if types, ok := plugins[categoryKey(category)]; ok {
		if p, ok := types[typeKey(capabilityType)]; ok {
			return p, nil
		}
	}
	return nil, fmt.Errorf("couldn't find a plugin to handle capability with category '%s' and type '%s'", category, capabilityType)
}

func categoryKey(category halkyon.CapabilityCategory) halkyon.CapabilityCategory {
	return halkyon.CapabilityCategory(strings.ToLower(category.String()))
}

func typeKey(capType halkyon.CapabilityType) halkyon.CapabilityType {
	return halkyon.CapabilityType(strings.ToLower(capType.String()))
}

func register(p *PluginClient) {
	category := p.GetCategory()
	categoryKey := categoryKey(category)
	if len(plugins) == 0 {
		plugins = make(pluginsRegistry, 7)
	}
	types, ok := plugins[categoryKey]
	if !ok {
		types = make(typeRegistry, 7)
		plugins[categoryKey] = types
	}
	capabilityTypes := p.GetTypes()
	for _, capabilityType := range capabilityTypes {
		t := capabilityType.Type
		typeKey := typeKey(t)
		plug, ok := types[typeKey]
		if ok {
			panic(fmt.Errorf("a plugin named '%s' is already registered for category '%s' / type '%s' pair", plug.Name(), category, t))
		}
		types[typeKey] = p
		p.log.Info(fmt.Sprintf("Registered plugin named '%s' for category '%s' / type '%s' pair", p.name, category, t))
	}
}
