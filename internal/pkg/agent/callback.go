package agent

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent"
	"github.com/cloudwego/eino/schema"
)

type Callbacks struct {
	io func(input string)
}

func (c *Callbacks) OnStart(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
	return ctx
}

func (c *Callbacks) OnEnd(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
	if info.Component == "ChatModel" {
		msg := output.(*model.CallbackOutput).Message
		c.io(fmt.Sprintf("--%s: %s", info.Name, msg.Content))
	}
	return ctx
}

func (c *Callbacks) OnError(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
	c.io(fmt.Sprintf("OnError: info(%+v), err(%s)", info, err.Error()))
	return ctx
}

func (c *Callbacks) OnStartWithStreamInput(ctx context.Context, info *callbacks.RunInfo,
	input *schema.StreamReader[callbacks.CallbackInput]) context.Context {
	return ctx
}

func (c *Callbacks) OnEndWithStreamOutput(ctx context.Context, info *callbacks.RunInfo,
	output *schema.StreamReader[callbacks.CallbackOutput]) context.Context {
	return ctx
}

func GetCallbackOpts(ioFunc func(input string)) agent.AgentOption {
	return agent.WithComposeOptions(compose.WithCallbacks(&Callbacks{io: ioFunc}))
}
