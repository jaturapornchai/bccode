import { describe, expect, it } from "vitest";
import {
  buildConnectionPayload,
  getFieldControl,
  mergeBackendConfig,
  normalizeBooleanValue,
  normalizeSetupBackendUrl,
  validateConfig,
  type ConfigItem,
} from "./setup-config";

describe("setup config helpers", () => {
  it("normalizes setup backend URLs to goapi", () => {
    expect(normalizeSetupBackendUrl("localhost:8888")).toBe("http://localhost:8888/goapi");
    expect(normalizeSetupBackendUrl("http://localhost:8888/goapi/")).toBe("http://localhost:8888/goapi");
  });

  it("merges legacy mongodb database_name into database", () => {
    const config = mergeBackendConfig([{ category: "mongodb", key: "database_name", value: "bcdev", is_secret: false }]);
    expect(config.mongodb.find((item) => item.key === "database")?.value).toBe("bcdev");
  });

  it("hides duplicated MongoDB production config from the settings UI", () => {
    const config = mergeBackendConfig([{ category: "mongodb_production", key: "uri", value: "mongodb://prod", is_secret: true }]);
    expect(config.mongodb_production).toBeUndefined();
  });

  it("builds kafka connection payload from server_url", () => {
    const payload = buildConnectionPayload("kafka", [{ category: "kafka", key: "server_url", value: "kafka:29092", isSecret: false, description: "" }], "12345");
    expect(payload).toMatchObject({ password: "12345", type: "kafka", host: "kafka", port: "29092" });
  });

  it("validates invalid ports", () => {
    const config = mergeBackendConfig([{ category: "postgresql", key: "port", value: "99999", is_secret: false }]);
    expect(validateConfig(config)[0]).toContain("postgresql.port");
  });

  it("uses radio control for PostgreSQL SSL mode", () => {
    const item = configItem("postgresql", "ssl_mode", "disable");
    const control = getFieldControl(item);
    expect(control?.type).toBe("radio");
    expect(control?.type === "radio" ? control.options.map((option) => option.value) : []).toContain("verify-full");
  });

  it("uses checkbox control for boolean service flags", () => {
    expect(getFieldControl(configItem("service", "enable_kafka", "true"))?.type).toBe("checkbox");
    expect(normalizeBooleanValue("TRUE")).toBe(true);
    expect(normalizeBooleanValue("false")).toBe(false);
  });
});

function configItem(category: string, key: string, value: string): ConfigItem {
  return { category, key, value, isSecret: false, description: "" };
}
