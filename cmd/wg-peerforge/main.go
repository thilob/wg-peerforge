package main

import (
	"fmt"

	"github.com/thilob/wg-peerforge/internal/ui"
)

func printBanner() {
	fmt.Println("wg-peerforge")
	fmt.Println("Server and peer manager foundation")
	fmt.Println()
}

func printNextSteps() {
	fmt.Println(ui.Placeholder())
	fmt.Println()
	fmt.Println("Planned next milestones:")
	fmt.Println("1. Persist structured server and peer data")
	fmt.Println("2. Render wg-quick configuration from that model")
	fmt.Println("3. Add interactive TUI screens for server and peer management")
}

func main() {
	printBanner()
	printNextSteps()
}
