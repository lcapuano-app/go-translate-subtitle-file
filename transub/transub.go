package transub

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	gtrans "github.com/lcapuano-app/go-googletrans"
)

type Options struct {
	RemoveCC     bool
	LanguageSrc  string
	OutputDir    string
	IsMainSub    bool
	RemoveOrigin bool
	Retries      int
	GTrans       GTransCfg
}
type withOptions = func(*Options)
type GTransCfg = gtrans.Config

type Transub struct {
	InputFile    string
	OutputFile   string
	LanguageDest string
	FileExt      string
	MetaStr      string
}

var opts Options

func WithRemoveCC(removeCC bool) func(*Options) {
	return func(opt *Options) {
		opt.RemoveCC = removeCC
	}
}

func WithMainSub(isMain bool) func(*Options) {
	return func(opt *Options) {
		opt.IsMainSub = isMain
	}
}

func WithRemoveOrigin(rmOrigin bool) func(*Options) {
	return func(opt *Options) {
		opt.RemoveOrigin = rmOrigin
	}
}

func WithLanguageSrc(langSrc string) func(*Options) {
	return func(opt *Options) {
		opt.LanguageSrc = langSrc
	}
}

func WithOutputDir(outputDir string) func(*Options) {
	return func(opt *Options) {
		opt.OutputDir = outputDir
	}
}

func WithGoogleTransCfg(cfg GTransCfg) func(*Options) {
	return func(opt *Options) {
		opt.GTrans = cfg
	}
}

func WithGoogleRetries(retries int) func(*Options) {
	return func(opt *Options) {
		if retries < 0 {
			retries = 0
		}
		if retries > 10 {
			log.Printf("limited to 10 retries. %d was requested", retries)
			retries = 10
		}
		opt.Retries = retries
	}
}

func New(filename, destLang string, options ...withOptions) *Transub {
	opts = Options{}
	opts.LanguageSrc = "auto"
	opts.Retries = 0
	tsub := Transub{}
	tsub.InputFile = filename
	tsub.FileExt = filepath.Ext(filename)
	tsub.OutputFile = ""

	for _, optFn := range options {
		optFn(&opts)
	}

	tsub.setLanguageDest(destLang)
	tsub.setOutputFilename()
	tsub.setMetaStr()

	return &tsub
}

func (ts *Transub) TranslateSSA() error {

	translateds, err := translateSSA(ts)
	if err != nil {
		fmt.Println(err)
		return err
	}

	if err = ts.CreateOutputFile(translateds); err != nil {
		return err
	}

	if err = ts.MarkOriginAsTrasnlated(); err != nil {
		return err
	}

	if err = ts.ManageOriginDestFiles(); err != nil {
		return err
	}

	return nil
}

func (ts *Transub) TranslasteSRT() error {
	fmt.Println("RAMO LA", ts.InputFile)
	translateds, err := translasteSRT(ts)
	if err != nil {
		fmt.Println(err)
		return err
	}

	if err = ts.CreateOutputFile(translateds); err != nil {
		return err
	}

	if err = ts.MarkOriginAsTrasnlated(); err != nil {
		return err
	}

	if err = ts.ManageOriginDestFiles(); err != nil {
		return err
	}
	// fileAsStrArr, err := []string{}, nil // ts.translatePrepare(".srt")
	// if err != nil {
	// 	return err
	// }

	// originals, translatables, err := extractSRTSpeechLines(fileAsStrArr, opts.RemoveCC)
	// if err != nil {
	// 	return err
	// }

	// joinedSpeeches := joinSpeechesByCharLimit(translatables)
	// if len(joinedSpeeches) == 0 {
	// 	err = fmt.Errorf("[transub] zero translatable lines in file. %s", ts.InputFile)
	// 	return err
	// }

	// if err = ts.updateSrcLang(joinedSpeeches); err != nil {
	// 	return err
	// }

	// translatedSpeeches := translateMany(joinedSpeeches, opts.LanguageSrc, ts.LanguageDest, opts.Retries)
	// translateds := rebuildAsOriginalLinesSRT(translatedSpeeches, originals)

	// if err = ts.CreateOutputFile(translateds); err != nil {
	// 	return err
	// }

	// if err = ts.MarkOriginAsTrasnlated(); err != nil {
	// 	return err
	// }

	// if err = ts.ManageOriginDestFiles(); err != nil {
	// 	return err
	// }

	return nil
}

// func (ts *Transub) translatePrepare(ext string) (fileLines []string, err error) {

// 	if ts.FileExt != ext {
// 		return fileLines, fmt.Errorf("this is not a %s file", ext)
// 	}

// 	if !IsValidSubtitleFile(ts.InputFile, ts.FileExt) {
// 		return fileLines, fmt.Errorf("unreachable %s file: %s", ext, ts.InputFile)
// 	}

// 	if outputFileExists(ts.OutputFile) {
// 		return fileLines, fmt.Errorf("output file %s already exists", ts.OutputFile)
// 	}

// 	fileLines, err = getFileStrLines(ts.InputFile)
// 	if err != nil {
// 		return fileLines, err
// 	}
// 	if len(fileLines) == 0 {
// 		return fileLines, fmt.Errorf("empty file")
// 	}

// 	if err = CheckForMetaStr(fileLines); err != nil {
// 		return fileLines, err
// 	}

// 	return fileLines, nil
// }

func (ts *Transub) updateSrcLang(sample []string) error {
	detectedSrcLang, err := detectSourceLanguage(sample[0])
	if err != nil {
		return err
	}

	if detectedSrcLang != opts.LanguageSrc && opts.LanguageSrc != "auto" {
		warn := fmt.Sprintf(
			"[Warning] - using '%s' as src language instead of '%s'",
			detectedSrcLang,
			opts.LanguageSrc,
		)
		log.Println(warn)
	}

	opts.LanguageSrc = detectedSrcLang

	if ts.LanguageDest == detectedSrcLang {
		err := fmt.Errorf(
			"[transub] this file (%s) was written into dest language. (%s)",
			ts.InputFile,
			detectedSrcLang,
		)
		return err
	}

	return nil
}

func (ts *Transub) CreateOutputFile(strLines []string) error {

	file, err := os.OpenFile(ts.OutputFile, fileEditFlag, 0666)
	if err != nil {
		return err
	}
	strLines = append(strLines, LN_BREAK+ts.MetaStr)
	datawriter := bufio.NewWriter(file)
	for _, text := range strLines {
		// if opts.RemoveCC && Validator.isCC(text) {
		// 	continue
		// }
		if _, err := datawriter.WriteString(text + "\n"); err != nil {
			log.Println(err)
		}

	}
	if err = datawriter.Flush(); err != nil {
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	return nil
}

func (ts *Transub) MarkOriginAsTrasnlated() error {
	filelines, err := getFileStrLines(ts.InputFile)
	if err != nil {
		return err
	}
	metaStr := fmt.Sprintf("%s;%s\n", META_TRASNLATED, opts.LanguageSrc)
	filelines = append(filelines, LN_BREAK+metaStr)

	if err = os.Remove(ts.InputFile); err != nil {
		return err
	}
	file, err := os.OpenFile(ts.InputFile, fileEditFlag, 0666)
	if err != nil {
		return err
	}
	datawriter := bufio.NewWriter(file)
	for _, text := range filelines {
		if opts.RemoveCC && ts.FileExt != ".ssa" {
			text = Validator.removeCC(text)
		}
		if _, err := datawriter.WriteString(text + "\n"); err != nil {
			log.Println(err)
		}

	}
	if err = datawriter.Flush(); err != nil {
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	return nil
}

func (ts *Transub) ManageOriginDestFiles() error {
	var err error

	// Keep translation and delete original file while changing the
	// translated file name to the original file name
	if opts.IsMainSub && opts.RemoveOrigin {
		err = os.Remove(ts.InputFile)
		if err != nil {
			return err
		}
		err = os.Rename(ts.OutputFile, ts.InputFile)
		if err != nil {
			return err
		}
		return nil
	}

	// Keep both files but change the translated file name to original file name
	// and rename the original file to file.language.srt
	if opts.IsMainSub && !opts.RemoveOrigin {
		noExtFilename := removeFileExtension(ts.InputFile, ts.FileExt)
		renamedPath := fmt.Sprintf("%s.%s%s", noExtFilename, opts.LanguageSrc, ts.FileExt)
		err = os.Rename(ts.InputFile, renamedPath)
		if err != nil {
			return err
		}
		err = os.Rename(ts.OutputFile, ts.InputFile)
		if err != nil {
			return err
		}
		return nil
	}

	// Keep translated file as filename.transLang.srt
	// and remove the original file
	if opts.RemoveOrigin {
		err = os.Remove(ts.InputFile)
		return err
	}

	// Keep booth otherwise
	return nil
}

func (ts *Transub) getSourceFileLines() (fileLines []string, err error) {

	if err := ts.validateTranslationSourceDest(); err != nil {
		return fileLines, err
	}

	fileLines, err = getFileStrLines(ts.InputFile)
	if err != nil {
		return fileLines, err
	}
	if len(fileLines) == 0 {
		return fileLines, fmt.Errorf("empty file")
	}

	if err = CheckForMetaStr(fileLines); err != nil {
		return fileLines, err
	}

	return fileLines, nil
}

func (ts *Transub) validateTranslationSourceDest() error {
	if !Validator.isReachableFile(ts.InputFile) {
		return fmt.Errorf("unreachable %s file: %s", ts.FileExt, ts.InputFile)
	}

	if Validator.isTranslatedFilename(ts.InputFile, ts.LanguageDest) {
		return fmt.Errorf("this file apears to be a translation: %s", ts.InputFile)
	}

	ok, err := Validator.isTextFile(ts.InputFile)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("this is not a valid subtitle file [%s]", ts.InputFile)
	}

	if _, err := os.Stat(ts.OutputFile); err == nil {
		return fmt.Errorf("output file %s already exists", ts.OutputFile)
	}

	return nil
}

func (ts *Transub) setOutputFilename() {
	outName, outDir := getOutputFileNameAndDir(
		ts.InputFile,
		ts.LanguageDest,
		ts.FileExt,
	)
	if len(opts.OutputDir) > 0 {
		ts.OutputFile = filepath.Join(opts.OutputDir, outName)
	} else {
		ts.OutputFile = filepath.Join(outDir, outName)
	}
}

func (ts *Transub) setLanguageDest(destLang string) {
	lang, err := gtrans.GetValidLanguageKey(destLang)
	if err != nil || lang == "auto" {
		log.Println(fmt.Errorf("invalid dest language: %s", destLang))
		ts.LanguageDest = "en"
	} else {
		ts.LanguageDest = lang
	}
}

func (ts *Transub) setMetaStr() {
	metaStr := fmt.Sprintf("\n%s;%s\n", META_TRASNLATED, ts.LanguageDest)
	ts.MetaStr = metaStr
}

func CheckForMetaStr(fileLines []string) error {
	err := fmt.Errorf(
		"file already translated. If this is a false positive, please, "+
			"locate and remove this text line: '%s' (should be the last entry)",
		META_TRASNLATED,
	)

	lastIdx := len(fileLines) - 1
	if strings.HasPrefix(fileLines[lastIdx], META_TRASNLATED) {

		return err
	}

	for i := lastIdx; i >= 0; i-- {
		ln := fileLines[i]
		if strings.HasPrefix(ln, META_TRASNLATED) {
			return err
		}
		if Validator.isIntStr(ln) {
			return nil
		}
	}

	return nil
}

func removeLangFromFilename(filename, ext string) string {
	delExt := strings.ReplaceAll(ext, ".", "")
	splited := strings.Split(filename, ".")
	if splited[len(splited)-1] == delExt {
		splited = splited[:len(splited)-1]
	}
	if len(splited) <= 1 {
		return filename
	}
	srcLang, results := splited[len(splited)-1], splited[:len(splited)-1]

	if _, err := gtrans.GetValidLanguageKey(srcLang); err != nil {
		return filename
	}
	return strings.Join(results, ".") + ext
}

func removeFileExtension(filename, ext string) string {
	idx := strings.LastIndex(filename, ext)
	filenameRmExt := filename[:idx] + strings.Replace(filename[idx:], ext, "", 1)
	return filenameRmExt
}

func getOutputFileNameAndDir(filename, dest, ext string) (fname, fdir string) {
	transSuffix := "." + dest + ext
	noLangFilename := removeLangFromFilename(filename, ext)
	noExtFilename := removeFileExtension(noLangFilename, ext)
	translatedFilename := noExtFilename + transSuffix
	fname = filepath.Base(translatedFilename)
	fdir = filepath.Dir(translatedFilename)
	return fname, fdir
}

// func extractSRTSpeechLines(fileAsStrArr []string, removeCC bool) (originals, translatables []string, err error) {
// 	normalizeLine := func(line string, index, accSize int) string {
// 		if Validator.isTranslatable(line, removeCC) {
// 			line = strings.ReplaceAll(line, ";", ",")
// 			if accSize == 0 {
// 				line = fmt.Sprintf("%d;%s", index, line)
// 			}
// 		} else {
// 			line = ""
// 		}
// 		return strings.TrimSpace(line)
// 	}

// 	originals = fileAsStrArr

// 	var acc []string
// 	for i := 0; i < len(originals); i++ {
// 		line := originals[i]
// 		if Validator.isIntStr(line) || Validator.isTimestampStr(line) {
// 			translatables = append(translatables, "")
// 			continue
// 		}
// 		if Validator.isLineBreak(line) {
// 			joined := strings.Join(acc, LN_SEP)
// 			translatables = append(translatables, joined)
// 			acc = []string{}
// 			continue
// 		}

// 		line = normalizeLine(line, i, len(acc))
// 		if len(line) > 0 {
// 			acc = append(acc, line)
// 		}

// 	}

// 	if len(acc) > 0 {
// 		joined := strings.Join(acc, LN_SEP)
// 		translatables = append(translatables, joined)
// 	}

// 	return originals, translatables, nil
// }

func getFileStrLines(filename string) ([]string, error) {
	var lines []string
	readFile, err := os.Open(filename)
	if err != nil {
		return lines, err
	}

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)
	for fileScanner.Scan() {
		line := strings.Trim(fileScanner.Text(), "\ufeff")
		line = strings.TrimSpace(line)
		lines = append(lines, line)
	}
	if err := readFile.Close(); err != nil {
		return lines, err
	}
	return lines, nil
}

// func joinSpeechesByCharLimit(speeches []string) []string {

// 	var chuncks []string
// 	var chunck string
// 	for _, line := range speeches {
// 		nextLen := len(line) + len(chunck)
// 		if nextLen >= gtransCharLimit {
// 			chuncks = append(chuncks, chunck)
// 			chunck = ""
// 		} else if len(line) > 0 {
// 			chunck += line + LN_BREAK
// 		}
// 	}
// 	chuncks = append(chuncks, chunck)
// 	return chuncks
// }

func detectSourceLanguage(text string) (string, error) {
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
	translator := gtrans.New(opts.GTrans)
	res, err := translator.DetectLanguage(sample, "auto")
	if err != nil {
		return "", err
	}
	if res.Confidence < 0.3 {
		return res.Src, err
	}
	return res.Src, nil
}

func translateMany(texts []string, src, dest string, retries int) []string {
	ch := make(chan string)
	var translateds []string
	var wg sync.WaitGroup
	wg.Add(len(texts))
	for _, text := range texts {
		gTranslator := *gtrans.New(opts.GTrans)
		go translateOneWG(text, src, dest, retries, gTranslator, &wg, ch)
	}
	go func() {
		wg.Wait()
		close(ch)
	}()
	for res := range ch {
		translateds = append(translateds, res)
	}
	return translateds
}

func translateOneWG(text, src, dest string, retries int, gTranslator gtrans.Translator, wg *sync.WaitGroup, ch chan<- string) {
	defer wg.Done()
	translatedText := translateOne(text, src, dest, retries, gTranslator)
	ch <- translatedText
}

func translateOne(text, src, dest string, retries int, gTranslator gtrans.Translator) string {
	result, err := gTranslator.Translate(text, src, dest)
	if err == nil {
		return result.Text
	}

	if err != nil && retries <= 0 {
		log.Println(err, "no retries attempts left, returning original text")
		return text
	}

	msg := fmt.Sprintf("will attempt with a diferent service url and user agent. attempts left [%d]", retries)
	log.Println(err, msg)
	retries--

	gTranslator = *gtrans.New(gtrans.Config{
		ServiceUrls: gtrans.GetDefaultServiceUrls(),
		UserAgent:   []string{},
		Proxy:       opts.GTrans.Proxy,
	})
	return translateOne(text, src, dest, retries, gTranslator)
}

// func rebuildAsOriginalLinesSRT(translatedSpeeches, originals []string) []string {
// 	getLineText := func(translation string) (int, string) {
// 		splited := strings.Split(translation, ";")
// 		if len(splited) < 1 {
// 			return -1, ""
// 		}
// 		idx, err := strconv.Atoi(strings.TrimSpace(splited[0]))
// 		if err != nil {
// 			return -1, ""
// 		}
// 		txt := strings.TrimSpace(splited[1])
// 		return idx, txt
// 	}

// 	for _, speechLine := range translatedSpeeches {
// 		subtitles := strings.Split(speechLine, LN_BREAK)
// 		if len(subtitles) == 0 {
// 			continue
// 		}
// 		for _, subtitle := range subtitles {
// 			idx, text := getLineText(subtitle)
// 			if idx == -1 {
// 				continue
// 			}

// 			subTexts := strings.Split(text, LN_SEP)
// 			if len(subTexts) == 0 {
// 				continue
// 			}
// 			if len(subTexts) == 1 {
// 				originals[idx] = text
// 				continue
// 			}
// 			for subIdx, subTxt := range subTexts {
// 				originals[idx+subIdx] = subTxt
// 			}
// 		}
// 	}

// 	return originals
// }
