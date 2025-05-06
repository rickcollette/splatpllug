// Package splatplug provides a zero‑dependency, cross‑platform plugin
// mechanism for Go applications. Plugins are standalone Go binaries
// that communicate with the host (loader) over stdio using JSON‑RPC.
// The host launches each plugin process, performs a semver‑safe
// handshake (including an arbitrary map[string]string of startup
// config), and then invokes exported symbols as RPC calls with
// timeouts and crash recovery.
//
// Basic host usage:
//
//     mgr := splatplug.NewManager()
//     if err := mgr.LoadAll("./plugins"); err != nil {
//         log.Fatal(err)
//     }
//     result, err := mgr.Lookup("myplugin", "DoThing", arg1, arg2)
//     // handle result or error
//
// Basic plugin stub:
//
//     func main() {
//         if os.Getenv("SPLATPLUG_MODE") == "serve" {
//             splatplug.RegisterSymbol("DoThing", func(args []interface{}) (interface{}, error) {
//                 // your code here
//             })
//             splatplug.Serve("myplugin", "v1.0.0")
//         }
//     }
//
// See https://pkg.go.dev/github.com/rickcollette/splatplug for full
// API docs and examples.
package splatplug
