package logan

import "strings"

var enStopWords = map[string]bool{
	"-":   true,
	";":   true,
	",":   true,
	".":   true,
	"#":   true,
	"$":   true,
	"%":   true,
	"&":   true,
	"'":   true,
	"(":   true,
	")":   true,
	"=":   true,
	"~":   true,
	"^":   true,
	"|":   true,
	"{":   true,
	"}":   true,
	":":   true,
	"+":   true,
	"<":   true,
	">":   true,
	"[":   true,
	"]":   true,
	"a":   true,
	"an":  true,
	"the": true,
}

func getDelimReplacer(separators string) *strings.Replacer {
	replacements := make([]string, 0, len(separators)*2)
	for _, char := range separators {
		replacements = append(replacements, string(char), " ")
	}
	return strings.NewReplacer(replacements...)
}
