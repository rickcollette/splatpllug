# splatplug

`splatplug` is a zero‑dependency, cross‑platform “plugin” system for Go, using JSON‑RPC over stdio.  It lets you write plugins as standalone Go binaries (no cgo, no shared libraries), and communicate with them via a simple RPC API—and even pass structured data at startup via the handshake.

---

## Table of Contents

- [splatplug](#splatplug)
  - [Table of Contents](#table-of-contents)
  - [Features](#features)
  - [Installation](#installation)
  - [Quickstart](#quickstart)
  - [Writing a Plugin](#writing-a-plugin)
    - [Plugin Stub](#plugin-stub)
    - [Registering Symbols](#registering-symbols)
    - [Accessing HostConfig](#accessing-hostconfig)
  - [Loader (Host) CLI](#loader-host-cli)
    - [Building with Makefile](#building-with-makefile)
    - [Listing Plugins](#listing-plugins)
    - [Calling Symbols](#calling-symbols)
  - [Handshake \& Passing Data](#handshake--passing-data)
    - [Extending HostInfo.Config](#extending-hostinfoconfig)
    - [Reading in the Plugin](#reading-in-the-plugin)
  - [Example Plugins](#example-plugins)
    - [Echo](#echo)
    - [Math (Add, Mul)](#math-addmul)
    - [Reverse](#reverse)
    - [Rolodex (Structured Data)](#rolodex-structured-data)
  - [API Reference](#api-reference)
    - [Core Types](#core-types)
    - [Manager](#manager)
    - [Plugin Core](#plugin-core)
  - [License](#license)

---

## Features

- **Cross‑platform**: works on Linux, macOS, Windows.  
- **Race‑detector friendly**: runs each plugin in its own process.  
- **No shared libraries**: distribute plugins as plain Go binaries.  
- **Version‑safe handshake**: semver check on startup.  
- **Structured config**: send arbitrary key/value maps at startup.  
- **Simple RPC**: call exported symbols with JSON‑RPC over stdio.  
- **Timeouts & crash recovery**: host kills unresponsive plugins.  

---

## Installation

```bash
go get github.com/rickcollette/splatplug@v1.0.0
```

Ensure your `$GOPATH/bin` or module bin dir is in your `PATH` if you build any CLI wrappers.

---

## Quickstart

1. Create a plugin under `examples/yourplugin/` with a `main.go` that registers symbols.  
2. Build it into `plugins/yourplugin` (no extension on Unix, `.exe` on Windows).  
3. Build the loader CLI into `bin/loader`.  
4. Run:

   ```bash
   ./bin/loader -list
   ./bin/loader -call yourplugin:YourSymbol arg1 arg2
   ```

---

## Writing a Plugin

### Plugin Stub

Every plugin binary needs a small `main.go`:

```go
package main

import (
    "os"
    "github.com/rickcollette/splatplug"
)

const (
    PluginName    = "yourplugin"
    PluginVersion = "v1.0.0"
)

func main() {
    // only enter RPC mode if SPLATPLUG_MODE=serve
    if os.Getenv("SPLATPLUG_MODE") == "serve" {
        registerSymbols()
        if err := splatplug.Serve(PluginName, PluginVersion); err != nil {
            os.Exit(1)
        }
    }
    // else: run standalone CLI or exit
}

func registerSymbols() {
    // call splatplug.RegisterSymbol here...
}
```

### Registering Symbols

Inside `registerSymbols()`, map names to Go functions:

```go
splatplug.RegisterSymbol("Echo", func(args []interface{}) (interface{}, error) {
    if len(args)!=1 {
        return nil, fmt.Errorf("needs one arg")
    }
    return args[0], nil
})
```

The function signature is always:

```go
func(args []interface{}) (interface{}, error)
```

Arguments come in as Go types (`float64` for JSON numbers, `string` for JSON strings, etc.).

### Accessing HostConfig

If your host (loader) passes structured data in the handshake, the core populates:

```go
// in plugin code, after Serve() handshake:
cfg := splatplug.HostConfig  // map[string]string
val := cfg["someKey"]
```

No other env‑vars or globals are needed.

---

## Loader (Host) CLI

Your loader is just another Go binary (under `examples/loader/main.go`) that:

1. Sets up any data to pass (e.g. rolodex JSON).  
2. Calls `mgr := splatplug.NewManager()`  
3. `mgr.LoadAll(dir)` to start each plugin in `serve` mode.  
4. `mgr.Lookup(pluginName, symbolName, args...)` to invoke.

### Building with Makefile

```makefile
BIN_DIR    := bin
PLUGIN_DIR := plugins
PLUGIN_EXAMPLES := echo math reverse rolodex

.PHONY: examples
examples: $(BIN_DIR)/loader $(PLUGIN_EXAMPLES:%=$(PLUGIN_DIR)/%)

$(BIN_DIR)/loader: examples/loader/main.go
    go build -o $@ $<

$(PLUGIN_DIR)/%: examples/%/main.go
    go build -o $@ $<

clean:
    rm -rf $(BIN_DIR) $(PLUGIN_DIR)
```

Run:

```bash
make clean examples
```

### Listing Plugins

```bash
./bin/loader -list
# → Loaded plugin: echo
# → Loaded plugin: math
# → Loaded plugin: reverse
# → Loaded plugin: rolodex
```

### Calling Symbols

```bash
./bin/loader -call echo:Echo "hello"
# → Result: hello
```

Arguments after `-call` are parsed for numbers vs strings.

---

## Handshake & Passing Data

### Extending HostInfo.Config

The core `manager.go` encodes a `HostInfo{Version, Config}`.  You can pack any JSON‑serializable map into `HostInfo.Config`.  For example, to ship a “rolodex”:

```go
rolodex := map[string]string{
    "rick": "...",
    "bob":  "...",
}
raw, _ := json.Marshal(rolodex)
hostConfig := map[string]string{
    "rolodex": string(raw),
}
enc.Encode(HostInfo{Version: HostVersion, Config: hostConfig})
```

### Reading in the Plugin

In your plugin’s `main()` before `Serve` completes:

```go
// HostConfig is global in plugin core:
data := splatplug.HostConfig["rolodex"]
var rolodex map[string]string
_ = json.Unmarshal([]byte(data), &rolodex)
```

Now your plugin can answer RPCs by looking up `rolodex[key]`.

---

## Example Plugins

### Echo

`examples/echo/main.go`  

```go
splatplug.RegisterSymbol("Echo", func(args []interface{}) (interface{}, error) {
    return args[0], nil
})
```

### Math (Add, Mul)

`examples/math/main.go`  

```go
splatplug.RegisterSymbol("Add", func(args []interface{}) (interface{}, error) { /* sum float64s */ })
splatplug.RegisterSymbol("Mul", func(args []interface{}) (interface{}, error) { /* prod float64s */ })
```

### Reverse

`examples/reverse/main.go`  

```go
splatplug.RegisterSymbol("Reverse", func(args []interface{}) (interface{}, error) {
    s := args[0].(string)
    return reverseString(s), nil
})
```

### Rolodex (Structured Data)

`examples/rolodex/main.go`  

```go
// Init reads `rolodex` JSON from splatplug.HostConfig:
data := splatplug.HostConfig["rolodex"]
_ = json.Unmarshal([]byte(data), &rolodexMap)

splatplug.RegisterSymbol("Rolodex", func(args []interface{}) (interface{}, error) {
    name := strings.ToLower(args[0].(string))
    return rolodexMap[name], nil
})
```

---

## API Reference

### Core Types

```go
type Info struct {
    Name    string
    Version string
    Config  map[string]string
}

type HostInfo struct {
    Version string
    Config  map[string]string
}

type CallRequest struct {
    Symbol string
    Args   []interface{}
}

type CallResponse struct {
    Result interface{}
    Error  string
}
```

### Manager

```go
func NewManager() *Manager

func (m *Manager) LoadAll(dirs ...string) error
func (m *Manager) Lookup(pluginName, symbol string, args ...interface{}) (interface{}, error)
func (m *Manager) Plugins() []string
func (m *Manager) Shutdown()
```

### Plugin Core

```go
var HostConfig map[string]string

func RegisterSymbol(name string, fn SymbolFunc)
func Serve(name, version string) error
```

---

## License

This project is released under the [BSD‑style license](/LICENSE).
