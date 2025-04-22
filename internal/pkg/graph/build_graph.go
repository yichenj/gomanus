package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"

	"github.com/yichenj/gomanus/internal/pkg/llm"
	pt "github.com/yichenj/gomanus/internal/pkg/prompt"
)

var modelBuilder = llm.NewArkModel

func BuildGraph(ctx context.Context, modelSetting llm.TypedModelSettings) (
	*compose.Graph[[]*schema.Message, *schema.Message], compose.Runnable[[]*schema.Message, *schema.Message], error) {
	g := compose.NewGraph[[]*schema.Message, *schema.Message](
		compose.WithGenLocalState(NewMessageState))

	err := addCoordinator(ctx, g, modelSetting[llm.ModelTypeLanguageModel])
	if err != nil {
		return nil, nil, err
	}
	err = addPlanner(ctx, g, modelSetting[llm.ModelTypeReasonModel])
	if err != nil {
		return nil, nil, err
	}
	err = addSupervisor(ctx, g, modelSetting[llm.ModelTypeLanguageModel])
	if err != nil {
		return nil, nil, err
	}
	for name, memberDesc := range crew {
		err = addTeamMember(ctx, g, name, memberDesc, modelSetting)
		if err != nil {
			return nil, nil, err
		}
	}
	err = addCoordinatorEdges(g)
	if err != nil {
		return nil, nil, err
	}
	err = addPlannerEdges(g)
	if err != nil {
		return nil, nil, err
	}
	err = addSupervisorEdges(g)
	if err != nil {
		return nil, nil, err
	}

	r, err := g.Compile(ctx, compose.WithMaxRunSteps(30))
	if err != nil {
		return nil, nil, fmt.Errorf("compile graph err(%w)", err)
	}

	return g, r, nil
}

func addCoordinator(ctx context.Context,
	g *compose.Graph[[]*schema.Message, *schema.Message], lm llm.ModelSetting) error {
	template := prompt.FromMessages(schema.GoTemplate,
		&schema.Message{
			Role:    schema.System,
			Content: pt.CoordinatorPrompt,
		},
	)
	coordinator, err := modelBuilder(ctx, lm)
	if err != nil {
		return fmt.Errorf("create coordinator model err(%w)", err)
	}
	err = g.AddChatModelNode("coordinator", coordinator, compose.WithStatePreHandler(
		func(ctx context.Context, in []*schema.Message, state *MessageState) (out []*schema.Message, err error) {
			state.messages = append(state.messages, in...)

			p, err := template.Format(ctx, map[string]any{
				"CURRENT_TIME": time.Now().Format(time.RFC3339),
			})
			if err != nil {
				return nil, fmt.Errorf("format prompt err(%w)", err)
			}
			msg := make([]*schema.Message, 0, len(p)+len(state.messages))
			msg = append(msg, p...)
			msg = append(msg, state.messages...)
			return msg, nil
		}),
		compose.WithNodeName("coordinator"),
	)
	if err != nil {
		return fmt.Errorf("add coordinator node err(%w)", err)
	}

	return nil
}

func addCoordinatorEdges(g *compose.Graph[[]*schema.Message, *schema.Message]) error {
	err := g.AddBranch("coordinator",
		compose.NewGraphBranch(func(ctx context.Context, in *schema.Message) (endNode string, err error) {
			if strings.Contains(in.Content, "handoff_to_planner()") {
				return "planner_input", nil
			}
			return compose.END, nil
		}, map[string]bool{
			"planner_input": true,
			compose.END:     true,
		}),
	)
	if err != nil {
		return fmt.Errorf("add coordinator post branch err(%w)", err)
	}
	err = g.AddEdge(compose.START, "coordinator")
	if err != nil {
		return fmt.Errorf("add START to coordinator edge err(%w)", err)
	}

	return nil
}

func addPlanner(ctx context.Context,
	g *compose.Graph[[]*schema.Message, *schema.Message], lm llm.ModelSetting) error {
	template := prompt.FromMessages(schema.GoTemplate,
		&schema.Message{
			Role:    schema.System,
			Content: pt.PlannerPrompt,
		},
	)
	teamMemberNames := make([]string, 0, len(crew))
	for each := range crew {
		teamMemberNames = append(teamMemberNames, each)
	}

	planner, err := modelBuilder(ctx, lm)
	if err != nil {
		return fmt.Errorf("create planner model err(%w)", err)
	}

	err = g.AddLambdaNode("planner_input", compose.ToList[*schema.Message]())
	if err != nil {
		return fmt.Errorf("add planner input node err(%w)", err)
	}
	err = g.AddChatModelNode("planner", planner, compose.WithStatePreHandler(
		func(ctx context.Context, in []*schema.Message, state *MessageState) (out []*schema.Message, err error) {
			p, err := template.Format(ctx, map[string]any{
				"CURRENT_TIME": time.Now().Format(time.RFC3339),
				"TEAM_MEMBERS": strings.Join(teamMemberNames, ","),
				"JSON_PREFIX":  "```json",
			})
			if err != nil {
				return nil, fmt.Errorf("format prompt err(%w)", err)
			}

			msg := make([]*schema.Message, 0, len(p)+len(state.messages))
			msg = append(msg, p...)
			msg = append(msg, state.messages...)
			return msg, nil
		}),
		compose.WithStatePostHandler(func(ctx context.Context, out *schema.Message, state *MessageState) (*schema.Message, error) {
			out.Role = schema.User
			out.Content = fmt.Sprintf(
				"Full plan is:\n\n<plan>\n%s\n</plan>\n\n*Please execute step by step.*",
				out.Content)
			out.Name = "planner"
			state.messages = append(state.messages, out)
			return out, nil
		}),
		compose.WithNodeName("planner"),
	)
	if err != nil {
		return fmt.Errorf("add planner node err(%w)", err)
	}
	return nil
}

func addPlannerEdges(g *compose.Graph[[]*schema.Message, *schema.Message]) error {
	err := g.AddEdge("planner_input", "planner")
	if err != nil {
		return fmt.Errorf("add planner input edge err(%w)", err)
	}
	err = g.AddEdge("planner", "supervisor_input")
	if err != nil {
		return fmt.Errorf("add planner to supvervisor edge err(%w)", err)
	}
	return nil
}

func addSupervisor(ctx context.Context,
	g *compose.Graph[[]*schema.Message, *schema.Message], lm llm.ModelSetting) error {
	template := prompt.FromMessages(schema.GoTemplate,
		&schema.Message{
			Role:    schema.System,
			Content: pt.SupervisorPrompt,
		},
	)

	teamMemberNames := make([]string, 0, len(crew))
	for each := range crew {
		teamMemberNames = append(teamMemberNames, each)
	}

	supervisor, err := modelBuilder(ctx, lm)
	if err != nil {
		return fmt.Errorf("create supervisor model err(%w)", err)
	}

	err = g.AddLambdaNode("supervisor_input", compose.ToList[*schema.Message]())
	if err != nil {
		return fmt.Errorf("add supervisor input node err(%w)", err)
	}
	err = g.AddChatModelNode("supervisor", supervisor, compose.WithStatePreHandler(
		func(ctx context.Context, in []*schema.Message, state *MessageState) (out []*schema.Message, err error) {
			p, err := template.Format(ctx, map[string]any{
				"CURRENT_TIME": time.Now().Format(time.RFC3339),
				"TEAM_MEMBERS": strings.Join(teamMemberNames, ","),
				"JSON_PREFIX":  "```json",
			})
			if err != nil {
				return nil, fmt.Errorf("format prompt err(%w)", err)
			}

			for _, each := range in {
				if _, exist := crew[each.Name]; !exist {
					continue
				}
				each.Role = schema.User
				each.Content = fmt.Sprintf(
					"Response from %s:\n\n<response>\n%s\n</response>\n\n*Please execute the next step.*",
					each.Name, each.Content)
				state.messages = append(state.messages, each)
			}

			msg := make([]*schema.Message, 0, len(p)+len(state.messages))
			msg = append(msg, p...)
			msg = append(msg, state.messages...)
			return msg, nil
		}), compose.WithStatePostHandler(
		func(ctx context.Context, in *schema.Message, state *MessageState) (out *schema.Message, err error) {
			in.Content = strings.ReplaceAll(in.Content, "```json", "")
			in.Content = strings.ReplaceAll(in.Content, "```", "")

			if strings.Contains(in.Content, "\"FINISH\"") {
				s := strings.LastIndex(state.messages[len(state.messages)-1].Content, "<response>") + len("<response>")
				if s == -1 {
					s = 0
				}
				t := strings.Index(state.messages[len(state.messages)-1].Content, "</response>")
				if t == -1 {
					t = len(state.messages[len(state.messages)-1].Content)
				}
				return &schema.Message{Role: schema.Assistant,
					Content: state.messages[len(state.messages)-1].Content[s:t]}, nil
			}
			return in, nil
		}),
		compose.WithNodeName("supervisor"),
	)
	if err != nil {
		return fmt.Errorf("add supervisor node err(%w)", err)
	}
	return nil
}

func addSupervisorEdges(g *compose.Graph[[]*schema.Message, *schema.Message]) error {
	nextNodeSet := map[string]bool{
		compose.END: true,
	}
	for each := range crew {
		nextNodeSet[each+"_input"] = true
	}

	err := g.AddEdge("supervisor_input", "supervisor")
	if err != nil {
		return fmt.Errorf("add supervisor input edge err(%w)", err)
	}

	err = g.AddBranch("supervisor",
		compose.NewGraphBranch(func(ctx context.Context, in *schema.Message) (endNode string, err error) {
			o := &struct {
				Next string `json:"next"`
			}{}
			err = json.Unmarshal([]byte(in.Content), &o)
			if err != nil {
				return compose.END, nil
			}
			if _, exist := nextNodeSet[o.Next+"_input"]; !exist {
				return compose.END, nil
			}
			return o.Next + "_input", nil
		}, nextNodeSet),
	)
	if err != nil {
		return fmt.Errorf("add supervisor post branch err(%w)", err)
	}
	return nil
}

func addTeamMember(ctx context.Context,
	g *compose.Graph[[]*schema.Message, *schema.Message], name string, desc MemberDesc,
	modelSetting llm.TypedModelSettings) error {
	template := prompt.FromMessages(schema.GoTemplate,
		&schema.Message{
			Role:    schema.System,
			Content: desc.Prompt,
		},
	)
	model, err := modelBuilder(ctx, modelSetting[desc.ModelType])
	if err != nil {
		return fmt.Errorf("create team member(%s) model err(%w)", name, err)
	}

	err = g.AddLambdaNode(name+"_input", compose.ToList[*schema.Message]())
	if err != nil {
		return fmt.Errorf("add team member(%s) input node err(%w)", name, err)
	}
	stateHandlers := []compose.GraphAddNodeOpt{
		compose.WithStatePreHandler(
			func(ctx context.Context, in []*schema.Message, state *MessageState) (out []*schema.Message, err error) {
				p, err := template.Format(ctx, map[string]any{
					"CURRENT_TIME": time.Now().Format(time.RFC3339),
				})
				if err != nil {
					return nil, fmt.Errorf("format prompt err(%w)", err)
				}

				msg := make([]*schema.Message, 0, len(p)+len(state.messages))
				msg = append(msg, p...)
				msg = append(msg, state.messages...)
				return msg, nil
			}),
		compose.WithStatePostHandler(
			func(ctx context.Context, in *schema.Message, state *MessageState) (out *schema.Message, err error) {
				in.Name = name
				return in, nil
			}),
		compose.WithNodeName(name),
	}
	if len(desc.Tools) > 0 {
		reactAgent, err := react.NewAgent(ctx, &react.AgentConfig{
			Model:       model,
			ToolsConfig: compose.ToolsNodeConfig{Tools: desc.Tools},
		})
		if err != nil {
			return fmt.Errorf("create react agent for team member(%s) err(%w)", name, err)
		}

		reactAgentGraph, opts := reactAgent.ExportGraph()
		err = g.AddGraphNode(name, reactAgentGraph, append(opts, stateHandlers...)...)
		if err != nil {
			return fmt.Errorf("add react agent graph for team member(%s) node err(%w)", name, err)
		}
	} else {
		err = g.AddChatModelNode(name, model, stateHandlers...)
		if err != nil {
			return fmt.Errorf("add team member(%s) node err(%w)", name, err)
		}
	}

	err = g.AddEdge(name+"_input", name)
	if err != nil {
		return fmt.Errorf("add team member(%s) input edge err(%w)", name, err)
	}
	err = g.AddEdge(name, "supervisor_input")
	if err != nil {
		return fmt.Errorf("add supervisor edge for team member(%s) err(%w)", name, err)
	}

	return nil
}
