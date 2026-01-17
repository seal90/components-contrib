package lb

import (
	"context"
	"log"

	"github.com/dapr/components-contrib/nameresolution"
	"github.com/dapr/kit/logger"
)

type DiscoveryClientServiceInstanceListSupplierFactory struct {
	logger logger.Logger
}

func NewDiscoveryClientServiceInstanceListSupplierFactory(logger logger.Logger) ServiceInstanceListSupplierFactory {
	return &DiscoveryClientServiceInstanceListSupplierFactory{
		logger: logger,
	}
}

func (s *DiscoveryClientServiceInstanceListSupplierFactory) Init(ctx context.Context, metadata nameresolution.Metadata) error {
	return nil
}

func (s *DiscoveryClientServiceInstanceListSupplierFactory) Create(id string, delegate ServiceInstanceListSupplier) ServiceInstanceListSupplier {
	return &DiscoveryClientServiceInstanceListSupplier{
		logger: s.logger,
	}
}

type DiscoveryClientServiceInstanceListSupplier struct {
	client DiscoveryClient
	logger logger.Logger
}

func (s *DiscoveryClientServiceInstanceListSupplier) InitWithClient(client DiscoveryClient) {
	s.client = client
}

func (s *DiscoveryClientServiceInstanceListSupplier) Instances(req nameresolution.ResolveRequest) ([]*ServiceInstance, error) {
	return s.client.Instances(req)
}

type ZonePreferenceServiceInstanceListSupplierFactory struct {
	logger logger.Logger
}

func NewZonePreferenceServiceInstanceListSupplierFactory(logger logger.Logger) ServiceInstanceListSupplierFactory {
	return &ZonePreferenceServiceInstanceListSupplierFactory{
		logger: logger,
	}
}

func (z *ZonePreferenceServiceInstanceListSupplierFactory) Init(ctx context.Context, metadata nameresolution.Metadata) error {
	return nil
}

func (z *ZonePreferenceServiceInstanceListSupplierFactory) Create(id string, delegate ServiceInstanceListSupplier) ServiceInstanceListSupplier {
	return &ZonePreferenceServiceInstanceListSupplier{
		DelegatingServiceInstanceListSupplier: DelegatingServiceInstanceListSupplier{
			delegate: delegate,
		},
		id:     id,
		logger: z.logger,
	}
}

type ZonePreferenceServiceInstanceListSupplier struct {
	DelegatingServiceInstanceListSupplier
	id     string
	logger logger.Logger
}

func (z *ZonePreferenceServiceInstanceListSupplier) Instances(req nameresolution.ResolveRequest) ([]*ServiceInstance, error) {

	allInstances, err := z.delegate.Instances(req)
	if err != nil {
		return nil, err
	}

	callerZone := req.Data["zone"]
	if callerZone == "" {
		return allInstances, nil
	}

	var sameZoneInstances []*ServiceInstance
	for _, inst := range allInstances {
		instZone := inst.Metadata["zone"]
		if instZone == callerZone {
			sameZoneInstances = append(sameZoneInstances, inst)
		}
	}

	if len(sameZoneInstances) > 0 {
		log.Printf("[ZoneAffinity] Found %d instances in zone '%s' for service '%s'",
			len(sameZoneInstances), callerZone, req.ID)
		return sameZoneInstances, nil
	}

	log.Printf("[ZoneAffinity] No instances in zone '%s', falling back to all %d instances",
		callerZone, len(allInstances))
	return allInstances, nil
}
