import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";
import { apiFetch, ApiClientError, type FilterParams } from "../../frontend/src/api/client";

// ---------------------------------------------------------------------------
// apiFetch — query parameter construction and error handling
// ---------------------------------------------------------------------------

describe("apiFetch", () => {
  const originalFetch = globalThis.fetch;

  beforeEach(() => {
    globalThis.fetch = vi.fn();
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
  });

  function mockFetchOk(body: unknown) {
    (globalThis.fetch as ReturnType<typeof vi.fn>).mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(body),
    });
  }

  function mockFetchError(status: number, body: { error: string }) {
    (globalThis.fetch as ReturnType<typeof vi.fn>).mockResolvedValue({
      ok: false,
      status,
      statusText: "Bad Request",
      json: () => Promise.resolve(body),
    });
  }

  it("builds correct URL without filters", async () => {
    mockFetchOk({ ok: true });
    await apiFetch("/status");
    expect(globalThis.fetch).toHaveBeenCalledWith("/api/v1/status");
  });

  it("maps preset period to query parameter", async () => {
    mockFetchOk([]);
    const filters: FilterParams = {
      period: { kind: "preset", preset: "last30" },
      tag: null,
    };
    await apiFetch("/queue/unreplied", filters);
    const url = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls[0][0] as string;
    expect(url).toContain("period=30d");
    expect(url).not.toContain("tag=");
  });

  it("maps allTime preset to period=all", async () => {
    mockFetchOk([]);
    const filters: FilterParams = {
      period: { kind: "preset", preset: "allTime" },
      tag: null,
    };
    await apiFetch("/queue/unreplied", filters);
    const url = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls[0][0] as string;
    expect(url).toContain("period=all");
  });

  it("maps custom range to from/to parameters", async () => {
    mockFetchOk([]);
    const filters: FilterParams = {
      period: { kind: "custom", range: { from: "2026-01-01", to: "2026-03-15" } },
      tag: null,
    };
    await apiFetch("/metrics/summary", filters);
    const url = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls[0][0] as string;
    expect(url).toContain("from=2026-01-01");
    expect(url).toContain("to=2026-03-15");
    expect(url).not.toContain("period=");
  });

  it("includes tag parameter when set", async () => {
    mockFetchOk([]);
    const filters: FilterParams = {
      period: { kind: "preset", preset: "allTime" },
      tag: "api",
    };
    await apiFetch("/queue/unreplied", filters);
    const url = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls[0][0] as string;
    expect(url).toContain("tag=api");
  });

  it("throws ApiClientError on HTTP error", async () => {
    mockFetchError(400, { error: "invalid period" });
    await expect(
      apiFetch("/queue/unreplied", { period: { kind: "preset", preset: "allTime" }, tag: null }),
    ).rejects.toThrow(ApiClientError);
  });

  it("includes error message from response body", async () => {
    mockFetchError(400, { error: "unknown tag: foo" });
    try {
      await apiFetch("/queue/unreplied", { period: { kind: "preset", preset: "allTime" }, tag: "foo" });
      expect.fail("should have thrown");
    } catch (err) {
      expect(err).toBeInstanceOf(ApiClientError);
      expect((err as ApiClientError).message).toBe("unknown tag: foo");
      expect((err as ApiClientError).status).toBe(400);
    }
  });

  it("returns parsed JSON on success", async () => {
    const data = { unrepliedCount: 5, untaggedCount: 2, oldestUnrepliedAgeDays: 14 };
    mockFetchOk(data);
    const result = await apiFetch("/queue/summary");
    expect(result).toEqual(data);
  });
});
