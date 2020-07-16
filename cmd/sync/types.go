package sync

import (
	"github.com/giantswarm/crsync/pkg/registry"
)

type getTagsJob struct {
	Src registry.Interface
	Dst registry.Interface

	ID   string
	Repo string
}

type retagJob struct {
	Src registry.Interface
	Dst registry.Interface

	ID   string
	Repo string
	Tag  string
}
