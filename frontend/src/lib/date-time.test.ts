import { describe, expect, it } from "vitest";
import { formatDefaultDate, formatDefaultDateTime, formatLocalDate, localTimeToUtcTime, normalizeTimeInput } from "@/lib/date-time";

describe("date-time helpers", () => {
  it("normalizes compact and clock time input", () => {
    expect(normalizeTimeInput("800")).toBe("08:00");
    expect(normalizeTimeInput("0830")).toBe("08:30");
    expect(normalizeTimeInput("8:45")).toBe("08:45");
  });

  it("converts branch local time to UTC+0 using explicit UTC offset", () => {
    expect(localTimeToUtcTime("08:00", "+07:00")).toEqual({ time: "01:00", dayOffset: 0 });
    expect(localTimeToUtcTime("00:30", "+07:00")).toEqual({ time: "17:30", dayOffset: -1 });
    expect(localTimeToUtcTime("23:30", "-02:00")).toEqual({ time: "01:30", dayOffset: 1 });
  });

  it("formats local dates with Buddhist Era or Christian Era", () => {
    expect(formatLocalDate("2026-05-18", "th", "buddhist")).toContain("2569");
    expect(formatLocalDate("2026-05-18", "th", "christian")).toContain("2026");
    expect(formatLocalDate("2026-05-18", "en", "christian")).toContain("2026");
  });

  it("formats project default date and timestamp displays consistently", () => {
    expect(formatDefaultDateTime("2026-05-29T16:30:19Z", {
      language: "th",
      yearType: "buddhist",
      timeZone: "Asia/Bangkok",
    })).toBe("29/05/2569 23:30:19");
    expect(formatDefaultDate("2026-05-29T16:30:19Z", {
      language: "th",
      yearType: "christian",
      timeZone: "Asia/Bangkok",
    })).toBe("29/05/2026");
    expect(formatDefaultDateTime("not-a-date")).toBe("not-a-date");
  });
});
