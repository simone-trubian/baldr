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
