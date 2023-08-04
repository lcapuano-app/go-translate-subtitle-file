package config

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	gtrans "github.com/lcapuano-app/go-googletrans"
)

type cliFlags struct {
	lang    string
	log     string
	src     string
	cc      bool
	retries int
}

type Config struct {
	CC               bool
	Lang             string
	LogLevel         string
	LogPath          string
	MonitorPaths     []string
	DoNotMonitor     bool
	KeepSrcFile      bool
	Retries          int
	SaveOutputAsMain bool
}

const (
	ConfigFilename  = "config.conf"
	lang_default    = "en"
	ccKey           = "CLOSE_CAPTIONS"
	ccValue         = "false"
	keepSrcKey      = "KEEP_SOURCE_FILE"
	keepSrcVal      = "true"
	logKey          = "LOG_PATH"
	logVal          = "path/to/log"
	langKey         = "LANG"
	langVal         = "pt"
	logLevelKey     = "LOG_LEVEL"
	logLevelVal     = "DEBUG"
	monitorPathKey  = "MONITOR_PATHS"
	monitorPathVal  = "path/to/monitor/folder, or/multiple/folders, comma/separated"
	retriesKey      = "RETRIES"
	retriesVal      = "0"
	saveDestMainKey = "SAVE_OUTPUT_AS_MAIN_FILE"
	saveDestMainVal = "false"
)

var cfg Config

func New() *Config {
	args, useCliCfg := getCliFlags()
	if useCliCfg {
		setConfigFromCliFlags(args)
	} else {
		setConfigFromFile()
	}
	return &cfg
}

func Get() *Config {
	return &cfg
}

func setConfigFromCliFlags(args cliFlags) {
	cfg.DoNotMonitor = true
	cfg.CC = args.cc
	cfg.Lang = args.lang
	cfg.LogPath = args.log
	cfg.LogLevel = "DEBUG"
	cfg.MonitorPaths = []string{args.src}
	cfg.Retries = args.retries
}

func getCliFlags() (cliFlags, bool) {
	langPtr := flag.String("lang", lang_default, "language that you want to translato to")
	srcPtr := flag.String("src", "", "path to srt file ")
	logPtr := flag.String("log", "", "path to log file")
	ccPtr := flag.Bool("cc", false, "keep close captions [CC]")
	rtPtr := flag.Int("rt", 0, "number of retries attempts")

	flag.Parse()

	var logPath string = *logPtr
	if len(*logPtr) == 0 {
		logPath = filepath.Dir(*srcPtr)
	}

	args := cliFlags{
		lang:    *langPtr,
		log:     logPath,
		src:     *srcPtr,
		cc:      *ccPtr,
		retries: *rtPtr,
	}

	if len(args.src) == 0 {
		return args, false
	}

	if len(args.lang) == 0 {
		args.lang = lang_default
	}

	if args.lang != lang_default {
		lang, err := gtrans.GetValidLanguageKey(args.lang)
		if err != nil || lang == "auto" {
			log.Println("invalid lang " + args.lang)
			args.lang = lang_default
		}
	}

	if args.retries < 0 {
		args.retries = 0
	}

	return args, true
}

func setConfigFromFile() {
	readFile, err := os.Open(ConfigFilename)
	if err != nil {
		createDefaultConfigFile()
		errMsg := configFileNotFound(fmt.Sprintf("Please fillout %s file.", ConfigFilename))
		log.Println(err)
		panic(errMsg)
	}

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)

	for fileScanner.Scan() {
		line := strings.TrimSpace(fileScanner.Text())
		configFileHandler(line)
	}

	if err = readFile.Close(); err != nil {
		log.Fatal(err)
	}
}

func configFileHandler(text string) *Config {

	splitText := func(ln string) (key, value string, err error) {
		splited := strings.Split(text, "=")
		if len(splited) < 2 {
			err = fmt.Errorf("invalid comand line")
			return "", "", err
		}
		key = strings.TrimSpace(splited[0])
		value = strings.TrimSpace(splited[1])
		return key, value, nil
	}

	key, value, err := splitText(text)

	if err != nil {
		log.Println(err)
		return &cfg
	}

	if strings.HasPrefix(key, monitorPathKey) {
		pathsStr := strings.Split(value, ",")
		for _, strPath := range pathsStr {
			strPath = strings.TrimSpace(strPath)
			path := filepath.FromSlash(strPath)
			cfg.MonitorPaths = append(cfg.MonitorPaths, path)
		}
		return &cfg
	}

	if strings.HasPrefix(key, langKey) {
		lang, err := gtrans.GetValidLanguageKey(value)
		if err != nil || lang == "auto" {
			cfg.Lang = lang_default
			log.Println(err)
		} else {
			cfg.Lang = lang
		}
		return &cfg
	}

	if strings.HasPrefix(key, logKey) {
		cfg.LogPath = value
		return &cfg
	}

	if strings.HasPrefix(key, logLevelKey) {
		cfg.LogLevel = strings.ToUpper(value)
		return &cfg
	}

	if strings.HasPrefix(key, ccKey) {
		closeCap, err := strconv.ParseBool(value)
		if err != nil {
			closeCap = false
		}
		cfg.CC = closeCap
		return &cfg
	}

	if strings.HasPrefix(key, keepSrcKey) {
		keepSrc, err := strconv.ParseBool(value)
		if err != nil {
			keepSrc = true
		}
		cfg.KeepSrcFile = keepSrc
		return &cfg
	}

	if strings.HasPrefix(key, retriesKey) {
		intVal, err := strconv.Atoi(value)
		if err != nil {
			intVal = 0
		}
		cfg.Retries = intVal
		return &cfg
	}

	if strings.HasPrefix(key, saveDestMainKey) {
		saveAsMain, err := strconv.ParseBool(value)
		if err != nil {
			saveAsMain = false
		}
		cfg.SaveOutputAsMain = saveAsMain
		return &cfg
	}

	return &cfg
}

func createDefaultConfigFile() {
	file, err := os.OpenFile(ConfigFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf(fmt.Sprintf("could not create %s file.", ConfigFilename))
		log.Panic(err)
	}

	datawriter := bufio.NewWriter(file)
	cfgs := []string{
		fmt.Sprintf("%s = %s", ccKey, ccValue),
		fmt.Sprintf("%s = %s", keepSrcKey, keepSrcVal),
		fmt.Sprintf("%s = %s", langKey, langVal),
		fmt.Sprintf("%s = %s", logLevelKey, logLevelVal),
		fmt.Sprintf("%s = %s", logKey, logVal),
		fmt.Sprintf("%s = %s", monitorPathKey, monitorPathVal),
		fmt.Sprintf("%s = %s", retriesKey, retriesVal),
		fmt.Sprintf("%s = %s", saveDestMainKey, saveDestMainVal),
	}

	for _, cfg := range cfgs {
		if _, err = datawriter.WriteString(cfg + "\n"); err != nil {
			log.Panic(err)
		}
	}

	if err = datawriter.Flush(); err != nil {
		log.Panic(err)
	}

	if err = file.Close(); err != nil {
		log.Panic(err)
	}
}

func configFileNotFound(customMsg string) string {
	br := "\n"
	tab := "    "
	msg := fmt.Sprintf("%s| %s |%s", tab, customMsg, tab)
	dashes := br + tab + strings.Repeat("-", (len(msg)-(len(tab)*2))) + tab + br
	return br + dashes + msg + dashes
}

func checkConfig() {

}
