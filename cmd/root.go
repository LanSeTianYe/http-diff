package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "http-diff",
	Short: "HTTP 响应对比工具",
	Long:  "HTTP 响应对比工具",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Printf("Error executing command: %v\n", err)
		os.Exit(1)
	}
}
