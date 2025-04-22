package tool

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/eino-ext/components/tool/duckduckgo"
	"github.com/cloudwego/eino-ext/components/tool/duckduckgo/ddgsearch"
	"github.com/cloudwego/eino/components/tool"
)

func EnsureSearchTool() tool.BaseTool {
	config := &duckduckgo.Config{
		ToolName:   "search_tool",
		MaxResults: 3, // Limit to return 3 results
		Region:     ddgsearch.RegionCN,
		DDGConfig: &ddgsearch.Config{
			Timeout:    30 * time.Second,
			Cache:      true,
			MaxRetries: 5,
		},
	}

	// Create search client
	t, err := duckduckgo.NewTool(context.Background(), config)
	if err != nil {
		panic(fmt.Errorf("NewTool of duckduckgo failed, err(%w)", err))
	}

	return t
}
