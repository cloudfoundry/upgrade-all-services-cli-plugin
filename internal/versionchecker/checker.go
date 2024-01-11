package versionchecker

import (
	"fmt"

	"github.com/hashicorp/go-version"
)

type checker struct {
	minimumVersionRequired *version.Version
}

func New(minimumRequiredVersion string) (*checker, error) {
	ver, err := version.NewSemver(minimumRequiredVersion)
	if err != nil {
		return nil, fmt.Errorf("incorrect minimum required version: %w", err)
	}

	return &checker{minimumVersionRequired: ver}, nil
}

func (c *checker) IsInstanceVersionLessThanMinimumRequired(instanceVersion string) (bool, error) {
	iv, err := version.NewSemver(instanceVersion)
	if err != nil {
		return false, fmt.Errorf("incorrect instance version: %w", err)
	}

	return iv.LessThan(c.minimumVersionRequired), nil
}