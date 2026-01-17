package lb

func init() {
	DefaultRegistry.RegisterLoadBalancerComponent(NewRoundRobinLoadBalancerFactory, "round-robin")
	DefaultRegistry.RegisterLoadBalancerComponent(NewRandomLoadBalancerFactory, "random")

	DefaultRegistry.RegisterServiceInstanceListSupplierComponent(NewDiscoveryClientServiceInstanceListSupplierFactory, "discovery-client")
	DefaultRegistry.RegisterServiceInstanceListSupplierComponent(NewZonePreferenceServiceInstanceListSupplierFactory, "zone-preference")

	DefaultRegistry.RegisterDiscoveryClientComponent(NewCompositeDiscoveryClient, "composite-discovery")
	DefaultRegistry.RegisterDiscoveryClientComponent(NewSimpleDiscoveryClient, "simple")

}
