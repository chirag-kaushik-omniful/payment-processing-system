import { FormEvent, useEffect, useState } from "react";
import { api } from "../api/client";
import type { AuditEvent, Payment } from "../api/types";
import { StatusBadge } from "../components/StatusBadge";

export function Admin() {
  const [payments, setPayments] = useState<Payment[]>([]);
  const [audit, setAudit] = useState<AuditEvent[]>([]);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [refundPaymentId, setRefundPaymentId] = useState("");
  const [refundAmount, setRefundAmount] = useState("");

  async function loadAdmin() {
    setError("");
    try {
      const [p, a] = await Promise.all([api.adminListPayments(), api.adminAudit()]);
      setPayments(p);
      setAudit(a);
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "Admin access denied — user needs admin role"
      );
    }
  }

  useEffect(() => {
    loadAdmin();
  }, []);

  async function handleReplay(id: string) {
    setError("");
    setSuccess("");
    try {
      await api.adminReplaySaga(id);
      setSuccess(`Replayed saga for ${id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Replay failed");
    }
  }

  async function handleRefund(e: FormEvent) {
    e.preventDefault();
    setError("");
    setSuccess("");
    try {
      await api.createRefund(refundPaymentId, parseFloat(refundAmount));
      setSuccess("Refund requested");
      loadAdmin();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Refund failed");
    }
  }

  return (
    <>
      <header className="page-header">
        <h1>Admin</h1>
        <p>Requires JWT with admin role — assign via auth DB or seed user</p>
      </header>

      {error && <div className="alert alert-error">{error}</div>}
      {success && <div className="alert alert-success">{success}</div>}

      <div className="card" style={{ marginBottom: "1.5rem", maxWidth: 480 }}>
        <h2 style={{ fontSize: "1.1rem", marginBottom: "1rem" }}>Refund (admin)</h2>
        <form onSubmit={handleRefund} className="form-grid">
          <input
            className="input mono"
            placeholder="Payment ID"
            value={refundPaymentId}
            onChange={(e) => setRefundPaymentId(e.target.value)}
            required
          />
          <input
            className="input"
            type="number"
            step="0.01"
            placeholder="Amount"
            value={refundAmount}
            onChange={(e) => setRefundAmount(e.target.value)}
            required
          />
          <button type="submit" className="btn btn-danger">Request refund</button>
        </form>
      </div>

      <div className="card" style={{ marginBottom: "1.5rem" }}>
        <div style={{ display: "flex", justifyContent: "space-between", marginBottom: "1rem" }}>
          <h2 style={{ fontSize: "1.1rem" }}>All payments</h2>
          <button type="button" className="btn btn-secondary" onClick={loadAdmin}>
            Refresh
          </button>
        </div>
        <table>
          <thead>
            <tr>
              <th>ID</th>
              <th>User</th>
              <th>Amount</th>
              <th>Status</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {payments.map((p) => (
              <tr key={p.id}>
                <td className="mono">{p.id.slice(0, 8)}…</td>
                <td className="mono">{p.user_id?.slice(0, 8)}…</td>
                <td>
                  {p.amount} {p.currency}
                </td>
                <td>
                  <StatusBadge status={p.status} />
                </td>
                <td>
                  <button
                    type="button"
                    className="btn btn-secondary"
                    style={{ padding: "0.35rem 0.6rem", fontSize: "0.8rem" }}
                    onClick={() => handleReplay(p.id)}
                  >
                    Replay saga
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <div className="card">
        <h2 style={{ fontSize: "1.1rem", marginBottom: "1rem" }}>Audit log</h2>
        {audit.length === 0 ? (
          <p style={{ color: "var(--text-muted)" }}>No audit events.</p>
        ) : (
          <table>
            <thead>
              <tr>
                <th>Action</th>
                <th>Resource</th>
                <th>Actor</th>
              </tr>
            </thead>
            <tbody>
              {audit.map((e) => (
                <tr key={e.id}>
                  <td>{e.action}</td>
                  <td className="mono">
                    {e.resource_type}/{e.resource_id?.slice(0, 8)}…
                  </td>
                  <td className="mono">{e.actor_id?.slice(0, 8) ?? "—"}…</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </>
  );
}
