package capability

import (
	"encoding/gob"
	"fmt"
	"github.com/prometheus/common/log"
	halkyon "halkyon.io/api/capability/v1beta1"
	framework "halkyon.io/operator-framework"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type PluginServer interface {
	Build(req PluginRequest, res *BuildResponse) error
	GetCategory(req PluginRequest, res *halkyon.CapabilityCategory) error
	GetDependentResourceTypes(req PluginRequest, res *[]schema.GroupVersionKind) error
	GetType(req PluginRequest, res *halkyon.CapabilityType) error
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

func NewPluginServer(capability PluginResource) PluginServer {
	return PluginServerImpl{capability}
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
	log.Info("server: GetCategory")
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

func (p PluginServerImpl) GetType(_ PluginRequest, res *halkyon.CapabilityType) error {
	log.Info("server: GetCategory")
	*res = p.capability.GetSupportedType()
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
	panic(fmt.Errorf("no dependent of type %v for plugin %v/%v", req.Target, p.capability.GetSupportedCategory(), p.capability.GetSupportedType()))
}

func requestedArg(dependent framework.DependentResource, req PluginRequest) runtime.Object {
	build, _ := dependent.Build(true)
	return req.getArg(build)
}

func init() {
	gob.Register(&unstructured.Unstructured{})
}
