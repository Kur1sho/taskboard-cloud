import { useState } from 'react'

export default function App() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [token, setToken] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string>("");


  const register = async () => {
    setError(null);
    setMessage("");

    try {
      const res = await fetch(
        "http://localhost:8001/auth/register?email=" + encodeURIComponent(email) + "&password=" + encodeURIComponent(password),
        { method: "POST" }
      );

      const data = await res.json();

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

      const res = await fetch("http://localhost:8001/auth/login", {
        method: "POST",
        body: form,
        headers: {
          "Content-Type": "application/x-www-form-urlencoded",
        },
      });

      const data = await res.json();

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

      <p style={{ marginTop: 12 }}>DEBUG: message="{message}" error="{error}"</p>

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

      {token && (
        <>
          <h3>Token:</h3>
          <textarea value={token} readOnly rows={4} cols={50} />
        </>
      )}
    </div>
  );
}