#!/bin/bash

BASE_URL=http://localhost:12434/engines/llama.cpp/v1
MODEL="ai/gemma3:latest"

curl -s ${BASE_URL}/chat/completions \
-H "Content-Type: application/json" \
-d '{
  "model": "'${MODEL}'",
  "messages": [
    {
      "role": "system",
      "content": "You are a helpful assistant. Respond with valid JSON."
    },
    {
      "role": "user", 
      "content": "Say hello to Bob and add 2 + 3"
    }
  ]
}'