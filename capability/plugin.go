package capability

import (
	"github.com/hashicorp/go-plugin"
	halkyon "halkyon.io/api/capability/v1beta1"
	framework "halkyon.io/operator-framework"
	"k8s.io/apimachinery/pkg/runtime"
	"net/rpc"
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

var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "HALKYON_CAPABILITY_PLUGIN",
	MagicCookieValue: "io.halkyon.capability.plugin",
}
