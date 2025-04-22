package tool

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/cloudwego/eino-ext/components/tool/duckduckgo"
	"github.com/cloudwego/eino/components/tool"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSearchTool(t *testing.T) {
	Convey("Testing SearchTool", t, func() {
		Convey("basic query", func() {
			searchTool, ok := EnsureSearchTool().(tool.InvokableTool)
			So(ok, ShouldBeTrue)

			ctx := context.Background()
			info, err := searchTool.Info(ctx)
			So(err, ShouldBeNil)
			So(info.Name, ShouldEqual, "search_tool")

			searchReq := &duckduckgo.SearchRequest{
				Query: "Golang programming development",
				Page:  1,
			}
			query, _ := json.Marshal(searchReq)
			resp, err := searchTool.InvokableRun(ctx, string(query))
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
		})
	})
}
