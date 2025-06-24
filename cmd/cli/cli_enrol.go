package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/canonical/lxd/shared"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/api/models/v1"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/config"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/database"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/database/store"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
)

type cmdEnrol struct {
	CFG             *config.Config
	DB              *database.DB
	flagDescription string
	flagExpire      string
}

// Command returns the subcommand for initializing a MicroCloud.
func (c *cmdEnrol) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "enrol",
		Short:   "Enrol a new remote cluster",
		Long:    "Generates a join token to be consumed by MicroCloud",
		Example: "microcloud-cluster-manager enrol my-cluster --expire 2042-05-23T17:00:00Z --description 'My Cluster in Quebec'",
		RunE:    c.Run,
	}

	cmd.Flags().StringVarP(&c.flagDescription, "description", "d", "", "Description for the cluster.")
	cmd.Flags().StringVarP(&c.flagExpire, "expire", "e", "", "Expiry for the join token.")

	return cmd
}

// Run execs the enrol command to create a new remote cluster join token.
func (c *cmdEnrol) Run(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return cmd.Help()
	}

	payload := models.RemoteClusterTokenPost{
		ClusterName: args[0],
		Description: c.flagDescription,
	}

	if c.flagExpire == "" {
		payload.Expiry = time.Now().Add(time.Hour * 24)
	} else {
		t, err := time.Parse(time.RFC3339, c.flagExpire)
		if err != nil {
			return err
		}
		if t.Before(time.Now()) {
			return fmt.Errorf("expire flag must be in the future, got %s", payload.Expiry)
		}
		payload.Expiry = t
	}

	secret, err := shared.RandomCryptoString()
	if err != nil {
		return err
	}

	// create the token to be sent
	cert, err := c.CFG.ClusterConnectorCert.PublicKeyX509()
	if err != nil {
		return err
	}

	// get the cluster-connector service address for the token payload
	clusterConnectorAddress := c.CFG.ClusterConnectorAddress

	token := models.RemoteClusterTokenBody{
		Secret:      secret,
		ExpiresAt:   payload.Expiry,
		Addresses:   []string{clusterConnectorAddress},
		ServerName:  payload.ClusterName,
		Fingerprint: shared.CertFingerprint(cert),
	}
	encodedToken, err := token.Encode()
	if err != nil {
		return err
	}

	//store token details in the database
	err = c.DB.Transaction(context.Background(), func(ctx context.Context, tx *sqlx.Tx) error {
		var err error
		isNameTaken, err := store.RemoteClusterExists(ctx, tx, payload.ClusterName)
		if err != nil {
			return err
		}
		if isNameTaken {
			return fmt.Errorf("cluster name already exists")
		}

		tokenData := store.RemoteClusterToken{
			ClusterName:  payload.ClusterName,
			Description:  payload.Description,
			EncodedToken: encodedToken,
			Expiry:       payload.Expiry,
			CreatedAt:    time.Now(),
		}
		_, err = store.CreateRemoteClusterToken(ctx, tx, tokenData)

		return err
	})

	if err != nil {
		return err
	}

	println("Join token for cluster created. To finish the enrollment, use the token below on any member of the MicroCloud you are going to join\n" + encodedToken)

	return nil
}
