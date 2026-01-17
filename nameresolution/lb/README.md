
```yaml
apiVersion: dapr.io/v1alpha1
kind: Configuration
metadata:
  name: appconfig
spec:
  nameResolution:
    component: "lb"
    version: "v1"
    configuration:
      loadBalancer: 
        - component: round-robin # or random
          configuration:
            hello: world
      serviceInstanceListSupplier: 
        - component: zonePreference
          configuration:
            hello: world
      discoveryClient:
        - component: simple
          configuration:
            instances: 
                hello:
                    - instanceID: 1
                      host: 127.0.0.1
                      port: 80
                      metadata: 
                        hello: world
```