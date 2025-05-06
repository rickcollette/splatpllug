// Package splatplug: core RPC types for plugin handshake and calls.
package splatplug

// Info is sent by the plugin to the host at startup.
type Info struct {
    // Name is the human‑readable plugin name.
    Name    string            `json:"name"`
    // Version is the plugin’s semver string. Must match HostVersion.
    Version string            `json:"version"`
    // Config is an optional plugin→host metadata map.
    Config  map[string]string `json:"config,omitempty"`
}

// CallRequest represents a JSON‑RPC invocation request.
type CallRequest struct {
    // Symbol is the name of the exported function or variable.
    Symbol string        `json:"symbol"`
    // Args is the list of arguments to pass; each is an interface{}.
    Args   []interface{} `json:"args"`
}

// CallResponse is the JSON‑RPC response.
type CallResponse struct {
    // Result is the returned value from the call (if any).
    Result interface{} `json:"result"`
    // Error is set if the call failed.
    Error  string      `json:"error"`
}

// HostInfo is sent by the host back to the plugin after version check.
type HostInfo struct {
    // Version echoes the host’s semver, must match Info.Version.
    Version string            `json:"version"`
    // Config is an optional host→plugin metadata map.
    Config  map[string]string `json:"config,omitempty"`
}

// HostVersion is the semver string of the host application.
const HostVersion = "v1.0.0"
