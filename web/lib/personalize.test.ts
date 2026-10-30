import { describe, expect, it } from "vitest";
import {
  currentHourBR,
  greeting,
  joinList,
  matchesConcern,
  missingAreasNotice,
  sortByPriority,
  welcomeMessage,
} from "./personalize";
import type { Alert } from "./types";

function alert(level: Alert["level"], type: string, resolved = false): Alert {
  return {
    id: `${level}-${type}`,
    property_id: "p",
    level,
    alert_type: type,
    period_start: "2026-10-30T00:00:00Z",
    period_end: "2026-10-31T00:00:00Z",
    triggered_at: "2026-10-30T00:00:00Z",
    resolved_at: resolved ? "2026-10-30T12:00:00Z" : null,
  };
}

describe("matchesConcern", () => {
  it("is true when an active alert has a chosen type", () => {
    expect(matchesConcern([alert("medium", "frost")], ["frost"])).toBe(true);
  });

  it("ignores resolved alerts, other types and empty concerns", () => {
    expect(matchesConcern([alert("medium", "frost", true)], ["frost"])).toBe(false);
    expect(matchesConcern([alert("medium", "heat_wave")], ["frost"])).toBe(false);
    expect(matchesConcern([alert("medium", "frost")], [])).toBe(false);
  });
});

describe("sortByPriority", () => {
  const quiet = { id: "quiet", alerts: [] as Alert[] };
  const mediumFrost = { id: "medium-frost", alerts: [alert("medium", "frost")] };
  const mediumHeat = { id: "medium-heat", alerts: [alert("medium", "heat_wave")] };
  const critical = { id: "critical", alerts: [alert("critical", "heavy_rain")] };

  it("orders by risk level first", () => {
    const sorted = sortByPriority([quiet, mediumHeat, critical], []);
    expect(sorted.map((i) => i.id)).toEqual(["critical", "medium-heat", "quiet"]);
  });

  it("puts chosen risk types ahead within the same level", () => {
    const sorted = sortByPriority([mediumHeat, mediumFrost], ["frost"]);
    expect(sorted.map((i) => i.id)).toEqual(["medium-frost", "medium-heat"]);
  });

  it("does not mutate the input", () => {
    const input = [quiet, critical];
    sortByPriority(input, []);
    expect(input.map((i) => i.id)).toEqual(["quiet", "critical"]);
  });
});

describe("greeting / currentHourBR", () => {
  it("greets by time of day using the first name", () => {
    expect(greeting("José da Silva", 8)).toBe("Bom dia, José");
    expect(greeting("José da Silva", 15)).toBe("Boa tarde, José");
    expect(greeting("José da Silva", 21)).toBe("Boa noite, José");
  });

  it("computes the hour in Brasília time", () => {
    expect(currentHourBR(new Date("2026-10-30T15:00:00Z"))).toBe(12);
    expect(currentHourBR(new Date("2026-10-30T02:30:00Z"))).toBe(23);
  });
});

describe("missingAreasNotice", () => {
  it("warns when fewer areas than announced are registered", () => {
    expect(missingAreasNotice("2-5", 1)).toBe(
      "Você disse ter de 2 a 5 áreas e há 1 cadastrada. Cadastre as outras para acompanhar todas."
    );
    expect(missingAreasNotice("6+", 3)).toBe(
      "Você disse ter mais de 5 áreas e há 3 cadastradas. Cadastre as outras para acompanhar todas."
    );
  });

  it("stays quiet when enough areas exist or the bucket is a single area", () => {
    expect(missingAreasNotice("2-5", 2)).toBeNull();
    expect(missingAreasNotice("1", 0)).toBeNull();
    expect(missingAreasNotice(undefined, 0)).toBeNull();
  });
});

describe("welcomeMessage / joinList", () => {
  it("joins lists in Portuguese", () => {
    expect(joinList([])).toBe("");
    expect(joinList(["geada"])).toBe("geada");
    expect(joinList(["geada", "seca"])).toBe("geada e seca");
    expect(joinList(["geada", "seca", "calor forte"])).toBe("geada, seca e calor forte");
  });

  it("tells producers what they will be alerted about and where", () => {
    expect(welcomeMessage({ profile: "producer", concerns: ["frost", "severe_water_stress"] }, "Campinas")).toBe(
      "Vamos te avisar de geada e seca em Campinas."
    );
    expect(welcomeMessage({ profile: "producer", concerns: [] }, "Campinas")).toBe(
      "Vamos te avisar de qualquer risco do clima em Campinas."
    );
    expect(welcomeMessage({ profile: "producer", concerns: ["frost"] })).toBe(
      "Vamos te avisar de geada na sua terra."
    );
  });

  it("tells technicians what the panel will highlight", () => {
    expect(welcomeMessage({ profile: "technician", concerns: ["heavy_rain"] })).toBe(
      "Vamos destacar chuva forte no seu painel."
    );
    expect(welcomeMessage({ profile: "technician", concerns: [] })).toBe(
      "Vamos acompanhar todos os riscos do clima no seu painel."
    );
  });
});
