# Baldr 🛡️

> *Balder the Beautiful, God of the summer sun,*
> *Fairest of all the Gods! Light from his forehead beamed,*
> *Runes were upon his tongue, As on the warrior’s sword.*

**Baldr** is a developer-centric infrastructure layer designed to bring observability, cost governance, and safety to the AI development lifecycle. It acts as an intelligent intermediary (Reverse Proxy) between your applications and upstream LLM providers.

---

## 🎯 The Mission

As organizations transition from GenAI experimentation to production, they face the "Black Box" challenge: financial unpredictability, silent quality drift, and compliance risks.

Baldr addresses this by standardizing the messy middle layer of AI development:

* **Unified Observability:** Per-key attribution and performance benchmarking to prevent "budget shocks."
* **Intelligent Guardrails:** Real-time toxic filtering and PII masking *before* data leaves your network.
* **Engineering Excellence:** A high-concurrency Go proxy ensuring sub-millisecond overhead and robust rate-limiting.

---

## 🏗 Architecture

Baldr employs a **Hexagonal Architecture** (Ports & Adapters) to decouple traffic management from guardrail logic.

### Target Data Flow

```mermaid
graph TD
    Client[Developer App] -->|1. HTTP Request| Proxy(Golang Proxy Service)
    
    subgraph "Internal Network"
        Proxy -->|2. Check Prompt| Guard[Python Guardrail Service]
        Guard -->|3. Verdict: OK/FAIL| Proxy
        Proxy -.->|7. Async Log| DB[(Postgres/Redis)]
    end
    
    subgraph "LLM Provider"
        Proxy -->|4. Generate If OK| LLM[OpenAI / Gemini]
        LLM -.->|5. Stream Response| Proxy
    end
    
    Proxy -.->|6. Stream Response| Client
```

### Component Status (v0.1 Walking Skeleton)

|Component|Tech Stack    |Status    |Responsibility|
|---------|:-------------|:---------|:-------------|
|Proxy    |Go 1.24       |🟢 Active |"Traffic control, Auth injection, SSE Streaming."|
|Guardrail|Python 3.12+  |🟢 Active |Basic Semantic Validation (Pydantic).|
|Storage  |Postgres/Redis|🔴 Roadmap|Async logging and Token attribution.|
|Infra    |Docker Compose|🟡 Local  |Local dev (Target: EKS/Terraform)|

---

## 📂 Repository Structure

```bash
.
├── docker-compose.yml   # Local orchestration
├── mise.toml            # Tool versioning (Go, Python, uv)
├── proxy/               # The Orchestrator (Go)
│   ├── cmd/             # Entrypoint (Composition Root)
│   ├── internal/        # Domain Logic & Adapters
│   ├── tests/           # Integration Tests (Testcontainers)
│   └── Dockerfile       # Multi-stage Go build
└── guardrail/           # The Sidecar (Python)
    ├── main.py          # FastAPI Service
    ├── uv.lock          # Dependency Lock
    └── Dockerfile       # Python runtime
```

---

## 🚀 Getting Started

We use Mise to ensure a deterministic development environment.

1. Prerequisites
    * Mise
    * Docker & Docker Compose
2. Installation
    >mise install
3. Configuration
    Create a .env configuration file. Ensure you populate the `GEMINI_API_KEY`
4. Run (Local)
    >docker-compose up --build
5. Verify
    Send a request through the proxy. It will be validated by the Guardrail Service before reaching Google Gemini.
    > curl -X POST <http://localhost:8080/chat/completions> \
    > -H "Content-Type: application/json" \
    > -H "Authorization: Bearer ANY_STRING_IS_FINE_HERE" \
    > -d '{
    > "model": "gemini-2.5-flash",
    > "messages": [
    > {"role": "user", "content": "Explain who is Baldr in a sentence."}
    > ],
    > "stream": true
    >}'

## 🧪 Testing

run with Mise: `mise run'test:int'`

## ⚠️ Known Issues / Roadmap

* Latency: The current synchronous HTTP call between Proxy and Guardrail adds serialization overhead. Future versions may move to gRPC or shared memory.
* Flaky Tests: Integration tests involving Testcontainers occasionally hang on CI due to race conditions in container startup.
