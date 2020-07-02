package cmd

import (
	"fmt"
	"os"

	"github.com/giantswarm/crsync/internal/env"
	"github.com/giantswarm/microerror"

	"github.com/spf13/cobra"
)

const (
	flagDstRegistryName     = "dst-name"
	flagDstRegistryUser     = "dst-user"
	flagDstRegistryPassword = "dst-password"
)

type flag struct {
	DstRegistryName     string
	DstRegistryUser     string
	DstRegistryPassword string
}

func (f *flag) Init(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.DstRegistryName, flagDstRegistryName, "", "Destination container registry name.")
	cmd.Flags().StringVar(&f.DstRegistryUser, flagDstRegistryUser, "", "Destination container registry user.")
	cmd.Flags().StringVar(&f.DstRegistryPassword, flagDstRegistryPassword, os.Getenv(env.DstRegistryPassword), fmt.Sprintf(`Destination container registry password. Defaults to %s environment variable.`, env.DstRegistryPassword))
}

func (f *flag) Validate() error {
	if f.DstRegistryName == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagDstRegistryName)
	}
	if f.DstRegistryUser == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagDstRegistryUser)
	}
	if f.DstRegistryPassword == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagDstRegistryPassword)
	}

	return nil
}
