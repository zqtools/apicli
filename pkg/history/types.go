package history

import (
"fmt"
"sort"
"strings"
"time"
)

// Entry represents a single request history entry
type Entry struct {
ID          string            `json:"id"`
Timestamp   time.Time         `json:"timestamp"`
Module      string            `json:"module"`
API         string            `json:"api"`
Parameters  map[string]string `json:"parameters"`
Request     Request           `json:"request"`
Response    Response          `json:"response,omitempty"`
}

// GetCommandLine returns the complete command line for this entry
func (e *Entry) GetCommandLine() string {
var params []string
for k, v := range e.Parameters {
params = append(params, fmt.Sprintf("--%s %q", k, v))
}
sort.Strings(params) // Sort for consistent output
return fmt.Sprintf("call %s.%s %s", e.Module, e.API, strings.Join(params, " "))
}

// Request represents the request details in a history entry
type Request struct {
Method      string            `json:"method"`
URL         string           `json:"url"`
Headers     map[string]string `json:"headers,omitempty"`
Body        string           `json:"body,omitempty"`
Form        map[string]string `json:"form,omitempty"`
QueryParams map[string]string `json:"query_params,omitempty"`
}

// Response represents the response details in a history entry
type Response struct {
StatusCode int               `json:"status_code"`
Headers    map[string]string `json:"headers,omitempty"`
Body       string           `json:"body"`
}

// History represents the collection of history entries
type History struct {
Entries []Entry `json:"entries"`
}
