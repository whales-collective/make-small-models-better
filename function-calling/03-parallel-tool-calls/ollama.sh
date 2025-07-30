#!/bin/bash
BASE_URL=http://localhost:11434/v1
#MODEL="ai/smollm3:latest"
MODEL="qwen2.5:3b"
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

echo "📝 Raw JSON response:"
echo "${JSON_RESULT}" | jq '.'

echo "🔍 Extracted function calls:"
echo "${JSON_RESULT}" | jq -r '.choices[0].message.tool_calls[]? | "Function: \(.function.name), Args: \(.function.arguments)"'

echo "📝 Extracted content from the response:"
echo "${JSON_RESULT}" | jq -r '.choices[0].message.content'

