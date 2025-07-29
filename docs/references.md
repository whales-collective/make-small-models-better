
## Tool Calling with Local LLMs: A Practical Evaluation

**Source:** [Docker Blog](https://www.docker.com/blog/local-llm-tool-calling-a-practical-evaluation/)  
**Author:** Ignasi Lopez Luna  
**Date:** June 30, 2025

### Summary
Docker's comprehensive evaluation of local LLM performance for tool calling scenarios, testing 21 models across 3,570 test cases. The study addresses the common question: "Which local model should I use for tool calling?" through systematic testing using a custom framework called `model-test`.

### Methodology
- **Initial Testing:** Manual testing with `chat2cart` shopping assistant revealed common issues with local models
- **Scalable Framework:** Built `model-test` for repeatable, measurable testing
- **Test Structure:** Real-world scenarios with multiple valid tool call sequences
- **Agent Loop:** Up to 5 rounds of interaction simulation
- **Hardware:** MacBook Pro M4 Max, 128GB RAM

### Key Findings

#### Top Performers (Tool Selection F1 Score)
1. **GPT-4**: 0.974 (5s latency) - benchmark reference
2. **Qwen 3 (14B-Q4_K_M)**: 0.971 (142s latency) - best local model
3. **Qwen 3 (14B-Q6_K)**: 0.943
4. **Claude 3 Haiku**: 0.933 (3.56s latency)
5. **Qwen 3 (8B-F16)**: 0.933 (84s latency) - best speed/accuracy balance

#### Common Local Model Issues
- **Eager invocation:** Calling tools for simple greetings
- **Wrong tool selection:** Searching when should add, removing from empty cart
- **Invalid arguments:** Missing or malformed parameters
- **Ignored responses:** Failing to respond to tool output

#### Quantization Impact
No significant difference observed between quantized and non-quantized variants, suggesting quantization is beneficial for resource reduction without accuracy loss.

### Test Metrics
- **Tool Invocation:** Did model realize tool was needed?
- **Tool Selection:** Correct tool choice and usage
- **Parameter Accuracy:** Correct tool call arguments
- **F1 Score:** Harmonic mean of precision and recall

### Conclusions

#### Recommendations
1. **Maximum Accuracy:** Qwen 3 (14B) or Qwen 3 (8B) - best local options
2. **Speed/Performance Balance:** Qwen 2.5 - good for real-time experiences
3. **Resource-Constrained:** LLaMA 3 Groq 7B - modest performance, low compute

#### Key Insights
- **Qwen family dominates** open-source tool calling performance
- **Trade-off exists** between accuracy and latency
- **Tool calling is core** to real-world GenAI applications
- **Testing framework** eliminates guesswork in model selection

### Underperformers
- **Watt 8B (quantized):** 0.484 F1 - struggled with parameter accuracy
- **LLaMA XLam 8B:** 0.570 F1 - missed correct tool paths

**Relevance to Project:** Critical insights for selecting and optimizing small models for tool calling scenarios. Demonstrates that model size isn't everything - Qwen 8B outperformed many larger models. The testing methodology could be adapted for evaluating small model improvements.