package main

import (
    "fmt"
    "os"

    "github.com/rickcollette/splatplug"
)

const (
    PluginName    = "reverse"
    PluginVersion = "v1.0.0"
)

func main() {
    if os.Getenv("SPLATPLUG_MODE") == "serve" {
        splatplug.RegisterSymbol("Reverse", func(args []interface{}) (interface{}, error) {
            if len(args) != 1 {
                return nil, fmt.Errorf("reverse requires exactly 1 argument")
            }
            s, ok := args[0].(string)
            if !ok {
                return nil, fmt.Errorf("reverse: argument %v is not a string", args[0])
            }
            runes := []rune(s)
            for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
                runes[i], runes[j] = runes[j], runes[i]
            }
            return string(runes), nil
        })

        if err := splatplug.Serve(PluginName, PluginVersion); err != nil {
            os.Exit(1)
        }
    }
}
