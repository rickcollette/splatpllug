package main

import (
    "fmt"
    "os"

    "github.com/rickcollette/splatplug"
)

const (
    PluginName    = "math"
    PluginVersion = "v1.0.0"
)

func main() {
    if os.Getenv("SPLATPLUG_MODE") == "serve" {
        // Add: sum all numeric args (JSON numbers decode as float64)
        splatplug.RegisterSymbol("Add", func(args []interface{}) (interface{}, error) {
            if len(args) < 2 {
                return nil, fmt.Errorf("Add requires at least 2 arguments")
            }
            sum := 0.0
            for _, a := range args {
                v, ok := a.(float64)
                if !ok {
                    return nil, fmt.Errorf("Add: argument %v is not a number", a)
                }
                sum += v
            }
            return sum, nil
        })

        // Mul: multiply all numeric args
        splatplug.RegisterSymbol("Mul", func(args []interface{}) (interface{}, error) {
            if len(args) < 2 {
                return nil, fmt.Errorf("Mul requires at least 2 arguments")
            }
            prod := 1.0
            for _, a := range args {
                v, ok := a.(float64)
                if !ok {
                    return nil, fmt.Errorf("Mul: argument %v is not a number", a)
                }
                prod *= v
            }
            return prod, nil
        })

        if err := splatplug.Serve(PluginName, PluginVersion); err != nil {
            os.Exit(1)
        }
    }
}
