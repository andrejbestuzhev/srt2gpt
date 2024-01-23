package main

import (
	"flag"
	"fmt"
	"srt2gpt/m/v2/app"
)

func main() {

	settings := app.NewSettings()
	settings.InputFile = flag.String("input", "", "Input file (srt only)")
	settings.OutputFile = flag.String("output", "", "Output file (srt only)")
	settings.ApiKey = flag.String("key", "", "OpenAI api key")
	settings.Prompt = flag.String("prompt", "", "Translation prompt")
	flag.Parse()

	err := app.CheckFile(settings)
	if err != nil {
		panic(err)
	}

	err, s := app.ParseStrings(settings)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Strings found: %d\n", len(s))
	err, translated := app.CallAPI(settings, &s)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Strings translated: %d\n", len(translated))
}
