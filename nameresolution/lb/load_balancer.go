package lb

import (
	"context"
	"errors"
	"math/rand"
	"sync/atomic"

	"github.com/dapr/components-contrib/nameresolution"
	"github.com/dapr/kit/logger"
)

type RoundRobinLoadBalancerFactory struct {
	logger logger.Logger
}

func NewRoundRobinLoadBalancerFactory(logger logger.Logger) LoadBalancerFactory {
	return &RoundRobinLoadBalancerFactory{
		logger: logger,
	}
}

func (r *RoundRobinLoadBalancerFactory) Init(ctx context.Context, metadata nameresolution.Metadata) error {
	return nil
}

func (r *RoundRobinLoadBalancerFactory) Create(id string, discovery ServiceInstanceListSupplier) LoadBalancer {
	return &RoundRobinLoadBalancer{
		id:        id,
		counter:   0,
		discovery: discovery,
		logger:    r.logger,
	}
}

type RoundRobinLoadBalancer struct {
	id        string
	counter   uint64
	discovery ServiceInstanceListSupplier
	logger    logger.Logger
}

func (lb *RoundRobinLoadBalancer) Choose(req nameresolution.ResolveRequest) (*ServiceInstance, error) {
	instances, err := lb.discovery.Instances(req)
	if err != nil || len(instances) == 0 {
		return nil, errors.New("no available instances for service: " + req.ID)
	}

	idx := atomic.AddUint64(&lb.counter, 1) % uint64(len(instances))
	return instances[idx], nil
}

type RandomLoadBalancerFactory struct {
	logger logger.Logger
}

func NewRandomLoadBalancerFactory(logger logger.Logger) LoadBalancerFactory {
	return &RandomLoadBalancerFactory{
		logger: logger,
	}
}

func (r *RandomLoadBalancerFactory) Init(ctx context.Context, metadata nameresolution.Metadata) error {
	return nil
}

func (r *RandomLoadBalancerFactory) Create(id string, discovery ServiceInstanceListSupplier) LoadBalancer {
	return &RandomLoadBalancer{
		id:        id,
		discovery: discovery,
		logger:    r.logger,
	}
}

type RandomLoadBalancer struct {
	id        string
	discovery ServiceInstanceListSupplier
	logger    logger.Logger
}

func (rlb *RandomLoadBalancer) Choose(req nameresolution.ResolveRequest) (*ServiceInstance, error) {

	instances, err := rlb.discovery.Instances(req)
	if err != nil || len(instances) == 0 {
		return nil, errors.New("no available instances for service: " + req.ID)
	}

	n := len(instances)
	if n == 0 {
		return nil, errors.New("no available instances for service: " + req.ID)
	}
	index := rand.Intn(n)
	return instances[index], nil
}
