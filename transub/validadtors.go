package transub

import (
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type validate struct {
}

var Validator validate

func (v validate) isTextFile(filename string) (ok bool, err error) {
	fBytes, err := os.ReadFile(filename)
	if err != nil {
		return false, err
	}
	cType := http.DetectContentType(fBytes)
	return strings.HasPrefix(cType, textPlainMIME), nil
}

func (v validate) isReachableFile(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func (v validate) isLineBreak(line string) bool {
	if len(line) == 0 || line == "\n" || line == "\r" || line == "\r\n" {
		return true
	}
	return false
}

func (v validate) isCC(line string) bool {
	return strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]")
}

func (v validate) removeCC(line string) string {
	if !v.isCC(line) {
		return line
	}
	re := regexp.MustCompile(`(?m)(\[.*?\])`)
	line = re.ReplaceAllString(line, "")
	return strings.TrimSpace(line)
}

func (v validate) isMusic(line string) bool {
	return strings.HasPrefix(line, "♪")
}

func (v validate) isTimestampStr(line string) bool {
	if len(line) < 2 {
		return false
	}
	asciiZero := 48
	asciiNine := 57
	firstChar := int(line[0])
	secndChar := int(line[1])
	if (firstChar >= asciiZero && firstChar <= asciiNine) &&
		(secndChar >= asciiZero && secndChar <= asciiNine) {
		return true
	}
	return false
}

func (v validate) isTranslatable(line string, removeCloseCaption bool) bool {
	notTextLine := v.isLineBreak(line) || v.isTimestampStr(line) || v.isIntStr(line) //|| v.isMusic(line)
	shouldTranslate := !notTextLine

	if removeCloseCaption {
		return shouldTranslate && !v.isCC(line)
	}
	return shouldTranslate
}

// func (v validate) getIntVal(line string) (int, bool) {
// 	val, err := strconv.Atoi(line)
// 	if err != nil {
// 		return val, false
// 	}
// 	return val, true
// }

func (v validate) isIntStr(line string) bool {
	_, err := strconv.Atoi(line)

	return err == nil
}

func (v validate) isTranslatedFilename(path, lang string) bool {
	filename := filepath.Base(path)
	ext := filepath.Ext(path)
	transSuffix := "." + lang + ext
	return strings.HasSuffix(filename, transSuffix)
}
