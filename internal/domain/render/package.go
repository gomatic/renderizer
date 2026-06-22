// Package render orchestrates the renderizer command.
//
// It defines the command's Config (the flags, injected seams, and parsed
// arguments the CLI binds) and Run (the orchestration entry point the CLI
// invokes). Run builds the template data context from command-line variables,
// settings files, and the environment; resolves which templates to render
// (explicit files, stdin, or a discovered default); and renders each by
// delegating to the reusable internal/template, internal/settings,
// internal/variables, and internal/environment packages. It contains no CLI,
// flag, or output-formatting logic. This is the domain tier: the seam between
// the app tier (internal/app) and the implementation packages.
package render
