package registry

type RegistryClient interface {
	//	Authorize() error
	ListRepositories() ([]string, error)
	//	ListTags(repositry string) ([]string, error)
}
