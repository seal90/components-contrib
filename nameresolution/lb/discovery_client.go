package lb

import (
	"context"
	"errors"
	"sync"

	"github.com/mitchellh/mapstructure"

	"github.com/dapr/components-contrib/nameresolution"
	"github.com/dapr/kit/logger"
)

type CompositeDiscoveryClient struct {
	clients []DiscoveryClient
	logger  logger.Logger
}

func NewCompositeDiscoveryClient(logger logger.Logger) DiscoveryClient {
	return &CompositeDiscoveryClient{
		logger: logger,
	}
}

func (c *CompositeDiscoveryClient) Init(ctx context.Context, metadata nameresolution.Metadata) error {

	return nil
}

func (c *CompositeDiscoveryClient) InitWithClient(clients []DiscoveryClient) {
	c.clients = clients
}

func (c *CompositeDiscoveryClient) Instances(req nameresolution.ResolveRequest) ([]*ServiceInstance, error) {
	if len(c.clients) == 0 {
		return nil, errors.New("no discovery clients registered")
	}

	var lastErr error
	for _, client := range c.clients {
		instances, err := client.Instances(req)
		if err != nil {
			lastErr = err
			continue
		}
		if len(instances) > 0 {
			return instances, nil
		}
		return instances, nil
	}

	return nil, lastErr
}

func (c *CompositeDiscoveryClient) Services() ([]string, error) {
	serviceSet := make(map[string]struct{})
	var lastErr error
	hasSuccess := false

	for _, client := range c.clients {
		services, err := client.Services()
		if err != nil {
			lastErr = err
			continue
		}
		hasSuccess = true
		for _, svc := range services {
			serviceSet[svc] = struct{}{}
		}
	}

	if !hasSuccess && lastErr != nil {
		return nil, lastErr
	}

	result := make([]string, 0, len(serviceSet))
	for svc := range serviceSet {
		result = append(result, svc)
	}
	return result, nil
}

type SampleInstance struct {
	InstanceID string            `mapstructure:"instanceID"`
	Host       string            `mapstructure:"host"`
	Port       int               `mapstructure:"port"`
	Metadata   map[string]string `mapstructure:"metadata"`
}

type SimpleDiscoveryClientMetadata struct {
	Instances map[string][]*SampleInstance
}

type SimpleDiscoveryClient struct {
	mu        sync.RWMutex
	instances map[string][]*ServiceInstance
	logger    logger.Logger
}

func NewSimpleDiscoveryClient(logger logger.Logger) DiscoveryClient {
	return &SimpleDiscoveryClient{
		logger: logger,
	}
}

func (s *SimpleDiscoveryClient) Init(ctx context.Context, metadata nameresolution.Metadata) error {

	var meta SimpleDiscoveryClientMetadata
	err := mapstructure.Decode(metadata.Configuration, &meta)
	if err != nil {
		return err
	}

	instances := meta.Instances

	instMap := make(map[string][]*ServiceInstance, len(instances))
	for svc, list := range instances {
		instList := make([]*ServiceInstance, len(list))
		for i, inst := range list {
			instList[i] = &ServiceInstance{
				InstanceID: inst.InstanceID,
				ServiceID:  svc,
				Host:       inst.Host,
				Port:       inst.Port,
				Metadata:   cloneStringMap(inst.Metadata),
			}
		}
		instMap[svc] = instList
	}
	s.instances = instMap
	return nil
}

func (s *SimpleDiscoveryClient) Instances(req nameresolution.ResolveRequest) ([]*ServiceInstance, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if insts, ok := s.instances[req.ID]; ok {
		copied := make([]*ServiceInstance, len(insts))
		copy(copied, insts)
		return copied, nil
	}
	return nil, errors.New("service not found: " + req.ID)
}

func (s *SimpleDiscoveryClient) Services() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	services := make([]string, 0, len(s.instances))
	for svc := range s.instances {
		services = append(services, svc)
	}
	return services, nil
}

func cloneStringMap(src map[string]string) map[string]string {
	if src == nil {
		return nil
	}
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
