/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"wget/internal/config"
	"wget/internal/downloader"

	"github.com/spf13/cobra"
)

var opts config.Options

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "wget [URL]",
	Short: "Download files from the web",
	Long: `wget is a command-line tool for downloading files from the web.
It supports features like background downloads, rate limiting, 
website mirroring, and batch downloads from a file.

Examples:
  wget https://example.com/file.zip
  wget -O myfile.zip https://example.com/file.zip
  wget --mirror https://example.com
  wget -i urls.txt`,
	Args: func(cmd *cobra.Command, args []string) error {
		if opts.InputFile != "" {
			return nil
		}
		if len(args) != 1 {
			return fmt.Errorf("expected URL")
		}
		opts.URL = args[0]
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if opts.Background && os.Getenv("WGET_BACKGROUND_CHILD") == "" {
			return runInBackground()
		}

		opts.ShouldRender = !opts.Background && !opts.Mirror

		if opts.InputFile != "" {
			return downloader.DownloadFromFile(&opts)
		}

		if opts.Mirror {
			return downloader.MirrorWebsite(&opts)
		}

		return downloader.DownloadOne(&opts)
	},
}

func runInBackground() error {
	logPath := nextBackgroundLogPath()
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	defer logFile.Close()

	exe, err := os.Executable()
	if err != nil {
		return err
	}

	cmd := exec.Command(exe, os.Args[1:]...)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Env = append(os.Environ(), "WGET_BACKGROUND_CHILD=1")

	if err := cmd.Start(); err != nil {
		return err
	}

	fmt.Printf("Continuing in background, pid %d.\n", cmd.Process.Pid)
	fmt.Printf("Output will be written to '%s'.\n", logPath)
	return nil
}

func nextBackgroundLogPath() string {
	base := "wget-log"
	if _, err := os.Stat(base); os.IsNotExist(err) {
		return base
	}

	for i := 1; ; i++ {
		candidate := fmt.Sprintf("%s.%d", base, i)
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.wget.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.

	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	// my args
	rootCmd.Flags().BoolVarP(&opts.Background, "background", "B", false, "download in background")
	rootCmd.Flags().StringVarP(&opts.Output, "output", "O", "", "save file with different name")
	rootCmd.Flags().StringVarP(&opts.Directory, "directory", "P", ".", "save file in directory")
	rootCmd.Flags().StringVar(&opts.RateLimit, "rate-limit", "", "limit speed, e.g. 300k or 2M")
	rootCmd.Flags().StringVarP(&opts.InputFile, "input-file", "i", "", "file containing URLs")
	rootCmd.Flags().BoolVar(&opts.Mirror, "mirror", false, "mirror website")
	rootCmd.Flags().StringSliceVarP(&opts.Reject, "reject", "R", nil, "reject file suffixes, e.g. jpg,gif")
	rootCmd.Flags().StringSliceVarP(&opts.Exclude, "exclude", "X", nil, "exclude paths, e.g. /img,/css")
	rootCmd.Flags().BoolVar(&opts.ConvertLinks, "convert-links", false, "convert links for offline viewing")
	rootCmd.Flags().IntVar(&opts.Timeout, "timeout", 30, "set timeout in seconds")
}
