package llm

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
)

func NewOpenAIModel(ctx context.Context, config ModelSetting) (model.ChatModel, error) {
	m, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		Model:   config.ModelName,
		BaseURL: config.BaseURL,
		APIKey:  config.APIKey,
	})
	if err != nil {
		return nil, fmt.Errorf(
			"create openai chat model(%+v) err(%w)", config, err)
	}
	return m, nil
}
