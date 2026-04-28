package commands

import "fmt"

func Unknown(arg string) {
	fmt.Printf("Unknown Command: %s\n\n", arg)
	fmt.Println("Available Commands:")
	fmt.Println("  cogito install          Install git hooks")
	fmt.Println("  cogito uninstall        Remove git hooks")
	fmt.Println("  cogito build-map        Scan project and build context map")
	fmt.Println("  cogito serve-mcp        Start the Model Context Protocol server")
	fmt.Println("  cogito --help           Show help documentation")
	fmt.Println("  cogito --version, -v    Show version info")
	fmt.Println("\nRun 'cogito' without arguments to open the UI.")
}
