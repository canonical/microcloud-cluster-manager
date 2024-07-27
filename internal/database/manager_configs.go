package database

//go:generate -command mapper lxd-generate db mapper -t manager_configs.mapper.go
//go:generate mapper reset
//
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e manager_config objects table=manager_configs
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e manager_config objects-by-Key table=manager_configs
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e manager_config id table=manager_configs
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e manager_config create table=manager_configs
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e manager_config delete-by-Key table=manager_configs
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e manager_config update table=manager_configs
//
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e manager_config GetMany table=manager_configs
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e manager_config ID table=manager_configs
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e manager_config Exists table=manager_configs
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e manager_config Create table=manager_configs
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e manager_config DeleteOne-by-Key table=manager_configs
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e manager_config Update table=manager_configs

// ManagerConfig configurations for the entire cluster.
type ManagerConfig struct {
	Key   string `db:"primary=yes"`
	ID    int
	Value string
}

// ManagerConfigFilter is a required struct for use with lxd-generate. It is used for filtering fields on database fetches.
type ManagerConfigFilter struct {
	Key *string
}
