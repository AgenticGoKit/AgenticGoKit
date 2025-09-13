package templates

const ConfigHelpersTemplate = `package config

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "time"

    "github.com/kunalkushwaha/agenticgokit/core"
)

// GetPath returns env override path or default CWD agentflow.toml
func GetPath() string {
    if env := os.Getenv("AGENTFLOW_CONFIG_PATH"); env != "" {
        return env
    }
    if wd, err := os.Getwd(); err == nil {
        return filepath.Join(wd, "agentflow.toml")
    }
    return "agentflow.toml"
}

// atomicWriteFile writes data atomically to path
func atomicWriteFile(path string, data []byte) error {
    dir := filepath.Dir(path)
    base := filepath.Base(path)
    tmp, err := os.CreateTemp(dir, base+".tmp-*")
    if err != nil { return err }
    tmpPath := tmp.Name()
    defer func() { _ = os.Remove(tmpPath) }()
    if _, err := tmp.Write(data); err != nil { _ = tmp.Close(); return err }
    if err := tmp.Sync(); err != nil { _ = tmp.Close(); return err }
    if err := tmp.Close(); err != nil { return err }
    return os.Rename(tmpPath, path)
}

// HandleGetRaw returns the contents of agentflow.toml
func HandleGetRaw(w http.ResponseWriter, r *http.Request) {
    path := GetPath()
    data, err := os.ReadFile(path)
    if err != nil {
        if os.IsNotExist(err) {
            http.Error(w, "agentflow.toml not found", http.StatusNotFound)
            return
        }
        http.Error(w, "Failed to read config", http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": map[string]any{"path": path, "size": len(data), "content": string(data)}})
}

// HandlePutRaw updates agentflow.toml after basic validation
func HandlePutRaw(w http.ResponseWriter, r *http.Request) {
    r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
    defer r.Body.Close()
    var body struct{ Toml string ` + "`json:\"toml\"`" + ` }
    dec := json.NewDecoder(r.Body)
    dec.DisallowUnknownFields()
    if err := dec.Decode(&body); err != nil {
        if err == io.EOF { http.Error(w, "Empty body", http.StatusBadRequest); return }
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    if strings.TrimSpace(body.Toml) == "" {
        http.Error(w, "Missing 'toml'", http.StatusBadRequest)
        return
    }
    tmpFile := filepath.Join(os.TempDir(), "agentflow-validate-"+fmt.Sprint(time.Now().UnixNano())+".toml")
    _ = os.WriteFile(tmpFile, []byte(body.Toml), 0644)
    parsed, err := core.LoadConfig(tmpFile)
    _ = os.Remove(tmpFile)
    if err != nil { http.Error(w, "TOML parse error: "+err.Error(), http.StatusBadRequest); return }
    if err := parsed.ValidateOrchestrationConfig(); err != nil { http.Error(w, "Config validation failed: "+err.Error(), http.StatusBadRequest); return }
    path := GetPath()
    if err := atomicWriteFile(path, []byte(body.Toml)); err != nil {
        http.Error(w, "Failed to write config", http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": map[string]any{"path": path, "updated": true}})
}

// BuildInfo builds the JSON response for GET /api/config
func BuildInfo(cfg *core.Config, port int) map[string]any {
    defaultOrch := false
    mode := cfg.Orchestration.Mode
    switch mode {
    case "sequential", "collaborative", "parallel", "loop", "mixed", "route":
        defaultOrch = true
    }
    return map[string]any{
        "server": map[string]any{"name": cfg.AgentFlow.Name, "port": port, "url": fmt.Sprintf("http://localhost:%d", port)},
        "features": map[string]any{"websocket": true, "streaming": true},
        "orchestration": map[string]any{
            "mode": mode,
            "default_enabled": defaultOrch,
            "sequential_agents": cfg.Orchestration.SequentialAgents,
            "collaborative_agents": cfg.Orchestration.CollaborativeAgents,
            "loop_agent": cfg.Orchestration.LoopAgent,
        },
    }
}
`
