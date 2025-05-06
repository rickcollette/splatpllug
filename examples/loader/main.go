package main

import (
    "encoding/json"
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
    // 0) prepare your rolodex data in Go
    rolodexData := map[string]string{
        "rick": "Name: Rick\nAge: 989\nAddress: DEADBEEF\nNumber: 9",
        "bob":  "Name: Bob\nAge: 19\nAddress: 01011101\nNumber: 901-555-1221",
    }
    if b, err := json.Marshal(rolodexData); err == nil {
        // inject into plugin’s env before LoadAll/Serve
        os.Setenv("ROLLODEX_DATA", string(b))
    }

    // 1) parse flags
    flag.Parse()

    // 2) load plugins
    mgr := splatplug.NewManager()
    if err := mgr.LoadAll(*pluginDir); err != nil {
        fmt.Fprintf(os.Stderr, "load plugins: %v\n", err)
        os.Exit(1)
    }

    // 3) list-only
    if *listOnly {
        for _, name := range mgr.Plugins() {
            fmt.Println("Loaded plugin:", name)
        }
        return
    }

    // 4) call mode
    if *call != "" {
        parts := strings.SplitN(*call, ":", 2)
        if len(parts) != 2 {
            fmt.Fprintln(os.Stderr, "call must be in pluginName:symbol format")
            os.Exit(1)
        }

        raw := flag.Args()
        if len(raw) != 1 {
            fmt.Fprintln(os.Stderr, "provide exactly one name argument")
            os.Exit(1)
        }
        name := raw[0]

        // Invoke the plugin
        res, err := mgr.Lookup(parts[0], parts[1], name)
        if err != nil {
            fmt.Fprintf(os.Stderr, "call error: %v\n", err)
            os.Exit(1)
        }
        fmt.Println(res)
        return
    }

    // 5) default: list
    for _, name := range mgr.Plugins() {
        fmt.Println("Loaded plugin:", name)
    }
}
