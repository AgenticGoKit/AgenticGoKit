// Package factory previously contained a legacy runner factory. It now defers to core APIs.
package factory

// This file intentionally left as a thin shim to avoid duplicate type definitions.
// All runner-related types and constructors now live in the core package and the default runner plugin.
// Import side effects via blank imports in your application to register plugins.

// No exported symbols are defined here anymore.
