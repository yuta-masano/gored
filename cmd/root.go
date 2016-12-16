package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// gometalinter がそうしろって。。。
const windows string = "windows"

// 終了ステータスコード。
const (
	exitOK int = iota
	exitNG
)

// RootCmd represents the base command when called without any subcommands.
var RootCmd = &cobra.Command{
	Use: "gored",
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&cfgFilePath,
		"config-file", "f",
		func() (defaultCfgFilePath string) {
			var configDir string
			if runtime.GOOS == windows {
				configDir = filepath.Join(os.Getenv("APPDATA"), "gored")
			} else {
				configDir = filepath.Join(os.Getenv("HOME"), ".config", "gored")
			}
			return filepath.Join(configDir, "config.yml")
		}(),
		"path to the config file")
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() int {
	viper.SetConfigFile(cfgFilePath)
	if err := viper.ReadInConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "failed in reading config file: %s", err)
		return exitNG
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		fmt.Fprintf(os.Stderr, "failed in setting config parameters: %s", err)
		return exitNG
	}
	for _, param := range []string{"Endpoint", "Apikey", "Trackers", "Priorities"} {
		if !viper.IsSet(param) {
			fmt.Fprintf(os.Stdout, "failed in reading config parameter: %s must be specified", param)
			return exitNG
		}
	}
	if err := RootCmd.Execute(); err != nil {
		return exitNG
	}
	return exitOK
}
