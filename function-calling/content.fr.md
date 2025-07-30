# Les petis modèles sont particuliérement mauvais pour le function calling, mais est-ce une fatalité ?

Expliquer aussi ce q'est le function callin: detecter dans un texte ....
faire un schéma ?

## Premier "tool call"


```bash
#!/bin/bash
DMR_BASE_URL=http://localhost:12434/engines/llama.cpp/v1
MODEL="ai/gemma3:latest"

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
  "tool_choice": "auto"
}
EOM

# Remove newlines from DATA 
DATA=$(echo ${DATA} | tr -d '\n')

JSON_RESULT=$(curl --silent ${DMR_BASE_URL}/chat/completions \
    -H "Content-Type: application/json" \
    -d "${DATA}"
)

echo "${JSON_RESULT}" | jq '.'
```



```bash
{
  "choices": [
    {
      "finish_reason": "tool_calls",
      "index": 0,
      "message": {
        "role": "assistant",
        "content": null,
        "tool_calls": [
          {
            "type": "function",
            "function": {
              "name": "say_hello",
              "arguments": "{\"name\":\"Jean-Luc Picard\"}"
            },
            "id": "r4A317EsTt415D2KclIBopexB55bjSLN"
          }
        ]
      }
    }
  ],
  "created": 1753848898,
  "model": "ai/gemma3:latest",
  "system_fingerprint": "b1-79e0b68",
  "object": "chat.completion",
  "usage": {
    "completion_tokens": 46,
    "prompt_tokens": 415,
    "total_tokens": 461
  },
  "id": "chatcmpl-P6On9Isd9ucANRp7s7iyV3k1OMGsxlVP",
  "timings": {
    "prompt_n": 415,
    "prompt_ms": 1077.902,
    "prompt_per_token_ms": 2.59735421686747,
    "prompt_per_second": 385.00717133839623,
    "predicted_n": 46,
    "predicted_ms": 1403.914,
    "predicted_per_token_ms": 30.51986956521739,
    "predicted_per_second": 32.765539769530044
  }
}
```

Alors avec un même modèle, parfois cela fonctionne, parfois non.



J'ai fait quelques tests avec certqins modèles, et il y a des modèles qui sont meilleurs que d'autres pour le "mono" function calling.

Mon test, pour chaque modèle, est 
```text
userQuestion := "Tell me why the sky is blue and then say hello to Jean-Luc Picard. I love pineapple pizza!"
userQuestion = "Where is Bob? Add 2 and 3. What is the capital of France?"
userQuestion = "The best pizza topping is pineapple. What is the capital of France? I love cooking." // trouver la capitale de la france n'existe pas dans la liste des tools à détecter
```

```text
- Model: ai/qwen2.5:0.5B-F16 Final Score: 6 Percentage: 66.66666666666666 %
- Model: ai/qwen2.5:1.5B-F16 Final Score: 3 Percentage: 33.33333333333333 %
- Model: ai/qwen2.5:3B-F16 Final Score: 9 Percentage: 100 %
- Model: ai/qwen2.5:latest Final Score: 9 Percentage: 100 %
- Model: ai/qwen3:0.6B-F16 Final Score: 3 Percentage: 33.33333333333333 %
- Model: hf.co/menlo/lucy-128k-gguf:q4_k_m Final Score: 9 Percentage: 100 %
- Model: ai/qwen3:latest Final Score: 9 Percentage: 100 %
- Model: ai/gemma3:latest Final Score: 6 Percentage: 66.66666666666666 %
- Model: ai/gemma3:1B-Q4_K_M Final Score: 9 Percentage: 100 %
- Model: ai/llama3.2:latest Final Score: 3 Percentage: 33.33333333333333 %
- Model: ai/mistral:latest Final Score: 3 Percentage: 33.33333333333333 %
- Model: ai/smollm3:latest Final Score: 9 Percentage: 100 %
``




## Loop
https://www.perplexity.ai/search/9a6e5a68-70e0-44b8-a457-39ffc489a719?login-source=oneTapThread&login-new=false

## Problématique

TODO: résumé du post d'Ignasi Lopez Luna


- On vérifie
- + Parallel tool calls

## Qu'est ce que le parallel tool calls

## Mais j'ai besoin du function calling pour des petits modèles

Pourquoi?
Rapidité
petites machines

### Solution



## Autre possibilité

Focal models