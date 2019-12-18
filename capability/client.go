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
	client   *rpc.Client
	name     string
	owner    *halkyon.Capability
	gpClient *plugin.Client
}

func (p *PluginClient) recordGoPluginClient(client *plugin.Client) {
	p.gpClient = client
}

type killableClient interface {
	Plugin
	recordGoPluginClient(client *plugin.Client)
}

func (p *PluginClient) Kill() {
	p.gpClient.Kill()
}

var _ Plugin = &PluginClient{}
var _ killableClient = &PluginClient{}

func (p *PluginClient) ReadyFor(owner *halkyon.Capability) framework.DependentResource {
	return &PluginClient{
		client: p.client,
		name:   p.name,
		owner:  owner,
	}
}

func (p *PluginClient) Fetch(helper *framework.K8SHelper) (runtime.Object, error) {
	into := framework.CreateEmptyUnstructured(p.GetGroupVersionKind())
	if err := helper.Client.Get(context.TODO(), types.NamespacedName{Name: p.Name(), Namespace: p.owner.GetNamespace()}, into); err != nil {
		return nil, err
	}
	return into, nil
}

func (p *PluginClient) GetTypeName() string {
	return p.name
}

func (p *PluginClient) ShouldWatch() bool {
	return true
}

func (p *PluginClient) CanBeCreatedOrUpdated() bool {
	return true
}

func (p *PluginClient) CreateOrUpdate(helper *framework.K8SHelper) error {
	return framework.CreateOrUpdate(p, helper)
}

func (p *PluginClient) ShouldBeOwned() bool {
	return true
}

func (p *PluginClient) OwnerStatusField() string {
	res := ""
	p.call("OwnerStatusField", PluginRequest{}, &res)
	return res
}

func (p *PluginClient) GetGroupVersionKind() schema.GroupVersionKind {
	var res schema.GroupVersionKind
	p.call("GetGroupVersionKind", PluginRequest{}, &res)
	return res
}

func (p *PluginClient) call(method string, args interface{}, result interface{}) {
	err := p.client.Call("Plugin."+method, args, result)
	if err != nil {
		log.Fatalf("error calling %s: %v", method, err)
	}
}

func (p *PluginClient) GetCategory() halkyon.CapabilityCategory {
	var cat halkyon.CapabilityCategory
	p.call("GetCategory", PluginRequest{}, &cat)
	return cat
}

func (p *PluginClient) GetWatchedResourcesTypes() []schema.GroupVersionKind {
	var res []schema.GroupVersionKind
	p.call("GetWatchedResourcesTypes", PluginRequest{}, &res)
	return res
}

func (p *PluginClient) GetType() halkyon.CapabilityType {
	var res halkyon.CapabilityType
	p.call("GetType", PluginRequest{}, &res)
	return res
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

func (p *PluginClient) Build() (runtime.Object, error) {
	b := &BuildResponse{}
	p.call("Build", nil, b)
	return b.Built, nil
}

func (p *PluginClient) IsReady(underlying runtime.Object) (ready bool, message string) {
	res := IsReadyResponse{}
	p.call("IsReady", underlying, &res)
	return res.Ready, res.Message
}

func (p *PluginClient) Name() string {
	res := ""
	p.call("Name", PluginRequest{Owner: p.owner}, &res)
	return res
}

func (p *PluginClient) NameFrom(underlying runtime.Object) string {
	res := ""
	p.call("NameFrom", underlying, &res)
	return res
}

func (p *PluginClient) Update(toUpdate runtime.Object) (bool, error) {
	res := UpdateResponse{}
	p.call("Update", toUpdate, &res)
	toUpdate = res.Updated
	return res.NeedsUpdate, res.Error
}

func (p *PluginClient) Owner() v1beta1.HalkyonResource {
	return p.owner
}

func (p *PluginClient) ShouldBeCheckedForReadiness() bool {
	return true
}
