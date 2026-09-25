import { describe, expect, it } from "vitest";
import { currentRiskLevel } from "./risk";
import type { Alert } from "./api";

function makeAlert(overrides: Partial<Alert>): Alert {
  return {
    id: "a1",
    property_id: "p1",
    level: "low",
    alert_type: "frost",
    period_start: "2026-01-01T00:00:00Z",
    period_end: "2026-01-02T00:00:00Z",
    triggered_at: "2026-01-01T00:00:00Z",
    resolved_at: null,
    ...overrides,
  };
}

describe("currentRiskLevel", () => {
  it("returns low when there are no alerts", () => {
    expect(currentRiskLevel([])).toBe("low");
  });

  it("returns low when all alerts are resolved", () => {
    const alerts = [makeAlert({ level: "critical", resolved_at: "2026-01-03T00:00:00Z" })];
    expect(currentRiskLevel(alerts)).toBe("low");
  });

  it("returns the level of the single active alert", () => {
    const alerts = [makeAlert({ level: "medium" })];
    expect(currentRiskLevel(alerts)).toBe("medium");
  });

  it("returns the level of the most recently triggered active alert", () => {
    const alerts = [
      makeAlert({ id: "a1", level: "medium", triggered_at: "2026-01-01T00:00:00Z" }),
      makeAlert({ id: "a2", level: "critical", triggered_at: "2026-01-02T00:00:00Z" }),
    ];
    expect(currentRiskLevel(alerts)).toBe("critical");
  });

  it("ignores resolved alerts even if more recent", () => {
    const alerts = [
      makeAlert({ id: "a1", level: "critical", triggered_at: "2026-01-01T00:00:00Z" }),
      makeAlert({
        id: "a2",
        level: "medium",
        triggered_at: "2026-01-02T00:00:00Z",
        resolved_at: "2026-01-03T00:00:00Z",
      }),
    ];
    expect(currentRiskLevel(alerts)).toBe("critical");
  });
});
