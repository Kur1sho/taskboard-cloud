import os
import importlib
from sqlalchemy import text
from fastapi.testclient import TestClient
from jose import jwt
from datetime import datetime, timedelta

JWT_SECRET = "ci-test-secret"
JWT_ALG = "HS256"


def make_token(email: str) -> str:
    payload = {"sub": email, "exp": datetime.utcnow() + timedelta(minutes=60)}
    return jwt.encode(payload, JWT_SECRET, algorithm=JWT_ALG)


def make_client():
    os.environ["JWT_SECRET"] = JWT_SECRET
    os.environ["DATABASE_URL"] = os.environ.get(
        "DATABASE_URL",
        "postgresql+psycopg://taskboard:taskboard@localhost:5432/taskboard_test",
    )

    tasks_main = importlib.import_module("main")

    with tasks_main.engine.begin() as conn:
        conn.execute(text("TRUNCATE TABLE tasks RESTART IDENTITY CASCADE;"))

    return TestClient(tasks_main.app)


def test_requires_auth():
    c = make_client()
    r = c.get("/tasks")
    assert r.status_code in (401, 403)  # HTTPBearer returns 403 if missing


def test_create_and_list_tasks_for_user():
    c = make_client()
    tkn = make_token("user@test.com")

    r = c.post("/tasks", params={"title": "First"}, headers={"Authorization": f"Bearer {tkn}"})
    assert r.status_code == 200, r.text
    created = r.json()
    assert created["title"] == "First"
    assert created["done"] is False

    r2 = c.get("/tasks", headers={"Authorization": f"Bearer {tkn}"})
    assert r2.status_code == 200, r2.text
    items = r2.json()
    assert isinstance(items, list)
    assert len(items) == 1
    assert items[0]["title"] == "First"


def test_user_cannot_see_other_users_tasks():
    c = make_client()
    tkn_a = make_token("a@test.com")
    tkn_b = make_token("b@test.com")

    c.post("/tasks", params={"title": "A1"}, headers={"Authorization": f"Bearer {tkn_a}"})
    c.post("/tasks", params={"title": "B1"}, headers={"Authorization": f"Bearer {tkn_b}"})

    ra = c.get("/tasks", headers={"Authorization": f"Bearer {tkn_a}"})
    rb = c.get("/tasks", headers={"Authorization": f"Bearer {tkn_b}"})

    assert [t["title"] for t in ra.json()] == ["A1"]
    assert [t["title"] for t in rb.json()] == ["B1"]


def test_update_toggle_done():
    c = make_client()
    tkn = make_token("user@test.com")

    created = c.post("/tasks", params={"title": "X"}, headers={"Authorization": f"Bearer {tkn}"}).json()
    task_id = created["id"]

    r = c.put(f"/tasks/{task_id}", json={"done": True}, headers={"Authorization": f"Bearer {tkn}"})
    assert r.status_code == 200, r.text
    assert r.json()["done"] is True


def test_delete_task():
    c = make_client()
    tkn = make_token("user@test.com")

    created = c.post("/tasks", params={"title": "To delete"}, headers={"Authorization": f"Bearer {tkn}"}).json()
    task_id = created["id"]

    r = c.delete(f"/tasks/{task_id}", headers={"Authorization": f"Bearer {tkn}"})
    assert r.status_code == 200, r.text
    assert r.json()["ok"] is True

    r2 = c.get("/tasks", headers={"Authorization": f"Bearer {tkn}"})
    assert r2.status_code == 200
    assert r2.json() == []


def test_cannot_update_other_users_task():
    c = make_client()
    tkn_a = make_token("a@test.com")
    tkn_b = make_token("b@test.com")

    created = c.post("/tasks", params={"title": "A1"}, headers={"Authorization": f"Bearer {tkn_a}"}).json()
    task_id = created["id"]

    r = c.put(f"/tasks/{task_id}", json={"done": True}, headers={"Authorization": f"Bearer {tkn_b}"})
    assert r.status_code == 404
