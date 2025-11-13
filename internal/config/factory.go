// Package config provides internal configuration factory functionality.
package config

import "github.com/agenticgokit/agenticgokit/core"

// NewValidator creates a new configuration validator that implements core.ConfigValidator
func NewValidator() core.ConfigValidator {
	return NewDefaultConfigValidator()
}
