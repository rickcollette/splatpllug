package splatplug

// Info is exchanged at startup handshake
type Info struct {
    Name    string `json:"name"`    // plugin name
    Version string `json:"version"` // semver, must match host
}

// CallRequest represents an invocation request
type CallRequest struct {
    Symbol string        `json:"symbol"`
    Args   []interface{} `json:"args"`
}

// CallResponse returns invocation result or error
type CallResponse struct {
    Result interface{} `json:"result"`
    Error  string      `json:"error"`
}

// HostInfo is sent back by host to plugin
type HostInfo struct {
    Version string `json:"version"`
}

// your host’s version constant
const HostVersion = "v1.0.0"