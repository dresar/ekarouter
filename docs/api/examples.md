# EkaRouter API Integration Examples & SDK Recipes

## 1. Overview

EkaRouter is engineered to be 100% wire-compatible with the official OpenAI API standard while offering universal proxying and multi-provider rotation. You can switch your existing applications to EkaRouter by simply altering the **API Base URL** and supplying an **EkaRouter API key**.

---

## 2. cURL Examples

### 2.1 Chat Completion (Standard Non-Streaming)
```bash
curl -X POST "http://localhost:8080/v1/chat/completions" \
  -H "Authorization: Bearer eka_live_1234567890abcdef" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [
      {"role": "system", "content": "You are a concise software architect."},
      {"role": "user", "content": "Explain AES-GCM in two sentences."}
    ],
    "temperature": 0.5
  }'
```

### 2.2 Chat Completion (Streaming Server-Sent Events)
```bash
curl -N -X POST "http://localhost:8080/v1/chat/completions" \
  -H "Authorization: Bearer eka_live_1234567890abcdef" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-5-sonnet-20241022",
    "messages": [
      {"role": "user", "content": "Count from 1 to 5 slowly."}
    ],
    "stream": true
  }'
```

### 2.3 Store an Encrypted Provider Credential
```bash
curl -X POST "http://localhost:8080/api/v1/credentials" \
  -H "Authorization: Bearer eka_live_1234567890abcdef" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Primary Production OpenAI Key",
    "provider_id": "openai",
    "secret_value": "sk-proj-super-secret-key-123456789",
    "environment": "production",
    "priority": 1
  }'
```

### 2.4 Invoking Upstream APIs via Universal Proxy
```bash
# EkaRouter transparently injects the stored credential and applies SSRF filters
curl -X GET "http://localhost:8080/api/v1/proxy/github/user/repos" \
  -H "Authorization: Bearer eka_live_1234567890abcdef" \
  -H "Accept: application/vnd.github.v3+json"
```

---

## 3. Node.js & TypeScript Examples

### 3.1 Official OpenAI SDK Drop-in
```typescript
import OpenAI from 'openai';

// Point the SDK to your EkaRouter instance
const openai = new OpenAI({
  baseURL: 'http://localhost:8080/v1',
  apiKey: process.env.EKAROUTER_API_KEY || 'eka_live_1234567890abcdef',
});

async function run() {
  const completion = await openai.chat.completions.create({
    model: 'gpt-4o',
    messages: [{ role: 'user', content: 'What makes EkaRouter unique?' }],
  });

  console.log(completion.choices[0].message.content);
}

run().catch(console.error);
```

### 3.2 Streaming with OpenAI SDK
```typescript
import OpenAI from 'openai';

const openai = new OpenAI({
  baseURL: 'http://localhost:8080/v1',
  apiKey: process.env.EKAROUTER_API_KEY || 'eka_live_1234567890abcdef',
});

async function streamResponse() {
  const stream = await openai.chat.completions.create({
    model: 'claude-3-5-sonnet-20241022',
    messages: [{ role: 'user', content: 'Write a quick deployment checklist.' }],
    stream: true,
  });

  for await (const chunk of stream) {
    process.stdout.write(chunk.choices[0]?.delta?.content || '');
  }
}

streamResponse().catch(console.error);
```

---

## 4. Python Examples

### 4.1 Official OpenAI Python SDK
```python
from openai import OpenAI

# Initialize client pointing to EkaRouter
client = OpenAI(
    base_url="http://localhost:8080/v1",
    api_key="eka_live_1234567890abcdef",
)

response = client.chat.completions.create(
    model="gpt-4o",
    messages=[
        {"role": "system", "content": "You are an automated code reviewer."},
        {"role": "user", "content": "Suggest improvements for Go error handling."}
    ],
)

print(response.choices[0].message.content)
```

### 4.2 Python Streaming
```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/v1",
    api_key="eka_live_1234567890abcdef",
)

stream = client.chat.completions.create(
    model="deepseek-v3",
    messages=[{"role": "user", "content": "Explain raft consensus in 100 words."}],
    stream=True,
)

for chunk in stream:
    if chunk.choices[0].delta.content is not None:
        print(chunk.choices[0].delta.content, end="", flush=True)
print()
```

---

## 5. Go Native Example

```go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func main() {
	payload := ChatRequest{
		Model: "gpt-4o",
		Messages: []ChatMessage{
			{Role: "user", Content: "Hello from Go native client!"},
		},
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "http://localhost:8080/v1/chat/completions", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer eka_live_1234567890abcdef")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	fmt.Println(string(respBytes))
}
```
