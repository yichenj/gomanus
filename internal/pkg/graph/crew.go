package graph

import (
	"github.com/cloudwego/eino/components/tool"

	"github.com/yichenj/gomanus/internal/pkg/llm"
	pt "github.com/yichenj/gomanus/internal/pkg/prompt"
	mtool "github.com/yichenj/gomanus/pkg/tool"
)

// crew defines the team members and their prompts.
var crew = map[string]MemberDesc{
	"researcher": {Prompt: pt.ResearcherPrompt, Tools: []tool.BaseTool{
		mtool.EnsureSearchTool(),
		mtool.EnsureCrawlTool(),
	}, ModelType: llm.ModelTypeLanguageModel},
	"reporter": {Prompt: pt.ReporterPrompt, ModelType: llm.ModelTypeMultiModalModel},
}

type MemberDesc struct {
	Prompt    string
	Tools     []tool.BaseTool
	ModelType llm.ModelType
}
