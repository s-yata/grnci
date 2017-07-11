package grnci

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"path"
	"time"
)

const (
	httpBufferSize = 1024 // Enough size to store the response header
)

// httpResponse is an HTTP response.
type httpResponse struct {
	resp    *http.Response // HTTP response
	plain   bool           // Whether or not the response is plain
	start   time.Time      // Start time
	elapsed time.Duration  // Elapsed time
	err     error          // Error response
	left    []byte         // Data left in buf
	buf     [1]byte        // Buffer for the next byte
}

// extractHTTPResponseHeader extracts the HTTP resonse header.
func extractHTTPResponseHeader(data []byte) (head, left []byte, err error) {
	left = bytes.TrimLeft(data[1:], " \t\r\n")
	if !bytes.HasPrefix(left, []byte("[")) {
		err = NewError(ResponseError, map[string]interface{}{
			"error": "The response does not contain a header.",
		})
		return
	}
	var i int
	stack := []byte{']'}
Loop:
	for i = 1; i < len(left); i++ {
		switch left[i] {
		case '[':
			stack = append(stack, ']')
		case '{':
			stack = append(stack, '}')
		case ']', '}':
			if left[i] != stack[len(stack)-1] {
				err = NewError(ResponseError, map[string]interface{}{
					"error": "The response header is broken.",
				})
				return
			}
			stack = stack[:len(stack)-1]
			if len(stack) == 0 {
				break Loop
			}
		case '"':
			for i++; i < len(left); i++ {
				if left[i] == '\\' {
					i++
					continue
				}
				if left[i] == '"' {
					break
				}
			}
		}
	}
	if len(stack) != 0 {
		err = NewError(ResponseError, map[string]interface{}{
			"error": "The response header is too long or broken.",
		})
		return
	}
	head = left[:i+1]
	left = bytes.TrimLeft(left[i+1:], " \t\r\n")
	if bytes.HasPrefix(left, []byte(",")) {
		left = bytes.TrimLeft(left[1:], " \t\r\n")
	}
	return
}

// parseHTTPResponseHeaderError parses the error information in the HTTP resonse header.
func parseHTTPResponseHeaderError(code int, elems []interface{}) error {
	err := NewError(code, nil)
	if len(elems) >= 1 {
		err = EnhanceError(err, map[string]interface{}{
			"message": elems[0],
		})
	}
	if len(elems) >= 2 {
		if locs, ok := elems[1].([]interface{}); ok {
			if len(locs) >= 1 {
				if grnLocs, ok := locs[0].([]interface{}); ok {
					if len(grnLocs) >= 1 {
						if f, ok := grnLocs[0].(string); ok {
							err = EnhanceError(err, map[string]interface{}{
								"function": f,
							})
						}
					}
					if len(grnLocs) >= 2 {
						if f, ok := grnLocs[1].(string); ok {
							err = EnhanceError(err, map[string]interface{}{
								"file": f,
							})
						}
					}
					if len(grnLocs) >= 3 {
						if f, ok := grnLocs[2].(float64); ok {
							err = EnhanceError(err, map[string]interface{}{
								"line": int(f),
							})
						}
					}
				}
			}
		}
	}
	return err
}

// parseHTTPResponseHeader parses the HTTP resonse header.
func parseHTTPResponseHeader(resp *http.Response, data []byte) (*httpResponse, error) {
	head, left, err := extractHTTPResponseHeader(data)
	if err != nil {
		return nil, err
	}

	var elems []interface{}
	if err := json.Unmarshal(head, &elems); err != nil {
		return nil, NewError(ResponseError, map[string]interface{}{
			"method": "json.Unmarshal",
			"error":  err.Error(),
		})
	}
	if len(elems) < 3 {
		return nil, NewError(ResponseError, map[string]interface{}{
			"method": "json.Unmarshal",
			"error":  "Too few elements in the response header.",
		})
	}
	f, ok := elems[0].(float64)
	if !ok {
		return nil, NewError(ResponseError, map[string]interface{}{
			"code":  elems[0],
			"error": "code must be a number.",
		})
	}
	code := int(f)
	f, ok = elems[1].(float64)
	if !ok {
		return nil, NewError(ResponseError, map[string]interface{}{
			"start": elems[1],
			"error": "start must be a number.",
		})
	}
	i, f := math.Modf(f)
	start := time.Unix(int64(i), int64(math.Floor(f*1000000+0.5))*1000).Local()
	f, ok = elems[2].(float64)
	if !ok {
		return nil, NewError(ResponseError, map[string]interface{}{
			"elapsed": elems[2],
			"error":   "elapsed must be a number.",
		})
	}
	elapsed := time.Duration(f * float64(time.Second))

	if code != 0 {
		err = parseHTTPResponseHeaderError(code, elems[3:])
	}

	return &httpResponse{
		resp:    resp,
		start:   start,
		elapsed: elapsed,
		err:     err,
		left:    left,
	}, nil
}

// newHTTPResponse returns a new httpResponse.
func newHTTPResponse(resp *http.Response, start time.Time) (*httpResponse, error) {
	buf := make([]byte, httpBufferSize)
	n := 0
	for n < len(buf) {
		m, err := resp.Body.Read(buf[n:])
		n += m
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, NewError(NetworkError, map[string]interface{}{
				"method": "http.Response.Body.Read",
				"error":  err.Error(),
			})
		}
	}
	data := bytes.TrimLeft(buf[:n], " \t\r\n")
	if bytes.HasPrefix(data, []byte("[")) {
		return parseHTTPResponseHeader(resp, data)
	}
	var err error
	code := resp.StatusCode
	if code != http.StatusOK {
		err = NewError(HTTPError, map[string]interface{}{
			"status": fmt.Sprintf("%d %s", code, http.StatusText(code)),
			"note":   "The response format is not JSON.",
		})
	}
	return &httpResponse{
		resp:    resp,
		plain:   true,
		start:   start,
		elapsed: time.Now().Sub(start),
		err:     err,
		left:    data,
	}, nil
}

func (r *httpResponse) Start() time.Time {
	return r.start
}

func (r *httpResponse) Elapsed() time.Duration {
	return r.elapsed
}

func (r *httpResponse) Read(p []byte) (n int, err error) {
	if len(r.left) != 0 {
		n = copy(p, r.left)
		r.left = r.left[n:]
		if len(r.left) != 0 {
			return
		}
	}
	var m int
	if n < len(p) {
		m, err = r.resp.Body.Read(p[n:])
		n += m
		if err != nil {
			if !r.plain && n > 0 && p[n-1] == ']' {
				n--
			}
			if err != io.EOF {
				err = NewError(NetworkError, map[string]interface{}{
					"method": "http.Response.Body.Read",
					"error":  err.Error(),
				})
			}
			return
		}
	}
	if r.plain || n == 0 || p[n-1] != ']' {
		return
	}
	m, err = r.resp.Body.Read(r.buf[:])
	if err == nil {
		r.left = r.buf[:m]
		return
	}
	if m == 0 {
		n--
	}
	if err != io.EOF {
		err = NewError(NetworkError, map[string]interface{}{
			"method": "http.Response.Body.Read",
			"error":  err.Error(),
		})
	}
	return
}

func (r *httpResponse) Close() error {
	io.Copy(ioutil.Discard, r.resp.Body)
	if err := r.resp.Body.Close(); err != nil {
		return NewError(NetworkError, map[string]interface{}{
			"method": "http.Response.Body.Close",
			"error":  err.Error(),
		})
	}
	return nil
}

func (r *httpResponse) Err() error {
	return r.err
}

// HTTPClient is a thread-safe HTTP client.
type HTTPClient struct {
	url    *url.URL
	client *http.Client
}

// NewHTTPClient returns a new HTTPClient.
// The expected address format is
// [scheme://][username[:password]@][host][:port][path][?query][#fragment].
// If client is nil, NewHTTPClient uses http.DefaultClient.
func NewHTTPClient(addr string, client *http.Client) (*HTTPClient, error) {
	a, err := ParseHTTPAddress(addr)
	if err != nil {
		return nil, err
	}
	url, err := url.Parse(a.String())
	if err != nil {
		return nil, NewError(AddressError, map[string]interface{}{
			"url":    a.String(),
			"method": "url.Parse",
			"error":  err.Error(),
		})
	}
	if client == nil {
		client = http.DefaultClient
	}
	return &HTTPClient{
		url:    url,
		client: client,
	}, nil
}

// Close does nothing.
func (c *HTTPClient) Close() error {
	return nil
}

// exec sends a command and receives a response.
func (c *HTTPClient) exec(name string, params map[string]string, body io.Reader) (Response, error) {
	start := time.Now()
	url := *c.url
	url.Path = path.Join(url.Path, name)
	if len(params) != 0 {
		query := url.Query()
		for k, v := range params {
			query.Add(k, v)
		}
		url.RawQuery = query.Encode()
	}
	if body == nil {
		resp, err := c.client.Get(url.String())
		if err != nil {
			return nil, NewError(NetworkError, map[string]interface{}{
				"url":    url.String(),
				"method": "http.Client.Get",
				"error":  err.Error(),
			})
		}
		return newHTTPResponse(resp, start)
	}
	resp, err := c.client.Post(url.String(), "application/json", body)
	if err != nil {
		return nil, NewError(NetworkError, map[string]interface{}{
			"url":    url.String(),
			"method": "http.Client.Post",
			"error":  err.Error(),
		})
	}
	return newHTTPResponse(resp, start)
}

// Exec assembles cmd and body into a Command and calls Query.
func (c *HTTPClient) Exec(cmd string, body io.Reader) (Response, error) {
	command, err := ParseCommand(cmd)
	if err != nil {
		return nil, err
	}
	command.SetBody(body)
	return c.Query(command)
}

// Invoke assembles name, params and body into a Command and calls Query.
func (c *HTTPClient) Invoke(name string, params map[string]interface{}, body io.Reader) (Response, error) {
	cmd, err := NewCommand(name, params)
	if err != nil {
		return nil, err
	}
	cmd.SetBody(body)
	return c.Query(cmd)
}

// Query sends a command and receives a response.
// It is the caller's responsibility to close the response.
func (c *HTTPClient) Query(cmd *Command) (Response, error) {
	if err := cmd.Check(); err != nil {
		return nil, err
	}
	return c.exec(cmd.Name(), cmd.Params(), cmd.Body())
}
