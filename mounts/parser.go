package mounts

import (
	"errors"
	"github.com/docker/docker/api/types"
)

// ErrVolumeTargetIsRoot is returned when the target destination is root.
// It's used by both LCOW and Linux parsers.
var ErrVolumeTargetIsRoot = errors.New("invalid specification: destination can't be '/'")

// Parser represents a platform specific parser for mount expressions
type Parser interface {
	ParseMountRaw(raw, volumeDriver string) (*types.MountPoint, error)
}
