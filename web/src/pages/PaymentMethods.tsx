import { FormEvent, useState } from "react";
import { api } from "../api/client";
import type { PaymentMethod } from "../api/types";

export function PaymentMethods() {
  const [methods, setMethods] = useState<PaymentMethod[]>([]);
  const [provider, setProvider] = useState("stripe");
  const [lastFour, setLastFour] = useState("4242");
  const [brand, setBrand] = useState("visa");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  async function handleCreate(e: FormEvent) {
    e.preventDefault();
    setError("");
    setSuccess("");
    try {
      const m = await api.createPaymentMethod({ provider, last_four: lastFour, brand });
      setMethods((prev) => [m, ...prev]);
      setSuccess(`Tokenized method ${m.id} (${m.brand} •••• ${m.last_four})`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create method");
    }
  }

  return (
    <>
      <header className="page-header">
        <h1>Payment methods</h1>
        <p>Tokenized cards — no PAN/CVV stored on platform</p>
      </header>

      {error && <div className="alert alert-error">{error}</div>}
      {success && <div className="alert alert-success">{success}</div>}

      <div className="card" style={{ maxWidth: 480, marginBottom: "1.5rem" }}>
        <h2 style={{ fontSize: "1.1rem", marginBottom: "1rem" }}>Add method</h2>
        <form onSubmit={handleCreate} className="form-grid">
          <div>
            <label className="label">Provider</label>
            <select className="select" value={provider} onChange={(e) => setProvider(e.target.value)}>
              <option value="stripe">Stripe</option>
              <option value="razorpay">Razorpay</option>
              <option value="paypal">PayPal</option>
            </select>
          </div>
          <div className="form-grid form-grid-2">
            <div>
              <label className="label">Last four</label>
              <input className="input" value={lastFour} onChange={(e) => setLastFour(e.target.value)} maxLength={4} />
            </div>
            <div>
              <label className="label">Brand</label>
              <input className="input" value={brand} onChange={(e) => setBrand(e.target.value)} />
            </div>
          </div>
          <button type="submit" className="btn btn-primary">Tokenize</button>
        </form>
      </div>

      {methods.length > 0 && (
        <div className="card">
          <h2 style={{ fontSize: "1.1rem", marginBottom: "1rem" }}>Session methods</h2>
          <table>
            <thead>
              <tr>
                <th>ID</th>
                <th>Provider</th>
                <th>Card</th>
              </tr>
            </thead>
            <tbody>
              {methods.map((m) => (
                <tr key={m.id}>
                  <td className="mono">{m.id.slice(0, 8)}…</td>
                  <td>{m.provider}</td>
                  <td>{m.brand} •••• {m.last_four}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </>
  );
}
