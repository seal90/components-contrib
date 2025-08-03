package configurestring

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/dapr/components-contrib/nameresolution"
	"github.com/dapr/kit/config"
	"github.com/dapr/kit/logger"
)

type address struct {
	IPV4         string            `json:"ipv4"`
	IPV6         string            `json:"ipv6"`
	Port         int               `json:"port,string"`
	ExtendedInfo map[string]string `json:"extendedInfo"`
}

type Resolver struct {
	appAddress map[string][]address
	configure  configStringMetadata
	// shutdown refreshes.
	runCtx    context.Context
	runCancel context.CancelFunc
	logger    logger.Logger
}

// NewResolver creates the instance of mDNS name resolver.
func NewResolver(logger logger.Logger) nameresolution.Resolver {
	runCtx, runCancel := context.WithCancel(context.Background())

	r := &Resolver{
		appAddress: make(map[string][]address),
		// shutdown refreshers
		runCtx:    runCtx,
		runCancel: runCancel,
		logger:    logger,
	}

	return r
}

// Init configString.
func (s *Resolver) Init(ctx context.Context, metadata nameresolution.Metadata) error {
	if metadata.Instance.AppID == "" {
		return errors.New("name is missing")
	}
	if metadata.Instance.Address == "" {
		return errors.New("address is missing")
	}
	if metadata.Instance.DaprInternalPort <= 0 {
		return errors.New("port is missing or invalid")
	}

	s.configure.InitWithMetadata(metadata)

	appAddrs, err := resolveAppAddress(s.configure)
	if err != nil {
		return err
	}

	s.appAddress = appAddrs
	return nil
}

func resolveAppAddress(configure configStringMetadata) (map[string][]address, error) {
	switch configure.valueType {
	case JSON_STRING_VALUE:
		return parseAddressConfig(configure.value)
	case YAML_STRING_VALUE:
		return parseAddressConfig(configure.value)
	case JSON_FILE_VALUE:
		// load data from path and then reload data every 2s
	case YAML_FILE_VALUE:
	}
	return nil, nil
}

func parseAddressConfig(rawConfig any) (map[string][]address, error) {
	rawConfig, err := config.Normalize(rawConfig)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(rawConfig)
	if err != nil {
		return nil, fmt.Errorf("error serializing to json: %w", err)
	}

	configuration := make(map[string][]address)
	err = json.Unmarshal(data, &configuration)
	if err != nil {
		return nil, fmt.Errorf("error deserializing to configSpec: %w", err)
	}
	return configuration, nil
}

// ResolveID resolves name to address.
func (s *Resolver) ResolveID(ctx context.Context, req nameresolution.ResolveRequest) (addr string, err error) {

	var eligibleAddresses []address
	appAdress, exists := s.appAddress[req.ID]
	if exists {
		virtualEnv, exists := req.Data["virtual-namespace"]
		if exists {
			eligibleAddresses = sameVirtualNamespace(appAdress, virtualEnv)
		}
	}
	if len(eligibleAddresses) == 0 {
		eligibleAddresses = noneVirtualNamespace(appAdress)
	}
	if len(eligibleAddresses) > 0 {
		appAddress := eligibleAddresses[rand.Int()%len(eligibleAddresses)]
		addr = appAddress.IPV4 + ":" + strconv.Itoa(appAddress.Port)
	}
	return addr, nil
}

func sameVirtualNamespace(addresses []address, reqVirtualEnv string) []address {
	var result []address
	for _, address := range addresses {
		if virtualEnv, exists := address.ExtendedInfo["virtual-namespace"]; exists {
			if reqVirtualEnv == virtualEnv {
				result = append(result, address)
			}
		}
	}
	return result
}

func noneVirtualNamespace(addresses []address) []address {
	var result []address
	for _, address := range addresses {
		if _, exists := address.ExtendedInfo["virtual-namespace"]; !exists {
			result = append(result, address)
		}
	}
	return result
}

// Close implements io.Closer.
func (s *Resolver) Close() (err error) {
	errs := make([]error, 0)
	return errors.Join(errs...)
}
