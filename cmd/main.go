package main

import (
	"fmt"
	"log/slog"
	_ "net/http/pprof"
	"os"

	"github.com/AbdeltwabMF/gomeet/configs"
	"github.com/AbdeltwabMF/gomeet/internal/googlecal"
	"github.com/AbdeltwabMF/gomeet/internal/localcal"
)

func main() {
	file, err := configs.InitLogger()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer file.Close()

	cfg, err := configs.LoadConfig()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	go googlecal.Monitor(cfg)
	go localcal.Monitor(cfg)

	select {}
}
