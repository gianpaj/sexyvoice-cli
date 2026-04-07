package ui

import (
	"strings"
)

const sexyVoiceLogo = "" +
	" ░▒▓███████▓▒░▒▓█▓▒░░▒▓█▓▒░\n" +
	"░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░\n" +
	"░▒▓█▓▒░       ░▒▓█▓▒▒▓█▓▒░ \n" +
	" ░▒▓██████▓▒░ ░▒▓█▓▒▒▓█▓▒░ \n" +
	"       ░▒▓█▓▒░ ░▒▓█▓▓█▓▒░  \n" +
	"       ░▒▓█▓▒░ ░▒▓█▓▓█▓▒░  \n" +
	"░▒▓███████▓▒░   ░▒▓██▓▒░"

const sexyVoiceLogoWidth = 19

func PaddedLogo() string {
	lines := strings.Split(sexyVoiceLogo, "\n")
	pad := "  "

	var b strings.Builder
	b.WriteString("\n")
	for _, line := range lines {
		b.WriteString(pad + line + "\n")
	}
	return b.String()
}

func UnboxedLogoPlain() string {
	return sexyVoiceLogo
}

func UnboxedLogoWidth() int {
	return sexyVoiceLogoWidth
}
