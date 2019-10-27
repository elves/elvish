package highlight

var transformerFor = map[string]string{
	barewordRegion:     "",
	singleQuotedRegion: "yellow",
	doubleQuotedRegion: "yellow",
	variableRegion:     "magenta",
	wildcardRegion:     "",
	tildeRegion:        "",

	commentRegion: "cyan",

	">":  "green",
	">>": "green",
	"<":  "green",
	"?>": "green",
	"|":  "green",
	"?(": "bold",
	"(":  "bold",
	")":  "bold",
	"[":  "bold",
	"]":  "bold",
	"{":  "bold",
	"}":  "bold",
	"&":  "bold",

	commandRegion: "green",
	keywordRegion: "yellow",
	errorRegion:   "bg-red",
}

var (
	transformerForGoodCommand = "green"
	transformerForBadCommand  = "red"
)
