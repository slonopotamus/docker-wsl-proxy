package mounts

import (
	"errors"
	"github.com/docker/docker/api/types"
	"path"

	"github.com/docker/docker/api/types/mount"
)

var lcowSpecificValidators mountValidator = func(m *mount.Mount) error {
	if path.Clean(m.Target) == "/" {
		return ErrVolumeTargetIsRoot
	}
	if m.Type == mount.TypeNamedPipe {
		return errors.New("Linux containers on Windows do not support named pipe mounts")
	}
	return nil
}

type LCOWParser struct {
	windowsParser
}

func (p *LCOWParser) ParseMountRaw(raw, volumeDriver string) (*types.MountPoint, error) {
	return p.parseMountRaw(raw, volumeDriver, rxLCOWDestination, false, lcowSpecificValidators)
}
