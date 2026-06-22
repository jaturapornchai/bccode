import { describe, expect, it } from "vitest";
import { formatDefaultDateTime, resolveWorkspaceDateTimeDisplayOptions } from "@/lib/date-time";
import type { WorkspaceSession } from "@/lib/workspace-models";

// Timezone Iron Rule: DB stores UTC+0, the FRONTEND renders timestamps in the
// active BRANCH timezone. This proves a single UTC instant renders to the
// correct branch-local wall clock for three different branch timezones, exactly
// as the system-settings audit screen now does via
// formatDefaultDateTime(value, resolveWorkspaceDateTimeDisplayOptions(workspace, language)).

function mockWorkspace(timezone: string): WorkspaceSession {
  return {
    // Force Gregorian year so the assertion is timezone-focused, not calendar-focused.
    shop: { holdingcode: "TEST" },
    branch: {
      guidfixed: "branch-test",
      timezone,
      yeartype: "christian",
    },
  };
}

// Fixed UTC instant near a day boundary (18:30 UTC -> crosses midnight in Asia).
const UTC_INSTANT = "2026-06-22T18:30:00.000Z";

describe("Timezone Iron Rule: branch-timezone rendering", () => {
  it("resolves the branch timezone from workspace.branch.timezone", () => {
    expect(resolveWorkspaceDateTimeDisplayOptions(mockWorkspace("Asia/Tokyo"), "en")).toEqual({
      language: "en",
      yearType: "christian",
      timeZone: "Asia/Tokyo",
    });
  });

  it("renders the SAME UTC instant in each branch local time", () => {
    const bangkok = formatDefaultDateTime(
      UTC_INSTANT,
      resolveWorkspaceDateTimeDisplayOptions(mockWorkspace("Asia/Bangkok"), "en"),
    );
    const tokyo = formatDefaultDateTime(
      UTC_INSTANT,
      resolveWorkspaceDateTimeDisplayOptions(mockWorkspace("Asia/Tokyo"), "en"),
    );
    const newYork = formatDefaultDateTime(
      UTC_INSTANT,
      resolveWorkspaceDateTimeDisplayOptions(mockWorkspace("America/New_York"), "en"),
    );

    // 18:30Z + 7h = 01:30 next day
    expect(bangkok).toBe("23/06/2026 01:30:00");
    // 18:30Z + 9h = 03:30 next day
    expect(tokyo).toBe("23/06/2026 03:30:00");
    // 18:30Z - 4h (EDT, June) = 14:30 same day
    expect(newYork).toBe("22/06/2026 14:30:00");

    // None of them equals the raw UTC wall clock (18:30) -> tz IS applied.
    expect(bangkok).not.toContain("18:30");
    expect(tokyo).not.toContain("18:30");
    expect(newYork).not.toContain("18:30");

    // All three differ from each other -> tz is genuinely applied, not ignored.
    expect(new Set([bangkok, tokyo, newYork]).size).toBe(3);
  });

  it("falls back to shopInfo.settings.timezone, then Asia/Bangkok", () => {
    const shopInfoOnly: WorkspaceSession = {
      shop: { holdingcode: "TEST" },
      branch: { guidfixed: "b", yeartype: "christian" },
      shopInfo: { settings: { timezone: "Asia/Tokyo" } },
    };
    expect(
      formatDefaultDateTime(UTC_INSTANT, resolveWorkspaceDateTimeDisplayOptions(shopInfoOnly, "en")),
    ).toBe("23/06/2026 03:30:00");

    const noTimezone: WorkspaceSession = {
      shop: { holdingcode: "TEST" },
      branch: { guidfixed: "b", yeartype: "christian" },
    };
    expect(resolveWorkspaceDateTimeDisplayOptions(noTimezone, "en").timeZone).toBe("Asia/Bangkok");
    expect(
      formatDefaultDateTime(UTC_INSTANT, resolveWorkspaceDateTimeDisplayOptions(noTimezone, "en")),
    ).toBe("23/06/2026 01:30:00");
  });
});
