package transub

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	gtrans "github.com/lcapuano-app/go-googletrans"
)

type Config struct {
	RemoveCloseCaption bool
	LanguageSrc        string
	LanguageDest       string
	OutputDir          string
}

type Transub struct {
	InputFilename  string
	OutputFilename string
	FileExt        string
	Conf           Config
}

type FileLines struct {
	Originals     []string
	Translatables []string
	Translateds   []string
	TempTrans     []string
}

var fileTextLines FileLines

const (
	LN_BREAK            = "\n"
	translatorLineLimit = 5_000
)

var translator gtrans.Translator

func TransubNew(filename string, config ...Config) *Transub {
	opts := getOptions(config...)
	ext := filepath.Ext(filename)
	ts := Transub{
		InputFilename:  filename,
		OutputFilename: "",
		FileExt:        ext,
		Conf:           opts,
	}
	translator = *gtrans.New()
	return &ts
}

func (ts *Transub) Translaste() error {
	if !isValidFile(ts.InputFilename, ts.FileExt) {
		return fmt.Errorf("invalid file: %s", ts.InputFilename)
	}
	if ts.translatedFileExists() {
		return fmt.Errorf("file %s already translated", filepath.Base(ts.InputFilename))
	}
	originals, translatables, err := ts.ExtractFileLines()
	if err != nil {
		return err
	}
	fileTextLines.Originals = originals
	fileTextLines.Translatables = translatables
	translatedChuncks, err := ts.translateFileLines(translatables)
	if err != nil {
		return err
	}

	fileTextLines.Translateds = replaceTranslatedLines(translatedChuncks, originals)
	ts.CreateOutputFile(fileTextLines)
	return nil
}

func (ts *Transub) RemoveLangFromFilename() string {
	delExt := strings.ReplaceAll(ts.FileExt, ".", "")
	splited := strings.Split(ts.InputFilename, ".")
	if splited[len(splited)-1] == delExt {
		splited = splited[:len(splited)-1]
	}
	if len(splited) <= 1 {
		return ts.InputFilename
	}
	srcLang, results := splited[len(splited)-1], splited[:len(splited)-1]
	if _, err := translator.GetValidLanguageKey(srcLang); err != nil {
		return ts.InputFilename
	}
	return strings.Join(results, ".") + ts.FileExt
}

func (ts *Transub) ExtractFileLines() (originals, translatables []string, err error) {
	readFile, err := os.Open(ts.InputFilename)
	if err != nil {
		return originals, translatables, err
	}
	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)
	lazyIdx := 0
	for fileScanner.Scan() {
		line := strings.Trim(fileScanner.Text(), "\ufeff")
		originals = append(originals, line)
		if isTranslatable(line, ts.Conf.RemoveCloseCaption) {
			line = strings.ReplaceAll(line, ";", ",")
			line = fmt.Sprintf("%d;%s", lazyIdx, line)
		} else {
			line = ""
		}
		translatables = append(translatables, line)
		lazyIdx++
	}
	readFile.Close()
	return originals, translatables, nil
}

func (ts *Transub) CreateOutputFile(fileTextLines FileLines) error {
	file, err := os.OpenFile(
		ts.OutputFilename,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0666,
	)
	if err != nil {
		return err
	}
	datawriter := bufio.NewWriter(file)
	for _, text := range fileTextLines.Translateds {
		if ts.Conf.RemoveCloseCaption && isCC(text) {
			continue
		}
		datawriter.WriteString(text + "\n")
	}
	datawriter.Flush()
	file.Close()
	return nil
}

func (ts *Transub) translatedFileExists() bool {
	if ts.Conf.LanguageDest == "auto" {
		return false
	}
	transSuffix := "." + ts.Conf.LanguageDest + ts.FileExt
	isTransled := strings.HasSuffix(ts.InputFilename, transSuffix)
	if isTransled {
		return true
	}
	outputFilename := ts.getOutputFilename()
	if _, err := os.Stat(outputFilename); err == nil {
		return true
	}

	return false
}

func (ts *Transub) getOutputFilename() string {
	if ts.Conf.LanguageDest == "auto" {
		return ""
	}
	transSuffix := "." + ts.Conf.LanguageDest + ts.FileExt
	noLangFilename := ts.RemoveLangFromFilename()
	noExtFilename := removeFileExtension(noLangFilename, ts.FileExt)
	translatedFilename := noExtFilename + transSuffix
	if len(ts.Conf.OutputDir) == 0 {
		ts.OutputFilename = translatedFilename
		return translatedFilename
	}

	ts.OutputFilename = filepath.Join(
		ts.Conf.OutputDir,
		filepath.Base(translatedFilename),
	)
	return ts.OutputFilename
}

func (ts *Transub) translateFileLines(lines []string) (translateds []string, err error) {
	transLines := spliptByTranslatorLim(lines)
	if len(transLines) == 0 {
		err = fmt.Errorf("[transub] zero translatable lines in file. %s", ts.InputFilename)
		return transLines, err
	}
	detectedLang := detectOriginLanguage(transLines[0])
	ts.Conf.LanguageSrc = detectedLang
	if ts.Conf.LanguageDest == detectedLang {
		err = fmt.Errorf("[transub] this file (%s) was written into dest language. (%s)", ts.InputFilename, detectedLang)
		return transLines, err
	}
	translatedChuncks := translateAllChuncks(transLines, ts.Conf.LanguageSrc, ts.Conf.LanguageDest)
	return translatedChuncks, nil
}

func replaceTranslatedLines(transRawLines, originals []string) []string {
	getLineText := func(translation string) (int, string) {
		splited := strings.Split(translation, ";")
		if len(splited) < 1 {
			return -1, ""
		}
		idx, err := strconv.Atoi(strings.TrimSpace(splited[0]))
		if err != nil {
			return -1, ""
		}
		txt := strings.TrimSpace(splited[1])
		return idx, txt
	}

	for _, rawLine := range transRawLines {
		subtitles := strings.Split(rawLine, LN_BREAK)
		if len(subtitles) == 0 {
			continue
		}
		for _, subtitle := range subtitles {
			idx, text := getLineText(subtitle)
			if idx == -1 {
				continue
			}
			originals[idx] = text
		}
	}
	return originals
}

func translateAllChuncks(chuncksToTranslate []string, src, dest string) []string {
	ch := make(chan string)
	var translatedChuncks []string
	var wg sync.WaitGroup
	wg.Add(len(chuncksToTranslate))
	for _, chunck := range chuncksToTranslate {
		go getTranslatedTextWG(chunck, src, dest, &wg, ch)
	}
	go func() {
		wg.Wait()
		close(ch)
	}()
	for res := range ch {
		translatedChuncks = append(translatedChuncks, res)
	}
	return translatedChuncks
}

func getTranslatedTextWG(text, src, dest string, wg *sync.WaitGroup, ch chan<- string) {
	defer wg.Done()
	result, err := translator.Translate(text, src, dest)
	if err != nil {
		log.Println(err)
		ch <- text
	} else {
		ch <- result.Text
	}
}

func detectOriginLanguage(text string) string {
	getSample := func(text string) string {
		sz := 80
		if len(text) <= sz {
			return text
		}
		textSlice := text[:sz]
		idx := len(textSlice) - 1
		for i := idx; i >= 0; i-- {
			strI := string(textSlice[i])
			if strI == " " || strI == LN_BREAK {
				idx = i
				break
			}
		}
		return textSlice[:idx]
	}
	sample := getSample(text)
	res, err := translator.DetectLanguage(sample, "auto")
	if err != nil {
		return "auto"
	}
	if res.Confidence < 0.5 {
		return "auto"
	}
	return res.Src
}

func getOptions(config ...Config) Config {
	var opts Config
	if len(config) > 0 {
		opts = overwriteOptions(config[0])
	} else {
		opts = getDefaultOptions()
	}
	return opts
}

func getDefaultOptions() Config {
	opts := Config{
		RemoveCloseCaption: false,
		LanguageSrc:        "auto",
		LanguageDest:       "en",
		OutputDir:          "",
	}
	return opts
}

func overwriteOptions(config Config) Config {
	opts := getDefaultOptions()
	if len(config.LanguageDest) > 0 {
		opts.LanguageDest = strings.ToLower(config.LanguageDest)
	}
	if len(config.LanguageSrc) > 0 {
		opts.LanguageSrc = strings.ToLower(config.LanguageSrc)
	}
	if len(config.OutputDir) > 0 {
		opts.OutputDir = config.OutputDir
	}
	opts.RemoveCloseCaption = config.RemoveCloseCaption
	return opts
}

func removeFileExtension(filename, ext string) string {
	idx := strings.LastIndex(filename, ext)
	filenameRmExt := filename[:idx] + strings.Replace(filename[idx:], ext, "", 1)
	return filenameRmExt
}

func spliptByTranslatorLim(translatables []string) []string {
	var chuncks []string
	var chunck string
	for _, line := range translatables {
		nextLen := len(line) + len(chunck)
		if nextLen >= translatorLineLimit {
			chuncks = append(chuncks, chunck)
			chunck = ""
		} else if len(line) > 0 {
			chunck += line + LN_BREAK
		}
	}
	chuncks = append(chuncks, chunck)
	return chuncks
}
