const statusClass: Record<string, string> = {
  SUCCESS: "badge-success",
  CAPTURED: "badge-success",
  CREATED: "badge-muted",
  AUTHORIZED: "badge-warning",
  PROCESSING: "badge-warning",
  FAILED: "badge-danger",
  REFUNDED: "badge-muted",
  DISPUTED: "badge-danger",
};

export function StatusBadge({ status }: { status: string }) {
  const cls = statusClass[status] ?? "badge-muted";
  return <span className={`badge ${cls}`}>{status}</span>;
}
