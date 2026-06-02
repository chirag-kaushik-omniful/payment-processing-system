#!/usr/bin/env bash
set -euo pipefail

ACTION="${1:-up}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DATABASE_URL="${DATABASE_URL:-postgres://payment:payment@localhost:5432/payment_platform?sslmode=disable}"

for svc in auth-service payment-service wallet-service ledger-service subscription-service; do
  MIGRATIONS="$ROOT/services/$svc/migrations"
  if [[ -d "$MIGRATIONS" ]]; then
    echo "==> $svc migrations ($ACTION)"
    for f in "$MIGRATIONS"/*.sql; do
      [[ -f "$f" ]] || continue
      echo "    applying $(basename "$f")"
      psql "$DATABASE_URL" -f "$f" -v ON_ERROR_STOP=1 || true
    done
  fi
done
