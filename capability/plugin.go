package capability

import (
	"fmt"
	"github.com/hashicorp/go-plugin"
	halkyon "halkyon.io/api/capability/v1beta1"
	"halkyon.io/api/v1beta1"
	framework "halkyon.io/operator-framework"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"net/rpc"
	"strings"
)

type Plugin interface {
	Name() string
	GetCategory() halkyon.CapabilityCategory
	GetTypes() []halkyon.CapabilityType
	ReadyFor(owner *halkyon.Capability) []framework.DependentResource
	Kill()
}

var _ plugin.Plugin = &GoPluginPlugin{}

type GoPluginPlugin struct {
	name     string
	Delegate PluginResource
}

func (p *GoPluginPlugin) Server(b *plugin.MuxBroker) (interface{}, error) {
	return NewPluginServer(p.Delegate), nil
}

func (p *GoPluginPlugin) Client(b *plugin.MuxBroker, client *rpc.Client) (interface{}, error) {
	return &PluginClient{name: p.name, client: client}, nil
}

type PluginRequest struct {
	Owner  v1beta1.HalkyonResource
	Target schema.GroupVersionKind
	Arg    *unstructured.Unstructured
}

func (p *PluginRequest) setArg(object runtime.Object) {
	u, ok := object.(*unstructured.Unstructured)
	if !ok {
		uns, e := framework.CreateUnstructuredObject(object, object.GetObjectKind().GroupVersionKind())
		if e != nil {
			panic(e)
		}
		u = uns.(*unstructured.Unstructured)
	}
	p.Arg = u
}

func (p *PluginRequest) getArg(object runtime.Object) runtime.Object {
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(p.Arg.Object, object)
	if err != nil {
		panic(err)
	}
	return object
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
	GetSupportedCategory() halkyon.CapabilityCategory
	GetSupportedTypes() []halkyon.CapabilityType
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
		typeKey := typeKey(capabilityType)
		plug, ok := types[typeKey]
		if ok {
			panic(fmt.Errorf("a plugin named '%s' is already registered for category '%s' / type '%s' pair", plug.Name(), category, capabilityType))
		}
		types[typeKey] = p
		p.log.Info(fmt.Sprintf("Registered plugin named '%s' for category '%s' / type '%s' pair", p.name, category, capabilityType))
	}
}

var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "HALKYON_CAPABILITY_PLUGIN",
	MagicCookieValue: "io.halkyon.capability.plugin",
}
