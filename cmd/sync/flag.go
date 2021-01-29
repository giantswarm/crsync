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
	flagDstRegistryName            = "dst-name"
	flagDstRegistryUser            = "dst-user"
	flagDstRegistryPassword        = "dst-password"
	flagSrcRegistryName            = "src-name"
	flagSrcRegistryUser            = "src-user"
	flagSrcRegistryPassword        = "src-password"
	flagLastModified               = "last-modified"
	flagLoop                       = "loop"
	flagIncludePrivateRepositories = "include-private-repositories"
	flagMetricsPort                = "metrics-port"
	flagQuayAPIToken               = "quay-api-token" // nolint
	flagSyncInterval               = "sync-interval"
)

type flag struct {
	DstRegistryName            string
	DstRegistryUser            string
	DstRegistryPassword        string
	SrcRegistryName            string
	SrcRegistryUser            string
	SrcRegistryPassword        string
	LastModified               time.Duration
	Loop                       bool
	IncludePrivateRepositories bool
	MetricsPort                int
	QuayAPIToken               string
	SyncInterval               int
}

func (f *flag) Init(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.DstRegistryName, flagDstRegistryName, "", `Destination container registry name. E.g.: "docker.io".`)
	cmd.Flags().StringVar(&f.DstRegistryUser, flagDstRegistryUser, "", `Destination container registry user.`)
	cmd.Flags().StringVar(&f.DstRegistryPassword, flagDstRegistryPassword, "", fmt.Sprintf(`Destination container registry password. Defaults to %s environment variable.`, env.DstRegistryPassword))
	cmd.Flags().StringVar(&f.SrcRegistryName, flagSrcRegistryName, "", `Source container registry name. E.g.: "quay.io".`)
	cmd.Flags().StringVar(&f.SrcRegistryUser, flagSrcRegistryUser, "", `Source container registry user.`)
	cmd.Flags().StringVar(&f.SrcRegistryPassword, flagSrcRegistryPassword, "", fmt.Sprintf(`Source container registry password. Defaults to %s environment variable.`, env.SrcRegistryPassword))
	cmd.Flags().DurationVar(&f.LastModified, flagLastModified, time.Hour, `Duration in time when source repository was last modified.`)
	cmd.Flags().BoolVar(&f.Loop, flagLoop, false, "Whether to run the job continuously.")
	cmd.Flags().BoolVar(&f.IncludePrivateRepositories, flagIncludePrivateRepositories, false, "Whether to synchronize private repositories.")
	cmd.Flags().IntVar(&f.MetricsPort, flagMetricsPort, 0, "Port on which metrics are served. 0 disables metrics.")
	cmd.Flags().StringVar(&f.QuayAPIToken, flagQuayAPIToken, "", fmt.Sprintf(`Quay container registry API token. Defaults to %s environment variable.`, env.QuayAPIToken))
	cmd.Flags().IntVar(&f.SyncInterval, flagSyncInterval, 30, "Interval(seconds) between two syncs when running in a loop.")

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
	if f.SrcRegistryName == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagSrcRegistryName)
	}
	if f.SrcRegistryUser == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagSrcRegistryUser)
	}
	if f.SrcRegistryPassword == "" {
		f.SrcRegistryPassword = os.Getenv(env.SrcRegistryPassword)
	}
	if f.SrcRegistryName == sourceRegistryName && f.QuayAPIToken == "" {
		f.QuayAPIToken = os.Getenv(env.QuayAPIToken)
	}

	return nil
}
