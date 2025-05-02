package main

import (
	"github.com/spf13/cobra"
)

type Config struct {
	Output             string
	ModulePath         string
	VolatileStatusFile string
	GoMod              string
	SrcFiles           []string
	StripPrefix        string
}

func cmd() *cobra.Command {
	var cfg Config

	command := &cobra.Command{
		Use:   "go_mod_tool",
		Short: "Create a Go module archive (.zip) for use with a Go proxy",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cfg)
		},
	}

	command.Flags().StringVar(&cfg.Output, "output", "", "Path to output .zip file")
	command.Flags().StringVar(&cfg.ModulePath, "module-path", "", "Module path (e.g., github.com/my_project)")
	command.Flags().StringVar(&cfg.VolatileStatusFile, "volatile-status-file", "", "Path to a file that will be stamped with the current timestamp")
	command.Flags().StringVar(&cfg.GoMod, "go-mod", "", "Path to go.mod file")
	command.Flags().StringSliceVar(&cfg.SrcFiles, "src", nil, "Path to a .go source file (can be repeated)")
	command.Flags().StringVar(&cfg.StripPrefix, "strip-prefix", "", "Prefix to strip from source file paths")

	// Mark required flags
	command.MarkFlagRequired("output")
	command.MarkFlagRequired("module-path")
	command.MarkFlagRequired("volatile-status-file")
	command.MarkFlagRequired("go-mod")
	command.MarkFlagRequired("src")

	return command
}
