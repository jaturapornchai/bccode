export type ConfigItem = {
  category: string;
  key: string;
  value: string;
  isSecret: boolean;
  description: string;
};

export type CategoryDef = {
  id: string;
  title: string;
  description: string;
  testType?: "mongodb" | "postgresql" | "clickhouse" | "kafka" | "redis" | "http";
  items: Array<Omit<ConfigItem, "value">>;
};

export type ConfigMap = Record<string, ConfigItem[]>;

export type TestResult = {
  status: "success" | "failed" | "testing";
  message: string;
  latencyMs?: number;
  errorCode?: string;
  database?: string;
};

export type AIProvider = {
  id: "openrouter" | "groq" | "deepseek" | "gemini";
  name: string;
  badge: string;
  registerUrl: string;
  description: string;
  freeModels: string[];
  paidModels: string[];
};

export type FieldOption = {
  value: string;
  label: string;
  description?: string;
};

export type FieldControl =
  | { type: "checkbox" }
  | { type: "radio"; options: FieldOption[] };

const hiddenCategoryIds = new Set(["mongodb_production"]);

export const SETUP_CATEGORY_DEFS: CategoryDef[] = [
  {
    id: "mongodb",
    title: "MongoDB",
    description: "ฐานข้อมูลหลักของ GoAPI และ system_config",
    testType: "mongodb",
    items: [
      setupItem("mongodb", "uri", true, "MongoDB Connection URI"),
      setupItem("mongodb", "database", false, "MongoDB Database Name"),
      setupItem("mongodb", "host", false, "MongoDB host แยกจาก URI ถ้ามี"),
      setupItem("mongodb", "port", false, "MongoDB port"),
      setupItem("mongodb", "username", false, "MongoDB username"),
      setupItem("mongodb", "password", true, "MongoDB password"),
    ],
  },
  {
    id: "postgresql",
    title: "PostgreSQL",
    description: "ฐานข้อมูล transaction/mainapi",
    testType: "postgresql",
    items: [
      setupItem("postgresql", "host", false, "PostgreSQL Host"),
      setupItem("postgresql", "port", false, "PostgreSQL Port"),
      setupItem("postgresql", "user", false, "PostgreSQL User"),
      setupItem("postgresql", "password", true, "PostgreSQL Password"),
      setupItem("postgresql", "db_name", false, "PostgreSQL Database Name"),
      setupItem("postgresql", "ssl_mode", false, "PostgreSQL SSL Mode"),
      setupItem("postgresql", "timezone", false, "PostgreSQL Timezone"),
      setupItem("postgresql", "logger_level", false, "PostgreSQL Logger Level"),
    ],
  },
  {
    id: "clickhouse",
    title: "ClickHouse",
    description: "ฐานข้อมูลรายงาน/BI",
    testType: "clickhouse",
    items: [
      setupItem("clickhouse", "host", false, "ClickHouse Host"),
      setupItem("clickhouse", "port", false, "ClickHouse Port"),
      setupItem("clickhouse", "user", false, "ClickHouse User"),
      setupItem("clickhouse", "password", true, "ClickHouse Password"),
      setupItem("clickhouse", "database_name", false, "ClickHouse Database Name"),
    ],
  },
  {
    id: "kafka",
    title: "Kafka",
    description: "Message broker สำหรับ consumer",
    testType: "kafka",
    items: [setupItem("kafka", "server_url", false, "Kafka broker host:port")],
  },
  {
    id: "service",
    title: "Service",
    description: "ค่าระบบและ runtime service",
    items: [
      setupItem("service", "enable_kafka", false, "เปิด/ปิด Kafka"),
      setupItem("service", "kafka_consumer_group_version", false, "Kafka consumer group version"),
      setupItem("service", "enable_clone_clickhouse", false, "เปิด/ปิด clone ClickHouse"),
      setupItem("service", "log_level", false, "Log level"),
      setupItem("service", "jwt_secret_key", true, "JWT Secret Key"),
      setupItem("service", "dev_api_mode", false, "Development API mode"),
      setupItem("service", "service_port", false, "Service port"),
      setupItem("service", "host_api", false, "Host API"),
      setupItem("service", "mode", false, "Runtime mode"),
      setupItem("service", "http_cors", false, "HTTP CORS"),
      setupItem("service", "cors_allowed_origins", false, "CORS allowed origins"),
      setupItem("service", "firebase_project_id", false, "Firebase project id"),
    ],
  },
  {
    id: "integrations",
    title: "Integrations",
    description: "AI provider, object storage และ external services",
    items: [
      setupItem("integrations", "ai_provider", false, "AI provider priority"),
      setupItem("integrations", "openrouter_api_key", true, "OpenRouter API Key"),
      setupItem("integrations", "openrouter_model", false, "OpenRouter model"),
      setupItem("integrations", "groq_api_key", true, "Groq API Key"),
      setupItem("integrations", "groq_model", false, "Groq model"),
      setupItem("integrations", "deepseek_api_key", true, "DeepSeek API Key"),
      setupItem("integrations", "deepseek_model", false, "DeepSeek model"),
      setupItem("integrations", "gemini_api_key", true, "Google Gemini API Key"),
      setupItem("integrations", "gemini_model", false, "Gemini model"),
      setupItem("integrations", "r2_account_id", false, "Cloudflare R2 Account ID"),
      setupItem("integrations", "r2_access_key_id", true, "Cloudflare R2 Access Key"),
      setupItem("integrations", "r2_secret_access_key", true, "Cloudflare R2 Secret Key"),
      setupItem("integrations", "r2_bucket_name", false, "Cloudflare R2 Bucket"),
      setupItem("integrations", "s3_endpoint", false, "S3 endpoint"),
      setupItem("integrations", "s3_public_endpoint", false, "S3 public endpoint"),
      setupItem("integrations", "s3_access_key_id", true, "S3 access key"),
      setupItem("integrations", "s3_secret_access_key", true, "S3 secret key"),
      setupItem("integrations", "s3_bucket_name", false, "S3 bucket"),
      setupItem("integrations", "thunder_api_key", true, "Thunder API Key"),
      setupItem("integrations", "brevo_api_key", true, "Brevo API Key"),
      setupItem("integrations", "brevo_from_email", false, "Brevo sender email"),
      setupItem("integrations", "brevo_from_name", false, "Brevo sender name"),
      setupItem("integrations", "bcweaviate_url", false, "BC Ai Account Weaviate URL"),
    ],
  },
  {
    id: "storage",
    title: "Storage",
    description: "ไฟล์และ object storage",
    items: [
      setupItem("storage", "data_path", false, "Local data path"),
      setupItem("storage", "data_uri", false, "Data URI"),
      setupItem("storage", "azure_account_name", false, "Azure account name"),
      setupItem("storage", "azure_account_key", true, "Azure account key"),
      setupItem("storage", "azure_container_name", false, "Azure container"),
      setupItem("storage", "azure_tenant_id", false, "Azure tenant id"),
      setupItem("storage", "s3_endpoint", false, "S3 endpoint"),
      setupItem("storage", "s3_public_endpoint", false, "S3 public endpoint"),
      setupItem("storage", "s3_access_key_id", true, "S3 access key"),
      setupItem("storage", "s3_secret_access_key", true, "S3 secret key"),
      setupItem("storage", "s3_bucket_name", false, "S3 bucket"),
    ],
  },
];

export const AI_PROVIDERS: AIProvider[] = [
  {
    id: "openrouter",
    name: "OpenRouter",
    badge: "ฟรี + 200+ model",
    registerUrl: "https://openrouter.ai/settings/keys",
    description: "รวมโมเดลหลายค่าย มีตัวฟรีจำนวนมาก",
    freeModels: [
      "nvidia/nemotron-3-nano-30b-a3b:free",
      "stepfun/step-3.5-flash:free",
      "upstage/solar-pro-3:free",
      "arcee-ai/trinity-large-preview:free",
      "arcee-ai/trinity-mini:free",
    ],
    paidModels: ["anthropic/claude-sonnet-4.6", "openai/gpt-4o", "google/gemini-2.5-pro"],
  },
  {
    id: "groq",
    name: "Groq",
    badge: "ฟรี + เร็วมาก",
    registerUrl: "https://console.groq.com/keys",
    description: "Inference เร็ว เหมาะกับงานตอบโต้",
    freeModels: [
      "meta-llama/llama-4-maverick-17b-128e-instruct",
      "qwen/qwen-3-32b",
      "deepseek-r1-distill-llama-70b",
      "llama-3.3-70b-versatile",
    ],
    paidModels: ["meta-llama/llama-4-maverick-17b-128e-instruct"],
  },
  {
    id: "deepseek",
    name: "DeepSeek",
    badge: "ราคาถูก",
    registerUrl: "https://platform.deepseek.com/api_keys",
    description: "คุณภาพสูง ราคาประหยัด",
    freeModels: [],
    paidModels: ["deepseek-chat", "deepseek-reasoner"],
  },
  {
    id: "gemini",
    name: "Google Gemini",
    badge: "Default fallback",
    registerUrl: "https://aistudio.google.com/apikey",
    description: "Provider fallback ถ้า provider อื่นใช้ไม่ได้",
    freeModels: ["gemini-2.5-flash", "gemini-2.5-flash-lite", "gemini-2.5-pro"],
    paidModels: ["gemini-2.5-pro", "gemini-2.5-flash"],
  },
];

export const fieldLabels: Record<string, string> = {
  ai_provider: "AI Provider",
  azure_account_key: "Azure Account Key",
  azure_account_name: "Azure Account Name",
  azure_container_name: "Azure Container",
  azure_tenant_id: "Azure Tenant ID",
  bcweaviate_url: "BC Ai Account Weaviate URL",
  brevo_api_key: "Brevo API Key",
  brevo_from_email: "Brevo From Email",
  brevo_from_name: "Brevo From Name",
  cors_allowed_origins: "CORS Allowed Origins",
  database: "Database",
  database_name: "Database Name",
  data_path: "Data Path",
  data_uri: "Data URI",
  db_name: "Database Name",
  deepseek_api_key: "DeepSeek API Key",
  deepseek_model: "DeepSeek Model",
  dev_api_mode: "Dev API Mode",
  enable_clone_clickhouse: "Enable Clone ClickHouse",
  enable_kafka: "Enable Kafka",
  firebase_project_id: "Firebase Project ID",
  gemini_api_key: "Gemini API Key",
  gemini_model: "Gemini Model",
  groq_api_key: "Groq API Key",
  groq_model: "Groq Model",
  host: "Host",
  host_api: "Host API",
  http_cors: "HTTP CORS",
  jwt_secret_key: "JWT Secret Key",
  kafka_consumer_group_version: "Kafka Consumer Group Version",
  logger_level: "Logger Level",
  log_level: "Log Level",
  mode: "Mode",
  openrouter_api_key: "OpenRouter API Key",
  openrouter_model: "OpenRouter Model",
  password: "Password",
  port: "Port",
  r2_access_key_id: "R2 Access Key",
  r2_account_id: "R2 Account ID",
  r2_bucket_name: "R2 Bucket",
  r2_secret_access_key: "R2 Secret Key",
  s3_access_key_id: "S3 Access Key",
  s3_bucket_name: "S3 Bucket",
  s3_endpoint: "S3 Endpoint",
  s3_public_endpoint: "S3 Public Endpoint",
  s3_secret_access_key: "S3 Secret Key",
  server_url: "Server URL",
  service_port: "Service Port",
  ssl_mode: "SSL Mode",
  thunder_api_key: "Thunder API Key",
  timezone: "Timezone",
  uri: "URI",
  user: "User",
  username: "Username",
};

export const booleanFieldKeys = new Set([
  "enable_kafka",
  "enable_clone_clickhouse",
  "dev_api_mode",
]);

export const radioFieldOptions: Record<string, FieldOption[]> = {
  logger_level: logLevelOptions(),
  log_level: logLevelOptions(),
  mode: [
    { value: "development", label: "Development", description: "เปิด debug behavior ตาม backend ConfigMode" },
    { value: "test", label: "Test", description: "ใช้กับ test/runtime เฉพาะกิจ" },
    { value: "production", label: "Production", description: "โหมด production" },
  ],
  ssl_mode: [
    { value: "disable", label: "Disable", description: "ไม่ใช้ SSL" },
    { value: "require", label: "Require", description: "บังคับใช้ SSL" },
    { value: "verify-ca", label: "Verify CA", description: "ตรวจ CA certificate" },
    { value: "verify-full", label: "Verify Full", description: "ตรวจ CA และ hostname" },
  ],
};

type BackendConfigEntry = {
  category?: unknown;
  key?: unknown;
  value?: unknown;
  is_secret?: unknown;
  description?: unknown;
};

export function createDefaultConfigMap(): ConfigMap {
  return Object.fromEntries(
    SETUP_CATEGORY_DEFS.map((category) => [
      category.id,
      category.items.map((item) => ({ ...item, value: "" })),
    ]),
  );
}

export function mergeBackendConfig(entries: unknown): ConfigMap {
  const configMap = createDefaultConfigMap();
  if (!Array.isArray(entries)) return configMap;

  for (const rawItem of entries) {
    const parsed = parseBackendConfigEntry(rawItem);
    if (!parsed) continue;
    if (hiddenCategoryIds.has(parsed.category)) continue;

    const remapped = remapLegacyConfigItem(parsed);
    const categoryItems = (configMap[remapped.category] ??= []);
    const current = categoryItems.find((item) => item.key === remapped.key);

    if (current) {
      current.value = remapped.value;
      current.isSecret = remapped.isSecret;
      current.description = remapped.description || current.description;
    } else {
      categoryItems.push(remapped);
    }
  }

  normalizeHostPort(configMap);
  return configMap;
}

export function updateConfigItem(configMap: ConfigMap, category: string, key: string, value: string): ConfigMap {
  return Object.fromEntries(
    Object.entries(configMap).map(([categoryId, items]) => [
      categoryId,
      categoryId === category
        ? items.map((item) => (item.key === key ? { ...item, value } : item))
        : items,
    ]),
  );
}

export function serializeConfig(configMap: ConfigMap): ConfigItem[] {
  return Object.values(configMap)
    .flat()
    .map((item) => ({
      category: item.category,
      key: item.key,
      value: item.value,
      isSecret: item.isSecret,
      description: item.description,
    }));
}

export function configMapToExportJson(configMap: ConfigMap): string {
  const data: Record<string, Record<string, string>> = {};
  for (const [category, items] of Object.entries(configMap)) {
    data[category] = {};
    for (const item of items) {
      data[category][item.key] = item.value;
    }
  }
  return JSON.stringify(data, null, 2);
}

export function importConfigJson(configMap: ConfigMap, jsonText: string): ConfigMap {
  const parsed = JSON.parse(jsonText) as unknown;
  if (!isRecord(parsed)) throw new Error("JSON ต้องเป็น object");

  let next = configMap;
  for (const [category, values] of Object.entries(parsed)) {
    if (!isRecord(values)) continue;
    for (const [key, value] of Object.entries(values)) {
      next = updateConfigItem(next, category, key, String(value ?? ""));
    }
  }
  return next;
}

export function validateConfig(configMap: ConfigMap): string[] {
  const errors: string[] = [];
  for (const [category, items] of Object.entries(configMap)) {
    for (const item of items) {
      const value = item.value.trim();
      if (!value || value === "***") continue;

      if (item.key === "port" || item.key === "service_port") {
        const port = Number(value);
        if (!Number.isInteger(port) || port < 1 || port > 65535) {
          errors.push(`${category}.${item.key}: port ไม่ถูกต้อง (${value})`);
        }
      }

      if ((item.key === "uri" || item.key === "server_url") && !looksLikeUriOrHostPort(value)) {
        errors.push(`${category}.${item.key}: URL ไม่ถูกต้อง (${value})`);
      }
    }
  }
  return errors;
}

export function getFieldControl(item: ConfigItem): FieldControl | null {
  if (booleanFieldKeys.has(item.key)) return { type: "checkbox" };

  const options = radioFieldOptions[item.key];
  if (options) return { type: "radio", options };

  return null;
}

export function normalizeBooleanValue(value: string): boolean {
  return ["true", "1", "yes", "y", "on"].includes(value.trim().toLowerCase());
}

export function buildConnectionPayload(category: string, items: ConfigItem[], setupPassword: string) {
  const categoryDef = SETUP_CATEGORY_DEFS.find((item) => item.id === category);
  const testType = categoryDef?.testType ?? category;
  const payload: Record<string, string> = {
    password: setupPassword,
    type: testType,
  };

  for (const item of items) {
    const value = item.value.trim();
    if (!value || value === "***") continue;

    switch (item.key) {
      case "uri":
        payload.uri = value;
        break;
      case "host":
        payload.host = value;
        break;
      case "port":
        payload.port = value;
        break;
      case "user":
      case "username":
        payload.user = value;
        break;
      case "password":
        payload.password2 = value;
        break;
      case "database":
      case "database_name":
      case "db_name":
        payload.database = value;
        break;
      case "server_url": {
        const hostPort = splitHostPort(value);
        payload.host = hostPort.host;
        if (hostPort.port) payload.port = hostPort.port;
        break;
      }
    }
  }

  return payload;
}

export function normalizeSetupBackendUrl(rawUrl: string): string {
  let value = rawUrl.trim();
  if (!value) return "";
  if (!/^https?:\/\//i.test(value)) value = `http://${value}`;

  try {
    const parsed = new URL(value);
    parsed.hash = "";
    parsed.search = "";
    parsed.pathname = parsed.pathname.replace(/\/+$/, "");
    if (!parsed.pathname.toLowerCase().endsWith("/goapi")) {
      parsed.pathname = `${parsed.pathname === "/" ? "" : parsed.pathname}/goapi`;
    }
    return parsed.toString().replace(/\/$/, "");
  } catch {
    return value;
  }
}

export function getCategoryDef(category: string): CategoryDef {
  return (
    SETUP_CATEGORY_DEFS.find((item) => item.id === category) ?? {
      id: category,
      title: category,
      description: "Config จาก backend",
      items: [],
    }
  );
}

function setupItem(category: string, key: string, isSecret: boolean, description: string): Omit<ConfigItem, "value"> {
  return { category, key, isSecret, description };
}

function logLevelOptions(): FieldOption[] {
  return [
    { value: "DEBUG", label: "Debug" },
    { value: "INFO", label: "Info" },
    { value: "SUCCESS", label: "Success" },
    { value: "WARN", label: "Warn" },
    { value: "ERROR", label: "Error" },
    { value: "FATAL", label: "Fatal" },
  ];
}

function parseBackendConfigEntry(rawItem: unknown): ConfigItem | null {
  if (!isRecord(rawItem)) return null;
  const item = rawItem as BackendConfigEntry;
  const category = typeof item.category === "string" ? item.category : "";
  const key = typeof item.key === "string" ? item.key : "";
  if (!category || !key) return null;

  return {
    category,
    key,
    value: typeof item.value === "string" ? item.value : String(item.value ?? ""),
    isSecret: item.is_secret === true || isSecretKey(key),
    description: typeof item.description === "string" ? item.description : "",
  };
}

function remapLegacyConfigItem(item: ConfigItem): ConfigItem {
  if (item.category === "mongodb" && item.key === "database_name") {
    return { ...item, key: "database" };
  }
  return item;
}

function normalizeHostPort(configMap: ConfigMap) {
  for (const items of Object.values(configMap)) {
    const hostItem = items.find((item) => item.key === "host");
    const portItem = items.find((item) => item.key === "port");
    const serverUrlItem = items.find((item) => item.key === "server_url");

    if (serverUrlItem && hostItem && portItem && serverUrlItem.value && !hostItem.value) {
      const hostPort = splitHostPort(serverUrlItem.value);
      hostItem.value = hostPort.host;
      if (hostPort.port) portItem.value = hostPort.port;
    }

    if (hostItem?.value && portItem) {
      const hostPort = splitHostPort(hostItem.value);
      if (hostPort.port) {
        hostItem.value = hostPort.host;
        if (!portItem.value) portItem.value = hostPort.port;
      }
    }

    if (portItem?.value) {
      portItem.value = portItem.value.replace(/[^0-9]/g, "");
    }
  }
}

function splitHostPort(value: string): { host: string; port: string } {
  const cleaned = value.trim().replace(/^https?:\/\//i, "").replace(/\/.*$/, "");
  const lastColonIndex = cleaned.lastIndexOf(":");
  if (lastColonIndex <= 0) return { host: cleaned, port: "" };

  const host = cleaned.slice(0, lastColonIndex);
  const port = cleaned.slice(lastColonIndex + 1);
  return /^\d+$/.test(port) ? { host, port } : { host: cleaned, port: "" };
}

function looksLikeUriOrHostPort(value: string): boolean {
  if (value.includes("://")) {
    try {
      new URL(value);
      return true;
    } catch {
      return false;
    }
  }
  return value.includes(":") || value.length > 0;
}

function isSecretKey(key: string): boolean {
  const normalized = key.toLowerCase();
  return (
    normalized.includes("password") ||
    normalized.includes("secret") ||
    normalized.includes("api_key") ||
    normalized.includes("account_key") ||
    normalized.includes("access_key")
  );
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}
