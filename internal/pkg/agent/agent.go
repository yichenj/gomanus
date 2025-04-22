package agent

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent"
	"github.com/cloudwego/eino/schema"

	"github.com/yichenj/gomanus/internal/pkg/graph"
	"github.com/yichenj/gomanus/internal/pkg/llm"
)

type Agent struct {
	graph    *compose.Graph[[]*schema.Message, *schema.Message]
	runnable compose.Runnable[[]*schema.Message, *schema.Message]
}

func NewAgent(ctx context.Context, modelSettings llm.TypedModelSettings) (*Agent, error) {
	g, r, err := graph.BuildGraph(ctx, modelSettings)
	if err != nil {
		return nil, fmt.Errorf("build graph err(%w)", err)
	}
	return &Agent{
		graph:    g,
		runnable: r,
	}, nil
}

// Generate generates a response from the agent.
func (r *Agent) Generate(ctx context.Context, input []*schema.Message, opts ...agent.AgentOption) (*schema.Message, error) {
	return r.runnable.Invoke(ctx, input, agent.GetComposeOptions(opts...)...)
}

// Stream calls the agent and returns a stream response.
func (r *Agent) Stream(ctx context.Context, input []*schema.Message, opts ...agent.AgentOption) (output *schema.StreamReader[*schema.Message], err error) {
	return r.runnable.Stream(ctx, input, agent.GetComposeOptions(opts...)...)
}

// ExportGraph exports the underlying graph from Agent, along with the []compose.GraphAddNodeOpt to be used when adding this graph to another graph.
func (r *Agent) ExportGraph() (compose.AnyGraph, []compose.GraphAddNodeOpt) {
	return r.graph, nil
}
