package splatplug

import (
    "encoding/json"
    "fmt"
    "os"
)

// SymbolFunc defines the signature for RPC‑exposed functions.
type SymbolFunc func(args []interface{}) (interface{}, error)

// registry holds all registered symbols in this plugin.
var registry = make(map[string]SymbolFunc)

// RegisterSymbol makes fn available under the given name.
func RegisterSymbol(name string, fn SymbolFunc) {
    registry[name] = fn
}

// Serve starts JSON‑RPC over stdio, but only if SPLATPLUG_MODE=serve.
func Serve(name, version string) error {
    if os.Getenv("SPLATPLUG_MODE") != "serve" {
        return nil
    }
    enc := json.NewEncoder(os.Stdout)
    dec := json.NewDecoder(os.Stdin)

    // 1) send plugin Info
    if err := enc.Encode(Info{Name: name, Version: version}); err != nil {
        return fmt.Errorf("handshake send failed: %w", err)
    }

    // 2) receive host Info
    var h HostInfo
    if err := dec.Decode(&h); err != nil {
        return fmt.Errorf("handshake recv failed: %w", err)
    }
    if h.Version != version {
        return fmt.Errorf("host v%q != plugin v%q", h.Version, version)
    }

    // 3) enter RPC loop
    for {
        var req CallRequest
        if err := dec.Decode(&req); err != nil {
            return fmt.Errorf("splatplug: failed to decode request: %w", err)
        }

        var resp CallResponse
        if fn, ok := registry[req.Symbol]; !ok {
            resp.Error = fmt.Sprintf("symbol %q not found", req.Symbol)
        } else {
            result, err := fn(req.Args)
            if err != nil {
                resp.Error = err.Error()
            } else {
                resp.Result = result
            }
        }

        if err := enc.Encode(resp); err != nil {
            return fmt.Errorf("splatplug: failed to encode response: %w", err)
        }
    }
}
