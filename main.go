package main

import (
	"github.com/riete/ws-tunnel/cmd"
	"github.com/spf13/cobra"
)

func main() {
	cobra.EnableTraverseRunHooks = true
	cmd.Execute()
}
