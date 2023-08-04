package transub

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
)

type sliptedDialogue struct {
	meta string
	text string
}

type ssaDialogues struct {
	texts    []string
	spliteds []sliptedDialogue
}

type ssaTranslate struct {
	originals     []string
	translateds   []string
	dialogues     ssaDialogues
	fileHeaders   []string
	translatables []string
}

type formatOK = int

func translateSSA(ts *Transub) (translateds []string, err error) {
	fileLines, err := ts.getSourceFileLines()
	if err != nil {
		return translateds, err
	}
	var ssa ssaTranslate
	ssa.originals = fileLines
	ssa.extractDialogues(fileLines, opts.RemoveCC, ssaParserUnknFormat)

	if err = ts.updateSrcLang(ssa.translatables); err != nil {
		log.Println(err, "I'll keep using '%s'", opts.LanguageSrc)
	}
	transDialogues := translateMany(ssa.translatables, opts.LanguageSrc, ts.LanguageDest, opts.Retries)
	ssa.mergeTranslatedToOriginal(transDialogues)

	return ssa.translateds, nil
}

func (ssa *ssaTranslate) extractDialogues(fileAsStrArr []string, removeCC bool, formatLen int) error {
	formatOK := false
	translatableLine := ""
	if formatLen >= 0 {
		ssa.fileHeaders = []string{}
		ssa.dialogues = ssaDialogues{}
		ssa.translatables = []string{}
	}

	updateTraslatable := func(dialogue sliptedDialogue) {
		nextLineLen := len(translatableLine) + len(dialogue.text)
		if nextLineLen >= gtransCharLimit {
			ssa.translatables = append(ssa.translatables, translatableLine)
			translatableLine = ""
		}
		if len(dialogue.text) > 0 {
			translatableLine += dialogue.text + LN_BREAK
		}
	}

	for idx, line := range ssa.originals {
		isDialogue := constCompare(line, ssaParserDialogueStr)
		if !isDialogue {
			ssa.appendDialogue("", "")
			ssa.fileHeaders = append(ssa.fileHeaders, line)
			continue
		}
		if formatLen < 0 {
			formatLen, formatOK = ssa.getFormatLen(idx)
			if !formatOK {
				break
			}
		}

		dialogue := getSliptedDialogue(line, formatLen, idx, removeCC)

		if opts.RemoveCC {
			dialogue.text = Validator.removeCC(dialogue.text)
			ssa.originals[idx] = dialogue.meta + "," + dialogue.text
		}
		ssa.appendDialogue(dialogue.meta, dialogue.text)
		updateTraslatable(dialogue)
	}

	// adds the last translatableLine that for loop could not catch
	ssa.translatables = append(ssa.translatables, translatableLine)

	// Everything is fine. Just return nil (no errors)
	if formatOK {
		return nil
	}

	// could not guess the ssa format (probably not a ssa file)
	if formatLen > 0 && !formatOK {
		return fmt.Errorf("could find or guess .ssa Dialogue format. This might not be an actual .ssa file")
	} else {
		// try to guess and returns it self with the best formatLen
		formatLen = ssa.guessFormatLen()
		return ssa.extractDialogues(fileAsStrArr, removeCC, formatLen)
	}
}

func (ssa *ssaTranslate) getFormatLen(idx int) (formatLen int, ok bool) {
	prevIdx := idx - 1
	if prevIdx < 0 {
		return 0, false
	}
	// eg: Format: Layer, Start, End, Style, Actor, MarginL, MarginR, MarginV, Effect, Text
	isFormatLine := constCompare(ssa.originals[prevIdx], ssaParserFormatStr)
	if !isFormatLine {
		return 0, false
	}

	formats := strings.Split(ssa.originals[prevIdx], ",")
	formatLen = len(formats)
	if formatLen == 0 {
		return 0, false
	}

	return formatLen, true
}

func (ssa *ssaTranslate) guessFormatLen() int {
	if len(ssa.originals) == 0 {
		return 0
	}
	bestGuess := math.MaxInt
	for _, line := range ssa.originals {
		isDialogue := constCompare(line, ssaParserDialogueStr)
		if !isDialogue {
			continue
		}
		splitedLen := len(strings.Split(line, ","))
		if splitedLen < bestGuess {
			bestGuess = splitedLen
		}
	}
	return bestGuess
}

func (ssa *ssaTranslate) appendDialogue(meta, text string) {
	ssa.dialogues.texts = append(ssa.dialogues.texts, text)
	ssa.dialogues.spliteds = append(ssa.dialogues.spliteds, sliptedDialogue{meta, text})
}

func getSliptedDialogue(line string, formatLen, lnIdx int, removeCC bool) sliptedDialogue {
	// Splits a dialogue line by commas
	splited := strings.Split(line, ",")
	if len(splited) < formatLen {
		return sliptedDialogue{meta: line, text: ""}
	}
	dialogueMeta := strings.Join(splited[:formatLen-1], ",")
	line = strings.Join(splited[formatLen-1:], ",")
	line = normalizeSpeechLine(line, lnIdx, removeCC)
	return sliptedDialogue{meta: dialogueMeta, text: line}
}

func normalizeSpeechLine(line string, index int, removeCC bool) string {
	if Validator.isMusic(line) {
		return line
	}
	if Validator.isTranslatable(line, removeCC) {
		line = strings.ReplaceAll(line, "\\N", LN_SEP)
		line = strings.ReplaceAll(line, ";", ",")
		line = fmt.Sprintf("%d;%s", index, line)
	} else {
		line = ""
	}
	return strings.TrimSpace(line)
}

func constCompare(text, constStr string) bool {
	text = strings.TrimSpace(strings.ToLower(text))
	return strings.HasPrefix(text, constStr)
}

func (ssa *ssaTranslate) mergeTranslatedToOriginal(transDialogues []string) {

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

	ssa.translateds = ssa.originals

	for _, dialogueLine := range transDialogues {
		dialogues := strings.Split(dialogueLine, LN_BREAK)
		if len(dialogues) == 0 {
			continue
		}
		for _, dialogue := range dialogues {
			idx, text := getLineText(dialogue)
			if idx == -1 {
				continue
			}

			transText := strings.ReplaceAll(text, LN_SEP, "\\N")
			dialogueMeta := ssa.dialogues.spliteds[idx].meta + ","
			ssa.translateds[idx] = dialogueMeta + transText
		}
	}
}
