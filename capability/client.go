package capability

import (
	"context"
	"encoding/gob"
	"fmt"
	"github.com/hashicorp/go-plugin"
	halkyon "halkyon.io/api/capability/v1beta1"
	"halkyon.io/api/v1beta1"
	framework "halkyon.io/operator-framework"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"log"
	"net/rpc"
	"os"
	"os/exec"
	"path/filepath"
)

type PluginClient struct {
	client      *rpc.Client
	name        string
	owner       *halkyon.Capability
	gpClient    *plugin.Client
	capCategory *halkyon.CapabilityCategory
	capType     *halkyon.CapabilityType
}

var _ Plugin = &PluginClient{}
var _ killableClient = &PluginClient{}

var emptyGVK = schema.GroupVersionKind{}

type killableClient interface {
	Plugin
	recordGoPluginClient(client *plugin.Client)
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

func (p *PluginClient) GetType() halkyon.CapabilityType {
	if p.capType == nil {
		var res halkyon.CapabilityType
		p.call("GetType", emptyGVK, &res)
		p.capType = &res
	}
	return *p.capType
}

func (p *PluginClient) GetWatchedResourcesTypes(owner *halkyon.Capability) []schema.GroupVersionKind {
	var res []schema.GroupVersionKind
	p.owner = owner
	p.call("GetWatchedResourcesTypes", emptyGVK, &res)
	return res
}

func (p *PluginClient) Kill() {
	p.gpClient.Kill()
}

func (p *PluginClient) ReadyFor(owner *halkyon.Capability) []framework.DependentResource {
	client := &PluginClient{
		client: p.client,
		name:   p.name,
		owner:  owner,
	}
	resourcesTypes := client.GetWatchedResourcesTypes(owner)
	depRes := make([]framework.DependentResource, 0, len(resourcesTypes))
	for _, rt := range resourcesTypes {
		depRes = append(depRes, &PluginDependentResource{client: client, gvk: rt, owner: owner})
	}
	return depRes
}

type PluginDependentResource struct {
	client *PluginClient
	config *framework.DependentResourceConfig
	gvk    schema.GroupVersionKind
	owner  v1beta1.HalkyonResource
	name   *string
}

var _ framework.DependentResource = &PluginDependentResource{}

func (p *PluginDependentResource) Name() string {
	if p.name == nil {
		name := ""
		p.client.call("Name", p.gvk, &name)
		p.name = &name
	}
	return *p.name
}

func (p PluginDependentResource) Owner() v1beta1.HalkyonResource {
	return p.owner
}

func (p PluginDependentResource) NameFrom(underlying runtime.Object) string {
	res := ""
	p.client.call("NameFrom", p.gvk, &res, underlying)
	return res
}

func (p PluginDependentResource) Fetch(helper *framework.K8SHelper) (runtime.Object, error) {
	into := framework.CreateEmptyUnstructured(p.GetConfig().GroupVersionKind)
	if err := helper.Client.Get(context.TODO(), types.NamespacedName{Name: p.Name(), Namespace: p.owner.GetNamespace()}, into); err != nil {
		return nil, err
	}
	return into, nil
}

func (p PluginDependentResource) Build(_ bool) (runtime.Object, error) {
	b := &BuildResponse{}
	p.client.call("Build", p.gvk, b)
	return b.Built, nil
}

func (p PluginDependentResource) Update(toUpdate runtime.Object) (bool, error) {
	res := UpdateResponse{}
	p.client.call("Update", p.gvk, &res, toUpdate)
	toUpdate = res.Updated
	return res.NeedsUpdate, res.Error
}

func (p PluginDependentResource) IsReady(underlying runtime.Object) (ready bool, message string) {
	res := IsReadyResponse{}
	p.client.call("IsReady", p.gvk, &res, underlying)
	return res.Ready, res.Message

}

func (p *PluginDependentResource) GetConfig() framework.DependentResourceConfig {
	if p.config == nil {
		config := &framework.DependentResourceConfig{}
		p.client.call("GetConfig", p.gvk, config)
		p.config = config
	}
	return *p.config
}

func NewPlugin(path string) (Plugin, error) {
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
	p := raw.(killableClient)
	p.recordGoPluginClient(client)
	return p, nil
}

func (p *PluginClient) call(method string, targetDependentType schema.GroupVersionKind, result interface{}, underlying ...runtime.Object) {
	if len(underlying) > 1 {
		log.Fatalf("error calling %s on %s plugin: call only accepts one extra argument, was given %v", method, p.name, underlying)
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
		log.Fatalf("error calling %s on %s plugin: %v", method, p.name, err)
	}
}

func init() {
	gob.Register(&halkyon.Capability{})
}
