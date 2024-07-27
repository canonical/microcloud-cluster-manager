package types

import "fmt"

// ManagerConfigs represents the Cluster Manager configs that are cluster wide.
type ManagerConfigs struct {
	Config map[string]string `json:"config" yaml:"config"`
}

// ValidManagerConfigKeys returns a map of valid manager config keys.
func ValidManagerConfigKeys() map[string]bool {
	return map[string]bool{
		"oidc.issuer":    true,
		"oidc.client.id": true,
		"oidc.audience":  true,
		"global.address": true,
	}
}

// ValidateManagerConfigKeys validates the given manager config keys that they exist in the map of valid config keys for Cluster Manager.
func ValidateManagerConfigKeys(config map[string]string) error {
	for k := range config {
		_, ok := ValidManagerConfigKeys()[k]
		if !ok {
			return fmt.Errorf("invalid config key: %s", k)
		}
	}

	return nil
}
