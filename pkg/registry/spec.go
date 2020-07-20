package registry

import "context"

type Interface interface {
	Login(ctx context.Context, user, password string) error
	Logout(ctx context.Context) error
	ListRepositories(ctx context.Context) ([]string, error)
	ListTags(ctx context.Context, repository string) ([]string, error)
	Name() string
	Pull(ctx context.Context, repo, tag string) error
	Push(ctx context.Context, repo, tag string) error
	RemoveImage(ctx context.Context, repo, tag string) error
}
