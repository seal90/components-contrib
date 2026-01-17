package lb

import (
	"fmt"
	"strings"

	"github.com/dapr/kit/logger"
)

type (
	LoadBalancerFactoryMethod                func(logger.Logger) LoadBalancerFactory
	ServiceInstanceListSupplierFactoryMethod func(logger.Logger) ServiceInstanceListSupplierFactory
	DiscoveryClientFactoryMethod             func(logger.Logger) DiscoveryClient

	// Registry handles registering and creating components.
	Registry struct {
		Logger                       logger.Logger
		loadBalancers                map[string]LoadBalancerFactoryMethod
		serviceInstanceListSuppliers map[string]ServiceInstanceListSupplierFactoryMethod
		discoveryClients             map[string]DiscoveryClientFactoryMethod
	}
)

var DefaultRegistry *Registry = NewRegistry()

// NewRegistry creates a name resolution registry.
func NewRegistry() *Registry {
	return &Registry{
		loadBalancers:                map[string]LoadBalancerFactoryMethod{},
		serviceInstanceListSuppliers: map[string]ServiceInstanceListSupplierFactoryMethod{},
		discoveryClients:             map[string]DiscoveryClientFactoryMethod{},
	}
}

func (s *Registry) RegisterLoadBalancerComponent(componentFactory LoadBalancerFactoryMethod, names ...string) {
	for _, name := range names {
		s.loadBalancers[createLoadBalancerFullName(name)] = componentFactory
	}
}

func (s *Registry) CreateLoadBalancer(name, version, logName string) (LoadBalancerFactory, error) {
	if method, ok := s.getLoadBalancer(createLoadBalancerFullName(name), version, logName); ok {
		return method(), nil
	}
	return nil, fmt.Errorf("couldn't find LoadBalancer %s/%s", name, version)
}

func (s *Registry) getLoadBalancer(name, version, logName string) (func() LoadBalancerFactory, bool) {
	if s.loadBalancers == nil {
		return nil, false
	}
	nameLower := strings.ToLower(name)
	versionLower := strings.ToLower(version)
	factoryFn, ok := s.loadBalancers[nameLower+"/"+versionLower]
	if ok {
		return s.wrapLoadBalancerFn(factoryFn, logName), true
	}
	if IsInitialVersion(versionLower) {
		factoryFn, ok = s.loadBalancers[nameLower]
		if ok {
			return s.wrapLoadBalancerFn(factoryFn, logName), true
		}
	}
	return nil, false
}

func (s *Registry) wrapLoadBalancerFn(componentFactory LoadBalancerFactoryMethod, logName string) func() LoadBalancerFactory {
	return func() LoadBalancerFactory {
		l := s.Logger
		if logName != "" && l != nil {
			l = l.WithFields(map[string]any{
				"component": logName,
			})
		}
		return componentFactory(l)
	}
}

func createLoadBalancerFullName(name string) string {
	return strings.ToLower("loadbalancer." + name)
}

// -----------
func (s *Registry) RegisterServiceInstanceListSupplierComponent(componentFactory ServiceInstanceListSupplierFactoryMethod, names ...string) {
	for _, name := range names {
		s.serviceInstanceListSuppliers[createServiceInstanceListSupplierFullName(name)] = componentFactory
	}
}

func (s *Registry) CreateServiceInstanceListSupplier(name, version, logName string) (ServiceInstanceListSupplierFactory, error) {
	if method, ok := s.getServiceInstanceListSupplier(createServiceInstanceListSupplierFullName(name), version, logName); ok {
		return method(), nil
	}
	return nil, fmt.Errorf("couldn't find ServiceInstanceListSupplier %s/%s", name, version)
}

func (s *Registry) getServiceInstanceListSupplier(name, version, logName string) (func() ServiceInstanceListSupplierFactory, bool) {
	if s.loadBalancers == nil {
		return nil, false
	}
	nameLower := strings.ToLower(name)
	versionLower := strings.ToLower(version)
	factoryFn, ok := s.serviceInstanceListSuppliers[nameLower+"/"+versionLower]
	if ok {
		return s.wrapServiceInstanceListSupplierFn(factoryFn, logName), true
	}
	if IsInitialVersion(versionLower) {
		factoryFn, ok = s.serviceInstanceListSuppliers[nameLower]
		if ok {
			return s.wrapServiceInstanceListSupplierFn(factoryFn, logName), true
		}
	}
	return nil, false
}

func (s *Registry) wrapServiceInstanceListSupplierFn(componentFactory ServiceInstanceListSupplierFactoryMethod, logName string) func() ServiceInstanceListSupplierFactory {
	return func() ServiceInstanceListSupplierFactory {
		l := s.Logger
		if logName != "" && l != nil {
			l = l.WithFields(map[string]any{
				"component": logName,
			})
		}
		return componentFactory(l)
	}
}

func createServiceInstanceListSupplierFullName(name string) string {
	return strings.ToLower("serviceinstancelistsupplier." + name)
}

// ---------------------------------------
func (s *Registry) RegisterDiscoveryClientComponent(componentFactory DiscoveryClientFactoryMethod, names ...string) {
	for _, name := range names {
		s.discoveryClients[createDiscoveryClientFullName(name)] = componentFactory
	}
}

func (s *Registry) CreateDiscoveryClient(name, version, logName string) (DiscoveryClient, error) {
	if method, ok := s.getDiscoveryClient(createDiscoveryClientFullName(name), version, logName); ok {
		return method(), nil
	}
	return nil, fmt.Errorf("couldn't find DiscoveryClient %s/%s", name, version)
}

func (s *Registry) getDiscoveryClient(name, version, logName string) (func() DiscoveryClient, bool) {
	if s.loadBalancers == nil {
		return nil, false
	}
	nameLower := strings.ToLower(name)
	versionLower := strings.ToLower(version)
	factoryFn, ok := s.discoveryClients[nameLower+"/"+versionLower]
	if ok {
		return s.wrapDiscoveryClientFn(factoryFn, logName), true
	}
	if IsInitialVersion(versionLower) {
		factoryFn, ok = s.discoveryClients[nameLower]
		if ok {
			return s.wrapDiscoveryClientFn(factoryFn, logName), true
		}
	}
	return nil, false
}

func (s *Registry) wrapDiscoveryClientFn(componentFactory DiscoveryClientFactoryMethod, logName string) func() DiscoveryClient {
	return func() DiscoveryClient {
		l := s.Logger
		if logName != "" && l != nil {
			l = l.WithFields(map[string]any{
				"component": logName,
			})
		}
		return componentFactory(l)
	}
}

func createDiscoveryClientFullName(name string) string {
	return strings.ToLower("discoveryclient." + name)
}

/*
package components

import (
	"github.com/dapr/components-contrib/nameresolution/sqlite"
	nrLoader "github.com/dapr/dapr/pkg/components/nameresolution"
)

func init() {
	nrLoader.DefaultRegistry.RegisterComponent(sqlite.NewResolver, "sqlite")
}
*/
