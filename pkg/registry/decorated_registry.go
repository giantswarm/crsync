package registry

import (
	"context"

	"github.com/giantswarm/microerror"
	"golang.org/x/time/rate"
)

type DecoratedRegistryConfig struct {
	RateLimiter DecoratedRegistryConfigRateLimiter
	Underlying  Interface
}

type DecoratedRegistryConfigRateLimiter struct {
	ListRepositories *rate.Limiter
	ListTags         *rate.Limiter
	Pull             *rate.Limiter
	Push             *rate.Limiter
}

type DecoratedRegistry struct {
	rateLimiter DecoratedRegistryConfigRateLimiter
	underlying  Interface
}

func NewDecoratedRegistry(config DecoratedRegistryConfig) (*DecoratedRegistry, error) {
	if config.RateLimiter.ListRepositories == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.RateLimiter.ListRepositories must not be empty", config)
	}
	if config.RateLimiter.ListTags == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.RateLimiter.ListTags must not be empty", config)
	}
	if config.RateLimiter.Pull == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.RateLimiter.Pull must not be empty", config)
	}
	if config.RateLimiter.Push == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.RateLimiter.Push must not be empty", config)
	}

	if config.Underlying == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Underlying must not be empty", config)
	}

	r := &DecoratedRegistry{
		rateLimiter: config.RateLimiter,
		underlying:  config.Underlying,
	}

	return r, nil
}

func (r *DecoratedRegistry) Login(ctx context.Context, user, password string) error {
	return microerror.Mask(r.underlying.Login(ctx, user, password))
}

func (r *DecoratedRegistry) Logout(ctx context.Context) error {
	return microerror.Mask(r.underlying.Logout(ctx))
}

func (r *DecoratedRegistry) ListRepositories(ctx context.Context) ([]string, error) {
	var err error

	err = r.rateLimiter.ListRepositories.Wait(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	rs, err := r.underlying.ListRepositories(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return rs, nil
}

func (r *DecoratedRegistry) ListTags(ctx context.Context, repository string) ([]string, error) {
	var err error

	err = r.rateLimiter.ListTags.Wait(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	ts, err := r.underlying.ListTags(ctx, repository)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return ts, nil
}

func (r DecoratedRegistry) Name() string {
	return r.underlying.Name()
}

func (r *DecoratedRegistry) Pull(ctx context.Context, repo, tag string) error {
	var err error

	err = r.rateLimiter.Pull.Wait(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.underlying.Pull(ctx, repo, tag)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *DecoratedRegistry) Push(ctx context.Context, repo, tag string) error {
	var err error

	err = r.rateLimiter.Push.Wait(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.underlying.Push(ctx, repo, tag)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *DecoratedRegistry) RemoveImage(ctx context.Context, repo, tag string) error {
	err := r.underlying.RemoveImage(ctx, repo, tag)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
