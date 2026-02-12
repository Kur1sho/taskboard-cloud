import os
from fastapi import FastAPI, Depends, HTTPException
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from jose import jwt, JWTError
from datetime import datetime
from sqlalchemy import create_engine, Column, Integer, String, Boolean, DateTime, select
from sqlalchemy.orm import sessionmaker, Session, DeclarativeBase
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from typing import Optional


JWT_SECRET = os.getenv("JWT_SECRET", "dev-secret-change-me")
JWT_ALGORITHM = "HS256"

app = FastAPI(title="Tasks Service")

origins = os.getenv(
    "CORS_ORIGINS",
    "http://localhost:5173,http://127.0.0.1:5173"
).split(",")

app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


bearer = HTTPBearer()

DATABASE_URL = os.getenv("DATABASE_URL")
if not DATABASE_URL:
    raise RuntimeError("DATABASE_URL is not set")

engine = create_engine(DATABASE_URL, pool_pre_ping=True)
SessionLocal = sessionmaker(bind=engine, autoflush=False, autocommit=False)

class Base(DeclarativeBase):
    pass


class Task(Base):
    __tablename__ = "tasks"
    id = Column(Integer, primary_key=True)
    owner_email = Column(String(320), index=True, nullable=False)
    title = Column(String(200), nullable=False)
    done = Column(Boolean, nullable=False, default=False)
    created_at = Column(DateTime, nullable=False, default=datetime.utcnow)

class TaskUpdate(BaseModel):
    title: Optional[str] = None
    done: Optional[bool] = None

def get_db():
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()

@app.on_event("startup")
def on_startup():
    Base.metadata.create_all(bind=engine)


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

# TASKS = {}

@app.get("/tasks")
def list_tasks(user_email: str = Depends(get_current_user_email), db: Session = Depends(get_db)):
    rows = db.execute(select(Task).where(Task.owner_email == user_email).order_by(Task.id.desc())).scalars().all()
    return [{"id": t.id, "title": t.title, "done": t.done} for t in rows]

@app.post("/tasks")
def create_task(title: str, user_email: str = Depends(get_current_user_email), db: Session = Depends(get_db)):
    t = Task(owner_email=user_email, title=title, done=False)
    db.add(t)
    db.commit()
    db.refresh(t)
    return {"id": t.id, "title": t.title, "done": t.done}

@app.put("/tasks/{task_id}")
def update_task(
    task_id: int,
    patch: TaskUpdate,
    user_email: str = Depends(get_current_user_email),
    db: Session = Depends(get_db),
):
    t = db.execute(
        select(Task).where(Task.id == task_id, Task.owner_email == user_email)
    ).scalar_one_or_none()

    if not t:
        raise HTTPException(status_code=404, detail="Task not found")

    if patch.title is not None:
        t.title = patch.title
    if patch.done is not None:
        t.done = patch.done

    db.commit()
    db.refresh(t)
    return {"id": t.id, "title": t.title, "done": t.done}


@app.delete("/tasks/{task_id}")
def delete_task(
    task_id: int,
    user_email: str = Depends(get_current_user_email),
    db: Session = Depends(get_db),
):
    t = db.execute(
        select(Task).where(Task.id == task_id, Task.owner_email == user_email)
    ).scalar_one_or_none()

    if not t:
        raise HTTPException(status_code=404, detail="Task not found")

    db.delete(t)
    db.commit()
    return {"ok": True}

