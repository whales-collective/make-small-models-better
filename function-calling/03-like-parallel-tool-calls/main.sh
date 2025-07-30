#!/bin/bash
BASE_URL=${MODEL_RUNNER_BASE_URL:-http://localhost:12434/engines/llama.cpp/v1}
#MODEL=${MODEL_QWEN2_5_MEDIUM:-"ai/qwen2.5:3B-F16"}
#MODEL=${MODEL_QWEN2_5_LARGE:-"ai/qwen2.5:latest"}
#MODEL=${MODEL_LUCY:-"hf.co/menlo/lucy-128k-gguf:q4_k_m"}
#MODEL=${MODEL_QWEN3_LARGE:-"ai/qwen3:latest"}
MODEL=${MODEL_GEMMA3:-"ai/gemma3:latest"}
#MODEL=${MODEL_GEMMA3_TINY:-"ai/gemma3:1B-Q4_K_M"}
#MODEL=${MODEL_SMOLLM3:-"ai/smollm3"}

read -r -d '' JSON_SCHEMA <<- EOM
{
  "json_schema": {
    "name": "function_calls",
    "description": "Function calls data structure",
    "schema": {
      "additionalProperties": false,
      "properties": {
        "function_calls": {
          "description": "Array of function calls to execute",
          "items": {
            "additionalProperties": false,
            "properties": {
              "arguments": {
                "description": "The arguments to pass to the function",
                "type": "object"
              },
              "name": {
                "description": "The name of the function to call",
                "type": "string"
              }
            },
            "required": [
              "name",
              "arguments"
            ],
            "type": "object"
          },
          "type": "array"
        }
      },
      "required": [
        "function_calls"
      ],
      "type": "object"
    }
  },
  "type": "json_schema"
}
EOM

# Tools index in JSON format
read -r -d '' TOOLS <<- EOM
[
  {
    "type": "function",
    "function": {
      "name": "add_two_numbers",
      "description": "Add two numbers together",
      "parameters": {
        "type": "object",
        "properties": {
          "number1": {
            "type": "number",
            "description": "The first number"
          },
          "number2": {
            "type": "number",
            "description": "The second number"
          }
        },
        "required": ["number1", "number2"]
      }
    }
  },
  {
    "type": "function",
    "function": {
      "name": "say_hello",
      "description": "Say hello to the given name",
      "parameters": {
        "type": "object",
        "properties": {
          "name": {
            "type": "string",
            "description": "The name to greet"
          }
        },
        "required": ["name"]
      }
    }
  }
]
EOM

read -r -d '' USER_MESSAGE <<- EOM
Tell me why the sky is blue and then say hello to Jean-Luc Picard. I love pineapple pizza!
Where is Bob? Add 2 and 3. What is the capital of France?
Say hello for me to Bob Morane and to Sam with fancy emojis. Add 5 and 10.
The best pizza topping is pineapple. What is the capital of France? I love cooking.
EOM

USER_MESSAGE=$(echo "${USER_MESSAGE}" | tr '\n' ' ')

read -r -d '' SYSTEM_INSTRUCTIONS <<- EOM
You are an AI assistant that identifies tool calls from user input. 

Available tools:
[AVAILABLE_TOOLS]
${TOOLS}
[/AVAILABLE_TOOLS]


Instructions:
1. Find ALL tool calls in the user input
2. For "add_two_numbers": extract numbers from patterns like "Add X and Y"
3. For "say_hello": extract names from patterns like "Say hello to NAME"
4. Only use exact tool names: "add_two_numbers" and "say_hello"

Respond with JSON only:
{
  "function_calls": [
    {"name": "tool_name", "arguments": {...}}
  ]
}

If no tools found: {"function_calls": []}
EOM

SYSTEM_INSTRUCTIONS=$(echo "${SYSTEM_INSTRUCTIONS}" | tr '\n' ' ' | sed 's/"/\\"/g')
USER_MESSAGE_ESCAPED=$(echo "${USER_MESSAGE}" | sed 's/"/\\"/g')

read -r -d '' DATA <<- EOM
{
  "model": "${MODEL}",
  "options": {
    "temperature": 0.0
  },
  "messages": [
    {
      "role": "system",
      "content": "${SYSTEM_INSTRUCTIONS}"
    },
    {
      "role": "user",  
      "content": "${USER_MESSAGE_ESCAPED}"
    }
  ]
}
EOM

JSON_RESULT=$(curl --silent ${BASE_URL}/chat/completions \
    -H "Content-Type: application/json" \
    -d "${DATA}"
)

echo -e "\n📝 Raw JSON response:\n"
echo "${JSON_RESULT}" | jq '.'

#echo "🔍 Extracted function calls:"
# echo "${JSON_RESULT}" | jq -r '.choices[0].message.tool_calls[]? | "Function: \(.function.name), Args: \(.function.arguments)"'

echo -e "\n📝 Extracted content from the response:\n"
echo "${JSON_RESULT}" | jq -r '.choices[0].message.content'
echo -e "\n"

