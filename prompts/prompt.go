package prompts

type Prompt struct {
	PathToPrompt string
	FollowUps    []string
}

var PromptRegistry = map[string]Prompt{
	"Behavioral": Prompt{
		PathToPrompt: "interviewer_prompt.md",
	},
	"Technical": Prompt{
		PathToPrompt: "technical_interview_prompt.md",
	},
}
