import { useEffect, useState } from "react";
import { api } from "../api/client";
import type { Wallet } from "../api/types";

export function WalletPage() {
  const [wallet, setWallet] = useState<Wallet | null>(null);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api
      .getWallet()
      .then(setWallet)
      .catch((err) => setError(err instanceof Error ? err.message : "Failed to load wallet"))
      .finally(() => setLoading(false));
  }, []);

  return (
    <>
      <header className="page-header">
        <h1>Wallet</h1>
        <p>Current balance from wallet-service (optimistic locking + Redis cache)</p>
      </header>

      {error && <div className="alert alert-error">{error}</div>}

      <div className="card" style={{ maxWidth: 420 }}>
        {loading ? (
          <p style={{ color: "var(--text-muted)" }}>Loading…</p>
        ) : wallet ? (
          <>
            <p className="label">Balance</p>
            <p style={{ fontSize: "2.5rem", fontWeight: 700, marginBottom: "1rem" }}>
              ${wallet.balance.toFixed(2)}
            </p>
            <p className="label">User ID</p>
            <p className="mono" style={{ marginBottom: "0.75rem" }}>{wallet.user_id}</p>
            <p className="label">Version (optimistic lock)</p>
            <p>{wallet.version}</p>
          </>
        ) : (
          <p style={{ color: "var(--text-muted)" }}>No wallet data.</p>
        )}
      </div>
    </>
  );
}
