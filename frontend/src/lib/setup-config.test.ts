import { describe, expect, it } from "vitest";
import {
  buildConnectionPayload,
  createDefaultConfigMap,
  getFieldControl,
  getCategoryDef,
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

  it("shows MongoDB config from backend entries in the settings UI", () => {
    const config = mergeBackendConfig([{ category: "mongodb", key: "databasename", value: "bcdev", issecret: false }]);
    expect(config.mongodb?.find((item) => item.key === "database")?.value).toBe("bcdev");
  });

  it("hides duplicated MongoDB production config from the settings UI", () => {
    const config = mergeBackendConfig([{ category: "mongodbproduction", key: "uri", value: "mongodb://prod", issecret: true }]);
    expect(config.mongodbproduction).toBeUndefined();
  });

  it("shows only the current MongoDB category in the setup UI", () => {
    const config = createDefaultConfigMap();
    expect(getCategoryDef("mongodb").title).toBe("MongoDB");
    expect(config.mongodb).toBeDefined();
    expect(config.mongodbdev).toBeUndefined();
    expect(config.mongodbuat).toBeUndefined();
    expect(config.mongodbpro).toBeUndefined();
  });

  it("hides future MongoDB UAT and PRO config from backend entries", () => {
    const config = mergeBackendConfig([
      { category: "mongodbuat", key: "uri", value: "mongodb://uat", issecret: true },
      { category: "mongodbpro", key: "uri", value: "mongodb://pro", issecret: true },
    ]);
    expect(config.mongodbuat).toBeUndefined();
    expect(config.mongodbpro).toBeUndefined();
  });

  it("builds kafka connection payload from serverurl", () => {
    const payload = buildConnectionPayload("kafka", [{ category: "kafka", key: "serverurl", value: "kafka:29092", isSecret: false, description: "" }], "12345");
    expect(payload).toMatchObject({ password: "12345", type: "kafka", host: "kafka", port: "29092" });
  });

  it("validates invalid ports", () => {
    const config = mergeBackendConfig([{ category: "postgresql", key: "port", value: "99999", issecret: false }]);
    expect(validateConfig(config)[0]).toContain("postgresql.port");
  });

  it("uses radio control for PostgreSQL SSL mode", () => {
    const item = configItem("postgresql", "sslmode", "disable");
    const control = getFieldControl(item);
    expect(control?.type).toBe("radio");
    expect(control?.type === "radio" ? control.options.map((option) => option.value) : []).toContain("verify-full");
  });

  it("uses checkbox control for boolean service flags", () => {
    expect(getFieldControl(configItem("service", "enablekafka", "true"))?.type).toBe("checkbox");
    expect(normalizeBooleanValue("TRUE")).toBe(true);
    expect(normalizeBooleanValue("false")).toBe(false);
  });
});

function configItem(category: string, key: string, value: string): ConfigItem {
  return { category, key, value, isSecret: false, description: "" };
}
