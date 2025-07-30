#!/bin/bash
BASE_URL=http://localhost:12434/engines/llama.cpp/v1
#MODEL="ai/smollm3:latest"
#MODEL="ai/qwen2.5:3B-F16"
#MODEL="ai/qwen2.5:latest"
MODEL="ai/gemma3:latest "

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

read -r -d '' SYSTEM_INSTRUCTIONS <<- EOM
You are an AI assistant with access to various tools. 
Your task is to analyze user input and identify ALL possible tool calls that can be made.
IMPORTANT: You must process the ENTIRE user input and identify ALL tool calls, not just the first few. 
Each line or request in the user input should be analyzed separately.
You have access to the following tools:
[AVAILABLE_TOOLS]
${TOOLS}
[/AVAILABLE_TOOLS]

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


RESPONSE FORMAT:
When you find tool calls, respond with a JSON array containing ALL identified tool calls:
${JSON_SCHEMA}

If no tool calls are found, respond with an empty array: []

CRITICAL: You must analyze the COMPLETE user input and identify ALL possible tool calls. Do not stop after finding the first few matches.
EOM

SYSTEM_INSTRUCTIONS=$(echo ${SYSTEM_INSTRUCTIONS} | tr -d '\n')

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
      "content": "${USER_MESSAGE}"
    }
  ],
  "response_format": {
    "type": "json_object",
    ${JSON_SCHEMA},
    "strict": true
  }

}
EOM

# Remove newlines from DATA 
DATA=$(echo ${DATA} | tr -d '\n')

echo "${DATA}"

JSON_RESULT=$(curl --silent ${BASE_URL}/chat/completions \
    -H "Content-Type: application/json" \
    -d "${DATA}"
)

echo "${JSON_RESULT}" 

# echo "📝 Raw JSON response:"
# echo "${JSON_RESULT}" | jq '.'

# echo "🔍 Extracted function calls:"
# echo "${JSON_RESULT}" | jq -r '.choices[0].message.tool_calls[]? | "Function: \(.function.name), Args: \(.function.arguments)"'

# echo "📝 Extracted content from the response:"
# echo "${JSON_RESULT}" | jq -r '.choices[0].message.content'

