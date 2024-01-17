package versionchecker

import (
	"fmt"

	"github.com/hashicorp/go-version"
)

type checker struct {
	minimumVersionRequired *version.Version
}

func New(minimumRequiredVersion *version.Version) (*checker, error) {
	return &checker{minimumVersionRequired: minimumRequiredVersion}, nil
}

func (c *checker) IsInstanceVersionLessThanMinimumRequired(instanceVersion string) (bool, error) {
	iv, err := version.NewSemver(instanceVersion)
	if err != nil {
		return false, fmt.Errorf("incorrect instance version: %w", err)
	}

	return iv.LessThan(c.minimumVersionRequired), nil
}
