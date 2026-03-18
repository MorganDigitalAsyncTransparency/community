// Spec: specs/dashboard/peak-activity.md (PA-23, PA-24)
// Tests: tests/dashboard/timezone-utils.unit.test.ts

interface CookieConsentModalProps {
  onAccept: () => void;
  onDeny: () => void;
}

export function CookieConsentModal({ onAccept, onDeny }: CookieConsentModalProps) {
  return (
    <div className="peak-consent-backdrop">
      <div className="peak-consent-dialog">
        <h3 className="peak-consent-title">Store timezone preferences?</h3>
        <p className="peak-consent-text">
          Your timezone selections will be stored in a browser cookie so they
          persist across visits. Only the timezone identifiers are stored — no
          personal data.
        </p>
        <div className="peak-consent-actions">
          <button className="peak-consent-btn peak-consent-btn-accept" onClick={onAccept}>
            Accept
          </button>
          <button className="peak-consent-btn peak-consent-btn-deny" onClick={onDeny}>
            Deny
          </button>
        </div>
      </div>
    </div>
  );
}
