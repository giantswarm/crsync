package registry

import "context"

type RegistryClient interface {
	Authorize(ctx context.Context, user, password string) error
	ListRepositories(ctx context.Context) ([]string, error)
	ListTags(ctx context.Context, repositry string) ([]string, error)
}
