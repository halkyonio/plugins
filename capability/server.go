package capability

import (
	halkyon "halkyon.io/api/capability/v1beta1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type PluginServer interface {
	Build(req PluginRequest, res *BuildResponse) error
	GetCategory(req PluginRequest, res *halkyon.CapabilityCategory) error
	GetWatchedResourcesTypes(req PluginRequest, res *[]schema.GroupVersionKind) error
	GetType(req PluginRequest, res *halkyon.CapabilityType) error
	IsReady(req PluginRequest, res *IsReadyResponse) error
	Name(req PluginRequest, res *string) error
	NameFrom(req PluginRequest, res *string) error
	Update(req PluginRequest, res *UpdateResponse) error
	GetGroupVersionKind(req PluginRequest, res *schema.GroupVersionKind) error
}

type PluginServerImpl struct {
	capability     PluginResource
	category       halkyon.CapabilityCategory
	capabilityType halkyon.CapabilityType
}

var _ PluginServer = &PluginServerImpl{}

func NewPluginServer(category halkyon.CapabilityCategory, capabilityType halkyon.CapabilityType, capability PluginResource) PluginServer {
	return PluginServerImpl{
		capability:     capability,
		category:       category,
		capabilityType: capabilityType,
	}
}

func (p PluginServerImpl) GetGroupVersionKind(req PluginRequest, res *schema.GroupVersionKind) error {
	p.capability.SetOwner(req.Owner)
	*res = p.capability.GetGroupVersionKind()
	return nil
}

func (p PluginServerImpl) Build(req PluginRequest, res *BuildResponse) error {
	p.capability.SetOwner(req.Owner)
	build, err := p.capability.Build()
	res.Built = build
	return err
}

func (p PluginServerImpl) GetCategory(req PluginRequest, res *halkyon.CapabilityCategory) error {
	p.capability.SetOwner(req.Owner)
	*res = p.category
	return nil
}

func (p PluginServerImpl) GetWatchedResourcesTypes(req PluginRequest, res *[]schema.GroupVersionKind) error {
	p.capability.SetOwner(req.Owner)
	*res = []schema.GroupVersionKind{p.capability.GetGroupVersionKind()}
	return nil
}

func (p PluginServerImpl) GetType(req PluginRequest, res *halkyon.CapabilityType) error {
	p.capability.SetOwner(req.Owner)
	*res = p.capabilityType
	return nil
}

func (p PluginServerImpl) IsReady(req PluginRequest, res *IsReadyResponse) error {
	p.capability.SetOwner(req.Owner)
	ready, message := p.capability.IsReady(req.Arg)
	*res = IsReadyResponse{
		Ready:   ready,
		Message: message,
	}
	return nil
}

func (p PluginServerImpl) Name(req PluginRequest, res *string) error {
	p.capability.SetOwner(req.Owner)
	name := p.capability.Name()
	*res = name
	return nil
}

func (p PluginServerImpl) NameFrom(req PluginRequest, res *string) error {
	p.capability.SetOwner(req.Owner)
	name := p.capability.NameFrom(req.Arg)
	*res = name
	return nil
}

func (p PluginServerImpl) Update(req PluginRequest, res *UpdateResponse) error {
	p.capability.SetOwner(req.Owner)
	update, err := p.capability.Update(req.Arg)
	*res = UpdateResponse{
		NeedsUpdate: update,
		Error:       err,
		Updated:     req.Arg,
	}
	return err
}
