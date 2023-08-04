package main

import (
	"fmt"

	"github.com/lcapuano-app/go-translate-subtitle-file/config"
	"github.com/lcapuano-app/go-translate-subtitle-file/dirmonitor"
	"github.com/lcapuano-app/go-translate-subtitle-file/logger"
	"github.com/lcapuano-app/go-translate-subtitle-file/transub"
)

func printWelcome() {
	fmt.Println()
	fmt.Println(" lcapuano.com.br                                                 ")
	fmt.Println("  __   __  ___    ___  __             __            ___  __   __  ")
	fmt.Println(" /__` |__)  |      |  |__)  /\\  |\\ | /__` |     /\\   |  /  \\ |__) ")
	fmt.Println(" .__/ |  \\  |      |  |  \\ /~~\\ | \\| .__/ |___ /~~\\  |  \\__/ |  \\")
	fmt.Println("                                                           v0.0.1")
	fmt.Println()
	fmt.Println(" use --help")
	fmt.Println()
}

func main() {
	printWelcome()
	cfg := config.New()
	logger.SetLogger(cfg.LogPath, cfg.LogLevel)

	if cfg.DoNotMonitor {
		translateOnce(cfg)
		return
	}

	dirmonitor.Setup(cfg)
	dirmonitor.Watch()

}

func translateOnce(cfg *config.Config) {
	if len(cfg.MonitorPaths) == 0 {
		return
	}
	ts := transub.New(
		cfg.MonitorPaths[0],
		cfg.Lang,
		transub.WithRemoveCC(!cfg.CC),
		transub.WithGoogleRetries(cfg.Retries),
	)
	if err := ts.TranslasteSRT(); err != nil {
		logger.Err(err)
	}
	logger.Info("done")
}
