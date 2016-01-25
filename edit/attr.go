package edit

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

var attrForType = map[TokenType]string{
	ParserError:  "31;3",
	Bareword:     "",
	SingleQuoted: "33",
	DoubleQuoted: "33",
	Variable:     "35",
	Sep:          "",
	/*
		parse.ItemSpace:             "36", // only applies to comments
		parse.ItemRedirLeader:       "32",
		parse.ItemStatusRedirLeader: "32",
		parse.ItemPipe:              "32",
		parse.ItemError:             "31",
		parse.ItemQuestionLParen:    "34;1",
		parse.ItemLParen:            "34;1",
		parse.ItemRParen:            "34;1",
		parse.ItemLBracket:          "34;1",
		parse.ItemRBracket:          "34;1",
		parse.ItemLBrace:            "34;1",
		parse.ItemRBrace:            "34;1",
		parse.ItemAmpersand:         "1",
		parse.ItemDollar:            "35",
	*/
}
