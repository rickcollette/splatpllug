package main

import (
	"fmt"
	"os"

	"github.com/rickcollette/splatplug"
)

const (
    PluginName    = "echo"
    PluginVersion = "v1.0.0"
)

func main() {
    // only run Serve when in RPC mode
    if os.Getenv("SPLATPLUG_MODE") == "serve" {
        splatplug.RegisterSymbol("Echo", func(args []interface{}) (interface{}, error) {
            if len(args) != 1 {
                return nil, fmt.Errorf("echo needs exactly 1 argument")
            }
            return args[0], nil
        })
        if err := splatplug.Serve(PluginName, PluginVersion); err != nil {
            os.Exit(1)
        }
    }
}