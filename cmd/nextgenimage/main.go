package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	verbose bool
	quiet   bool
)

var rootCmd = &cobra.Command{
	Use:   "nextgenimage",
	Short: "Convert traditional web images to next-gen formats",
	Long: `nextgenimage converts traditional web image formats (JPEG, PNG, GIF) 
to next-generation formats (WebP, AVIF) following best practices.`,
	Version: version,
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVar(&quiet, "quiet", false, "Suppress all output except errors")

	rootCmd.AddCommand(webpCmd)
	rootCmd.AddCommand(avifCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}