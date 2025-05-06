// examples/loader/main.go
package main

import (
    "flag"
    "fmt"
    "os"
    "strings"

    "github.com/rickcollette/splatplug"
)

var (
    pluginDir = flag.String("plugins", "./plugins", "directory of plugin executables")
    listOnly  = flag.Bool("list", false, "just list available plugins")
    call      = flag.String("call", "", "pluginName:symbol to invoke")
)

func main() {
    // 1) parse flags
    flag.Parse()

    // 2) load plugins from the directory specified by -plugins
    mgr := splatplug.NewManager()
    if err := mgr.LoadAll(*pluginDir); err != nil {
        fmt.Fprintf(os.Stderr, "load plugins: %v\n", err)
        os.Exit(1)
    }

    // 3) if -list, just print names
    if *listOnly {
        for _, name := range mgr.Plugins() {
            fmt.Println("Loaded plugin:", name)
        }
        return
    }

    // 4) if -call, invoke the given symbol
    if *call != "" {
        parts := strings.SplitN(*call, ":", 2)
        if len(parts) != 2 {
            fmt.Fprintln(os.Stderr, "call must be in pluginName:symbol format")
            os.Exit(1)
        }
        res, err := mgr.Lookup(parts[0], parts[1])
        if err != nil {
            fmt.Fprintf(os.Stderr, "call error: %v\n", err)
            os.Exit(1)
        }
        fmt.Printf("Result: %v\n", res)
        return
    }

    // 5) default -> list plugins
    for _, name := range mgr.Plugins() {
        fmt.Println("Loaded plugin:", name)
    }
}
