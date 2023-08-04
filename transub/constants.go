package transub

import "os"

const (
	LN_BREAK                      = "\n"
	LN_SEP                        = " /// "
	META_TRASNLATED               = "meta=translated"
	gtransCharLimit               = 5_000
	fileEditFlag                  = os.O_APPEND | os.O_CREATE | os.O_WRONLY
	textPlainMIME                 = "text/plain"
	ssaParserEvtsStr              = "[events]"
	ssaParserFormatStr            = "format:"
	ssaParserDialogueStr          = "dialogue:"
	ssaParserUnknFormat  formatOK = -1
)
