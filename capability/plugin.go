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
)

type Plugin interface {
	Name() string
	GetCategory() halkyon.CapabilityCategory
	GetType() halkyon.CapabilityType
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
	GetSupportedType() halkyon.CapabilityType
	GetDependentResourcesWith(owner v1beta1.HalkyonResource) []framework.DependentResource
}

type typeRegistry map[halkyon.CapabilityType]Plugin
type pluginsRegistry map[halkyon.CapabilityCategory]typeRegistry

var plugins pluginsRegistry

func GetPluginFor(category halkyon.CapabilityCategory, capabilityType halkyon.CapabilityType) (Plugin, error) {
	if types, ok := plugins[category]; ok {
		if p, ok := types[capabilityType]; ok {
			return p, nil
		}
	}
	return nil, fmt.Errorf("couldn't find a plugin to handle capability with category '%s' and type '%s'", category, capabilityType)
}

func register(p Plugin) {
	category := p.GetCategory()
	if len(plugins) == 0 {
		plugins = make(pluginsRegistry, 7)
	}
	types, ok := plugins[category]
	if !ok {
		types = make(typeRegistry, 7)
		plugins[category] = types
	}
	capabilityType := p.GetType()
	plug, ok := types[capabilityType]
	if ok {
		panic(fmt.Errorf("a plugin named '%s' is already registered for category '%s' / type '%s' pair", plug.Name(), category, capabilityType))
	}
	types[capabilityType] = p
}

var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "HALKYON_CAPABILITY_PLUGIN",
	MagicCookieValue: "io.halkyon.capability.plugin",
}
