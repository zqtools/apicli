package client

import (
"bytes"
"encoding/json"
"fmt"
"io"
"mime/multipart"
"net/http"
"net/http/httputil"
"os"
"path/filepath"
"strings"
"time"

"github.com/google/uuid"
"github.com/zqtools/apicli/pkg/config"
"github.com/zqtools/apicli/pkg/history"
"github.com/zqtools/apicli/pkg/template"
)

// Client handles HTTP request execution
type Client struct {
httpClient *http.Client
verbose    bool
renderer   *template.Renderer
history   *history.Manager
modulePath string
apiName    string
}

// NewClient creates a new API client
func NewClient(params map[string]interface{}, verbose bool, historyManager *history.Manager, modulePath, apiName string) *Client {
return &Client{
httpClient: &http.Client{},
verbose:    verbose,
renderer:   template.NewRenderer(params),
history:   historyManager,
modulePath: modulePath,
apiName:    apiName,
}
}

// ExecuteRequest executes an API request based on the given specification
func (c *Client) ExecuteRequest(spec config.RequestSpec) (string, error) {
// Initialize history entry
historyEntry := history.Entry{
ID:         uuid.New().String(),
Timestamp:  time.Now(),
Module:     c.modulePath,
API:        c.apiName,
Parameters: make(map[string]string),
Request: history.Request{
Method:      spec.Method,
Headers:     make(map[string]string),
},
}

// Save parameters
for k, v := range c.renderer.GetParams() {
if str, ok := v.(string); ok {
historyEntry.Parameters[k] = str
} else {
historyEntry.Parameters[k] = fmt.Sprintf("%v", v)
}
}
// Render URL template
url, err := c.renderer.Render(spec.URL)
if err != nil {
return "", fmt.Errorf("rendering URL template: %w", err)
}

var req *http.Request
var bodyErr error

// Create request based on specification
switch {
case spec.BodyFile != "":
req, bodyErr = c.createRequestFromFile(spec.Method, url, spec.BodyFile)
case len(spec.Form) > 0:
req, bodyErr = c.createFormRequest(spec.Method, url, spec.Form)
case spec.Body != "":
req, bodyErr = c.createBodyRequest(spec.Method, url, spec.Body)
default:
req, bodyErr = http.NewRequest(spec.Method, url, nil)
}

if bodyErr != nil {
return "", fmt.Errorf("creating request: %w", bodyErr)
}

// Add query parameters
if len(spec.Params) > 0 {
queryParams := make(map[string]string)
if err := c.addQueryParams(req, spec.Params, queryParams); err != nil {
return "", err
}
historyEntry.Request.QueryParams = queryParams
}

// Add headers
if err := c.addHeaders(req, spec.Headers, historyEntry.Request.Headers); err != nil {
return "", err
}

// Print request details if verbose mode is enabled
// Save request URL after all parameters are added
historyEntry.Request.URL = req.URL.String()

// Save request body if present
if spec.Body != "" {
historyEntry.Request.Body = spec.Body
} else if len(spec.Form) > 0 {
historyEntry.Request.Form = spec.Form
}

if c.verbose {
c.dumpRequest(req)
}

// Execute request
resp, err := c.httpClient.Do(req)
if err != nil {
return "", fmt.Errorf("sending request: %w", err)
}
defer resp.Body.Close()

// Initialize response in history entry
historyEntry.Response = history.Response{
StatusCode: resp.StatusCode,
Headers:    make(map[string]string),
}

// Copy response headers
for k, v := range resp.Header {
if len(v) > 0 {
historyEntry.Response.Headers[k] = v[0]
}
}

// Print response details if verbose mode is enabled
if c.verbose {
c.dumpResponse(resp)
}

// Read and format response
respStr, err := c.formatResponse(resp)
if err != nil {
return "", err
}

// Save response body
historyEntry.Response.Body = respStr

// Record history if manager is available
if c.history != nil {
if err := c.history.AddEntry(historyEntry); err != nil {
// Just log the error, don't fail the request
fmt.Fprintf(os.Stderr, "Warning: Failed to record history: %v\n", err)
}
}

return respStr, nil
}

func (c *Client) createRequestFromFile(method, url, filepath string) (*http.Request, error) {
path, err := c.renderer.Render(filepath)
if err != nil {
return nil, fmt.Errorf("rendering body file path template: %w", err)
}

file, err := os.Open(path)
if err != nil {
return nil, fmt.Errorf("opening file: %w", err)
}
defer file.Close()

return http.NewRequest(method, url, file)
}

func (c *Client) createFormRequest(method, url string, form map[string]string) (*http.Request, error) {
body := &bytes.Buffer{}
writer := multipart.NewWriter(body)

for field, valueTmpl := range form {
value, err := c.renderer.Render(valueTmpl)
if err != nil {
return nil, fmt.Errorf("rendering form field template: %w", err)
}

// Try to open as file first
if file, err := os.Open(value); err == nil {
defer file.Close()
part, err := writer.CreateFormFile(field, filepath.Base(value))
if err != nil {
return nil, fmt.Errorf("creating form file: %w", err)
}
if _, err := io.Copy(part, file); err != nil {
return nil, fmt.Errorf("copying file content: %w", err)
}
} else {
// Not a file, write as regular field
if err := writer.WriteField(field, value); err != nil {
return nil, fmt.Errorf("writing form field: %w", err)
}
}
}

if err := writer.Close(); err != nil {
return nil, fmt.Errorf("closing multipart writer: %w", err)
}

req, err := http.NewRequest(method, url, body)
if err != nil {
return nil, err
}

req.Header.Set("Content-Type", writer.FormDataContentType())
return req, nil
}

func (c *Client) createBodyRequest(method, url, body string) (*http.Request, error) {
renderedBody, err := c.renderer.Render(body)
if err != nil {
return nil, fmt.Errorf("rendering body template: %w", err)
}
return http.NewRequest(method, url, bytes.NewBufferString(renderedBody))
}

func (c *Client) addQueryParams(req *http.Request, params []config.QueryParam, historyParams map[string]string) error {
q := req.URL.Query()
for _, param := range params {
value, err := c.renderer.Render(param.Value)
if err == nil && value != "" {
q.Add(param.Name, value)
historyParams[param.Name] = value
}
}
req.URL.RawQuery = q.Encode()
return nil
}

func (c *Client) addHeaders(req *http.Request, headers map[string]string, historyHeaders map[string]string) error {
for key, valueTmpl := range headers {
value, err := c.renderer.Render(valueTmpl)
if err != nil {
return fmt.Errorf("rendering header template: %w", err)
}
req.Header.Set(key, value)
historyHeaders[key] = value
}
return nil
}

func (c *Client) dumpRequest(req *http.Request) {
dump, err := httputil.DumpRequestOut(req, true)
if err == nil {
fmt.Printf("\n>>> Request:\n%s\n\n", string(dump))
}
}

func (c *Client) dumpResponse(resp *http.Response) {
dump, err := httputil.DumpResponse(resp, true)
if err == nil {
fmt.Printf("\n<<< Response:\n%s\n\n", string(dump))
}
}

func (c *Client) formatResponse(resp *http.Response) (string, error) {
body, err := io.ReadAll(resp.Body)
if err != nil {
return "", fmt.Errorf("reading response: %w", err)
}

if isJSONResponse(resp.Header) {
var prettyJSON bytes.Buffer
if err := json.Indent(&prettyJSON, body, "", "  "); err == nil {
return prettyJSON.String(), nil
}
}

return string(body), nil
}

func isJSONResponse(header http.Header) bool {
contentType := header.Get("Content-Type")
return strings.Contains(contentType, "application/json")
}
