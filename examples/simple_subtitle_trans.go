package main

import (
	"log"

	transub "github.com/lcapuano-app/go-translate-subtitle-file"
)

func main() {
	opts := transub.Config{
		RemoveCloseCaption: true,
		LanguageSrc:        "en",
		LanguageDest:       "pt",
		OutputDir:          "",
	}
	ts := transub.TransubNew("subtitle.srt", opts)
	if err := ts.Translaste(); err != nil {
		log.Fatal(err)
	}

}
