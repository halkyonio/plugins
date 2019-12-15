package capability

import (
	"github.com/hashicorp/go-plugin"
	halkyon "halkyon.io/api/capability/v1beta1"
	framework "halkyon.io/operator-framework"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"net/rpc"
)

type Plugin interface {
	framework.DependentResource
	GetCategory() halkyon.CapabilityCategory
	GetType() halkyon.CapabilityType
	GetWatchedResourcesTypes() []schema.GroupVersionKind
	ReadyFor(owner framework.Resource) framework.DependentResource
	Init(scheme *runtime.Scheme)
	Kill()
}

var _ plugin.Plugin = &GoPluginPlugin{}

type GoPluginPlugin struct {
	Delegate PluginResource
}

func (p *GoPluginPlugin) Server(b *plugin.MuxBroker) (interface{}, error) {
	return NewPluginServer(p.Delegate), nil
}

func (p *GoPluginPlugin) Client(b *plugin.MuxBroker, client *rpc.Client) (interface{}, error) {
	return &PluginClient{client: client}, nil
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

type InitResponse struct {
	TypesToRegister []runtime.Object
	GroupVersion    schema.GroupVersion
}

type PluginResource interface {
	framework.DependentResource
	SetOwner(owner framework.Resource)
	GetSupportedCategory() halkyon.CapabilityCategory
	GetSupportedType() halkyon.CapabilityType
	Init() InitResponse
}

type TypeRegistry map[halkyon.CapabilityType]bool
type CategoryRegistry map[halkyon.CapabilityCategory]TypeRegistry

var SupportedCategories CategoryRegistry
var Plugins []Plugin

var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "HALKYON_CAPABILITY_PLUGIN",
	MagicCookieValue: "io.halkyon.capability.plugin",
}
