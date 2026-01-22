from typing import List, Optional, Any, Dict
from pydantic import BaseModel
from fastapi import FastAPI


# 1. Mirror the OpenAI structure (simplified)
class Message(BaseModel):
    role: str
    content: str


class OpenAIChatRequest(BaseModel):
    model: str
    messages: List[Message]
    stream: bool = False

    # Allow other fields (temperature, etc.) without crashing
    class Config:
        extra = "allow"


class ValidationResponse(BaseModel):
    allowed: bool
    reason: str = ""
    # Return the FULL modified request body to send to OpenAI
    sanitized_input: Optional[Dict[str, Any]] = None


app = FastAPI()


@app.post("/validate", response_model=ValidationResponse)
async def validate_prompt(req: OpenAIChatRequest):
    # Extract the last user message
    last_message = req.messages[-1].content

    print(f"[Guardrail] Scanning: {last_message[:20]}...")

    if "ATTACK" in last_message:
        return ValidationResponse(
            allowed=False, reason="Malicious keyword detected"
        )

    # Simulate PII masking
    if "password" in last_message:
        # Modifying the request object directly
        req.messages[-1].content = last_message.replace(
            "password", "[REDACTED]"
        )

        # Return the ENTIRE updated JSON structure
        return ValidationResponse(
            allowed=True,
            sanitized_input=req.model_dump(),  # Dumps the full OpenAI-compatible JSON
        )

    return ValidationResponse(allowed=True, sanitized_input=None)


"""
from typing import Optional, Any
from fastapi import FastAPI
from pydantic import BaseModel
import time
import os

app = FastAPI()


class ValidationRequest(BaseModel):
    prompt: str


class ValidationResponse(BaseModel):
    allowed: bool
    reason: str = ""
    sanitized_input: Optional[Any] = None  # Allows returning a JSON object


@app.post("/validate", response_model=ValidationResponse)
async def validate_prompt(req: ValidationRequest):
    # Simulate ML Latency (configurable via ENV for stress testing)
    latency = float(os.getenv("SIMULATED_LATENCY_MS", 50)) / 1000.0
    time.sleep(latency)

    print(f"[Guardrail] Scanning: {req.prompt[:20]}...")

    if "ATTACK" in req.prompt:
        return ValidationResponse(
            allowed=False, reason="Malicious keyword detected"
        )

    # Simulate PII masking
    sanitized = req.prompt.replace("password", "[REDACTED]")

    return ValidationResponse(allowed=True, sanitized_input=sanitized)
"""
