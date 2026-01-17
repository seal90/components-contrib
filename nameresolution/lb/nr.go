package lb

import (
	"context"
	"errors"
	"strconv"
	"sync"

	"github.com/mitchellh/mapstructure"

	"github.com/dapr/components-contrib/nameresolution"
	"github.com/dapr/kit/logger"
)

type LoadBalancerMetadata struct {
	Component     string `mapstructure:"component"`
	Configuration any    `mapstructure:"configuration" json:"configuration"`
}

type ServiceInstanceListSupplierMetadata struct {
	Component     string `mapstructure:"component" json:"component"`
	Configuration any    `mapstructure:"configuration" json:"configuration"`
}

type DiscoveryClientMetadata struct {
	Component     string `mapstructure:"component" json:"component"`
	Configuration any    `mapstructure:"configuration" json:"configuration"`
}

type ResolverMetadata struct {
	LoadBalancer                LoadBalancerMetadata                  `mapstructure:"loadBalancer"`
	ServiceInstanceListSupplier []ServiceInstanceListSupplierMetadata `mapstructure:"serviceInstanceListSupplier"`
	DiscoveryClient             []DiscoveryClientMetadata             `mapstructure:"discoveryClient"`
}

type Resolver struct {
	loadBalancers map[string]LoadBalancer
	mu            sync.RWMutex

	metadata         nameresolution.Metadata
	resolverMetadata ResolverMetadata

	discoveryClient                            DiscoveryClient
	serviceInstanceListSupplierFactories       []ServiceInstanceListSupplierFactory
	discoveryClientServiceInstanceListSupplier ServiceInstanceListSupplier
	loadBalancerFactory                        LoadBalancerFactory

	logger logger.Logger
}

func NewResolver(logger logger.Logger) nameresolution.Resolver {
	return &Resolver{
		logger: logger,
	}
}

func (m *Resolver) Init(ctx context.Context, metadata nameresolution.Metadata) error {

	m.loadBalancers = make(map[string]LoadBalancer)
	var meta ResolverMetadata
	err := mapstructure.Decode(metadata.Configuration, &meta)
	// err := kitmd.DecodeMetadata(metadata.Configuration, &meta)
	if err != nil {
		return err
	}
	m.metadata = metadata
	m.resolverMetadata = meta

	err = m.InitDiscoveryClient(ctx, metadata)
	if err != nil {
		return err
	}
	err = m.InitServiceInstanceListSupplier(ctx, metadata)
	if err != nil {
		return err
	}
	err = m.InitLoadBalancer(ctx, metadata)
	if err != nil {
		return err
	}

	return nil
}

func (m *Resolver) InitDiscoveryClient(ctx context.Context, metadata nameresolution.Metadata) error {
	discoveryClientMetadata := m.resolverMetadata.DiscoveryClient

	var discoveryClients []DiscoveryClient
	for _, meta := range discoveryClientMetadata {
		if meta.Component == "composite-discovery" {
			return errors.New("DiscoveryClient cann't config Component with composite-discovery")
		}
		// TODO version
		discoveryClient, err := DefaultRegistry.CreateDiscoveryClient(meta.Component, "", meta.Component)
		if err != nil {
			return err
		}

		m := nameresolution.Metadata{
			Instance:      metadata.Instance,
			Configuration: meta.Configuration,
		}
		err = discoveryClient.Init(ctx, m)
		if err != nil {
			return err
		}
		discoveryClients = append(discoveryClients, discoveryClient)
	}

	compositeDiscoveryClientInterface, err := DefaultRegistry.CreateDiscoveryClient("composite-discovery", "", "composite-discovery")
	if err != nil {
		return err
	}
	compositeDiscoveryClient := compositeDiscoveryClientInterface.(*CompositeDiscoveryClient)
	compositeDiscoveryClient.InitWithClient(discoveryClients)
	m.discoveryClient = compositeDiscoveryClient

	return nil
}

func (m *Resolver) InitServiceInstanceListSupplier(ctx context.Context, metadata nameresolution.Metadata) error {
	serviceInstanceListSupplierMetadata := m.resolverMetadata.ServiceInstanceListSupplier

	var serviceInstanceListSupplierFactories []ServiceInstanceListSupplierFactory
	for _, meta := range serviceInstanceListSupplierMetadata {
		if meta.Component == "discovery-client" {
			return errors.New("ServiceInstanceListSupplier cann't config Component with discovery-client")
		}
		// TODO version
		serviceInstanceListSupplierFactory, err := DefaultRegistry.CreateServiceInstanceListSupplier(meta.Component, "", meta.Component)
		if err != nil {
			return err
		}

		m := nameresolution.Metadata{
			Instance:      metadata.Instance,
			Configuration: meta.Configuration,
		}
		err = serviceInstanceListSupplierFactory.Init(ctx, m)
		if err != nil {
			return err
		}
		serviceInstanceListSupplierFactories = append(serviceInstanceListSupplierFactories, serviceInstanceListSupplierFactory)
	}
	m.serviceInstanceListSupplierFactories = serviceInstanceListSupplierFactories

	discoveryClientServiceInstanceListSupplierInterfaceFactory, err := DefaultRegistry.CreateServiceInstanceListSupplier("discovery-client", "", "discovery-client")
	if err != nil {
		return err
	}
	discoveryClientServiceInstanceListSupplierFactroy := discoveryClientServiceInstanceListSupplierInterfaceFactory.(*DiscoveryClientServiceInstanceListSupplierFactory)
	discoveryClientServiceInstanceListSupplier := discoveryClientServiceInstanceListSupplierFactroy.Create("", nil).(*DiscoveryClientServiceInstanceListSupplier)
	discoveryClientServiceInstanceListSupplier.InitWithClient(m.discoveryClient)
	m.discoveryClientServiceInstanceListSupplier = discoveryClientServiceInstanceListSupplier

	return nil
}

func (m *Resolver) InitLoadBalancer(ctx context.Context, metadata nameresolution.Metadata) error {
	loadBalancerMetadata := m.resolverMetadata.LoadBalancer
	loadBalancerFactory, err := DefaultRegistry.CreateLoadBalancer(loadBalancerMetadata.Component, "", loadBalancerMetadata.Component)
	if err != nil {
		return nil
	}
	loadBalancerFactory.Init(ctx, nameresolution.Metadata{
		Instance:      metadata.Instance,
		Configuration: loadBalancerMetadata.Configuration,
	})
	m.loadBalancerFactory = loadBalancerFactory

	return nil
}

func (m *Resolver) ResolveID(parentCtx context.Context, req nameresolution.ResolveRequest) (string, error) {

	loadBalancer := m.loadBalancers[req.ID]
	if loadBalancer == nil {
		m.mu.Lock()
		defer m.mu.Unlock()
		loadBalancer = m.loadBalancers[req.ID]
		if loadBalancer == nil {
			loadBalancer = m.buildLoadBalancer(req.ID)
			m.loadBalancers[req.ID] = loadBalancer
		}
	}
	instance, err := loadBalancer.Choose(req)
	if err != nil {
		return "", err
	}

	return instance.Host + ":" + strconv.Itoa(instance.Port), nil
}

func (m *Resolver) buildLoadBalancer(id string) LoadBalancer {
	serviceInstanceListSupplier := m.discoveryClientServiceInstanceListSupplier
	for _, suppler := range m.serviceInstanceListSupplierFactories {
		serviceInstanceListSupplier = suppler.Create(id, serviceInstanceListSupplier)
	}

	return m.loadBalancerFactory.Create(id, serviceInstanceListSupplier)
}

func (m *Resolver) Close() error {
	// TODO
	return nil
}
