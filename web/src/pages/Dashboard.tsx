import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { api } from "../api/client";
import { useAuth } from "../context/AuthContext";

export function Dashboard() {
  const { user } = useAuth();
  const [health, setHealth] = useState<string>("—");
  const [paymentCount, setPaymentCount] = useState<number | null>(null);
  const [balance, setBalance] = useState<number | null>(null);
  useEffect(() => {
    api.health().then((h) => setHealth(h.status)).catch(() => setHealth("unreachable"));

    api
      .listPayments()
      .then((p) => setPaymentCount(p.length))
      .catch(() => setPaymentCount(null));

    api
      .getWallet()
      .then((w) => setBalance(w.balance))
      .catch(() => setBalance(null));
  }, []);

  return (
    <>
      <header className="page-header">
        <h1>Dashboard</h1>
        <p>Welcome back, {user?.email}</p>
      </header>

      <div className="form-grid form-grid-2" style={{ marginBottom: "1.5rem" }}>
        <div className="card">
          <p className="label">API Gateway</p>
          <p style={{ fontSize: "1.5rem", fontWeight: 700 }}>{health}</p>
        </div>
        <div className="card">
          <p className="label">Your payments</p>
          <p style={{ fontSize: "1.5rem", fontWeight: 700 }}>
            {paymentCount ?? "—"}
          </p>
        </div>
        <div className="card">
          <p className="label">Wallet balance</p>
          <p style={{ fontSize: "1.5rem", fontWeight: 700 }}>
            {balance != null ? `$${balance.toFixed(2)}` : "—"}
          </p>
        </div>
        <div className="card">
          <p className="label">Quick actions</p>
          <div style={{ display: "flex", gap: "0.5rem", flexWrap: "wrap", marginTop: "0.5rem" }}>
            <Link to="/payments" className="btn btn-primary">
              New payment
            </Link>
            <Link to="/wallet" className="btn btn-secondary">
              View wallet
            </Link>
          </div>
        </div>
      </div>
    </>
  );
}
