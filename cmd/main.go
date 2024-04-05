package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/AbdeltwabMF/gomeet/configs"
	"github.com/AbdeltwabMF/gomeet/internal/googlecal"
	"github.com/AbdeltwabMF/gomeet/internal/localcal"
)

func main() {
	f, err := configs.OpenLog(os.O_CREATE | os.O_TRUNC | os.O_WRONLY)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer f.Close()
	configs.InitLogger(f)

	cfg, err := configs.LoadConfig()
	if err != nil {
		slog.Error(err.Error(), slog.Any("func", configs.CallerInfo()))
		os.Exit(1)
	}

	go googlecal.Monitor(cfg)
	go localcal.Monitor(cfg)

	select {}
}
