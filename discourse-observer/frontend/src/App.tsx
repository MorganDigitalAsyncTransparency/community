// Spec: specs/dashboard/queue-visibility.md
// Tests: tests/dashboard/queue-visibility.unit.test.ts

import "./App.css";
import { MOCK_DATA } from "./mock/data";
import { SummaryCards } from "./components/SummaryCards";
import { UnrepliedTable } from "./components/UnrepliedTable";
import { UntaggedTable } from "./components/UntaggedTable";

function formatSyncTime(isoString: string): string {
  return new Date(isoString).toLocaleString(undefined, {
    dateStyle: "short",
    timeStyle: "short",
  });
}

export function App() {
  return (
    <div className="app">
      <header className="app-header">
        <h1 className="app-title">discourse-observer</h1>
        <span className="app-sync-status">
          Last synced: {formatSyncTime(MOCK_DATA.lastSyncedAt)}
        </span>
      </header>

      <main className="app-content">
        <SummaryCards data={MOCK_DATA} />

        <section className="app-section">
          <h2 className="app-section-title">Awaiting reply</h2>
          <UnrepliedTable topics={MOCK_DATA.unrepliedTopics} />
        </section>

        <section className="app-section">
          <h2 className="app-section-title">Untagged topics</h2>
          <UntaggedTable topics={MOCK_DATA.untaggedTopics} />
        </section>
      </main>
    </div>
  );
}
