#!/bin/bash
: <<'COMMENT'
Install llama.cpp on macOS:
```bash
brew install llama.cpp
llama-server --version
# Load and run the model:
llama-server -hf unsloth/Qwen3-4B-GGUF:Q4_K_M --port 10000 --jinja
```
COMMENT

BASE_URL=http://localhost:10000/v1
MODEL="unsloth_Qwen3-4B-GGUF_Qwen3-4B-Q4_K_M.gguf"


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
  "parallel_tool_calls": true,
  "tool_choice": "auto"
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

echo -e "\n📝 Extracted content from the response:\n"
echo "${JSON_RESULT}" | jq -r '.choices[0].message.content'
echo -e "\n"

