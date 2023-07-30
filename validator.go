package transub

import (
	"os"
	"strconv"
	"strings"
)

func isValidFile(filename, ext string) bool {
	if !strings.HasSuffix(filename, ext) {
		return false
	}
	if _, err := os.Stat(filename); err != nil {
		return false
	}
	return true
}

func isSkip(line string) bool {
	if len(line) == 0 || line == "\n" || line == "\r" || line == "\r\n" {
		return true
	}
	return false
}

func isCC(line string) bool {
	if strings.Contains(line, "[") ||
		strings.Contains(line, "]") ||
		strings.HasPrefix(line, "♪") {
		return true
	}
	return false
}

func isTimestamp(line string) bool {
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

func isTranslatable(line string, removeCloseCaption bool) bool {
	notTextLine := isSkip(line) || isTimestamp(line) || isIntLine(line)
	shouldTranslate := !notTextLine

	if removeCloseCaption {
		//fmt.Println("-------", line, isCC(line), shouldTranslate, shouldTranslate && !isCC(line))
		return shouldTranslate && !isCC(line)
	}
	return shouldTranslate
}

func isIntLine(line string) bool {
	_, err := strconv.Atoi(line)
	if err != nil {
		return false
	}
	return true
}
