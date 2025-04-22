package graph

import (
	"context"

	"github.com/cloudwego/eino/schema"
)

type MessageState struct {
	messages []*schema.Message
}

func NewMessageState(_ context.Context) *MessageState {
	return &MessageState{
		messages: make([]*schema.Message, 0),
	}
}
