package edit

// Styles for UI.
var (
	attrForPrompt            = ""
	attrForRprompt           = "7"
	attrForCompleted         = ";4"
	attrForMode              = "1;3;35"
	attrForTip               = ""
	attrForCurrentCompletion = ";7"
	attrForCompletedHistory  = "4"
	attrForSelectedFile      = ";7"
)

// Styles for semantic coloring.
var (
	styleForBadCommand  = "31;3"
	styleForBadVariable = "31;3"
)

var attrForType = map[TokenType]string{
	ParserError:  "31;3",
	Bareword:     "",
	SingleQuoted: "33",
	DoubleQuoted: "33",
	Variable:     "35",
	Sep:          "",
}

var styleForSep = map[byte]string{
	// unknown : "31",
	'#': "36",

	'>': "32",
	'<': "32",
	'?': "32", // applies to both ?( and ?>
	// "?(": "34;1",
	'|': "32",

	'(': "1",
	')': "1",
	'[': "1",
	']': "1",
	'{': "1",
	'}': "1",

	'&': "1",
}
