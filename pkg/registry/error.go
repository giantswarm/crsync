package registry

import (
	"github.com/giantswarm/microerror"
)

var executionFailedError = &microerror.Error{
	Kind: "executionFailedError",
}

// IsExecutionFailed asserts executionFailedError.
func IsExecutionFailed(err error) bool {
	return microerror.Cause(err) == executionFailedError
}

var unknownContainerRegistryKindError = &microerror.Error{
	Kind: "unknownContainerRegistryKind",
}

// IsUnknownAuthMethodError asserts unknownContainerRegistryKind
func IsUnknownContainerRegistryKind(err error) bool {
	return microerror.Cause(err) == unknownContainerRegistryKindError
}
