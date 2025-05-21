package main

import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/zqtools/apicli/pkg/api"
    "github.com/zqtools/apicli/pkg/config"
)

func main() {
    // Initialize configuration directory
    apiDir, err := config.InitUserConfigDir()
    if err != nil {
        fmt.Printf("Error initializing configuration directory: %v\n", err)
        os.Exit(1)
    }

    // Load or create user configuration
    configPath := filepath.Join(apiDir, "config")
    userConfig, err := config.LoadOrCreateUserConfig(configPath)
    if err != nil {
        fmt.Printf("Error loading user configuration: %v\n", err)
        os.Exit(1)
    }

    // Create CLI instance
    cli, err := api.NewCLI(userConfig, apiDir)
    if err != nil {
        fmt.Printf("Error creating CLI: %v\n", err)
        os.Exit(1)
    }

    // Execute command
    if err := cli.Execute(os.Args[1:]); err != nil {
        fmt.Printf("Error executing command: %v\n", err)
        os.Exit(1)
    }
}
