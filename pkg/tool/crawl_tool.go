package tool

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type JinaClient struct {
	client  *http.Client
	baseURL string
}

type CrawlRequest struct {
	URL string `json:"url" jsonschema_description:"The url to crawl"`
}

func NewJinaClient() *JinaClient {
	return &JinaClient{
		client:  &http.Client{},
		baseURL: "https://r.jina.ai/",
	}
}

func (j *JinaClient) Crawl(_ context.Context, cr *CrawlRequest) (output string, err error) {
	accessURL := j.baseURL + cr.URL
	req, err := http.NewRequest(http.MethodGet, accessURL, nil)
	if err != nil {
		return "", fmt.Errorf("new request(GET:%s) err(%w)", accessURL, err)
	}

	resp, err := j.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("do request(GET:%s) err(%w)", accessURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("o response(GET:%s) with http status code(%d)", accessURL, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response(GET:%s) body err(%w)", accessURL, err)
	}
	return string(body), nil
}

func EnsureCrawlTool() tool.BaseTool {
	c := NewJinaClient()
	t, err := utils.InferTool("crawl_tool", "a web content crawl tool powered by jina.ai", c.Crawl)
	if err != nil {
		panic(fmt.Errorf("NewTool of Jina failed, err(%w)", err))
	}
	return t
}
