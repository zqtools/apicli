# apicli

A command-line interface tool for managing and executing API calls with configurable parameters and templates.

## Features

- Configurable API definitions using YAML
- Support for multiple API modules and endpoints
- Parameter templating and validation
- File upload support
- Environment-specific configurations
- Request history tracking

## Installation

```bash
go install github.com/zqtools/apicli@latest
```

Or clone and build from source:

```bash
git clone https://github.com/zqtools/apicli.git
cd apicli
go build
```

## Project Structure

```
.
├── main.go              # Main application entry point
├── apis.yaml           # API definitions file
├── pkg/
│   ├── api/           # API handling and CLI implementation
│   ├── client/        # HTTP client implementation
│   ├── config/        # Configuration management
│   ├── history/       # Request history tracking
│   └── template/      # Template processing
```

## Configuration

APIs are defined in YAML format. Here's an example structure:

```yaml
modules:
  user:
    description: User management APIs
    params:
      - name: token
        type: string
        required: true
        description: Auth token
    request:
      headers:
        Authorization: Bearer ${token}
    modules:
      profile:
        apis:
          get_settings:
            params:
              - name: id
                type: string
                required: true
            request:
              method: GET
              url: https://api.example.com/users/${id}/settings
```

## Usage Examples

1. Get user settings:
```bash
apicli user profile get_settings --token="your-token" --id="user123"
```

2. Update user settings:
```bash
apicli user profile update_settings \
  --token="your-token" \
  --id="user123" \
  --settings='{"theme":"dark"}'
```

3. Upload user avatar:
```bash
apicli user profile upload_avatar --token="your-token" --id="user123" --file="./avatar.jpg"
```

4. Admin operations:
```bash
apicli admin roles list --admin_token="admin-token" --env="prod"
apicli admin roles assign --admin_token="admin-token" --env="prod" --user_id="user123" --role="admin"
```

## Dependencies

- Go 1.20 or higher
- gopkg.in/yaml.v3
- github.com/google/uuid

## License

[Add your license information here]
