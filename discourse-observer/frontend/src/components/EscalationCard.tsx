// Spec: specs/api/escalations.md (EP-12)
// Tests: backend/api/escalations_contract_test.go

import type { Escalations } from "../api/types";

interface EscalationCardProps {
  data: Escalations;
}

export function EscalationCard({ data }: EscalationCardProps) {
  const rateDisplay = data.rate !== null
    ? `${(data.rate * 100).toFixed(1)}%`
    : "–";

  return (
    <div className="response-card">
      <span className="response-card-value">{rateDisplay}</span>
      <span className="response-card-label">
        Escalation rate ({data.total} topics)
      </span>
    </div>
  );
}
