package adk

import "context"

// Agent runs iterative model-tool loops in a single, self-contained runtime.
type Agent struct {
	model         Model
	tools         *ToolRegistry
	systemPrompt  string
	maxIterations int
}

func NewAgent(model Model, tools *ToolRegistry, systemPrompt string, maxIterations int) *Agent {
	if maxIterations <= 0 {
		maxIterations = 8
	}
	return &Agent{model: model, tools: tools, systemPrompt: systemPrompt, maxIterations: maxIterations}
}

func (a *Agent) RunTurn(ctx context.Context, history []Message, userInput string) (Message, []ToolResult, error) {
	state := SessionState{
		SystemPrompt: a.systemPrompt,
		Messages:     append(append([]Message{}, history...), Message{Role: "user", Content: userInput}),
	}
	allResults := make([]ToolResult, 0)

	for i := 0; i < a.maxIterations; i++ {
		resp, err := a.model.Respond(ctx, state)
		if err != nil {
			return Message{}, allResults, err
		}
		if len(resp.ToolCalls) == 0 {
			return Message{Role: "assistant", Content: resp.AssistantMessage}, allResults, nil
		}

		results := make([]ToolResult, 0, len(resp.ToolCalls))
		for _, tc := range resp.ToolCalls {
			r := a.tools.Execute(ctx, tc)
			results = append(results, r)
			allResults = append(allResults, r)
		}
		state.LastToolRuns = results
	}

	return Message{Role: "assistant", Content: "Stopped after max iterations."}, allResults, nil
}
