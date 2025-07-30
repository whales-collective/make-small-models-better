
# Function Calling with Parallel Tool Calls

This project tests how well different language models can identify and extract function calls from natural language text using a structured JSON schema approach.

## GenerateResponseFormat Function

The `GenerateResponseFormat()` function programmatically creates a JSON schema that enforces the structure of function calls returned by the language model.

### Purpose

Instead of relying on the model's native function calling capabilities, this approach uses a **structured JSON schema** to:

1. **Control output format** - Forces the model to return function calls in a specific JSON structure
2. **Ensure consistency** - All models return the same format regardless of their native capabilities
3. **Enable validation** - The schema validates that responses contain required fields
4. **Support parallel calls** - Multiple function calls can be returned in a single response

### Schema Structure

The generated schema enforces this structure:

```json
{
  "function_calls": [
    {
      "name": "function_name",
      "arguments": {
        "param1": "value1",
        "param2": "value2"
      }
    }
  ]
}
```

### Usage with OpenAI Client

The function returns an `openai.ChatCompletionNewParamsResponseFormatUnion` that is used directly in the chat completion parameters:

```go
responseFormat := GenerateResponseFormat()

params := openai.ChatCompletionNewParams{
    Messages: []openai.ChatCompletionMessageParamUnion{
        openai.SystemMessage(systemPrompt),
        openai.UserMessage(userInput),
    },
    Model:          modelName,
    ResponseFormat: responseFormat,  // Enforces JSON schema
    Temperature:    openai.Opt(0.0),
}

completion, err := client.Chat.Completions.New(ctx, params)
```

### Benefits

- **Model-agnostic**: Works with any model that supports structured output
- **Validation**: Automatically validates response structure
- **Maintainable**: Schema definition is centralized and reusable
- **Type-safe**: Go compiler ensures correct schema construction

### System Prompt Integration

The schema is also included in the system prompt to help the model understand the expected output format, providing both programmatic enforcement and instructional guidance.

