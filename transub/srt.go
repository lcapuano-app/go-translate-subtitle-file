package transub

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

type srtTranslate struct {
	originals     []string
	translateds   []string
	translatables []string
	transChuncks  []string
}

func translasteSRT(ts *Transub) ([]string, error) {

	fileLines, err := ts.getSourceFileLines()
	if err != nil {
		return fileLines, err
	}

	var srt srtTranslate
	srt.originals = fileLines
	srt.extractSpeechLines(opts.RemoveCC)
	if err = ts.updateSrcLang(srt.transChuncks); err != nil {
		log.Println(err, "I'll keep using '%s'", opts.LanguageSrc)
	}
	transSpeeches := translateMany(srt.transChuncks, opts.LanguageSrc, ts.LanguageDest, opts.Retries)
	srt.mergeTranslatedToOriginal(transSpeeches)
	return srt.translateds, nil
}

func (srt *srtTranslate) extractSpeechLines(removeCC bool) {
	translatableLine := ""

	normalizeLine := func(line string, index, accSize int) string {
		if Validator.isTranslatable(line, removeCC) {
			line = strings.ReplaceAll(line, ";", ",")
			if accSize == 0 {
				line = fmt.Sprintf("%d;%s", index, line)
			}
		} else {
			line = ""
		}
		return strings.TrimSpace(line)
	}

	updateTransChuncks := func(text string) {
		nextLineLen := len(translatableLine) + len(text)
		if nextLineLen >= gtransCharLimit {
			srt.transChuncks = append(srt.transChuncks, translatableLine)
			translatableLine = ""
		}
		if len(text) > 0 {
			translatableLine += text + LN_BREAK
		}
	}

	var acc []string
	for idx, line := range srt.originals {
		if Validator.isIntStr(line) || Validator.isTimestampStr(line) {
			srt.translatables = append(srt.translatables, "")
			continue
		}
		if Validator.isLineBreak(line) {
			joined := strings.Join(acc, LN_SEP)
			srt.translatables = append(srt.translatables, joined)
			updateTransChuncks(joined)
			acc = []string{}
			continue
		}

		line = normalizeLine(line, idx, len(acc))
		if len(line) > 0 {
			acc = append(acc, line)
		}

	}
	if len(acc) > 0 {
		joined := strings.Join(acc, LN_SEP)
		srt.translatables = append(srt.translatables, joined)
		updateTransChuncks(joined)
	}
}

func (srt *srtTranslate) mergeTranslatedToOriginal(transSpeeches []string) {
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

	srt.translateds = srt.originals

	for _, speechLine := range transSpeeches {
		subtitles := strings.Split(speechLine, LN_BREAK)
		if len(subtitles) == 0 {
			continue
		}
		for _, subtitle := range subtitles {
			idx, text := getLineText(subtitle)
			if idx == -1 {
				continue
			}

			subTexts := strings.Split(text, LN_SEP)
			if len(subTexts) == 0 {
				continue
			}
			if len(subTexts) == 1 {
				srt.translateds[idx] = text
				continue
			}
			for subIdx, subTxt := range subTexts {
				srt.translateds[idx+subIdx] = subTxt
			}
		}
	}
}
