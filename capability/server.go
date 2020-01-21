package capability

import (
	"encoding/gob"
	"fmt"
	"github.com/hashicorp/go-plugin"
	halkyon "halkyon.io/api/capability/v1beta1"
	framework "halkyon.io/operator-framework"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"os"
	"path/filepath"
)

type PluginServer interface {
	Build(req PluginRequest, res *BuildResponse) error
	GetCategory(req PluginRequest, res *halkyon.CapabilityCategory) error
	GetDependentResourceTypes(req PluginRequest, res *[]schema.GroupVersionKind) error
	GetTypes(req PluginRequest, res *[]TypeInfo) error
	IsReady(req PluginRequest, res *IsReadyResponse) error
	Name(req PluginRequest, res *string) error
	NameFrom(req PluginRequest, res *string) error
	Update(req PluginRequest, res *UpdateResponse) error
	GetConfig(req PluginRequest, res *framework.DependentResourceConfig) error
}

type PluginServerImpl struct {
	capability PluginResource
}

func (p PluginServerImpl) GetConfig(req PluginRequest, res *framework.DependentResourceConfig) error {
	resource := p.dependentResourceFor(req)
	*res = resource.GetConfig()
	return nil
}

var _ PluginServer = &PluginServerImpl{}

func StartPluginServerFor(resources ...PluginResource) {
	pluginName := filepath.Base(os.Args[0])
	p, err := NewAggregatePluginResource(resources...)
	if err != nil {
		panic(err)
	}
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: Handshake,
		Plugins:         map[string]plugin.Plugin{pluginName: &GoPluginPlugin{Delegate: p}},
	})
}

func (p PluginServerImpl) Build(req PluginRequest, res *BuildResponse) error {
	resource := p.dependentResourceFor(req)
	build, err := resource.Build(false)
	if err != nil {
		return err
	}
	res.Built, err = framework.CreateUnstructuredObject(build, req.Target)
	return err
}

func (p PluginServerImpl) GetCategory(_ PluginRequest, res *halkyon.CapabilityCategory) error {
	*res = p.capability.GetSupportedCategory()
	return nil
}

func (p PluginServerImpl) GetDependentResourceTypes(req PluginRequest, res *[]schema.GroupVersionKind) error {
	dependents := p.capability.GetDependentResourcesWith(req.Owner)
	*res = make([]schema.GroupVersionKind, 0, len(dependents))
	for _, dependent := range dependents {
		*res = append(*res, dependent.GetConfig().GroupVersionKind)
	}
	return nil
}

func (p PluginServerImpl) GetTypes(req PluginRequest, res *[]TypeInfo) error {
	*res = p.capability.GetSupportedTypes()
	return nil
}

func (p PluginServerImpl) IsReady(req PluginRequest, res *IsReadyResponse) error {
	resource := p.dependentResourceFor(req)
	ready, message := resource.IsReady(requestedArg(resource, req))
	*res = IsReadyResponse{
		Ready:   ready,
		Message: message,
	}
	return nil
}

func (p PluginServerImpl) Name(req PluginRequest, res *string) error {
	resource := p.dependentResourceFor(req)
	*res = resource.Name()
	return nil
}

func (p PluginServerImpl) NameFrom(req PluginRequest, res *string) error {
	resource := p.dependentResourceFor(req)
	*res = resource.NameFrom(requestedArg(resource, req))
	return nil
}

func (p PluginServerImpl) Update(req PluginRequest, res *UpdateResponse) error {
	resource := p.dependentResourceFor(req)
	update, err := resource.Update(requestedArg(resource, req))
	*res = UpdateResponse{
		NeedsUpdate: update,
		Error:       err,
		Updated:     req.Arg,
	}
	return err
}

func (p PluginServerImpl) dependentResourceFor(req PluginRequest) framework.DependentResource {
	dependents := p.capability.GetDependentResourcesWith(req.Owner)
	for _, dependent := range dependents {
		if dependent.GetConfig().GroupVersionKind == req.Target {
			return dependent
		}
	}
	panic(fmt.Errorf("no dependent of type %v for plugin %v/%v", req.Target, p.capability.GetSupportedCategory(), p.capability.GetSupportedTypes()))
}

func requestedArg(dependent framework.DependentResource, req PluginRequest) runtime.Object {
	build, _ := dependent.Build(true)
	return req.getArg(build)
}

func init() {
	gob.Register(&unstructured.Unstructured{})
}
