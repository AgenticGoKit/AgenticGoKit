# Tool Development in AgenticGoKit

## Overview

Tools are the building blocks that extend agent capabilities beyond text generation. This tutorial covers how to develop custom tools in AgenticGoKit, from simple utilities to complex integrations with external services.

By creating custom tools, you can enable your agents to perform specific tasks, access specialized information, and interact with external systems in a controlled manner.

## Prerequisites

- Understanding of [MCP Overview](README.md)
- Basic knowledge of Go interfaces and error handling
- Familiarity with [State Management](../core-concepts/state-management.md)

## Tool Interface

In AgenticGoKit, all tools implement the `Tool` interface:

```go
type Tool interface {
    // Name returns the tool's unique identifier
    Name() string
    
    // Description provides information about the tool's functionality
    Description() string
    
    // ParameterSchema defines the expected input parameters
    ParameterSchema() map[string]ParameterDefinition
    
    // Execute runs the tool with the provided parameters
    Execute(ctx context.Context, params map[string]interface{}) (interface{}, error)
}

type ParameterDefinition struct {
    Type        string      `json:"type"`        // string, number, boolean, array, object
    Description string      `json:"description"` // Parameter description
    Required    bool        `json:"required"`    // Whether parameter is required
    Default     interface{} `json:"default"`     // Default value if not provided
    Enum        []string    `json:"enum"`        // Possible values (optional)
}
```

## Creating a Basic Tool

### 1. Simple Calculator Tool

```go
package tools

import (
    "context"
    "fmt"
    "strconv"
    
    "github.com/kunalkushwaha/agenticgokit/core"
)

// CalculatorTool provides basic arithmetic operations
type CalculatorTool struct{}

func NewCalculatorTool() *CalculatorTool {
    return &CalculatorTool{}
}

func (t *CalculatorTool) Name() string {
    return "calculator"
}

func (t *CalculatorTool) Description() string {
    return "Performs basic arithmetic operations (add, subtract, multiply, divide)"
}

func (t *CalculatorTool) ParameterSchema() map[string]core.ParameterDefinition {
    return map[string]core.ParameterDefinition{
        "operation": {
            Type:        "string",
            Description: "The arithmetic operation to perform",
            Required:    true,
            Enum:        []string{"add", "subtract", "multiply", "divide"},
        },
        "a": {
            Type:        "number",
            Description: "First operand",
            Required:    true,
        },
        "b": {
            Type:        "number",
            Description: "Second operand",
            Required:    true,
        },
    }
}

func (t *CalculatorTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    // Extract operation
    operation, ok := params["operation"].(string)
    if !ok {
        return nil, fmt.Errorf("operation must be a string")
    }
    
    // Extract operands and convert to float64
    a, err := getFloat(params["a"])
    if err != nil {
        return nil, fmt.Errorf("invalid first operand: %w", err)
    }
    
    b, err := getFloat(params["b"])
    if err != nil {
        return nil, fmt.Errorf("invalid second operand: %w", err)
    }
    
    // Perform operation
    var result float64
    switch operation {
    case "add":
        result = a + b
    case "subtract":
        result = a - b
    case "multiply":
        result = a * b
    case "divide":
        if b == 0 {
            return nil, fmt.Errorf("division by zero")
        }
        result = a / b
    default:
        return nil, fmt.Errorf("unsupported operation: %s", operation)
    }
    
    // Return result
    return map[string]interface{}{
        "result": result,
    }, nil
}

// Helper function to convert interface{} to float64
func getFloat(value interface{}) (float64, error) {
    switch v := value.(type) {
    case float64:
        return v, nil
    case float32:
        return float64(v), nil
    case int:
        return float64(v), nil
    case int64:
        return float64(v), nil
    case string:
        return strconv.ParseFloat(v, 64)
    default:
        return 0, fmt.Errorf("cannot convert %T to float64", value)
    }
}
```

### 2. Weather Information Tool

```go
package tools

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
    "time"
    
    "github.com/kunalkushwaha/agenticgokit/core"
)

// WeatherTool provides weather information for a location
type WeatherTool struct {
    apiKey     string
    httpClient *http.Client
}

func NewWeatherTool(apiKey string) *WeatherTool {
    return &WeatherTool{
        apiKey: apiKey,
        httpClient: &http.Client{
            Timeout: 10 * time.Second,
        },
    }
}

func (t *WeatherTool) Name() string {
    return "weather"
}

func (t *WeatherTool) Description() string {
    return "Gets current weather information for a specified location"
}

func (t *WeatherTool) ParameterSchema() map[string]core.ParameterDefinition {
    return map[string]core.ParameterDefinition{
        "location": {
            Type:        "string",
            Description: "City name or location",
            Required:    true,
        },
        "units": {
            Type:        "string",
            Description: "Temperature units (metric, imperial, standard)",
            Required:    false,
            Default:     "metric",
            Enum:        []string{"metric", "imperial", "standard"},
        },
    }
}

func (t *WeatherTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    // Extract location
    location, ok := params["location"].(string)
    if !ok || location == "" {
        return nil, fmt.Errorf("location must be a non-empty string")
    }
    
    // Extract units (with default)
    units := "metric"
    if unitsParam, ok := params["units"].(string); ok && unitsParam != "" {
        units = unitsParam
    }
    
    // Build API URL
    apiURL := fmt.Sprintf(
        "https://api.openweathermap.org/data/2.5/weather?q=%s&units=%s&appid=%s",
        url.QueryEscape(location),
        url.QueryEscape(units),
        url.QueryEscape(t.apiKey),
    )
    
    // Create request
    req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }
    
    // Execute request
    resp, err := t.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("weather API request failed: %w", err)
    }
    defer resp.Body.Close()
    
    // Check status code
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("weather API returned status %d", resp.StatusCode)
    }
    
    // Parse response
    var weatherData map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&weatherData); err != nil {
        return nil, fmt.Errorf("failed to parse weather data: %w", err)
    }
    
    // Extract relevant information
    result := map[string]interface{}{
        "location": location,
        "units":    units,
    }
    
    // Extract temperature
    if main, ok := weatherData["main"].(map[string]interface{}); ok {
        if temp, ok := main["temp"].(float64); ok {
            result["temperature"] = temp
        }
        if humidity, ok := main["humidity"].(float64); ok {
            result["humidity"] = humidity
        }
    }
    
    // Extract weather description
    if weather, ok := weatherData["weather"].([]interface{}); ok && len(weather) > 0 {
        if firstWeather, ok := weather[0].(map[string]interface{}); ok {
            if description, ok := firstWeather["description"].(string); ok {
                result["description"] = description
            }
        }
    }
    
    return result, nil
}
```

## Best Practices

### 1. Tool Design Principles

- **Single Responsibility**: Each tool should do one thing well
- **Clear Interface**: Define clear parameter schemas and return values
- **Robust Error Handling**: Provide meaningful error messages
- **Statelessness**: Prefer stateless tools when possible
- **Security**: Validate inputs and limit access to sensitive operations
- **Performance**: Optimize for speed and resource usage
- **Documentation**: Provide clear descriptions and examples

### 2. Parameter Schema Design

- **Required vs. Optional**: Only mark parameters as required if they're truly necessary
- **Defaults**: Provide sensible defaults for optional parameters
- **Validation**: Use enums and type constraints to prevent errors
- **Documentation**: Clearly describe each parameter's purpose and format
- **Consistency**: Use consistent naming and types across tools

### 3. Error Handling

- **Specific Errors**: Return specific error messages that explain what went wrong
- **Context**: Include context in error messages (e.g., parameter names)
- **Recovery**: Implement graceful recovery from transient errors
- **Logging**: Log errors for debugging and monitoring
- **User-Friendly**: Make error messages understandable to end users

### 4. Testing Tools

```go
package tools_test

import (
    "context"
    "testing"
    
    "github.com/kunalkushwaha/agenticgokit/tools"
    "github.com/stretchr/testify/assert"
)

func TestCalculatorTool(t *testing.T) {
    calculator := tools.NewCalculatorTool()
    ctx := context.Background()
    
    // Test addition
    result, err := calculator.Execute(ctx, map[string]interface{}{
        "operation": "add",
        "a":         5,
        "b":         3,
    })
    
    assert.NoError(t, err)
    assert.Equal(t, 8.0, result.(map[string]interface{})["result"])
    
    // Test division by zero
    _, err = calculator.Execute(ctx, map[string]interface{}{
        "operation": "divide",
        "a":         5,
        "b":         0,
    })
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "division by zero")
}
```

## Conclusion

Developing custom tools is a powerful way to extend agent capabilities in AgenticGoKit. By following the patterns and best practices in this tutorial, you can create tools that are robust, secure, and easy to use.

Key takeaways:
- Implement the `Tool` interface for all custom tools
- Use parameter schemas to define and validate inputs
- Follow best practices for error handling and testing
- Design tools with single responsibility and clear interfaces

## Next Steps

- [Tool Integration](tool-integration.md) - Learn how to integrate tools with agents
- [Advanced Tool Patterns](advanced-tool-patterns.md) - Explore complex tool usage patterns
- [Error Handling](../core-concepts/error-handling.md) - Implement robust error handling

## Further Reading

- [API Reference: MCP](../../reference/api/agent.md#mcp)
- [Examples: Tool Usage](../../examples/)
- [Advanced Patterns](../advanced/README.md) - Advanced multi-agent patterns
