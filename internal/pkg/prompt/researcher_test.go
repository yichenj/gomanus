package prompt

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/yichenj/gomanus/internal/pkg/llm"
)

const (
	basicModelName = "doubao-1-5-pro-32k-250115"

	userInput = "Calculate the influence index of DeepSeek R1 on HuggingFace. " +
		"This index can be designed by considering a weighted sum of factors such as followers, downloads, and likes\n"

	plannerInput = `
{
  "thought": "I need to calculate the influence index of DeepSeek R1 on HuggingFace by designing a weighted formula that incorporates factors like followers, downloads, and likes. This requires gathering data, determining weight values, and performing calculations.",
  "title": "DeepSeek R1 Influence Index Analysis",
  "steps": [
    {
      "agent_name": "researcher",
      "title": "Data Collection",
      "description": "Retrieve quantitative metrics for DeepSeek R1 from HuggingFace: 1. Number of followers 2. Total downloads 3. Likes/upvotes. Compare with benchmark values from similar models to establish context.",
      "note": "Verify data freshness and source reliability"
    },
    {
      "agent_name": "researcher",
      "title": "Weight Determination",
      "description": "Research common weighting practices for influence metrics in AI communities. Propose weight distribution (e.g., followers: 40%, downloads: 50%, likes: 10%) based on platform engagement patterns."
    },
    {
      "agent_name": "researcher",
      "title": "Index Calculation",
      "description": "Compute influence index using formula: Influence Index = (Followers × W₁) + (Downloads × W₂) + (Likes × W₃). Normalize values if required before applying weights."
    },
    {
      "agent_name": "reporter",
      "title": "Final Analysis Report",
      "description": "Present formatted results with: 1. Raw metric data 2. Weight justification 3. Calculated index 4. Comparative analysis 5. Limitations (e.g., temporal variations, platform-specific biases)"
    }
  ]
}
	`
)

func TestResearcher(t *testing.T) {
	Convey("Testing Researcher Prompt", t, func() {
		ctx := context.Background()

		systemPromptTplt := prompt.FromMessages(schema.GoTemplate,
			&schema.Message{
				Role:    schema.System,
				Content: ResearcherPrompt,
			},
		)
		sysPrompts, err := systemPromptTplt.Format(ctx, map[string]any{
			"CURRENT_TIME": time.Now().Format(time.RFC3339),
		})
		So(err, ShouldBeNil)

		prompts := make([]*schema.Message, 0)
		prompts = append(prompts, sysPrompts...)
		prompts = append(prompts, []*schema.Message{
			{
				Role:    schema.User,
				Content: userInput,
			},
			{
				Role:    schema.User,
				Content: plannerInput,
			},
		}...)

		model, err := llm.NewArkModel(ctx, llm.ModelSetting{
			ModelName: basicModelName,
			APIKey:    os.Getenv("API_KEY"),
		})
		So(err, ShouldBeNil)

		Convey("basic language model generate", func() {
			resp, err := model.Generate(ctx, prompts)
			So(err, ShouldBeNil)
			So(resp.Content, ShouldNotBeBlank)
			fmt.Println(resp.Content)
		})

		Convey("basic language model stream", func() {
			streamResult, err := model.Stream(ctx, prompts)
			So(err, ShouldBeNil)
			defer streamResult.Close()
			for {
				chunk, err := streamResult.Recv()
				if errors.Is(err, io.EOF) {
					break
				}
				So(err, ShouldBeNil)
				fmt.Print(chunk.Content)
			}
		})
	})
}
