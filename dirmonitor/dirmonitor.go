package dirmonitor

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/lcapuano-app/go-translate-subtitle-file/config"
	"github.com/lcapuano-app/go-translate-subtitle-file/logger"
	"github.com/lcapuano-app/go-translate-subtitle-file/transub"
)

var cfg *config.Config

func Setup(config *config.Config) {
	cfg = config
	var paths []string
	for _, baseDir := range cfg.MonitorPaths {
		filenames := findFilesPathsToTranslate(baseDir, cfg.Lang)
		paths = append(paths, filenames...)
	}
	translateBatch(paths)
}

func Watch() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Panic(err)
	}
	defer watcher.Close()
	done := make(chan bool)
	go monitorLoop(watcher)
	addMonitorPathsWatchers(watcher)
	<-done
}

func findFilesPathsToTranslate(baseDir, lang string) []string {
	var subtitlePaths []string
	filepath.WalkDir(baseDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			logger.Err(err)
			return err
		}

		isSrtFile := filepath.Ext(d.Name()) == ".srt"
		isSSaFile := filepath.Ext(d.Name()) == ".ssa"
		isMkvFile := filepath.Ext(d.Name()) == ".mkv"
		if isSSaFile || isSrtFile {
			subtitlePaths = append(subtitlePaths, path)
		}
		if isMkvFile {
			fmt.Println("FILE, p", path)
		}

		return nil
	})

	return subtitlePaths
}

func monitorLoop(watcher *fsnotify.Watcher) {
	for {
		select {
		case event := <-watcher.Events:
			// if event.Has(fsnotify.Write) && event.Name == conf.Filename {
			//     cfg = conf.ReloadConfig()
			//     break
			// }
			if event.Has(fsnotify.Create) {
				time.Sleep(time.Second)
				translateOne(event.Name)
			}

		case err := <-watcher.Errors:
			logger.Err(err)
		}
	}
}

func addMonitorPathsWatchers(watcher *fsnotify.Watcher) {
	for _, monitorPath := range cfg.MonitorPaths {
		filepath.Walk(monitorPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				logger.Err(err)
			}
			if info.IsDir() {
				if err := watcher.Add(path); err != nil {
					logger.Err(err)
				}
			}
			return nil
		})
	}
}

func translateBatch(paths []string) {
	var wg sync.WaitGroup
	wg.Add(len(paths))

	for _, path := range paths {
		go translateOneWG(&wg, path)
	}
	wg.Wait()
}

func translateOneWG(wg *sync.WaitGroup, filename string) error {
	defer wg.Done()
	return translateOne(filename)
}

func translateOne(filename string) error {
	ts := transub.New(
		filename,
		cfg.Lang,
		transub.WithRemoveCC(!cfg.CC),
		transub.WithMainSub(cfg.SaveOutputAsMain),
		transub.WithRemoveOrigin(!cfg.KeepSrcFile),
	)

	ext := filepath.Ext(filename)
	var err error
	if ext == ".srt" {
		// if err := ts.TranslasteSRT(); err != nil {
		//     logger.Err(err)
		//     return err
		// }
		fmt.Println("VOLTAR O SRT HEIN!!")
		//return ts.TranslasteSRT()
		return err
	}
	if ext == ".ssa" {
		return ts.TranslateSSA()
	}

	return fmt.Errorf("invalid extension - this should never hapen")
}
