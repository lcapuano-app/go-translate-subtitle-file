package transub

import (
	"testing"
)

func TestTransub_Translate(t *testing.T) {

	filename := "examples/subtitle.srt"
	destLanguage := "portuguese"
	tr := New(
		filename,
		destLanguage,
		WithLanguageSrc("en"),
		WithRemoveCC(true),
		WithMainSub(true),
		WithRemoveOrigin(false),
		WithGoogleRetries(3),
		//WithOutputDir("another/dir/to/output/file"),
	)

	err := tr.TranslasteSRT()
	if err != nil {
		t.Fatal(err)
	}

}
