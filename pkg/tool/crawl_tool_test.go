package tool

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/cloudwego/eino/components/tool"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCrawlTool(t *testing.T) {
	Convey("Testing CrawlTool", t, func() {
		Convey("basic query", func() {
			crawlTool, ok := EnsureCrawlTool().(tool.InvokableTool)
			So(ok, ShouldBeTrue)

			ctx := context.Background()
			info, err := crawlTool.Info(ctx)
			So(err, ShouldBeNil)
			So(info.Name, ShouldEqual, "crawl_tool")

			crawURL := &CrawlRequest{
				URL: "https://go.dev/",
			}
			query, _ := json.Marshal(crawURL)
			resp, err := crawlTool.InvokableRun(ctx, string(query))
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
		})
	})
}
