package types

// MemberConfigPatch represents the payload required to update configs for a single site manager member.
type MemberConfigPatch struct {
	HTTPSAddress    string `json:"https_address,omitempty"`
	ExternalAddress string `json:"external_address,omitempty"`
}

// MemberConfig represents config data for a single site manager member, which includes the member name.
type MemberConfig struct {
	Target string `json:"target"`
	MemberConfigPatch
}
