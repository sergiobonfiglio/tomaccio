package app

import (
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"

	"github.com/sergiobonfiglio/tomaccio/internal/config"
	"github.com/sergiobonfiglio/tomaccio/internal/download"
	"github.com/sergiobonfiglio/tomaccio/internal/search"
	"github.com/sergiobonfiglio/tomaccio/internal/watched"
	tomagnetlib "github.com/sergiobonfiglio/tomagnet/pkg/tomagnet"
)

var (
	versionOverride = ""
	readBuildInfo  = debug.ReadBuildInfo
)

type commandEnv struct {
	configPath      string
	loadConfig      func(string) (*config.Config, error)
	downloader      func(*config.Config) (download.Downloader, error)
	searchProviders func(*config.Config) ([]search.Provider, []search.ProviderError)
	watchedProvider func(*config.Config) (watched.Provider, error)
	definitionsSync func() (tomagnetlib.DefinitionsMetadata, error)
}

func NewRootCommand() *cobra.Command {
	env := &commandEnv{}
	root := &cobra.Command{
		Use:     "tomaccio",
		Short:   "Taste Oriented Media Assistant Accio",
		Version: buildVersion(),
	}
	root.SetVersionTemplate("{{printf \"%s\\n\" .Version}}")
	root.PersistentFlags().StringVar(&env.configPath, "config", "./config.yaml", "path to YAML config file")
	root.AddCommand(env.downloadCommand(), env.searchCommand(), env.watchedCommand(), env.definitionsCommand(), versionCommand(root))
	return root
}

func versionCommand(root *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the tomaccio version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(cmd.OutOrStdout(), root.Version)
		},
	}
}

func buildVersion() string {
	if versionOverride != "" {
		return versionOverride
	}
	if info, ok := readBuildInfo(); ok && info != nil && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}

func (e *commandEnv) load(command string) (*config.Config, error) {
	load := e.loadConfig
	if load == nil {
		load = config.Load
	}
	cfg, err := load(e.configPath)
	if err != nil {
		return nil, err
	}
	if err := cfg.Validate(command); err != nil {
		return nil, err
	}
	setupLogging(cfg.App.LogLevel)
	return cfg, nil
}

func setupLogging(level string) {
	lvl := new(slog.LevelVar)
	switch strings.ToLower(level) {
	case "debug":
		lvl.Set(slog.LevelDebug)
	default:
		lvl.Set(slog.LevelInfo)
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: lvl})))
}
