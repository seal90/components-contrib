package lb

import (
	"context"

	"github.com/dapr/components-contrib/nameresolution"
)

type ServiceInstance struct {
	InstanceID string
	ServiceID  string
	Host       string
	Port       int
	Metadata   map[string]string
}

type LoadBalancer interface {
	Choose(req nameresolution.ResolveRequest) (*ServiceInstance, error)
}

type LoadBalancerFactory interface {
	Init(ctx context.Context, metadata nameresolution.Metadata) error
	Create(id string, serviceInstanceListSupplier ServiceInstanceListSupplier) LoadBalancer
}

type ServiceInstanceListSupplier interface {
	Instances(req nameresolution.ResolveRequest) ([]*ServiceInstance, error)
}

type ServiceInstanceListSupplierFactory interface {
	Init(ctx context.Context, metadata nameresolution.Metadata) error
	Create(id string, delegate ServiceInstanceListSupplier) ServiceInstanceListSupplier
}

type DelegatingServiceInstanceListSupplier struct {
	delegate ServiceInstanceListSupplier
}

func NewDelegatingServiceInstanceListSupplier(
	delegate ServiceInstanceListSupplier,
) *DelegatingServiceInstanceListSupplier {
	return &DelegatingServiceInstanceListSupplier{
		delegate: delegate,
	}
}

func (d *DelegatingServiceInstanceListSupplier) Init(ctx context.Context, metadata nameresolution.Metadata) error {
	return nil
}

func (d *DelegatingServiceInstanceListSupplier) Delegate() ServiceInstanceListSupplier {
	return d.delegate
}

func (d *DelegatingServiceInstanceListSupplier) Instances(req nameresolution.ResolveRequest) ([]*ServiceInstance, error) {
	return d.delegate.Instances(req)
}

type DiscoveryClient interface {
	Init(ctx context.Context, metadata nameresolution.Metadata) error

	Instances(req nameresolution.ResolveRequest) ([]*ServiceInstance, error)

	Services() ([]string, error)
}
