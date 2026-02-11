import os
from fastapi import FastAPI, Depends, HTTPException
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from jose import jwt, JWTError

JWT_SECRET = os.getenv("JWT_SECRET", "dev-secret-change-me")
JWT_ALGORITHM = "HS256"

app = FastAPI(title="Tasks Service")

bearer = HTTPBearer()

def get_current_user_email(
    creds: HTTPAuthorizationCredentials = Depends(bearer),
) -> str:
    token = creds.credentials
    try:
        payload = jwt.decode(token, JWT_SECRET, algorithms=[JWT_ALGORITHM])
        sub = payload.get("sub")
        if not sub:
            raise HTTPException(status_code=401, detail="Invalid token")
        return sub
    except JWTError:
        raise HTTPException(status_code=401, detail="Invalid token")

@app.get("/health")
def health():
    return {"status": "ok"}

TASKS = {}

@app.get("/tasks")
def list_tasks(user_email: str = Depends(get_current_user_email)):
    return TASKS.get(user_email, [])

@app.post("/tasks")
def create_task(title: str, user_email: str = Depends(get_current_user_email)):
    TASKS.setdefault(user_email, [])
    TASKS[user_email].append({"title": title})
    return {"ok": True}
