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

  it("shows only Pure PostgreSQL as database in default config map", () => {
    const config = createDefaultConfigMap();
    expect(getCategoryDef("postgresql").title).toBe("PostgreSQL");
    expect(config.postgresql).toBeDefined();
    expect(config.mongodb).toBeUndefined();
    expect(config.clickhouse).toBeUndefined();
    expect(config.kafka).toBeUndefined();
    expect(config.redis).toBeUndefined();
  });

  it("merges backend PostgreSQL config entries into config map", () => {
    const config = mergeBackendConfig([
      { category: "postgresql", key: "dbname", value: "bcai_projection", issecret: false },
      { category: "postgresql", key: "host", value: "postgres", issecret: false },
    ]);
    expect(config.postgresql?.find((item) => item.key === "dbname")?.value).toBe("bcai_projection");
    expect(config.postgresql?.find((item) => item.key === "host")?.value).toBe("postgres");
  });

  it("builds postgresql connection payload correctly", () => {
    const payload = buildConnectionPayload(
      "postgresql",
      [
        { category: "postgresql", key: "host", value: "postgres", isSecret: false, description: "" },
        { category: "postgresql", key: "port", value: "5432", isSecret: false, description: "" },
        { category: "postgresql", key: "user", value: "postgres", isSecret: false, description: "" },
        { category: "postgresql", key: "dbname", value: "bcai_projection", isSecret: false, description: "" },
      ],
      "12345",
    );
    expect(payload).toMatchObject({
      password: "12345",
      type: "postgresql",
      host: "postgres",
      port: "5432",
      user: "postgres",
      database: "bcai_projection",
    });
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
    expect(getFieldControl(configItem("service", "devapimode", "true"))?.type).toBe("checkbox");
    expect(normalizeBooleanValue("TRUE")).toBe(true);
    expect(normalizeBooleanValue("false")).toBe(false);
  });
});

function configItem(category: string, key: string, value: string): ConfigItem {
  return { category, key, value, isSecret: false, description: "" };
}
