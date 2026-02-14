import { useEffect, useState } from 'react'

type Task = { id: number; title: string; done: boolean }

export default function App() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [token, setToken] = useState("");
  const [tasks, setTasks] = useState<Task[]>([]);
  const [newTitle, setNewTitle] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string>("");

  const AUTH_URL = import.meta.env.VITE_AUTH_URL ?? "http://localhost:8001";
  const TASKS_URL = import.meta.env.VITE_TASKS_URL ?? "http://localhost:8002";

  const fetchTasks = async (tkn = token) => {
    setError(null);
    setMessage("");
    if (!tkn) return;

    try {
      const res = await fetch(`${TASKS_URL}/tasks/`, {
        headers: { Authorization: `Bearer ${tkn}` },
      });
      const data = await res.json().catch(() => ({}));

      if (!res.ok) {
        setError(data?.detail ?? "Failed to load tasks");
        return;
      }

      setTasks(Array.isArray(data) ? data : []);
    } catch {
      setError("Cannot reach tasks service");
    }
  };


  const register = async () => {
    setError(null);
    setMessage("");
    try {
      const res = await fetch(`${AUTH_URL}/auth/register?email=${encodeURIComponent(email)}&password=${encodeURIComponent(password)}`, {
        method: "POST",
      });

      const data = await res.json().catch(() => ({}));

      if (!res.ok) {
        setError(data?.detail ?? "Register failed");
        return;
      }

      setMessage("Registered successfully");
    } catch (e) {
      setError("Cannot reach auth service");
    }
  };



  const login = async () => {
    setError(null);
    setMessage("");

    try {
      const form = new URLSearchParams();
      form.append("username", email);
      form.append("password", password);

      const res = await fetch(`${AUTH_URL}/auth/login`, {
        method: "POST",
        body: form,
        headers: { "Content-Type": "application/x-www-form-urlencoded" },
      });

      const data = await res.json().catch(() => ({}));

      if (!res.ok) {
        setError(data?.detail ?? "Login failed");
        return;
      }

      setToken(data.access_token);
      setMessage("Login successful");
    } catch (e) {
      setError("Cannot reach auth service");
    }
  };

  const logout = () => {
    setToken("");
    setTasks([]);
    setNewTitle("");
    setError(null);
    setMessage("Logged out");
  };

  const createTask = async () => {
    setError(null);
    setMessage("");
    if (!token) return;

    const title = newTitle.trim();
    if (!title) {
      setError("Task title cannot be empty");
      return;
    }

    try {
      const res = await fetch(
        `${TASKS_URL}/tasks/?title=${encodeURIComponent(title)}`,
        {
          method: "POST",
          headers: { Authorization: `Bearer ${token}` },
        }
      );

      const data = await res.json().catch(() => ({}));

      if (!res.ok) {
        setError(data?.detail ?? "Create task failed");
        return;
      }

      setNewTitle("");
      setTasks((prev) => [data as Task, ...prev]);
      setMessage("Task created");
    } catch {
      setError("Cannot reach tasks service");
    }
  };

  const toggleDone = async (task: Task) => {
    setError(null);
    setMessage("");
    if (!token) return;

    try {
      const res = await fetch(`${TASKS_URL}/tasks/${task.id}`, {
        method: "PUT",
        headers: {
          Authorization: `Bearer ${token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ done: !task.done }),
      });

      const data = await res.json().catch(() => ({}));
      if (!res.ok) {
        setError(data?.detail ?? "Update task failed");
        return;
      }

      setTasks((prev) => prev.map((t) => (t.id === task.id ? data : t)));
    } catch {
      setError("Cannot reach tasks service");
    }
  };

  const deleteTask = async (taskId: number) => {
    setError(null);
    setMessage("");
    if (!token) return;

    try {
      const res = await fetch(`${TASKS_URL}/tasks/${taskId}`, {
        method: "DELETE",
        headers: { Authorization: `Bearer ${token}` },
      });

      const data = await res.json().catch(() => ({}));
      if (!res.ok) {
        setError(data?.detail ?? "Delete task failed");
        return;
      }

      setTasks((prev) => prev.filter((t) => t.id !== taskId));
      setMessage("Task deleted");
    } catch {
      setError("Cannot reach tasks service");
    }
  };

  useEffect(() => { // future
    if (token) fetchTasks();
  }, []);


  return (
    <div style={{
      padding: 40,
      background: "#111827",
      borderRadius: 12,
      width: 320,
      boxShadow: "0 10px 25px rgba(0,0,0,0.4)",
      textAlign: "center"
    }}>
      <h1>Taskboard</h1>
      {!token && (
        <>
          <input
            placeholder="Email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
          />
          <br /><br />

          <input
            placeholder="Password"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
          <br /><br />

          <button onClick={register}>Register</button>
          <button onClick={login}>Login</button>
        </>
      )}

      {token && (
        <>
          <div style={{ display: "flex", gap: 8, justifyContent: "center" }}>
            <button onClick={() => fetchTasks()}>Refresh</button>
            <button onClick={logout}>Logout</button>
          </div>

          <div style={{ marginTop: 16 }}>
            <input
              placeholder="New task title"
              value={newTitle}
              onChange={(e) => setNewTitle(e.target.value)}
            />
            <button onClick={createTask} style={{ marginLeft: 8 }}>
              Add
            </button>
          </div>

          <div style={{ marginTop: 16, textAlign: "left" }}>
            {tasks.length === 0 ? (
              <p style={{ opacity: 0.8 }}>No tasks yet.</p>
            ) : (
              tasks.map((t) => (
                <div
                  key={t.id}
                  style={{
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "space-between",
                    padding: "8px 10px",
                    borderRadius: 8,
                    background: "rgba(255,255,255,0.06)",
                    marginBottom: 8,
                  }}
                >
                  <label style={{ display: "flex", gap: 10, cursor: "pointer" }}>
                    <input
                      type="checkbox"
                      checked={t.done}
                      onChange={() => toggleDone(t)}
                    />
                    <span style={{ textDecoration: t.done ? "line-through" : "none" }}>
                      {t.title}
                    </span>
                  </label>

                  <button onClick={() => deleteTask(t.id)}>Delete</button>
                </div>
              ))
            )}
          </div>
        </>
      )}

      {error && (
        <div style={{ marginTop: 12, color: "#fc0c0c" }}>
          {error}
        </div>
      )}


      {message && (
        <p style={{ marginTop: 12, color: "#86efac" }}>
          {message}
        </p>
      )}
    </div>
  );
}