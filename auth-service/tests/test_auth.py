import os
import importlib
from sqlalchemy import text
from fastapi.testclient import TestClient
from jose import jwt

JWT_SECRET = "ci-test-secret"
JWT_ALG = "HS256"


def make_client():
    # IMPORTANT: env must be set BEFORE importing main.py (it creates engine at import time)
    os.environ["JWT_SECRET"] = JWT_SECRET
    os.environ["DATABASE_URL"] = os.environ.get(
        "DATABASE_URL",
        "postgresql+psycopg://taskboard:taskboard@localhost:5432/taskboard_test",
    )

    auth_main = importlib.import_module("main")

    # clean table
    with auth_main.engine.begin() as conn:
        conn.execute(text("TRUNCATE TABLE users RESTART IDENTITY CASCADE;"))

    return TestClient(auth_main.app)


def test_register_ok():
    c = make_client()
    r = c.post("/auth/register", params={"email": "a@test.com", "password": "123456"})
    assert r.status_code == 200, r.text
    assert r.json()["email"] == "a@test.com"


def test_register_duplicate():
    c = make_client()
    c.post("/auth/register", params={"email": "a@test.com", "password": "123456"})
    r2 = c.post("/auth/register", params={"email": "a@test.com", "password": "123456"})
    assert r2.status_code == 400
    assert r2.json()["detail"] == "Email already registered"


def test_register_invalid_email():
    c = make_client()
    r = c.post("/auth/register", params={"email": "not-an-email", "password": "123456"})
    # FastAPI/Pydantic validation error
    assert r.status_code == 422


def test_register_password_too_short():
    c = make_client()
    r = c.post("/auth/register", params={"email": "b@test.com", "password": "123"})
    assert r.status_code == 422


def test_login_ok_returns_jwt_with_sub():
    c = make_client()
    c.post("/auth/register", params={"email": "a@test.com", "password": "123456"})

    r = c.post(
        "/auth/login",
        data={"username": "a@test.com", "password": "123456"},
        headers={"Content-Type": "application/x-www-form-urlencoded"},
    )
    assert r.status_code == 200, r.text
    token = r.json()["access_token"]

    payload = jwt.decode(token, JWT_SECRET, algorithms=[JWT_ALG])
    assert payload["sub"] == "a@test.com"
    assert "exp" in payload


def test_login_invalid_credentials():
    c = make_client()
    r = c.post(
        "/auth/login",
        data={"username": "nope@test.com", "password": "123456"},
        headers={"Content-Type": "application/x-www-form-urlencoded"},
    )
    assert r.status_code == 401
