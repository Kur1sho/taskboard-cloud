import os
from datetime import datetime, timedelta

from fastapi import FastAPI, HTTPException, Depends
from fastapi.security import OAuth2PasswordRequestForm
from jose import jwt
from passlib.context import CryptContext

JWT_SECRET = os.getenv("JWT_SECRET", "dev-secret-change-me")
JWT_ALGORITHM = "HS256"
JWT_EXPIRE_MINUTES = 60

pwd = CryptContext(schemes=["pbkdf2_sha256"], deprecated="auto")

app = FastAPI(title="Auth Service")

def create_access_token(subject: str) -> str:
    expire = datetime.utcnow() + timedelta(minutes=JWT_EXPIRE_MINUTES)
    payload = {"sub": subject, "exp": expire}
    return jwt.encode(payload, JWT_SECRET, algorithm=JWT_ALGORITHM)

USERS = {}

@app.get("/health")
def health():
    return {"status": "ok"}

@app.post("/auth/register")
def register(email: str, password: str):
    if email in USERS:
        raise HTTPException(status_code=400, detail="Email already registered")
    USERS[email] = pwd.hash(password)
    return {"email": email}

@app.post("/auth/login")
def login(form: OAuth2PasswordRequestForm = Depends()):
    email = form.username
    password = form.password
    if email not in USERS or not pwd.verify(password, USERS[email]):
        raise HTTPException(status_code=401, detail="Invalid credentials")
    token = create_access_token(email)
    return {"access_token": token, "token_type": "bearer"}
