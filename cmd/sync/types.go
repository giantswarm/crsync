package sync

import "github.com/giantswarm/crsync/pkg/registry"

type retagJob struct {
	Src registry.Interface
	Dst registry.Interface

	ID   string
	Repo string
	Tag  string
}
