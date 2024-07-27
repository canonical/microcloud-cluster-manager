package database

//go:generate -command mapper lxd-generate db mapper -t manager_member_configs.mapper.go
//go:generate mapper reset
//
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e manager_member_config objects table=manager_member_configs
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e manager_member_config objects-by-Target table=manager_member_configs
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e manager_member_config id table=manager_member_configs
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e manager_member_config create table=manager_member_configs
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e manager_member_config delete-by-Target table=manager_member_configs
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e manager_member_config update table=manager_member_configs
//
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e manager_member_config GetMany
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e manager_member_config ID
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e manager_member_config Exists
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e manager_member_config Create
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e manager_member_config DeleteOne-by-Target
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e manager_member_config Update

// ManagerMemberConfig represents configs for a single Cluster Manager member.
type ManagerMemberConfig struct {
	Target          string `db:"primary=yes"`
	ID              int
	HTTPSAddress    string
	ExternalAddress string
}

// ManagerMemberConfigFilter is a required struct for use with lxd-generate. It is used for filtering fields on database fetches.
type ManagerMemberConfigFilter struct {
	Target *string
}
