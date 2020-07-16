package sync

import (
	"fmt"
	"os"
	"time"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/crsync/internal/env"

	"github.com/spf13/cobra"
)

const (
	flagDstRegistryName     = "dst-name"
	flagDstRegistryUser     = "dst-user"
	flagDstRegistryPassword = "dst-password"
	flagLastModified        = "last-modified"
	flagLoop                = "loop"
)

type flag struct {
	DstRegistryName     string
	DstRegistryUser     string
	DstRegistryPassword string
	LastModified        time.Duration
	Loop                bool
}

func (f *flag) Init(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.DstRegistryName, flagDstRegistryName, "", `Destination container registry name. E.g.: "docker.io".`)
	cmd.Flags().StringVar(&f.DstRegistryUser, flagDstRegistryUser, "", `Destination container registry user.`)
	cmd.Flags().StringVar(&f.DstRegistryPassword, flagDstRegistryPassword, "", fmt.Sprintf(`Destination container registry password. Defaults to %s environment variable.`, env.DstRegistryPassword))
	cmd.Flags().DurationVar(&f.LastModified, flagLastModified, time.Hour, `Duration in time when source repository was last modified.`)
	cmd.Flags().BoolVar(&f.Loop, flagLoop, false, "Whether to run the job continuously.")
}

func (f *flag) Validate() error {
	if f.DstRegistryName == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagDstRegistryName)
	}
	if f.DstRegistryUser == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagDstRegistryUser)
	}
	if f.DstRegistryPassword == "" {
		f.DstRegistryPassword = os.Getenv(env.DstRegistryPassword)
	}
	if f.DstRegistryPassword == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagDstRegistryPassword)
	}

	return nil
}
