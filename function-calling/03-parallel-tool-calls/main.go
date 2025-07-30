package main

import (
	"context"
	"fmt"
	"os"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

func GetToolsIndex() []openai.ChatCompletionToolParam {
	sayHelloTool := openai.ChatCompletionToolParam{
		Function: openai.FunctionDefinitionParam{
			Name:        "say_hello",
			Description: openai.String("Say hello to the given person name"),
			Parameters: openai.FunctionParameters{
				"type": "object",
				"properties": map[string]interface{}{
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
				"properties": map[string]interface{}{
					"number1": map[string]string{
						"type": "number",
					},
					"number2": map[string]string{
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

// MODEL_RUNNER_BASE_URL=http://localhost:12434 go run main.go
func main() {
	ctx := context.Background()

	// Docker Model Runner base URL
	baseURL := os.Getenv("MODEL_RUNNER_BASE_URL")

	type Model struct {
		Name  string
		Score int
	}

	models := []Model{
		// {Name: os.Getenv("MODEL_QWEN2_5_TINY"), Score: 0},
		// {Name: os.Getenv("MODEL_QWEN2_5_SMALL"), Score: 0},
		//{Name: os.Getenv("MODEL_QWEN2_5_MEDIUM"), Score: 0},
		// {Name: os.Getenv("MODEL_QWEN2_5_LARGE"), Score: 0},
		// {Name: os.Getenv("MODEL_QWEN3_TINY"), Score: 0},
		{Name: os.Getenv("MODEL_LUCY"), Score: 0},
		// {Name: os.Getenv("MODEL_QWEN3_LARGE"), Score: 0},
		// {Name: os.Getenv("MODEL_GEMMA3"), Score: 0},
		// {Name: os.Getenv("MODEL_GEMMA3_TINY"), Score: 0},
		// {Name: os.Getenv("MODEL_LLAMA3_2"), Score: 0},
		// {Name: os.Getenv("MODEL_MISTRAL"), Score: 0},
		//{Name: os.Getenv("MODEL_SMOLLM3"), Score: 0},
	}

	client := openai.NewClient(
		option.WithBaseURL(baseURL),
		option.WithAPIKey(""),
	)

	//userQuestion := openai.UserMessage("Say hello to Jean-Luc Picard")
	systemInstructions := `
	You are an AI assistant with access to a tools index. 
	Your task is to analyze user input and identify ALL possible tool calls that can be made.
	IMPORTANT: You must process the ENTIRE user input and identify ALL tool calls, not just the first few. 
	Each line or request in the user input should be analyzed separately.
	You have access to the tools index.

	INSTRUCTIONS:
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

	CRITICAL: You must analyze the COMPLETE user input and identify ALL possible tool calls. Do not stop after finding the first few matches.
	`

	detectToolCall := func(model string, userQuestion string, theNumberOfExpectedCalls int) int {

		success := 0

		params := openai.ChatCompletionNewParams{
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(systemInstructions),
				openai.UserMessage(userQuestion),
			},
			ParallelToolCalls: openai.Bool(true),
			Tools:             GetToolsIndex(),
			Model:             model,
			Temperature:       openai.Opt(0.0),
		}

		// Make completion request
		completion, err := client.Chat.Completions.New(ctx, params)
		if err != nil {
			fmt.Println("🔴 Model:", model, "Error:", err)
			return success
		}

		fmt.Println("🟪 Response:\n", completion.Choices[0].Message.Content)

		toolCalls := completion.Choices[0].Message.ToolCalls

		// Return early if there are no tool calls
		if len(toolCalls) == 0 {
			fmt.Println("🔴 Model:", model, "No function call detected but expected:", theNumberOfExpectedCalls)
			return success
		}

		// Display the tool call(s)
		for _, toolCall := range toolCalls {

			fmt.Println("🟢 Model:", model, "Function call detected:", toolCall.Function.Name, "with arguments:", toolCall.Function.Arguments)
			success += 1
		}
		// if success == theNumberOfExpectedCalls {
		// 	fmt.Println("🟢 Model:", model, "All expected function calls detected:", theNumberOfExpectedCalls)
		// 	success = 1
		// } else {
		// 	fmt.Println("🔴 Model:", model, "Expected", theNumberOfExpectedCalls, "but got:", success)
		// 	success = 0
		// }
		return success

	}

	numberOfIterations := 1

	userQuestion := `
	Tell me why the sky is blue and then say hello to Jean-Luc Picard. I love pineapple pizza!
	Where is Bob? Add 2 and 3. What is the capital of France?
	Say hello for me to Bob Morane and to Sam with fancy emojis. Add 5 and 10.
	The best pizza topping is pineapple. What is the capital of France? I love cooking.
	`
	nbToolCallExpectedPerIteration := 2

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
