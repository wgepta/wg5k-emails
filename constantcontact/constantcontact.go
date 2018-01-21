package constantcontact

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	apiVersion     = "v2"
	defaultBaseURL = "https://api.constantcontact.com/" + apiVersion + "/"
	userAgent      = "go-constantcontact/" + apiVersion
)

// A Client manages communication with the API.
type Client struct {
	//authorization settings
	accessToken string
	apiKey      string

	client *http.Client // HTTP client used to communicate with the API.

	// Base URL for API requests. BaseURL should
	// always be specified with a trailing slash.
	BaseURL *url.URL

	// User agent used when communicating with the API.
	UserAgent string

	common service // Reuse a single struct instead of allocating one for each service on the heap.

	Lists    *ListService
	Contacts *ContactService
}

type metaResponse struct {
	Meta struct {
		Pagination struct {
			NextLink string `json:"next_link,omitempty"`
		} `json:"pagination,omitempty"`
	} `json:"meta,omitempty"`
}

//Response is an http response with functionality added for the CC API
//Namely, pagination
type Response struct {
	*http.Response

	Next string
}

type service struct {
	client *Client
}

// NewClient returns a new GitHub API client. If a nil httpClient is
// provided, http.DefaultClient will be used. To use API methods which require
// authentication, provide an http.Client that will perform the authentication
// for you (such as that provided by the golang.org/x/oauth2 library).
func NewClient(httpClient *http.Client, apiKey string, accessToken string) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	baseURL, _ := url.Parse(defaultBaseURL)

	c := &Client{
		client:      httpClient,
		BaseURL:     baseURL,
		UserAgent:   userAgent,
		apiKey:      apiKey,
		accessToken: accessToken,
	}
	c.common.client = c
	c.Lists = (*ListService)(&c.common)
	c.Contacts = (*ContactService)(&c.common)
	return c
}

// NewRequest creates an API request. A relative URL can be provided in urlStr,
// in which case it is resolved relative to the BaseURL of the Client.
// Relative URLs should always be specified without a preceding slash. If
// specified, the value pointed to by body is JSON encoded and included as the
// request body.
func (c *Client) NewRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	if !strings.HasSuffix(c.BaseURL.Path, "/") {
		return nil, fmt.Errorf("BaseURL must have a trailing slash, but %q does not", c.BaseURL)
	}
	u, err := c.BaseURL.Parse(urlStr)

	// Add api key to params
	v := u.Query()
	v.Set("api_key", c.apiKey)
	u.RawQuery = v.Encode()

	if err != nil {
		return nil, err
	}

	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		err := enc.Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)

	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}
	return req, nil
}

// Do sends an API request and returns the API response. The API response is
// JSON decoded and stored in the value pointed to by v, or returned as an
// error if an API error has occurred. If v implements the io.Writer
// interface, the raw response body will be written to v, without attempting to
// first decode it. If rate limit is exceeded and reset time is in the future,
// Do returns *RateLimitError immediately without making a network API call.
//
// The provided ctx must be non-nil. If it is canceled or times out,
// ctx.Err() will be returned.
func (c *Client) Do(ctx context.Context, req *http.Request, v interface{}) (*Response, error) {
	req = req.WithContext(ctx)

	resp, err := c.client.Do(req)
	if err != nil {
		// If we got an error, and the context has been canceled,
		// the context's error is probably more useful.
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		return nil, err
	}

	// Store the body for reuse
	defer resp.Body.Close()
	var b bytes.Buffer
	_, err = io.Copy(&b, resp.Body)
	if err != nil {
		return nil, err
	}

	// Create the CC response type
	response, err := newResponse(resp, &b)
	if err != nil {
		return nil, err
	}

	// fmt.Printf("%v\n", b.String())

	// Use the same response body to create the actual data
	// for this response
	mainReader := bytes.NewReader(b.Bytes())
	err = json.NewDecoder(mainReader).Decode(&v)
	if err == io.EOF {
		err = nil // ignore EOF errors caused by empty response body
	}

	return response, err
}

func newResponse(r *http.Response, body *bytes.Buffer) (*Response, error) {
	response := &Response{Response: r}

	headerReader := bytes.NewReader(body.Bytes())
	var meta *metaResponse
	err := json.NewDecoder(headerReader).Decode(&meta)
	if err != nil {
		// If we can't decode the meta data, this won't have pagination
		// information.  Just return the response without the pagination
		// stuff
		response.Next = ""
		return response, nil
	}

	response.Next = meta.Meta.Pagination.NextLink

	return response, err
}
