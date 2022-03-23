package mounts

import (
	"github.com/pkg/errors"
)

func errInvalidMode(mode string) error {
	return errors.Errorf("invalid mode: %v", mode)
}

func errInvalidSpec(spec string) error {
	return errors.Errorf("invalid volume specification: '%s'", spec)
}
