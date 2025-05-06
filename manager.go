package splatplug

import (
    "context"
    "encoding/json"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "runtime"
    "sync"
    "time"
)

// Plugin represents one running plugin subprocess.
type Plugin struct {
    Name    string
    Version string
    enc     *json.Encoder
    dec     *json.Decoder
    mu      sync.Mutex
    cmd     *exec.Cmd
}

// Manager loads and communicates with multiple plugins.
type Manager struct {
    mu      sync.Mutex
    plugins map[string]*Plugin
}

// NewManager creates a fresh Manager instance.
func NewManager() *Manager {
    return &Manager{plugins: make(map[string]*Plugin)}
}

// acceptable executable extensions by OS.
var exeExts = map[string][]string{
    "windows": {".exe"},
    "darwin":  {""},
    "linux":   {""},
}

// LoadAll scans the given dirs (default "./plugins") for executables,
// starts each in RPC serve mode, and performs the semver/config handshake.
func (m *Manager) LoadAll(dirs ...string) error {
    if len(dirs) == 0 {
        dirs = []string{"./plugins"}
    }
    for _, dir := range dirs {
        entries, err := os.ReadDir(dir)
        if err != nil {
            if os.IsNotExist(err) {
                continue
            }
            return err
        }
        for _, fi := range entries {
            if fi.IsDir() {
                continue
            }
            name := fi.Name()
            extOK := false
            for _, ext := range exeExts[runtime.GOOS] {
                if ext != "" && filepath.Ext(name) == ext {
                    extOK = true
                }
                if ext == "" {
                    info, err := fi.Info()
                    if err != nil {
                        continue
                    }
                    if info.Mode()&0o111 != 0 {
                        extOK = true
                    }
                }
            }
            if !extOK {
                continue
            }
            path := filepath.Join(dir, name)
            if err := m.spawnAndHandshake(path); err != nil {
                return fmt.Errorf("plugin %q: %w", path, err)
            }
        }
    }
    return nil
}

// spawnAndHandshake starts the plugin process, sends/receives Info and HostInfo
// (including Config), then registers the Plugin in m.plugins.
func (m *Manager) spawnAndHandshake(path string) error {
    cmd := exec.Command(path)
    cmd.Env = append(os.Environ(), "SPLATPLUG_MODE=serve")

    stdout, err := cmd.StdoutPipe()
    if err != nil {
        return err
    }
    stdin, err := cmd.StdinPipe()
    if err != nil {
        return err
    }
    if err := cmd.Start(); err != nil {
        return err
    }

    dec := json.NewDecoder(stdout)
    enc := json.NewEncoder(stdin)

    // 1) read plugin Info
    var pInfo Info
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()
    errCh := make(chan error, 1)
    go func() { errCh <- dec.Decode(&pInfo) }()
    select {
    case err := <-errCh:
        if err != nil {
            cmd.Process.Kill()
            return fmt.Errorf("read plugin Info: %w", err)
        }
    case <-ctx.Done():
        cmd.Process.Kill()
        return fmt.Errorf("timeout waiting for plugin Info")
    }

    // 2) version check
    if pInfo.Version != HostVersion {
        cmd.Process.Kill()
        return fmt.Errorf("plugin %s v%q != host v%q", pInfo.Name, pInfo.Version, HostVersion)
    }

    // 3) prepare host Config (customize as needed)
    hostConfig := map[string]string{
        "db_host":   os.Getenv("DB_HOST"),
        "log_level": os.Getenv("LOG_LEVEL"),
    }

    // 4) send HostInfo back
    if err := enc.Encode(HostInfo{Version: HostVersion, Config: hostConfig}); err != nil {
        cmd.Process.Kill()
        return fmt.Errorf("send host Info: %w", err)
    }

    // 5) register in Manager
    pl := &Plugin{
        Name:    pInfo.Name,
        Version: pInfo.Version,
        enc:     enc,
        dec:     dec,
        cmd:     cmd,
    }
    m.mu.Lock()
    m.plugins[pl.Name] = pl
    m.mu.Unlock()
    return nil
}

// Lookup invokes the given symbol on pluginName with args, returning the result.
// It applies timeouts, kills hung processes, and cleans up crashed plugins.
func (m *Manager) Lookup(pluginName, symbol string, args ...interface{}) (interface{}, error) {
    m.mu.Lock()
    pl, ok := m.plugins[pluginName]
    m.mu.Unlock()
    if !ok {
        return nil, fmt.Errorf("plugin %q not loaded", pluginName)
    }

    pl.mu.Lock()
    defer pl.mu.Unlock()

    // Encode request with timeout
    req := CallRequest{Symbol: symbol, Args: args}
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    errCh := make(chan error, 1)
    go func() { errCh <- pl.enc.Encode(req) }()
    select {
    case err := <-errCh:
        if err != nil {
            return nil, err
        }
    case <-ctx.Done():
        pl.cmd.Process.Kill()
        m.cleanup(pluginName)
        return nil, ctx.Err()
    }

    // Decode response with timeout
    respCh := make(chan CallResponse, 1)
    errCh2 := make(chan error, 1)
    go func() {
        var resp CallResponse
        err := pl.dec.Decode(&resp)
        if err == nil {
            respCh <- resp
        }
        errCh2 <- err
    }()
    select {
    case resp := <-respCh:
        if resp.Error != "" {
            return nil, fmt.Errorf("%s", resp.Error)
        }
        return resp.Result, nil
    case err := <-errCh2:
        m.cleanup(pluginName)
        return nil, fmt.Errorf("plugin %q decode error: %w", pluginName, err)
    case <-ctx.Done():
        pl.cmd.Process.Kill()
        m.cleanup(pluginName)
        return nil, ctx.Err()
    }
}

// Plugins returns the list of currently‑loaded plugin names.
func (m *Manager) Plugins() []string {
    m.mu.Lock()
    defer m.mu.Unlock()
    names := make([]string, 0, len(m.plugins))
    for n := range m.plugins {
        names = append(names, n)
    }
    return names
}

// Shutdown cleanly terminates all plugin processes.
func (m *Manager) Shutdown() {
    m.mu.Lock()
    defer m.mu.Unlock()
    for _, pl := range m.plugins {
        pl.cmd.Process.Signal(os.Interrupt)
        done := make(chan struct{})
        go func(cmd *exec.Cmd) {
            cmd.Wait()
            close(done)
        }(pl.cmd)
        select {
        case <-done:
        case <-time.After(1 * time.Second):
            pl.cmd.Process.Kill()
        }
    }
    m.plugins = make(map[string]*Plugin)
}

// cleanup removes a crashed or timed‑out plugin from the registry.
func (m *Manager) cleanup(name string) {
    m.mu.Lock()
    delete(m.plugins, name)
    m.mu.Unlock()
}
