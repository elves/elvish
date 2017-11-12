package edit

import "github.com/elves/elvish/edit/prompt"

var (
	_ = RegisterVariable("prompt", prompt.PromptVariable)
	_ = RegisterVariable("rprompt", prompt.RpromptVariable)
	_ = RegisterVariable("rprompt-persistent", prompt.RpromptPersistentVariable)
	_ = RegisterVariable("-prompts-max-wait", prompt.MaxWaitVariable)
)
