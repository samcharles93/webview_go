// Package webview provides Go bindings for the upstream webview library.
//
// The package stays close to the intentionally small upstream C API while
// adding Go-focused ergonomics such as typed errors, constructor options, and
// JSON-backed JavaScript interop helpers.
//
// Start with NewWithOptions for new code. The legacy New and NewWindow helpers
// remain available for compatibility.
package webview
