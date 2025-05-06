// cmd/host/main.go
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/rickcollette/splatplug"
)

var (
    pluginDir = flag.String("plugins", "./bin", "directory of plugin executables")
    listOnly  = flag.Bool("list", false, "just list available plugins")
    call      = flag.String("call", "", "pluginName:symbol to invoke")
)

func main() {
    flag.Parse()
    mgr := splatplug.NewManager()
    if err := mgr.LoadAll(*pluginDir); err != nil {
        fmt.Fprintf(os.Stderr, "failed to load plugins: %v\n", err)
        os.Exit(1)
    }

    if *listOnly {
        fmt.Println("Loaded plugins:")
        for name := range mgr.Plugins() {
            fmt.Println(" –", name)
        }
        return
    }

    if *call != "" {
        parts := strings.SplitN(*call, ":", 2)
        if len(parts) != 2 {
            fmt.Fprintln(os.Stderr, "call must be in pluginName:symbol format")
            os.Exit(1)
        }
        res, err := mgr.Lookup(parts[0], parts[1])
        if err != nil {
            fmt.Fprintf(os.Stderr, "error calling %s: %v\n", *call, err)
            os.Exit(1)
        }
        fmt.Printf("Result: %v\n", res)
        return
    }

    fmt.Println("Nothing to do. Use –list or –call.")
}
