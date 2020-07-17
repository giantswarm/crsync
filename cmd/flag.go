package cmd

import "github.com/spf13/cobra"

const (
	flagMetricsPort = "metrics-port"
)

type flag struct {
	MetricsPort int
}

func (f *flag) Init(cmd *cobra.Command) {
	cmd.Flags().IntVar(&f.MetricsPort, flagMetricsPort, 0, "Port on which metrics are served. 0 disables metrics.")
}

func (f *flag) Validate() error {
	return nil
}
