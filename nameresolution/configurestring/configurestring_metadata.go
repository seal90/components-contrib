package configurestring

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/dapr/components-contrib/nameresolution"
	"github.com/dapr/kit/config"
)

const (
	JSON_STRING_VALUE = "jsonStringValue"
	YAML_STRING_VALUE = "yamlStringValue"
	JSON_FILE_VALUE   = "jsonFileValue"
	YAML_FILE_VALUE   = "yamlFileValue"
)

type configStringMetadata struct {
	valueType string
	value     any

	// Instance properties - these are passed by the runtime
	appID       string
	namespace   string
	hostAddress string
	port        int
}

func (m *configStringMetadata) InitWithMetadata(meta nameresolution.Metadata) error {

	// Set and validate the instance properties
	m.appID = meta.Instance.AppID
	if m.appID == "" {
		return errors.New("name is missing")
	}
	m.hostAddress = meta.Instance.Address
	if m.hostAddress == "" {
		return errors.New("address is missing")
	}
	m.port = meta.Instance.DaprInternalPort
	if m.port == 0 {
		return errors.New("port is missing or invalid")
	}
	m.namespace = meta.Instance.Namespace // Can be empty

	configData, error := parseConfig(meta.Configuration)
	if error != nil {
		return error
	}
	m.valueType = configData.ValueType
	m.value = configData.Value
	return nil
}

type configData struct {
	ValueType string `json:"valueType"`
	Value     any    `json:"value"`
}

func parseConfig(rawConfig any) (configData, error) {
	var result configData
	rawConfig, err := config.Normalize(rawConfig)
	if err != nil {
		return result, err
	}

	data, err := json.Marshal(rawConfig)
	if err != nil {
		return result, fmt.Errorf("error serializing to json: %w", err)
	}

	configuration := configData{}
	err = json.Unmarshal(data, &configuration)
	if err != nil {
		return result, fmt.Errorf("error deserializing to configSpec: %w", err)
	}
	return configuration, nil
}
