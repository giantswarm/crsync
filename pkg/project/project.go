package project

var (
	description = "CLI tool to sync images between two registries"
	gitSHA      = "n/a"
	name        = "crsync"
	source      = "https://github.com/giantswarm/crsync"
	version     = "0.10.1-dev"
)

func Description() string {
	return description
}

func GitSHA() string {
	return gitSHA
}

func Name() string {
	return name
}

func Source() string {
	return source
}

func Version() string {
	return version
}
