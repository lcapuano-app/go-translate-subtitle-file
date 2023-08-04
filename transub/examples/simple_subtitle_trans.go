package main

import (
	"log"

	"github.com/lcapuano-app/go-translate-subtitle-file/transub"
)

func main() {
	filename := "subtitle.srt"
	destLanguage := "pt"
	ts := transub.New(
		filename,
		destLanguage,
		transub.WithLanguageSrc("en"),
		transub.WithRemoveCC(true),
		transub.WithMainSub(true),
		transub.WithRemoveOrigin(false),
		transub.WithGoogleRetries(3),
		//WithOutputDir("another/dir/to/output/file"),
	)

	err := ts.TranslasteSRT()
	if err != nil {
		log.Fatal(err)
	}

}
