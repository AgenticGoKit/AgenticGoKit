package embedding

// Bridge package to ensure internal embedding factories are registered.
// This package exists so consumer binaries (including scaffolded apps)
// can import a stable public path and trigger initialization in the
// internal/embedding package which registers embedding factories with core.

import (
	_ "github.com/agenticgokit/agenticgokit/internal/embedding"
)

// No additional code required; blank import triggers init in internal/embedding

