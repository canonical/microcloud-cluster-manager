package types

// MemberConfigKey represents the valid keys for Cluster Manager member configs.
type MemberConfigKey string

// Valid member config keys.
const (
	HTTPSAddress    MemberConfigKey = "https_address"
	ExternalAddress MemberConfigKey = "external_address"
)

// MemberConfigPatch represents the payload required to update configs for a single Cluster Manager member.
type MemberConfigPatch struct {
	Config map[MemberConfigKey]string `json:"config" yaml:"config"`
}

// MemberConfig represents config data for a single Cluster Manager member, which includes the member name.
type MemberConfig struct {
	Target string `json:"target"`
	MemberConfigPatch
}

// ValidMemberConfigKeys returns a map of valid member config keys.
func ValidMemberConfigKeys() map[MemberConfigKey]bool {
	return map[MemberConfigKey]bool{
		HTTPSAddress:    true,
		ExternalAddress: true,
	}
}
