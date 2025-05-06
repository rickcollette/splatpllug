package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/rickcollette/splatplug"
)

const (
    PluginName    = "rolodex"
    PluginVersion = "v1.0.0"
)

var rolodex map[string]string

func init() {
    // 1) read the JSON blob from the env var set by loader
    rolodex = make(map[string]string)
    if data := os.Getenv("ROLLODEX_DATA"); data != "" {
        _ = json.Unmarshal([]byte(data), &rolodex)
    }

    // 2) register the Rolodex symbol
    splatplug.RegisterSymbol("Rolodex", func(args []interface{}) (interface{}, error) {
        if len(args) != 1 {
            return nil, fmt.Errorf("rolodex needs exactly 1 argument")
        }
        key, ok := args[0].(string)
        if !ok {
            return nil, fmt.Errorf("rolodex: argument %v is not a string", args[0])
        }
        entry, found := rolodex[strings.ToLower(key)]
        if !found {
            return nil, fmt.Errorf("rolodex: %q not found", key)
        }
        return entry, nil
    })
}

func main() {
    // only run RPC‐serve when invoked by splatplug
    if os.Getenv("SPLATPLUG_MODE") == "serve" {
        if err := splatplug.Serve(PluginName, PluginVersion); err != nil {
            os.Exit(1)
        }
    }
}
