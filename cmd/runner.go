package cmd

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"

	"github.com/giantswarm/crsync/pkg/project"
)

type runner struct {
	flag   *flag
	logger micrologger.Logger
	stdout io.Writer
	stderr io.Writer
}

func (r *runner) PersistentPreRun(cmd *cobra.Command, args []string) error {
	fmt.Printf("Version = %#q\n", project.Version())
	fmt.Printf("Git SHA = %#q\n", project.GitSHA())
	fmt.Printf("Command = %#q\n", cmd.Name())
	fmt.Println()

	return nil
}

func (r *runner) Run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	err := r.flag.Validate()
	if err != nil {
		return microerror.Mask(err)
	}

	if r.flag.MetricsPort != 0 {
		go func() {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("serving metrics at :%d", r.flag.MetricsPort))
			http.Handle("/metrics", promhttp.HandlerFor(
				prometheus.DefaultGatherer,
				promhttp.HandlerOpts{},
			))
			err := http.ListenAndServe(fmt.Sprintf(":%d", r.flag.MetricsPort), nil)
			if err != nil {
				r.logger.LogCtx(ctx, "level", "error", "message", "failed serving metrics", "stack", microerror.Mask(err))
			}
		}()
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "metrics disabled")
	}

	err = r.run(ctx, cmd, args)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *runner) run(ctx context.Context, cmd *cobra.Command, args []string) error {
	err := cmd.Help()
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
