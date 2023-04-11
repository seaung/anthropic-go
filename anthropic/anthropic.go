package anthropic

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

const (
	baseURL     = "https://api.anthropic.com"
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

func NewClient(client *http.Client, apiKey string) *Client {
	if client == nil {
		client = http.DefaultClient
	}
	return &Client{
		APIKey: apiKey,
		mux:    &sync.Mutex{},
	}
}

func NewEnvClient(client *http.Client) *Client {
	return &Client{
		Client: client,
		APIKey: os.Getenv("ANTHROPIC_API_KEY"),
	}
}

func (c *Client) SetDebug(debug bool) {
	c.mux.Lock()
	defer c.mux.Unlock()

	c.Debug = debug
}

func (c *Client) SetProxy(uri string) {
	c.mux.Lock()
	defer c.mux.Unlock()

	u, err := url.Parse(uri)
	if err != nil {
		return
	}

	c.Client.Transport = &http.Transport{
		Proxy: http.ProxyURL(u),
	}
}

func (c *Client) SetTimeout(timeout int) {
	c.mux.Lock()
	defer c.mux.Unlock()

	c.Client.Timeout = time.Duration(timeout)
}

func (c *Client) newRequest(body io.Reader) (*http.Request, error) {
	uri := fmt.Sprintf("%s/v1/complete", baseURL)
	req, err := http.NewRequest(http.MethodPost, uri, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Client", clientID)
	req.Header.Add("X-API-Key", c.APIKey)

	return req, nil
}

func (c *Client) do(cxt context.Context, request *http.Request) (*http.Response, error) {
	if cxt != nil {
		request = request.WithContext(cxt)
	}

	if c.Debug {
	}

	return c.Client.Do(request)
}

func (c *Client) dumpRequest(request *http.Request) {}

func (c *Client) Completion(cxt context.Context) (*CompletionResponse, error) {
	var completion CompletionResponse
	return &completion, nil
}
