package logan

import "strings"

var enStopWords = map[string]string{
	"-":   "",
	";":   "",
	",":   "",
	".":   "",
	"#":   "",
	"$":   "",
	"%":   "",
	"&":   "",
	"'":   "",
	"(":   "",
	")":   "",
	"=":   "",
	"~":   "",
	"^":   "",
	"|":   "",
	"{":   "",
	"}":   "",
	":":   "",
	"+":   "",
	"<":   "",
	">":   "",
	"[":   "",
	"]":   "",
	"a":   "",
	"an":  "",
	"the": "",
}

func getDelimReplacer(separators string) *strings.Replacer {
	replacements := make([]string, 0, len(separators)*2)
	for _, char := range separators {
		replacements = append(replacements, string(char), " ")
	}
	return strings.NewReplacer(replacements...)
}
