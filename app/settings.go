package app

type Settings struct {
	InputFile  *string
	OutputFile *string
	Prompt     *string
	ApiKey     *string
	ApiURL     string
}

func NewSettings() *Settings {
	settings := Settings{}
	settings.ApiURL = "https://api.openai.com/v1/chat/completions"
	return &settings
}
