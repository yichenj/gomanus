package llm

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/components/model"
)

func NewArkModel(ctx context.Context, config ModelSetting) (model.ChatModel, error) {
	m, err := ark.NewChatModel(ctx, &ark.ChatModelConfig{
		Model:  config.ModelName,
		APIKey: config.APIKey,
	})
	if err != nil {
		return nil, fmt.Errorf(
			"create ark chat model(%+v) err(%w)", config, err)
	}

	return m, nil
}
