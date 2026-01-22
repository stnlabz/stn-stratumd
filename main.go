// File: main.go
// Version 1.7

package main

import (
    "database/sql"
    "io/ioutil"
    "log"

    _ "github.com/mattn/go-sqlite3"
    "gopkg.in/yaml.v3"

    "stratumd/core"
)

// AppConfig holds config.yaml fields
type AppConfig struct {
    Port         int    `yaml:"port"`
    ChainRPC     string `yaml:"chain_rpc"`
    WorkInterval int    `yaml:"work_interval"`
}

func main() {
    // 1) Load configuration
    raw, err := ioutil.ReadFile("config/config.yaml")
    if err != nil {
        log.Fatalf("could not read config: %v", err)
    }
    var appCfg AppConfig
    if err := yaml.Unmarshal(raw, &appCfg); err != nil {
        log.Fatalf("invalid config: %v", err)
    }

    // 2) Initialize database in ./data/pool.db
    db, err := sql.Open("sqlite3", "file:data/pool.db?cache=shared&mode=rwc")
    if err != nil {
        log.Fatalf("failed to open DB: %v", err)
    }
    defer db.Close()
    core.SetDB(db)

    // 3) Create and inject RPC client
    rpcClient := core.NewRPCClient(appCfg.ChainRPC)
    core.SetRPCClient(rpcClient)

    // 4) Inject server config
    serverCfg := core.Config{
        Port:         appCfg.Port,
        WorkInterval: appCfg.WorkInterval,
        ChainRPC:     appCfg.ChainRPC,
    }
    core.SetConfig(serverCfg)

    // 5) Start Stratum server
    core.Run()
}
