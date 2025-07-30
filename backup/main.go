package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
)

// Generate JSON Format for the response of the chat completion
func GenerateResponseFormat() openai.ChatCompletionNewParamsResponseFormatUnion {

	return openai.ChatCompletionNewParamsResponseFormatUnion{
		OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
			Type: "json_schema",
			JSONSchema: shared.ResponseFormatJSONSchemaJSONSchemaParam{
				Name:        "function_calls",
				Description: openai.String("Function calls data structure"),
				Schema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"function_calls": map[string]any{
							"type": "array",
							"items": map[string]any{
								"type": "object",
								"properties": map[string]any{
									"name": map[string]any{
										"type":        "string",
										"description": "The name of the function to call",
									},
									"arguments": map[string]any{
										"type":        "object",
										"description": "The arguments to pass to the function",
									},
								},
								"required":             []string{"name", "arguments"},
								"additionalProperties": false,
							},
							"description": "Array of function calls to execute",
						},
					},
					"required":             []string{"function_calls"},
					"additionalProperties": false,
				},
			},
		},
	}
}

// GetToolsIndex returns the list of tools available for the model
// In a usual scenario, it's used with a tool completion (a chat completion with tools enabled)
func GetToolsIndex() []openai.ChatCompletionToolParam {
	sayHelloTool := openai.ChatCompletionToolParam{
		Function: openai.FunctionDefinitionParam{
			Name:        "say_hello",
			Description: openai.String("Say hello to the given person name"),
			Parameters: openai.FunctionParameters{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]string{
						"type": "string",
					},
				},
				"required": []string{"name"},
			},
		},
	}
	addTwoNumbersTool := openai.ChatCompletionToolParam{
		Function: openai.FunctionDefinitionParam{
			Name:        "add_two_numbers",
			Description: openai.String("Add two numbers together"),
			Parameters: openai.FunctionParameters{
				"type": "object",
				"properties": map[string]any{
					"number1": map[string]any{
						"type": "number",
					},
					"number2": map[string]any{
						"type": "number",
					},
				},
				"required": []string{"number1", "number2"},
			},
		},
	}
	return []openai.ChatCompletionToolParam{
		sayHelloTool,
		addTwoNumbersTool,
	}
}

// GetPrettyPrintJSON returns a pretty-printed JSON string from the given data
func GetPrettyPrintJSON(data any) (string, error) {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error marshalling data to JSON: %w", err)
	}
	return string(jsonBytes), nil
}

func main() {
	ctx := context.Background()

	baseURL := os.Getenv("MODEL_RUNNER_BASE_URL")

	type Model struct {
		Name  string
		Score int
		SuccessWithToolCalls bool
	}

	models := []Model{
		//{Name: os.Getenv("MODEL_QWEN2_5_TINY"), Score: 0},
		{Name: os.Getenv("MODEL_QWEN2_5_SMALL"), Score: 0},
		//{Name: os.Getenv("MODEL_QWEN2_5_MEDIUM"), Score: 0},
		//{Name: os.Getenv("MODEL_QWEN2_5_LARGE"), Score: 0},
		//{Name: os.Getenv("MODEL_QWEN3_TINY"), Score: 0},
		//{Name: os.Getenv("MODEL_LUCY"), Score: 0},
		//{Name: os.Getenv("MODEL_QWEN3_LARGE"), Score: 0},
	}

	client := openai.NewClient(
		option.WithBaseURL(baseURL),
		option.WithAPIKey(""),
	)

	responseFormat := GenerateResponseFormat()
	toolsIndex := GetToolsIndex()

	// Generate the system instructions content to guide the model to detect all tool calls
	systemMessageContent, err := GenerateSystemInstructionFromToolsCatalog(responseFormat, toolsIndex)
	if err != nil {
		fmt.Println("Error generating system message content:", err)
		return
	}
	// Save the system message content to a file for debugging
	if err := os.WriteFile("system_message_content.txt", []byte(systemMessageContent), 0644); err != nil {
		fmt.Println("Error writing system message content to file:", err)
		return
	}


	detectToolCall := func(model string, userQuestion string, theNumberOfExpectedCalls int) int {

		success := 0

		params := openai.ChatCompletionNewParams{
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(systemMessageContent),
				openai.UserMessage(userQuestion),
			},
			//ParallelToolCalls: openai.Bool(true),
			//Tools:             GetToolsIndex(),
			Model:          model,
			Temperature:    openai.Opt(0.0),
			ResponseFormat: responseFormat,
			TopP:           openai.Opt(0.8),
		}

		// Create context with 20-second timeout
		timeoutCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
		defer cancel()

		// Make completion request
		completion, err := client.Chat.Completions.New(timeoutCtx, params)
		if err != nil {
			fmt.Println("🔴 Model:", model, "Error:", err)
			return success
		}

		if len(completion.Choices) == 0 {
			fmt.Println("🔴 Model:", model, "No choices returned from chat completion")
			return success
		}
		result := completion.Choices[0].Message.Content
		if result == "" {
			fmt.Println("🔴 Model:", model, "No content returned from chat completion")
			return success
		}

		type Command struct {
			Name      string         `json:"name"`
			Arguments map[string]any `json:"arguments"`
		}

		type FunctionCalls struct {
			FunctionCalls []Command `json:"function_calls"`
		}

		var commands FunctionCalls

		var prettyJSON map[string]interface{}
		if err := json.Unmarshal([]byte(result), &prettyJSON); err == nil {
			if formatted, err := json.MarshalIndent(prettyJSON, "", "  "); err == nil {
				fmt.Println("Pretty JSON result:\n", string(formatted))
			}
		} else {
			fmt.Println("Raw JSON result:", result)
		}

		errJson := json.Unmarshal([]byte(result), &commands)
		if errJson != nil {
			fmt.Println("🔴 Model:", model, "Error unmarshalling JSON result:", errJson)
			return success
		}
		if len(commands.FunctionCalls) == 0 {
			fmt.Println("🔴 Model:", model, "No commands found in the JSON result, expected:", theNumberOfExpectedCalls)
			return success
		}

		// search if command.Name exists in the tools index
		functions := map[string]func(any) (any, error){
			"add_two_numbers": func(args any) (any, error) {
				number1, _ := args.(map[string]any)["number1"].(float64)
				number2, _ := args.(map[string]any)["number2"].(float64)
				// if !ok1 || !ok2 {
				// 	return nil, fmt.Errorf("invalid arguments for add_two_numbers")
				// }
				return number1 + number2, nil
			},

			"say_hello": func(args any) (any, error) {
				name, ok := args.(map[string]any)["name"].(string)
				if !ok {
					return nil, fmt.Errorf("invalid arguments for say_hello")
				}
				return fmt.Sprintf("Hello, %s!", name), nil
			},
		}

		for _, command := range commands.FunctionCalls {
			fmt.Println("  - Command:", command.Name, "with arguments:", command.Arguments)
			if function, exists := functions[command.Name]; exists {
				result, err := function(command.Arguments)
				if err != nil {
					fmt.Println("🔴 Model:", model, "Error executing command", command.Name, "with arguments", command.Arguments, ":", err)
					//return success
					//continue // Skip to the next command 🤔
				} else {
					fmt.Println("🟢 Model:", model, " Executed command: ", command.Name, " with arguments: ", command.Arguments, " and result: ", result)
					success += 1
				}

			} else {
				fmt.Println("🔴 Model:", model, "No function defined for command:", command.Name)
			}
		}

		return success

	}

	numberOfIterations := 1

	userQuestion := `
	Tell me why the sky is blue and then say hello to Jean-Luc Picard. I love pineapple pizza!
	Where is Bob? Add 2 and 3. What is the capital of France?
	Say hello for me to Bob Morane and to Sam with fancy emojis. Add 5 and 10.
	The best pizza topping is pineapple. What is the capital of France? I love cooking.
	`
	nbToolCallExpectedPerIteration := 5
	//nbToolCallExpectedPerIteration := 1

	for i := range numberOfIterations {
		fmt.Println(i, ". Running detection for models...")

		for j, model := range models {
			fmt.Println("🔵 Model:", model)
			success := detectToolCall(model.Name, userQuestion, nbToolCallExpectedPerIteration)

			models[j].Score += success
			fmt.Println("🟣 Model:", model.Name, "Score:", models[j].Score)

		}
	}

	fmt.Println("Final scores:")
	for _, model := range models {
		fmt.Println("- Model:", model.Name, "Final Score:", model.Score, "Percentage:", float64(model.Score)/float64(numberOfIterations*nbToolCallExpectedPerIteration)*100, "%")
	}
	fmt.Println("Done!")

}

func GenerateSystemInstructionFromToolsCatalog(responseFormat openai.ChatCompletionNewParamsResponseFormatUnion, toolsIndex []openai.ChatCompletionToolParam) (string, error) {
	systemContentIntroduction := `You are an AI assistant with access to various tools. 
	Your task is to analyze user input and identify ALL possible tool calls that can be made.
	IMPORTANT: You must process the ENTIRE user input and identify ALL tool calls, not just the first few. 
	Each line or request in the user input should be analyzed separately.
	You have access to the following tools:
	`

	// make a JSON String from the content of tools
	toolsJson, err := json.Marshal(toolsIndex)
	if err != nil {
		return "", err
	}
	toolsContent := "\n[AVAILABLE_TOOLS]\n" + string(toolsJson) + "\n[/AVAILABLE_TOOLS]\n"

	systemContentInstructions := `INSTRUCTIONS:
	1. Read the ENTIRE user input carefully
	2. Process each line/request separately
	3. For each request, check if it matches any tool description from the tools index in the [AVAILABLE_TOOLS] section
	4. If multiple tool calls are needed, include ALL of them in your response
	5. NEVER stop processing until you've analyzed the complete input

	TOOL MATCHING RULES:
	- Match tool calls based on the "description" field of each tool
	- Use the exact "Name" field from the tool definition -> be focused on the "name" field
	- Provide all required arguments as specified in the tool's properties fields
	- If the number of arguments is not the same as the tool's properties, ignore that tool call and do not include it in the response
	- If the tool call is not found in the tools index, ignore it and do not include it in the response

	USAGE OF add_two_numbers:
	For the add_two_numbers tool, extract number1 and number2 from these patterns:
	- "Add X and Y" → number1: X, number2: Y
	- "X + Y" → number1: X, number2: Y  
	- "Add X to Y" → number1: X, number2: Y
	- "Sum of X and Y" → number1: X, number2: Y
	- Always use the EXACT numbers found in the text, not random values

	USAGE OF say_hello:
	For say_hello tool, extract names from these patterns:
	- "Say hello to NAME" → name: "NAME"
	- "Hello NAME" → name: "NAME"
	- "Greet NAME" → name: "NAME"
	- "Say hi to NAME" → name: "NAME"
	- Extract the EXACT name mentioned after "to", "hello", or greeting words
	- Names can include spaces (e.g., "Jean-Luc Picard", "Bob Morane")

	STRICT VALIDATION:
	- ONLY use tools that exist in the [AVAILABLE_TOOLS] section
	- Tool names MUST match exactly: "say_hello" and "add_two_numbers" ONLY
	- Do NOT create calls for non-existent tools like "send_message", "operation", "greetings", etc.
	- If a request cannot be fulfilled by available tools, ignore it completely


	RESPONSE FORMAT:
	When you find tool calls, respond with a JSON array containing ALL identified tool calls:

	`
	// jsonStringResponseFormat, err := GenerateResponseFormat().MarshalJSON()
	// if err != nil {
	// 	fmt.Println("Error marshalling response format:", err)
	// 	return "", err
	// }

	jsonStringResponseFormat, _ := GetPrettyPrintJSON(responseFormat)

	endOfSystemContentInstructions := `
	If no tool calls are found, respond with an empty array: []

	CRITICAL: You must analyze the COMPLETE user input and identify ALL possible tool calls. Do not stop after finding the first few matches.
	`
	//return systemContentIntroduction + toolsContent + systemContentInstructions  + endOfSystemContentInstructions, nil

	return systemContentIntroduction + toolsContent + systemContentInstructions + string(jsonStringResponseFormat) + endOfSystemContentInstructions, nil
}

/*
	"function_calls": [
		{
			"name": "<exact_tool_name_from_catalog>",
			"arguments": {
				"<parameter_name>": "<parameter_value>"
			}
		},
		{
			"name": "<next_tool_name>",
			"arguments": {
				"<parameter_name>": "<parameter_value>"
			}
		},
		etc...
	]


	{
		"type": "json_schema",
		"json_schema": {
		"name": "function_calls",
		"description": "Function calls data structure",
		"schema": {
			"type": "object",
			"properties": {
			"function_calls": {
				"type": "array",
				"items": {
				"type": "object",
				"properties": {
					"name": {
					"type": "string",
					"description": "The name of the function to call"
					},
					"arguments": {
					"type": "object",
					"description": "The arguments to pass to the function"
					}
				},
				"required": ["name", "arguments"],
				"additionalProperties": false
				},
				"description": "Array of function calls to execute"
			}
			},
			"required": ["function_calls"],
			"additionalProperties": false
		}
		}
	}



*/
