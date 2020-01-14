package capability

import (
	"context"
	"encoding/gob"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/hashicorp/go-plugin"
	halkyon "halkyon.io/api/capability/v1beta1"
	"halkyon.io/api/v1beta1"
	framework "halkyon.io/operator-framework"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"net/rpc"
	"os"
	"os/exec"
	"path/filepath"
)

type Plugin interface {
	Name() string
	GetCategory() halkyon.CapabilityCategory
	GetTypes() []halkyon.CapabilityType
	ReadyFor(owner *halkyon.Capability) []framework.DependentResource
	Kill()
}

type PluginClient struct {
	client      *rpc.Client
	name        string
	owner       *halkyon.Capability
	gpClient    *plugin.Client
	capCategory *halkyon.CapabilityCategory
	capType     *[]halkyon.CapabilityType
	log         logr.Logger
}

var _ Plugin = &PluginClient{}
var _ killableClient = &PluginClient{}

var emptyGVK = schema.GroupVersionKind{}

type killableClient interface {
	Plugin
	recordGoPluginClient(client *plugin.Client)
}

func (p *PluginClient) Name() string {
	return p.name
}

func (p *PluginClient) recordGoPluginClient(client *plugin.Client) {
	p.gpClient = client
}

func (p *PluginClient) GetCategory() halkyon.CapabilityCategory {
	if p.capCategory == nil {
		var cat halkyon.CapabilityCategory
		p.call("GetCategory", emptyGVK, &cat)
		p.capCategory = &cat
	}
	return *p.capCategory
}

func (p *PluginClient) GetTypes() []halkyon.CapabilityType {
	if p.capType == nil {
		res := []halkyon.CapabilityType{}
		p.call("GetTypes", emptyGVK, &res)
		p.capType = &res
	}
	return *p.capType
}

func (p *PluginClient) Kill() {
	p.gpClient.Kill()
}

func (p *PluginClient) ReadyFor(owner *halkyon.Capability) []framework.DependentResource {
	client := &PluginClient{
		client: p.client,
		name:   p.name,
		log:    p.log,
		owner:  owner,
	}
	resourcesTypes := []schema.GroupVersionKind{}
	client.call("GetDependentResourceTypes", emptyGVK, &resourcesTypes)
	depRes := make([]framework.DependentResource, 0, len(resourcesTypes))
	for _, rt := range resourcesTypes {
		depRes = append(depRes, &PluginDependentResource{client: client, gvk: rt, owner: owner})
	}
	return depRes
}

func NewPlugin(path string, log logr.Logger) (Plugin, error) {
	name := filepath.Base(path)

	// We're a host. Start by launching the plugin process.
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: Handshake,
		Plugins:         map[string]plugin.Plugin{name: &GoPluginPlugin{name: name}},
		Cmd:             exec.Command(path),
	})

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}

	// Request the plugin
	raw, err := rpcClient.Dispense(name)
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}
	p := raw.(*PluginClient)
	p.log = log
	p.recordGoPluginClient(client)

	register(p)

	return p, nil
}

func (p *PluginClient) call(method string, targetDependentType schema.GroupVersionKind, result interface{}, underlying ...runtime.Object) {
	if len(underlying) > 1 {
		p.log.Error(fmt.Errorf("error calling %s on %s plugin", method, p.name), fmt.Sprintf("call only accepts one extra argument, was given %v", underlying))
	}
	request := PluginRequest{}
	if p.owner != nil {
		request.Owner = p.owner
	}
	if !targetDependentType.Empty() {
		request.Target = targetDependentType
	}
	if len(underlying) == 1 {
		request.setArg(underlying[0])
	}
	err := p.client.Call("Plugin."+method, request, result)
	if err != nil {
		p.log.Error(err, fmt.Sprintf("error calling %s on %s plugin", method, p.name))
	}
}

func init() {
	gob.Register(&halkyon.Capability{})
}
