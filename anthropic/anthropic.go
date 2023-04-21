package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	baseURL     = "https://api.anthropic.com"
	clientID    = "anthropic-golang/0.4.3"
	apiPrompt   = "\n\nAssistant:"
	HumanPrompt = "\n\nHuman:"
)

var ReadBodyError = errors.New("could not read error response.")

type ErrorHandler func(*http.Response) error

type Client struct {
	mux    *sync.Mutex
	apiKey string
	Client *http.Client
	Debug  bool
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
		apiKey: apiKey,
		mux:    &sync.Mutex{},
	}
}

func NewEnvClient(client *http.Client) *Client {
	return &Client{
		Client: client,
		apiKey: os.Getenv("ANTHROPIC_API_KEY"),
	}
}

func (c *Client) SetDebug(debug bool) {
	c.mux.Lock()
	defer c.mux.Unlock()

	c.Debug = debug
}

func (c *Client) Setproxy(uri string) {
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

func (c *Client) newRequest(uri string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodPost, uri, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Client", clientID)
	req.Header.Add("X-API-Key", c.apiKey)

	return req, nil
}

func (c *Client) do(cxt context.Context, request *http.Request) (*http.Response, error) {
	if cxt != nil {
		request = request.WithContext(cxt)
	}

	if c.Debug {
		c.dumpRequest(request)
	}

	return c.Client.Do(request)
}

func (c *Client) dumpRequest(request *http.Request) {
	var body string
	if request.Body != nil {
		byteContent, err := ioutil.ReadAll(request.Body)
		if err != nil {
			log.Printf("[DEBUG] failed to read body : %s\n", err)
		}

		request.Body = ioutil.NopCloser(bytes.NewReader(byteContent))
		body = string(byteContent)
	}

	message := fmt.Sprintf("%s %s %s", request.Method, request.URL.String(), body)
	log.Printf("[DEBUG] client request : %s\n", message)
}

func (c *Client) parseResponseContent(destination interface{}, body io.Reader) error {
	var err error
	if word, ok := destination.(io.Writer); ok {
		_, err = io.Copy(word, body)
	} else {
		decoder := json.NewDecoder(body)
		err = decoder.Decode(destination)
	}

	return err
}

func (c *Client) DoWithErrorHanding(ctx context.Context, request *http.Request, destination interface{}, errHandler ErrorHandler) error {
	response, err := c.do(ctx, request)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return errHandler(response)
	}

	if destination == nil {
		return nil
	}

	return c.parseResponseContent(destination, response.Body)
}

func (c *Client) Do(ctx context.Context, request *http.Request, destination interface{}) error {
	return c.DoWithErrorHanding(ctx, request, destination, getErrorFromResponse)
}

func getErrorFromResponse(r *http.Response) error {
	rs := new(struct {
		Error string `json:"error"`
	})
	msg, err := ioutil.ReadAll(r.Body)
	if err == nil {
		if err := json.Unmarshal(msg, rs); err == nil {
			return errors.New(rs.Error)
		}
		return errors.New(strings.TrimSpace(string(msg)))
	}
	return ReadBodyError
}

func (c *Client) Completion(ctx context.Context, parameters url.Values) (*CompletionResponse, error) {
	var completion CompletionResponse
	uri := fmt.Sprintf("%s/v1/complete", baseURL)

	req, err := c.newRequest(uri, strings.NewReader(parameters.Encode()))
	if err != nil {
		return nil, err
	}

	if err := c.Do(ctx, req, &completion); err != nil {
		return nil, err
	}
	return &completion, nil
}
