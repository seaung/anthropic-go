package anthropic

import (
	"context"
	"io"
	"net/http"
	"sync"
)

const (
	baseURL     = "https://api.anthropic.com/"
	clientID    = "anthropic-golang/0.4.3"
	apiPrompt   = "\n\nAssistant:"
	HumanPrompt = "\n\nHuman:"
)

type Client struct {
	APIKey   string
	ProxyURl string
	Client   *http.Client
	Debug    bool
	Header   http.Header
	mux      *sync.Mutex
}

type CompletionResponse struct {
	Completion string `json:"completion"`
	Stop       string `json:"stop"`
	StopReason string `json:"stop_reason"`
	Truncated  bool   `json:"truncated"`
	Exception  string `json:"exception"`
	LogID      string `json:"log_id"`
}

func NewClient() *Client {
	return &Client{}
}

func NewEnvClient() *Client {
	return &Client{}
}

func (c *Client) SetDebug(debug bool) {
	c.mux.Lock()
	defer c.mux.Unlock()

	c.Debug = debug
}

func (c *Client) newRequest(params interface{}, body io.Reader) (*http.Request, error) {
	return nil, nil
}

func (c *Client) NewRequest() {}

func (c *Client) Do() {}

func (c *Client) do() {}

func (c *Client) Completion(cxt context.Context) {}
