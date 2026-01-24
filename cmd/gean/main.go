package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/devylongs/gean/config"
	"github.com/devylongs/gean/node"
)

var cli struct {
	GenesisDir string `arg:"" required:"" help:"Path to genesis directory"`
	NodeID     string `help:"Node ID from validators.yaml (e.g., gean_0)"`
	Listen     string `default:"/ip4/0.0.0.0/tcp/9000" help:"Listen multiaddr"`
	LogLevel   string `default:"info" enum:"debug,info,warn,error" help:"Log level"`
}

func main() {
	kong.Parse(&cli,
		kong.Name("gean"),
		kong.Description("Lean Ethereum consensus client"),
	)

	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ gean ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	// Setup logger
	level := slog.LevelInfo
	switch cli.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))

	// Load configuration
	cfg, validatorIndices, bootnodes, err := config.Load(cli.GenesisDir, cli.NodeID)
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Build node config
	nodeCfg := &node.Config{
		GenesisTime:    cfg.GenesisTime,
		ValidatorCount: cfg.ValidatorCount,
		ListenAddrs:    []string{cli.Listen},
		Bootnodes:      bootnodes,
		Logger:         logger,
	}

	if len(validatorIndices) > 0 {
		nodeCfg.ValidatorIndex = &validatorIndices[0]
		logger.Info("running as validator", "index", validatorIndices[0])
	}

	logger.Info("loaded config",
		"genesis_time", cfg.GenesisTime,
		"validators", cfg.ValidatorCount,
		"bootnodes", len(bootnodes),
	)

	// Create and start node
	n, err := node.New(context.Background(), nodeCfg)
	if err != nil {
		logger.Error("failed to create node", "error", err)
		os.Exit(1)
	}

	n.Start()
	logger.Info("gean running", "slot", n.CurrentSlot(), "peers", n.PeerCount())

	// Wait for shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	logger.Info("shutting down...")
	n.Stop()
}
