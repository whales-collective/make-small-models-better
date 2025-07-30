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

func main() {
	ctx := context.Background()

	// Docker Model Runner base URL
	baseURL := os.Getenv("MODEL_RUNNER_BASE_URL")

	type Model struct {
		Name  string
		Score int
	}

	models := []Model{
		{Name: os.Getenv("MODEL_QWEN2_5_TINY"), Score: 0}, 
		{Name: os.Getenv("MODEL_QWEN2_5_SMALL"), Score: 0}, 
		{Name: os.Getenv("MODEL_QWEN2_5_MEDIUM"), Score: 0}, 
		{Name: os.Getenv("MODEL_QWEN2_5_LARGE"), Score: 0}, 
		{Name: os.Getenv("MODEL_QWEN3_TINY"), Score: 0}, 
		{Name: os.Getenv("MODEL_LUCY"), Score: 0}, 
		{Name: os.Getenv("MODEL_QWEN3_LARGE"), Score: 0},
		{Name: os.Getenv("MODEL_GEMMA3"), Score: 0},  
		{Name: os.Getenv("MODEL_GEMMA3_TINY"), Score: 0}, 
		{Name: os.Getenv("MODEL_LLAMA3_2"), Score: 0}, 
		{Name: os.Getenv("MODEL_MISTRAL"), Score: 0}, 
		{Name: os.Getenv("MODEL_SMOLLM3"), Score: 0}, 
	}

	client := openai.NewClient(
		option.WithBaseURL(baseURL),
		option.WithAPIKey(""),
	)

	//userQuestion := openai.UserMessage("Say hello to Jean-Luc Picard")

	detectToolCall := func(model string, userQuestion string, theToolNameShouldBe string) int {

		success := 0

		params := openai.ChatCompletionNewParams{
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.UserMessage(userQuestion),
			},
			ParallelToolCalls: openai.Bool(false),
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

		toolCalls := completion.Choices[0].Message.ToolCalls

		// Return early if there are no tool calls
		if len(toolCalls) == 0 {
			if theToolNameShouldBe != "no_tool_call_expected" {
				fmt.Println("🔴 Model:", model, "No function call detected but expected:", theToolNameShouldBe)
			} else {
				fmt.Println("🟢 Model:", model, "No function call (not in the tools index)")
				success = 1
			}
			return success
		}

		// Display the tool call(s)
		for _, toolCall := range toolCalls {

			if toolCall.Function.Name != theToolNameShouldBe {
				fmt.Println("🟡 Model:", model, "Function call detected:", toolCall.Function.Name, "but expected:", theToolNameShouldBe)
			} else {
				fmt.Println("🟢 Model:", model, "Function call detected:", toolCall.Function.Name, "with arguments:", toolCall.Function.Arguments)
				success = 1
			}
		}
		return success

	}

	numberOfIterations := 3

	for i := range numberOfIterations {
		fmt.Println(i, ". Running detection for models...")

		for j, model := range models {
			fmt.Println("🔵 Model:", model)
			userQuestion := "Tell me why the sky is blue and then say hello to Jean-Luc Picard. I love pineapple pizza!"
			success1 := detectToolCall(model.Name, userQuestion, "say_hello")

			userQuestion = "Where is Bob? Add 2 and 3. What is the capital of France?"
			success2 := detectToolCall(model.Name, userQuestion, "add_two_numbers")

			userQuestion = "The best pizza topping is pineapple. What is the capital of France? I love cooking."
			success3 := detectToolCall(model.Name, userQuestion, "no_tool_call_expected")

			models[j].Score += success1 + success2 + success3
			fmt.Println("🟣 Model:", model.Name, "Score:", models[j].Score)

		}
	}
	fmt.Println("Final scores:")
	for _, model := range models {
		fmt.Println("- Model:", model.Name, "Final Score:", model.Score, "Percentage:", float64(model.Score)/float64(numberOfIterations*3)*100, "%")
	}
	fmt.Println("Done!")

}
