#!/bin/bash

# Script to restructure scout-mcp into proper package layout
set -e

echo "Restructuring scout-mcp into scout package..."

# Create new directory structure
mkdir -p cmd scout

# Move and rename files to scout package
echo "Moving files to scout package..."
mv config.go scout/
mv const.go scout/
mv logger.go scout/
mv mcp_server.go scout/
mv types.go scout/
mv util.go scout/
mv args.go scout/

# Create new cmd/main.go (minimal)
cat > cmd/main.go << 'EOF'
package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/mikeschinkel/scout-mcp/scout"
)

func main() {
	var server *scout.MCPServer
	var err error
	var args scout.Args

	// Initialize logger
	scout.SetLogger(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	args, err = scout.ParseArgs()
	if err != nil {
		scout.Logger().Error("Error parsing arguments", "error", err)
		os.Exit(1)
	}

	if args.IsInit {
		err = scout.CreateDefaultConfig(args.ConfigArgs)
		if err != nil {
			scout.Logger().Error("Failed to create config", "error", err)
			os.Exit(1)
		}
		return
	}

	server, err = scout.NewMCPServer(args.AdditionalPaths, args.ServerOpts)
	if err != nil {
		if len(args.AdditionalPaths) == 0 && !args.ServerOpts.OnlyMode {
			showUsageError()
			os.Exit(1)
		}
		scout.Logger().Error("Failed to create server", "error", err)
		os.Exit(1)
	}

	scout.Logger().Info("Scout-MCP File Operations Server")
	scout.Logger().Info("Whitelisted directories:")
	for dir := range server.WhitelistedDirs() {
		scout.Logger().Info("Directory", "path", dir)
	}

	err = server.StartMCP()
	if err != nil {
		scout.Logger().Error("MCP server failed", "error", err)
		os.Exit(1)
	}
}

func showUsageError() {
	var displayPath string

	displayPath = filepath.Join("~", scout.ConfigBaseDirName, scout.ConfigDirName, scout.ConfigFileName)

	fmt.Printf(`ERROR: No whitelisted directories specified.

Usage options:
  %[1]s <path>              Add path to config file paths
  %[1]s --only <path>       Use only the specified path
  %[1]s init                Create empty config file
  %[1]s init <path>         Create config with custom initial path

Config file location: %[2]s

Examples:
  %[1]s ~/MyProjects
  %[1]s --only /tmp/safe-dir
  %[1]s init ~/Code
`, scout.AppName, displayPath)
}
EOF

# Update all scout package files to use package scout
echo "Updating package declarations..."
for file in scout/*.go; do
    sed -i.bak '1s/package main/package scout/' "$file"
    rm "$file.bak"
done

# Remove old main.go
rm -f main.go

# Update go.mod to point to cmd
echo "Updating go.mod..."
if [ -f go.mod ]; then
    # Add replace directive if go.mod exists at root
    {
      echo
      echo "// For local development"
      echo "replace github.com/mikeschinkel/scout-mcp/scout => ./scout"
    } >> go.mod
fi

# Create cmd/go.mod
cat > cmd/go.mod << 'EOF'
module github.com/mikeschinkel/scout-mcp/cmd

go 1.24

require github.com/mikeschinkel/scout-mcp/scout v0.0.0

replace github.com/mikeschinkel/scout-mcp/scout => ../scout
EOF

# Create scout/go.mod
cat > scout/go.mod << 'EOF'
module github.com/mikeschinkel/scout-mcp/scout

go 1.24

require github.com/mark3labs/mcp-go v0.0.0
EOF

echo "Done! Structure is now:"
echo "cmd/main.go       - Main entry point"
echo "scout/*.go        - Application logic"
echo ""
echo "To build: cd cmd && go build -o scout-mcp"
echo ""
echo "Note: You'll need to:"
echo "1. Export necessary functions/types from scout package (capitalize names)"
echo "2. Add Logger() and SetLogger() functions to scout package"
echo "3. Add WhitelistedDirs() method to MCPServer"
echo "4. Verify all imports are correct"