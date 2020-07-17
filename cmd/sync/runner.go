package sync

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"golang.org/x/time/rate"

	"github.com/giantswarm/crsync/internal/key"
	"github.com/giantswarm/crsync/pkg/azurecr"
	"github.com/giantswarm/crsync/pkg/dockerhub"
	"github.com/giantswarm/crsync/pkg/quay"
	"github.com/giantswarm/crsync/pkg/registry"
)

const (
	sourceRegistryName = "quay.io"

	listBurst     = 1
	pullPushBurst = 30
)

type runner struct {
	flag   *flag
	logger micrologger.Logger
	stdout io.Writer
	stderr io.Writer
}

func (r *runner) Run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	err := r.flag.Validate()
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.run(ctx, cmd, args)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *runner) run(ctx context.Context, cmd *cobra.Command, args []string) error {
	var err error

	fmt.Printf("Source registry       = %#q\n", sourceRegistryName)
	fmt.Printf("Destination registry  = %#q\n", r.flag.DstRegistryName)

	var srcRegistryClient registry.RegistryClient
	{
		c := quay.Config{
			Namespace:    key.Namespace,
			LastModified: r.flag.LastModified,
		}

		srcRegistryClient, err = quay.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var srcRegistry registry.Interface
	{
		c := registry.Config{
			Name:           sourceRegistryName,
			RegistryClient: srcRegistryClient,
		}

		srcRegistry, err = registry.New(c)
		if err != nil {
			return microerror.Mask(err)
		}

		srcRegistry, err = newDecoratedRegistry(srcRegistry)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var dstRegistryClient registry.RegistryClient
	{
		switch registryName := r.flag.DstRegistryName; {
		case registryName == "docker.io":
			c := dockerhub.Config{}

			dstRegistryClient, err = dockerhub.New(c)
			if err != nil {
				return microerror.Mask(err)
			}
		case strings.HasSuffix(registryName, "azurecr.io"):
			c := azurecr.Config{
				RegistryName: r.flag.DstRegistryName,
			}

			dstRegistryClient, err = azurecr.New(c)
			if err != nil {
				return microerror.Mask(err)
			}

		default:
			return microerror.Maskf(executionFailedError, "unknown container registry %#q", registryName)
		}
	}

	var dstRegistry registry.Interface
	{
		config := registry.Config{
			Name:           r.flag.DstRegistryName,
			RegistryClient: dstRegistryClient,
		}

		dstRegistry, err = registry.New(config)
		if err != nil {
			return microerror.Mask(err)
		}

		dstRegistry, err = newDecoratedRegistry(dstRegistry)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	if !r.flag.Loop {
		err := r.sync(ctx, srcRegistry, dstRegistry)
		if err != nil {
			return microerror.Mask(err)
		}

		return nil
	}

	if r.flag.MetricsPort != 0 {
		go func() {
			fmt.Printf("Serving metrics at :%d", r.flag.MetricsPort)
			http.Handle("/metrics", promhttp.HandlerFor(
				prometheus.DefaultGatherer,
				promhttp.HandlerOpts{},
			))
			err := http.ListenAndServe(fmt.Sprintf(":%d", r.flag.MetricsPort), nil)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed serving metrics: %v", microerror.Mask(err))
			}
		}()
	} else {
		fmt.Println("Metrics disabled")
	}

	for {
		start := time.Now()

		err := r.sync(ctx, srcRegistry, dstRegistry)
		if err != nil {
			fmt.Fprintf(os.Stderr, "\nSync error:\n%s\n\n", microerror.JSON(err))
		} else {
			fmt.Printf("\nTook %s\n", time.Since(start))
		}

		time.Sleep(10 * time.Second)
	}
}

func (r *runner) sync(ctx context.Context, srcRegistry, dstRegistry registry.Interface) error {
	var err error

	fmt.Println()
	fmt.Printf("Logging in destination registry...\n")
	err = dstRegistry.Login(ctx, r.flag.DstRegistryUser, r.flag.DstRegistryPassword)
	if err != nil {
		return microerror.Mask(err)
	}
	defer func(ctx context.Context) {
		fmt.Println()
		fmt.Printf("Logging out of destination registry...\n")
		_ = dstRegistry.Logout(ctx)
	}(ctx)

	// getTagsJobCh channel has buffer 4 times bigger listing tags/repos burst to not starve
	// processing.
	getTagsJobCh := make(chan getTagsJob, listBurst*4)

	// retagJobCh channel has buffer 2 times bigger push/pull burst to not starve
	// processing.
	retagJobCh := make(chan retagJob, pullPushBurst*2)

	processGetTagsJobErrCh := make(chan error)
	processRetagJobsErrCh := make(chan error)
	go func(ctx context.Context) {
		processGetTagsJobErrCh <- r.processGetTagsJobs(ctx, getTagsJobCh, retagJobCh)
	}(ctx)
	go func(ctx context.Context) {
		processRetagJobsErrCh <- r.processRetagJobs(ctx, retagJobCh)
	}(ctx)
	defer close(processGetTagsJobErrCh)
	defer close(processRetagJobsErrCh)

	fmt.Println()
	fmt.Printf("Reading list of repositories to sync from source registry...\n")
	reposToSync, err := srcRegistry.ListRepositories(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	fmt.Printf("There are %d repositories to sync.\n", len(reposToSync))
	if len(reposToSync) > 0 {
		fmt.Println()
	}

	for repoIndex, repo := range reposToSync {
		job := getTagsJob{
			Src: srcRegistry,
			Dst: dstRegistry,

			ID:   fmt.Sprintf("Repository [%d/%d] = %#q", repoIndex+1, len(reposToSync), repo),
			Repo: repo,
		}

		select {
		case <-ctx.Done():
			return microerror.Mask(ctx.Err())
		case err := <-processGetTagsJobErrCh:
			errorsTotal.Inc()
			return microerror.Mask(err)
		case err := <-processRetagJobsErrCh:
			errorsTotal.Inc()
			return microerror.Mask(err)
		case getTagsJobCh <- job:
			// Job added.
		}
	}

	// Wait for job processing to finish.
	{
		close(getTagsJobCh)
		err = <-processGetTagsJobErrCh
		if err != nil {
			return microerror.Mask(err)
		}

		close(retagJobCh)
		err = <-processRetagJobsErrCh
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func (r *runner) processGetTagsJobs(ctx context.Context, jobCh <-chan getTagsJob, resultCh chan retagJob) error {
	var wg sync.WaitGroup
	defer wg.Wait()

	for {
		select {
		case <-ctx.Done():
			return microerror.Mask(ctx.Err())
		case job, ok := <-jobCh:
			if !ok {
				return nil
			}

			wg.Add(1)

			go func(ctx context.Context, job getTagsJob) {
				defer wg.Done()
				start := time.Now()

				fmt.Printf("%s: Getting list of tags to sync...\n", job.ID)

				tags, err := r.processGetTagsJob(ctx, job)
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s: Failed to get list of tags to sync: %s\n", job.ID, microerror.JSON(err))
					return
				}

				fmt.Printf("%s: Scheduling %d tags to sync...\n", job.ID, len(tags))

				for i, t := range tags {
					j := retagJob{
						Src: job.Src,
						Dst: job.Dst,

						ID:   fmt.Sprintf("%s: Tag [%d/%d] = %#q", job.ID, i+1, len(tags), t),
						Repo: job.Repo,
						Tag:  t,
					}

					select {
					case <-ctx.Done():
						fmt.Fprintf(os.Stderr, "%s: Cancelled while scheduling %d/%d job: %s\n", job.ID, i+1, len(tags), microerror.JSON(err))
					case resultCh <- j:
						// ok
					}
				}

				fmt.Printf("%s: Done (took %s)\n", job.ID, time.Since(start))
			}(ctx, job)
		}
	}
}

func (r *runner) processRetagJobs(ctx context.Context, jobCh <-chan retagJob) error {
	var wg sync.WaitGroup
	defer wg.Wait()

	for {
		select {
		case <-ctx.Done():
			return microerror.Mask(ctx.Err())
		case job, ok := <-jobCh:
			if !ok {
				return nil
			}

			wg.Add(1)

			go func(ctx context.Context, job retagJob) {
				defer wg.Done()
				start := time.Now()

				fmt.Printf("%s: Retagging...\n", job.ID)

				err := r.processRetagJob(ctx, job)
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s: Failed to retag: %s\n", job.ID, microerror.JSON(err))
					return
				}

				fmt.Printf("%s: Done (took %s)\n", job.ID, time.Since(start))
			}(ctx, job)
		}
	}
}

func (r *runner) processGetTagsJob(ctx context.Context, job getTagsJob) ([]string, error) {
	var srcTags, dstTags []string

	eg := new(errgroup.Group)
	eg.Go(func() error {
		var err error
		srcTags, err = job.Src.ListTags(ctx, job.Repo)
		tagsTotal.WithLabelValues(job.Src.Name(), job.Repo).Set(float64(len(srcTags)))
		return microerror.Mask(err)
	})
	eg.Go(func() error {
		var err error
		dstTags, err = job.Dst.ListTags(ctx, job.Repo)
		tagsTotal.WithLabelValues(job.Dst.Name(), job.Repo).Set(float64(len(srcTags)))
		return microerror.Mask(err)
	})
	err := eg.Wait()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	tags := sliceDiff(srcTags, dstTags)

	return tags, nil
}

func (r *runner) processRetagJob(ctx context.Context, job retagJob) error {
	err := job.Src.Pull(ctx, job.Repo, job.Tag)
	if err != nil {
		return microerror.Mask(err)
	}

	err = registry.RetagImage(job.Repo, job.Tag, sourceRegistryName, r.flag.DstRegistryName)
	if err != nil {
		// Try to remove the image by best effort in case of error.
		_ = job.Src.RemoveImage(ctx, job.Repo, job.Tag)
		return microerror.Mask(err)
	}

	err = job.Src.RemoveImage(ctx, job.Repo, job.Tag)
	if err != nil {
		return microerror.Mask(err)
	}

	err = job.Dst.Push(ctx, job.Repo, job.Tag)
	if err != nil {
		// Try to remove the image by best effort in case of error.
		_ = job.Dst.RemoveImage(ctx, job.Repo, job.Tag)
		return microerror.Mask(err)
	}

	err = job.Dst.RemoveImage(ctx, job.Repo, job.Tag)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func newDecoratedRegistry(reg registry.Interface) (*registry.DecoratedRegistry, error) {
	c := registry.DecoratedRegistryConfig{
		RateLimiter: registry.DecoratedRegistryConfigRateLimiter{
			ListRepositories: rate.NewLimiter(rate.Every(5*time.Second), listBurst),
			ListTags:         rate.NewLimiter(rate.Every(1*time.Second), listBurst),
			Pull:             rate.NewLimiter(rate.Every(100*time.Millisecond), pullPushBurst),
			Push:             rate.NewLimiter(rate.Every(100*time.Millisecond), pullPushBurst),
		},
		Underlying: reg,
	}

	r, err := registry.NewDecoratedRegistry(c)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return r, nil
}

func sliceDiff(s1, s2 []string) []string {
	var result []string

	for _, e1 := range s1 {
		found := false
		for _, e2 := range s2 {
			if e1 == e2 {
				found = true
				break
			}
		}
		if !found {
			result = append(result, e1)
		}
	}

	return result
}
