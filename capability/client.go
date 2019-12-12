package capability

import (
	"context"
	"github.com/natefinch/pie"
	halkyon "halkyon.io/api/capability/v1beta1"
	framework "halkyon.io/operator-framework"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"log"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"path/filepath"
)

type PluginClient struct {
	client *rpc.Client
	name   string
	owner  framework.Resource
}

var _ Plugin = &PluginClient{}

func (p *PluginClient) ReadyFor(owner framework.Resource) framework.DependentResource {
	return &PluginClient{
		client: p.client,
		name:   p.name,
		owner:  owner,
	}
}

func (p *PluginClient) Fetch(helper *framework.K8SHelper) (runtime.Object, error) {
	into, err := helper.Scheme.New(p.GetGroupVersionKind())
	if err != nil {
		return nil, err
	}
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
	err := p.client.Call(p.name+"."+method, args, result)
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
	client, err := pie.StartProviderCodec(jsonrpc.NewClientCodec, os.Stderr, path)
	if err != nil {
		return nil, err
	}
	return &PluginClient{client: client, name: filepath.Base(path)}, nil
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

func (p *PluginClient) Owner() framework.Resource {
	panic("implement me")
}

func (p *PluginClient) ShouldBeCheckedForReadiness() bool {
	return true
}
