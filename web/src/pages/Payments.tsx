import { FormEvent, useCallback, useEffect, useState } from "react";
import { api } from "../api/client";
import type { Payment } from "../api/types";
import { StatusBadge } from "../components/StatusBadge";

export function Payments() {
  const [payments, setPayments] = useState<Payment[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [amount, setAmount] = useState("100");
  const [currency, setCurrency] = useState("USD");
  const [provider, setProvider] = useState("stripe");
  const [captureMode, setCaptureMode] = useState("auto");
  const [idempotencyKey, setIdempotencyKey] = useState("");
  const [lookupId, setLookupId] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const list = await api.listPayments();
      setPayments(list);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load payments");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  async function handleCreate(e: FormEvent) {
    e.preventDefault();
    setError("");
    setSuccess("");
    try {
      const key = idempotencyKey.trim() || undefined;
      const p = await api.createPayment(
        {
          amount: parseFloat(amount),
          currency,
          provider,
          capture_mode: captureMode,
        },
        key
      );
      setSuccess(`Payment created: ${p.id} (${p.status})`);
      setIdempotencyKey("");
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Create failed");
    }
  }

  async function handleCapture(id: string) {
    setError("");
    try {
      await api.capturePayment(id);
      setSuccess(`Captured payment ${id}`);
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Capture failed");
    }
  }

  async function handleLookup() {
    if (!lookupId.trim()) return;
    setError("");
    try {
      const p = await api.getPayment(lookupId.trim());
      setSuccess(`Found: ${p.id} — ${p.status} — ${p.amount} ${p.currency}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Not found");
    }
  }

  return (
    <>
      <header className="page-header">
        <h1>Payments</h1>
        <p>Create and manage payments through the API gateway</p>
      </header>

      {error && <div className="alert alert-error">{error}</div>}
      {success && <div className="alert alert-success">{success}</div>}

      <div className="form-grid form-grid-2" style={{ marginBottom: "1.5rem" }}>
        <div className="card">
          <h2 style={{ fontSize: "1.1rem", marginBottom: "1rem" }}>Create payment</h2>
          <form onSubmit={handleCreate} className="form-grid">
            <div className="form-grid form-grid-2">
              <div>
                <label className="label">Amount</label>
                <input
                  className="input"
                  type="number"
                  step="0.01"
                  min="0.01"
                  value={amount}
                  onChange={(e) => setAmount(e.target.value)}
                  required
                />
              </div>
              <div>
                <label className="label">Currency</label>
                <select className="select" value={currency} onChange={(e) => setCurrency(e.target.value)}>
                  <option value="USD">USD</option>
                  <option value="EUR">EUR</option>
                  <option value="INR">INR</option>
                </select>
              </div>
            </div>
            <div className="form-grid form-grid-2">
              <div>
                <label className="label">Provider</label>
                <select className="select" value={provider} onChange={(e) => setProvider(e.target.value)}>
                  <option value="stripe">Stripe</option>
                  <option value="razorpay">Razorpay</option>
                  <option value="paypal">PayPal</option>
                </select>
              </div>
              <div>
                <label className="label">Capture mode</label>
                <select className="select" value={captureMode} onChange={(e) => setCaptureMode(e.target.value)}>
                  <option value="auto">Auto</option>
                  <option value="manual">Manual (hold)</option>
                </select>
              </div>
            </div>
            <div>
              <label className="label">Idempotency-Key (optional)</label>
              <input
                className="input mono"
                value={idempotencyKey}
                onChange={(e) => setIdempotencyKey(e.target.value)}
                placeholder="uuid — reuse to test dedup"
              />
            </div>
            <button type="submit" className="btn btn-primary">
              Create payment
            </button>
          </form>
        </div>

        <div className="card">
          <h2 style={{ fontSize: "1.1rem", marginBottom: "1rem" }}>Lookup by ID</h2>
          <div className="form-grid">
            <input
              className="input mono"
              placeholder="Payment UUID"
              value={lookupId}
              onChange={(e) => setLookupId(e.target.value)}
            />
            <button type="button" className="btn btn-secondary" onClick={handleLookup}>
              Get payment
            </button>
          </div>
        </div>
      </div>

      <div className="card">
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "1rem" }}>
          <h2 style={{ fontSize: "1.1rem" }}>Your payments</h2>
          <button type="button" className="btn btn-secondary" onClick={load} disabled={loading}>
            Refresh
          </button>
        </div>
        {loading ? (
          <p className="text-muted">Loading…</p>
        ) : payments.length === 0 ? (
          <p style={{ color: "var(--text-muted)" }}>No payments yet.</p>
        ) : (
          <table>
            <thead>
              <tr>
                <th>ID</th>
                <th>Amount</th>
                <th>Status</th>
                <th>Provider</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {payments.map((p) => (
                <tr key={p.id}>
                  <td className="mono">{p.id.slice(0, 8)}…</td>
                  <td>
                    {p.amount} {p.currency}
                  </td>
                  <td>
                    <StatusBadge status={p.status} />
                  </td>
                  <td>{p.provider || "—"}</td>
                  <td>
                    {p.status === "AUTHORIZED" && (
                      <button
                        type="button"
                        className="btn btn-secondary btn-sm"
                        style={{ width: "auto", padding: "0.35rem 0.6rem", fontSize: "0.8rem" }}
                        onClick={() => handleCapture(p.id)}
                      >
                        Capture
                      </button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </>
  );
}
