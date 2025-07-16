package main

import (
	"fmt"
	"runtime"
)

// Version information - update these when releasing new versions
const (
	Version   = "1.2.0"
	BuildDate = "2025-01-16"
	GitCommit = "dev"
)

// VersionInfo holds all version-related information
type VersionInfo struct {
	Version   string
	BuildDate string
	GitCommit string
	GoVersion string
	OS        string
	Arch      string
}

// GetVersionInfo returns complete version information
func GetVersionInfo() VersionInfo {
	return VersionInfo{
		Version:   Version,
		BuildDate: BuildDate,
		GitCommit: GitCommit,
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
}

// PrintVersion prints the version information in a user-friendly format
func PrintVersion() {
	info := GetVersionInfo()
	fmt.Printf("Hani Markdown Editor v%s\n", info.Version)
	fmt.Printf("Built: %s\n", info.BuildDate)
	if info.GitCommit != "dev" {
		fmt.Printf("Commit: %s\n", info.GitCommit)
	}
	fmt.Printf("Go: %s\n", info.GoVersion)
	fmt.Printf("Platform: %s/%s\n", info.OS, info.Arch)
}

// PrintVersionShort prints just the version number
func PrintVersionShort() {
	fmt.Printf("v%s\n", Version)
}

// PrintHelp prints usage information
func PrintHelp() {
	fmt.Printf("Hani - A TUI Markdown Editor v%s\n\n", Version)
	fmt.Println("USAGE:")
	fmt.Println("  hani [filename]     Start editor with optional file")
	fmt.Println("  hani -v, --version  Show version information")
	fmt.Println("  hani -h, --help     Show this help message")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  hani                Create a new markdown file")
	fmt.Println("  hani README.md      Edit an existing file")
	fmt.Println("  hani document.md    Create or edit document.md")
	fmt.Println()
	fmt.Println("KEY BINDINGS:")
	fmt.Println("  Tab/Shift+Tab       Switch between editor and preview")
	fmt.Println("  Ctrl+S              Save file")
	fmt.Println("  Ctrl+Q              Quit application")
	fmt.Println("  i                   Enter insert mode")
	fmt.Println("  Esc                 Return to normal mode")
	fmt.Println("  h,j,k,l             Navigate (left, down, up, right)")
	fmt.Println("  w,b,e               Word movements")
	fmt.Println("  0,$                 Line beginning/end")
	fmt.Println("  gg,G                File beginning/end")
	fmt.Println("  o,O                 Insert new line")
	fmt.Println("  x,dd                Delete operations")
	fmt.Println()
	fmt.Println("For more information, visit: https://github.com/your-username/hani")
}
