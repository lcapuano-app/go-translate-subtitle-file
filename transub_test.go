package transub

import (
	"testing"
)

func TestTransub_Translate(t *testing.T) {
	opts := Config{
		RemoveCloseCaption: true,
		LanguageSrc:        "en",
		LanguageDest:       "pt",
		OutputDir:          "",
	}
	ts := TransubNew("path/to/subtitle.srt", opts)
	if err := ts.Translaste(); err != nil {
		t.Fatal(err)
	}

}
