package splatplug

import (
    "encoding/json"
    "fmt"
    "os"
)

// HostConfig holds the host→plugin metadata sent during handshake.
// Plugins can read HostConfig["key"] to get structured startup data.
var HostConfig map[string]string

// SymbolFunc is the signature for any RPC‑exposed function in a plugin.
type SymbolFunc func(args []interface{}) (interface{}, error)

// registry maps symbol names to the functions registered by the plugin.
var registry = make(map[string]SymbolFunc)

// RegisterSymbol makes fn available under the given name.
// Plugins call this in init or main before Serve().
func RegisterSymbol(name string, fn SymbolFunc) {
    registry[name] = fn
}

// Serve enters RPC mode over stdio if SPLATPLUG_MODE=serve.
// It sends the plugin’s Info, receives HostInfo, populates HostConfig,
// and then loops decoding CallRequest → encoding CallResponse.
//
// Returns an error if the handshake fails or if any RPC encode/decode fails.
func Serve(name, version string) error {
    // only run RPC serve when told to
    if os.Getenv("SPLATPLUG_MODE") != "serve" {
        return nil
    }
    enc := json.NewEncoder(os.Stdout)
    dec := json.NewDecoder(os.Stdin)

    // 1) send plugin Info (no Config by default)
    if err := enc.Encode(Info{Name: name, Version: version}); err != nil {
        return fmt.Errorf("handshake send failed: %w", err)
    }

    // 2) receive HostInfo
    var h HostInfo
    if err := dec.Decode(&h); err != nil {
        return fmt.Errorf("handshake recv failed: %w", err)
    }
    if h.Version != version {
        return fmt.Errorf("host v%q != plugin v%q", h.Version, version)
    }

    // 3) save host‑provided config
    HostConfig = h.Config

    // 4) main RPC loop
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

        if err := enc.Encode(&resp); err != nil {
            return fmt.Errorf("splatplug: failed to encode response: %w", err)
        }
    }
}
