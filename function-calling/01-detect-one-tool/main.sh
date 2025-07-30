#!/bin/bash
BASE_URL=${MODEL_RUNNER_BASE_URL:-http://localhost:12434/engines/llama.cpp/v1}
#MODEL=${MODEL_QWEN2_5_MEDIUM:-"ai/qwen2.5:3B-F16"}
#MODEL=${MODEL_QWEN2_5_LARGE:-"ai/qwen2.5:latest"}
#MODEL=${MODEL_LUCY:-"hf.co/menlo/lucy-128k-gguf:q4_k_m"}
#MODEL=${MODEL_QWEN3_LARGE:-"ai/qwen3:latest"}
#MODEL=${MODEL_GEMMA3:-"ai/gemma3:latest"}
MODEL=${MODEL_GEMMA3_TINY:-"ai/gemma3:1B-Q4_K_M"}
#MODEL=${MODEL_SMOLLM3:-"ai/smollm3"}

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
Tell me why the sky is blue 
and then say hello to Jean-Luc Picard. 
I love pineapple pizza!
EOM

read -r -d '' DATA <<- EOM
{
  "model": "${MODEL}",
  "options": {
    "temperature": 0.0
  },
  "messages": [
    {
      "role": "user",
      "content": "${USER_MESSAGE}"
    }
  ],
  "tools": ${TOOLS},
  "parallel_tool_calls": false,
}
EOM

# Remove newlines from DATA 
DATA=$(echo ${DATA} | tr -d '\n')

JSON_RESULT=$(curl --silent ${BASE_URL}/chat/completions \
    -H "Content-Type: application/json" \
    -d "${DATA}"
)

echo -e "\n📝 Raw JSON response:\n"
echo "${JSON_RESULT}" | jq '.'

echo -e "\n🔍 Extracted function calls:\n"
echo "${JSON_RESULT}" | jq -r '.choices[0].message.tool_calls[]? | "Function: \(.function.name), Args: \(.function.arguments)"'
echo -e "\n"
