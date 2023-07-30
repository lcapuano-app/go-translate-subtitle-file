
# Translate Subtitle File

  

A simple subtitle translator using [my own fork](https://github.com/lcapuano-app/go-googletrans) of [go-googletrans](https://github.com/Conight/go-googletrans).

  

## Quick Start Example

  
#### Input


```text
// some_subtitle_file_example.srt
3

00:00:16,500 --> 00:00:19,500

[tense jazzy music]

  

4

00:00:19,500 --> 00:00:26,542

♪ ♪

  

5

00:01:48,083 --> 00:01:50,792

- Hello world!

```

  

### Simple translate

```go

package main

  

import (
	"log"

	transub "github.com/lcapuano-app/go-translate-subtitle-file"
)

  

func  main() {
	opts := transub.Config{
		RemoveCloseCaption: true,
		LanguageSrc: "en",
		LanguageDest: "pt",
		OutputDir: "",
	}

	ts := transub.TransubNew("some_subtitle_file_example.srt", opts)

	if  err := ts.Translaste(); err != nil {
		log.Fatal(err)
	}
}

```

  

#### Output

```text
// some_subtitle_file_example.pt.srt

3

00:00:16,500 --> 00:00:19,500

  

4

00:00:19,500 --> 00:00:26,542

  

5

00:01:48,083 --> 00:01:50,792

- Olá mundo!

```