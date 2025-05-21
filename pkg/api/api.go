package api

import (
"flag"
"fmt"
"strconv"
"strings"

"github.com/zqtools/apicli/pkg/client"
"github.com/zqtools/apicli/pkg/config"
"github.com/zqtools/apicli/pkg/history"
"github.com/zqtools/apicli/pkg/template"
)

// CLI represents the command-line interface for the API client
type CLI struct {
config    *config.Config
verbose   *bool
force     *bool
userConfig *config.UserConfig
history   *history.Manager
}

// NewCLI creates a new CLI instance
func NewCLI(userConfig *config.UserConfig, apiDir string) (*CLI, error) {
// Load API configuration
apiConfig, err := config.LoadConfig(userConfig.APIConfigPath)
if err != nil {
return nil, fmt.Errorf("loading API config: %w", err)
}

// Initialize history manager
historyManager, err := history.NewManager(apiDir)
if err != nil {
return nil, fmt.Errorf("initializing history manager: %w", err)
}

// Create CLI instance
cli := &CLI{
config:     apiConfig,
verbose:    flag.Bool("verbose", false, "Show request details"),
force:      flag.Bool("force", false, "Skip confirmation for non-GET requests"),
userConfig: userConfig,
history:    historyManager,
}

return cli, nil
}

// Execute processes command line arguments and executes the API request
func (c *CLI) Execute(args []string) error {
// Create a new FlagSet for global flags
globalFlags := flag.NewFlagSet("global", flag.ExitOnError)
verbose := globalFlags.Bool("verbose", false, "Show request details")
force := globalFlags.Bool("force", false, "Skip confirmation for non-GET requests")

// Find the position of the first non-flag argument
cmdStart := 0
for i, arg := range args {
if !strings.HasPrefix(arg, "-") {
cmdStart = i
break
}
}

// Parse global flags
if err := globalFlags.Parse(args[:cmdStart]); err != nil {
return fmt.Errorf("parsing global flags: %w", err)
}

// Update CLI flags
c.verbose = verbose
c.force = force

// Get remaining arguments
remaining := args[cmdStart:]
if len(remaining) == 0 {
c.printUsage()
return fmt.Errorf("no command specified")
}

// Handle commands
switch remaining[0] {
case "call":
return c.handleCallCommand(remaining[1:])
case "history":
return c.handleHistoryCommand(remaining[1:])
default:
// For backward compatibility, treat as a call command
return c.handleCallCommand(remaining)
}
}

func (c *CLI) handleCallCommand(args []string) error {
if len(args) < 2 {
c.printUsage()
return fmt.Errorf("insufficient arguments for call command")
}

// Parse module path and API name
modulePath := strings.Split(args[0], ".")
apiName := args[1]

// Get module chain info
moduleParams, moduleReqs, apiSpec, err := config.CollectModuleInfo(c.config.Modules, modulePath, apiName)
if err != nil {
return fmt.Errorf("collecting module info: %w", err)
}

// Create parameter flags
apiFlags := flag.NewFlagSet("call", flag.ExitOnError)
paramFlags := make(map[string]*string)

// Define flags for all parameters
for _, param := range moduleParams {
paramFlags[param.Name] = apiFlags.String(param.Name, "", param.Description)
}
for _, param := range apiSpec.Params {
paramFlags[param.Name] = apiFlags.String(param.Name, "", param.Description)
}

// Parse API-specific flags
if err := apiFlags.Parse(args[2:]); err != nil {
return fmt.Errorf("parsing parameters: %w", err)
}

// Collect and validate parameter values
paramValues := make(map[string]interface{})
allParams := append(moduleParams, apiSpec.Params...)
for _, param := range allParams {
value := *paramFlags[param.Name]
if err := c.validateParam(param, value); err != nil {
return err
}
if value != "" {
paramValues[param.Name] = c.convertParamValue(param.Type, value)
}
}

// Merge request configurations
mergedReq := config.MergeRequestConfigs(moduleReqs, &apiSpec.Request)

// Confirm non-GET requests unless forced
if !*c.force && mergedReq.Method != "GET" {
if confirmed := c.confirmRequest(mergedReq, paramValues); !confirmed {
return fmt.Errorf("operation cancelled by user")
}
}

// Create and execute request
apiClient := client.NewClient(paramValues, *c.verbose, c.history, strings.Join(modulePath, "."), apiName)
response, err := apiClient.ExecuteRequest(*mergedReq)
if err != nil {
return fmt.Errorf("executing request: %w", err)
}

fmt.Println(response)
return nil
}

func (c *CLI) validateParam(param config.ParamDef, value string) error {
if param.Required && value == "" {
return fmt.Errorf("parameter '%s' is required", param.Name)
}
if value != "" {
if err := template.ValidateType(value, param.Type); err != nil {
return fmt.Errorf("parameter '%s': %w", param.Name, err)
}
}
return nil
}

func (c *CLI) convertParamValue(paramType, value string) interface{} {
switch {
case paramType == "integer":
// Already validated, so conversion should succeed
val, _ := strconv.Atoi(value)
return val
case paramType == "boolean":
return value == "true"
default:
return value
}
}

func (c *CLI) handleHistoryCommand(args []string) error {
if len(args) == 0 {
c.printUsage()
return fmt.Errorf("no history subcommand specified")
}

switch args[0] {
case "list":
return c.handleHistoryList(args[1:])
case "show":
return c.handleHistoryShow(args[1:])
case "clear":
return c.handleHistoryClear()
default:
return fmt.Errorf("unknown history subcommand: %s", args[0])
}
}

func (c *CLI) handleHistoryList(args []string) error {
listFlags := flag.NewFlagSet("history list", flag.ExitOnError)
limit := listFlags.Int("limit", 10, "Maximum number of entries to show")
if err := listFlags.Parse(args); err != nil {
return err
}

entries, err := c.history.ListEntries(*limit)
if err != nil {
return fmt.Errorf("listing history: %w", err)
}

if len(entries) == 0 {
fmt.Println("No history entries found")
return nil
}

fmt.Printf("Last %d API calls:\n\n", len(entries))
for _, entry := range entries {
fmt.Printf("Time: %s\n", entry.Timestamp.Format("2006-01-02 15:04:05"))
fmt.Printf("Command: apicli %s\n", entry.GetCommandLine())
fmt.Printf("ID: %s\n", entry.ID)
fmt.Printf("Status: %d\n", entry.Response.StatusCode)
fmt.Println(strings.Repeat("-", 80))
}

return nil
}

func (c *CLI) handleHistoryShow(args []string) error {
if len(args) == 0 {
return fmt.Errorf("no history entry ID specified")
}

entry, err := c.history.GetEntry(args[0])
if err != nil {
return err
}

fmt.Printf("API Call Details:\n")
fmt.Printf("Time: %s\n", entry.Timestamp.Format("2006-01-02 15:04:05"))
fmt.Printf("Module: %s\n", entry.Module)
fmt.Printf("API: %s\n", entry.API)
fmt.Printf("\nRequest:\n")
fmt.Printf("Method: %s\n", entry.Request.Method)
fmt.Printf("URL: %s\n", entry.Request.URL)
if len(entry.Request.Headers) > 0 {
fmt.Printf("\nHeaders:\n")
for k, v := range entry.Request.Headers {
fmt.Printf("  %s: %s\n", k, v)
}
}
if entry.Request.Body != "" {
fmt.Printf("\nBody:\n%s\n", entry.Request.Body)
}
if len(entry.Request.Form) > 0 {
fmt.Printf("\nForm Data:\n")
for k, v := range entry.Request.Form {
fmt.Printf("  %s: %s\n", k, v)
}
}
if len(entry.Request.QueryParams) > 0 {
fmt.Printf("\nQuery Parameters:\n")
for k, v := range entry.Request.QueryParams {
fmt.Printf("  %s: %s\n", k, v)
}
}

fmt.Printf("\nResponse:\n")
fmt.Printf("Status: %d\n", entry.Response.StatusCode)
if len(entry.Response.Headers) > 0 {
fmt.Printf("\nHeaders:\n")
for k, v := range entry.Response.Headers {
fmt.Printf("  %s: %s\n", k, v)
}
}
fmt.Printf("\nBody:\n%s\n", entry.Response.Body)

return nil
}

func (c *CLI) handleHistoryClear() error {
fmt.Print("Are you sure you want to clear all history? [y/N] ")
var answer string
fmt.Scanln(&answer)
if strings.ToLower(answer) != "y" {
fmt.Println("Operation cancelled")
return nil
}

if err := c.history.ClearHistory(); err != nil {
return fmt.Errorf("clearing history: %w", err)
}

fmt.Println("History cleared successfully")
return nil
}

func (c *CLI) confirmRequest(req *config.RequestSpec, params map[string]interface{}) bool {
fmt.Printf("\nAbout to send %s request to %s\n", req.Method, req.URL)

renderer := template.NewRenderer(params)

if req.Body != "" {
body, err := renderer.Render(req.Body)
if err == nil {
fmt.Printf("With body:\n%s\n", body)
}
} else if len(req.Form) > 0 {
fmt.Println("With form data:")
for k, v := range req.Form {
if val, err := renderer.Render(v); err == nil {
fmt.Printf("  %s: %s\n", k, val)
}
}
}

fmt.Print("\nDo you want to proceed? [y/N] ")

var answer string
fmt.Scanln(&answer)
return strings.ToLower(answer) == "y"
}

func (c *CLI) printUsage() {
fmt.Println("Usage: apicli [command] [options]")
fmt.Println("\nCommands:")
fmt.Println("  call MODULE[.SUBMODULE] API [parameters]  Call an API endpoint")
fmt.Println("  history list [--limit N]                  List recent API calls")
fmt.Println("  history show ID                          Show details of a specific API call")
fmt.Println("  history clear                            Clear API call history")
fmt.Println("\nOptions:")
fmt.Println("  --verbose\tShow request details")
fmt.Println("  --force\tSkip confirmation for non-GET requests")

if c.config != nil {
fmt.Println("\nAvailable modules:")
c.printModuleTree(c.config.Modules, "", "")
}
}

func (c *CLI) printModuleTree(modules map[string]config.Module, prefix, indent string) {
for name, module := range modules {
fmt.Printf("%s%s%s\t%s\n", indent, prefix, name, module.Description)

if len(module.Params) > 0 {
fmt.Printf("%s  Module Parameters:\n", indent)
for _, param := range module.Params {
required := ""
if param.Required {
required = " (required)"
}
fmt.Printf("%s    --%s\t%s [%s]%s\n", indent, param.Name, param.Description, param.Type, required)
}
}

if module.Request != nil && len(module.Request.Headers) > 0 {
fmt.Printf("%s  Module Headers:\n", indent)
for key, value := range module.Request.Headers {
if strings.Contains(value, "${") {
fmt.Printf("%s    %s: %s (template)\n", indent, key, value)
} else {
fmt.Printf("%s    %s: %s\n", indent, key, value)
}
}
}

if len(module.Modules) > 0 {
newIndent := indent + "  "
newPrefix := name + "."
c.printModuleTree(module.Modules, newPrefix, newIndent)
}

if len(module.APIs) > 0 {
fmt.Printf("%s  APIs:\n", indent)
for apiName, api := range module.APIs {
fmt.Printf("%s    %s (%s %s)\n", indent, apiName, api.Request.Method, api.Request.URL)
if len(api.Params) > 0 {
fmt.Printf("%s      Parameters:\n", indent)
for _, param := range api.Params {
required := ""
if param.Required {
required = " (required)"
}
fmt.Printf("%s        --%s\t%s [%s]%s\n", indent, param.Name, param.Description, param.Type, required)
}
}
}
fmt.Println()
}
}
}
