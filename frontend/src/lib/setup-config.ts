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

const hiddenCategoryIds = new Set(["mongodbdev", "mongodbuat", "mongodbpro", "mongodbproduction", "mongodbproductionlegacy"]);

export const SETUP_CATEGORY_DEFS: CategoryDef[] = [
  {
    id: "mongodb",
    title: "MongoDB",
    description: "ฐานข้อมูลหลักสำหรับ CRUD เอกสาร และ master data",
    testType: "mongodb",
    items: [
      setupItem("mongodb", "uri", true, "MongoDB Connection URI"),
      setupItem("mongodb", "database", false, "MongoDB Database Name"),
      setupItem("mongodb", "host", false, "MongoDB Host"),
      setupItem("mongodb", "port", false, "MongoDB Port"),
      setupItem("mongodb", "username", false, "MongoDB Username"),
      setupItem("mongodb", "password", true, "MongoDB Password"),
    ],
  },
  {
    id: "postgresql",
    title: "PostgreSQL",
    description: "ฐานข้อมูลประมวลผล relation, posting, balance และบัญชี",
    testType: "postgresql",
    items: [
      setupItem("postgresql", "host", false, "PostgreSQL Host"),
      setupItem("postgresql", "port", false, "PostgreSQL Port"),
      setupItem("postgresql", "user", false, "PostgreSQL User"),
      setupItem("postgresql", "password", true, "PostgreSQL Password"),
      setupItem("postgresql", "dbname", false, "PostgreSQL Database Name"),
      setupItem("postgresql", "sslmode", false, "PostgreSQL SSL Mode"),
      setupItem("postgresql", "timezone", false, "PostgreSQL Timezone"),
      setupItem("postgresql", "loggerlevel", false, "PostgreSQL Logger Level"),
    ],
  },
  {
    id: "clickhouse",
    title: "ClickHouse",
    description: "ฐานข้อมูล BI/analytics สำหรับรายงาน",
    testType: "clickhouse",
    items: [
      setupItem("clickhouse", "host", false, "ClickHouse Host"),
      setupItem("clickhouse", "port", false, "ClickHouse Port"),
      setupItem("clickhouse", "user", false, "ClickHouse User"),
      setupItem("clickhouse", "password", true, "ClickHouse Password"),
      setupItem("clickhouse", "databasename", false, "ClickHouse Database Name"),
    ],
  },
  {
    id: "kafka",
    title: "Kafka",
    description: "Message broker สำหรับ consumer",
    testType: "kafka",
    items: [setupItem("kafka", "serverurl", false, "Kafka broker host:port")],
  },
  {
    id: "service",
    title: "Service",
    description: "ค่าระบบและ runtime service",
    items: [
      setupItem("service", "enablekafka", false, "เปิด/ปิด Kafka"),
      setupItem("service", "kafkaconsumergroupversion", false, "Kafka consumer group version"),
      setupItem("service", "enablecloneclickhouse", false, "เปิด/ปิด clone ClickHouse"),
      setupItem("service", "loglevel", false, "Log level"),
      setupItem("service", "jwtsecretkey", true, "JWT Secret Key"),
      setupItem("service", "devapimode", false, "Development API mode"),
      setupItem("service", "serviceport", false, "Service port"),
      setupItem("service", "hostapi", false, "Host API"),
      setupItem("service", "mode", false, "Runtime mode"),
      setupItem("service", "httpcors", false, "HTTP CORS"),
      setupItem("service", "corsallowedorigins", false, "CORS allowed origins"),
      setupItem("service", "firebaseprojectid", false, "Firebase project id"),
    ],
  },
  {
    id: "integrations",
    title: "Integrations",
    description: "AI provider, object storage และ external services",
    items: [
      setupItem("integrations", "aiprovider", false, "AI provider priority"),
      setupItem("integrations", "openrouterapikey", true, "OpenRouter API Key"),
      setupItem("integrations", "openroutermodel", false, "OpenRouter model"),
      setupItem("integrations", "groqapikey", true, "Groq API Key"),
      setupItem("integrations", "groqmodel", false, "Groq model"),
      setupItem("integrations", "deepseekapikey", true, "DeepSeek API Key"),
      setupItem("integrations", "deepseekmodel", false, "DeepSeek model"),
      setupItem("integrations", "geminiapikey", true, "Google Gemini API Key"),
      setupItem("integrations", "geminimodel", false, "Gemini model"),
      setupItem("integrations", "r2accountid", false, "Cloudflare R2 Account ID"),
      setupItem("integrations", "r2accesskeyid", true, "Cloudflare R2 Access Key"),
      setupItem("integrations", "r2secretaccesskey", true, "Cloudflare R2 Secret Key"),
      setupItem("integrations", "r2bucketname", false, "Cloudflare R2 Bucket"),
      setupItem("integrations", "s3endpoint", false, "S3 endpoint"),
      setupItem("integrations", "s3publicendpoint", false, "S3 public endpoint"),
      setupItem("integrations", "s3accesskeyid", true, "S3 access key"),
      setupItem("integrations", "s3secretaccesskey", true, "S3 secret key"),
      setupItem("integrations", "s3bucketname", false, "S3 bucket"),
      setupItem("integrations", "thunderapikey", true, "Thunder API Key"),
      setupItem("integrations", "brevoapikey", true, "Brevo API Key"),
      setupItem("integrations", "brevofromemail", false, "Brevo sender email"),
      setupItem("integrations", "brevofromname", false, "Brevo sender name"),
      setupItem("integrations", "bcweaviateurl", false, "BC Ai Account Weaviate URL"),
    ],
  },
  {
    id: "storage",
    title: "Storage",
    description: "ไฟล์และ object storage",
    items: [
      setupItem("storage", "datapath", false, "Local data path"),
      setupItem("storage", "datauri", false, "Data URI"),
      setupItem("storage", "azureaccountname", false, "Azure account name"),
      setupItem("storage", "azureaccountkey", true, "Azure account key"),
      setupItem("storage", "azurecontainername", false, "Azure container"),
      setupItem("storage", "azuretenantid", false, "Azure tenant id"),
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
    registerUrl: "https://platform.deepseek.com/apikeys",
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
  aiprovider: "AI Provider",
  azureaccountkey: "Azure Account Key",
  azureaccountname: "Azure Account Name",
  azurecontainername: "Azure Container",
  azuretenantid: "Azure Tenant ID",
  bcweaviateurl: "BC Ai Account Weaviate URL",
  brevoapikey: "Brevo API Key",
  brevofromemail: "Brevo From Email",
  brevofromname: "Brevo From Name",
  corsallowedorigins: "CORS Allowed Origins",
  database: "Database",
  databasename: "Database Name",
  datapath: "Data Path",
  datauri: "Data URI",
  dbname: "Database Name",
  deepseekapikey: "DeepSeek API Key",
  deepseekmodel: "DeepSeek Model",
  devapimode: "Dev API Mode",
  enablecloneclickhouse: "Enable Clone ClickHouse",
  enablekafka: "Enable Kafka",
  firebaseprojectid: "Firebase Project ID",
  geminiapikey: "Gemini API Key",
  geminimodel: "Gemini Model",
  groqapikey: "Groq API Key",
  groqmodel: "Groq Model",
  host: "Host",
  hostapi: "Host API",
  httpcors: "HTTP CORS",
  jwtsecretkey: "JWT Secret Key",
  kafkaconsumergroupversion: "Kafka Consumer Group Version",
  loggerlevel: "Logger Level",
  loglevel: "Log Level",
  mode: "Mode",
  openrouterapikey: "OpenRouter API Key",
  openroutermodel: "OpenRouter Model",
  password: "Password",
  port: "Port",
  r2accesskeyid: "R2 Access Key",
  r2accountid: "R2 Account ID",
  r2bucketname: "R2 Bucket",
  r2secretaccesskey: "R2 Secret Key",
  s3accesskeyid: "S3 Access Key",
  s3bucketname: "S3 Bucket",
  s3endpoint: "S3 Endpoint",
  s3publicendpoint: "S3 Public Endpoint",
  s3secretaccesskey: "S3 Secret Key",
  serverurl: "Server URL",
  serviceport: "Service Port",
  sslmode: "SSL Mode",
  thunderapikey: "Thunder API Key",
  timezone: "Timezone",
  uri: "URI",
  user: "User",
  username: "Username",
};

export const booleanFieldKeys = new Set([
  "enablekafka",
  "enablecloneclickhouse",
  "devapimode",
]);

export const radioFieldOptions: Record<string, FieldOption[]> = {
  loggerlevel: logLevelOptions(),
  loglevel: logLevelOptions(),
  mode: [
    { value: "development", label: "Development", description: "เปิด debug behavior ตาม backend ConfigMode" },
    { value: "test", label: "Test", description: "ใช้กับ test/runtime เฉพาะกิจ" },
    { value: "production", label: "Production", description: "โหมด production" },
  ],
  sslmode: [
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
  issecret?: unknown;
  description?: unknown;
};

export function createDefaultConfigMap(): ConfigMap {
  return Object.fromEntries(
    SETUP_CATEGORY_DEFS
      .filter((category) => !hiddenCategoryIds.has(category.id))
      .map((category) => [
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
  const nextConfigMap = Object.fromEntries(
    Object.entries(configMap).map(([categoryId, items]) => [
      categoryId,
      categoryId === category
        ? items.map((item) => (item.key === key ? { ...item, value } : item))
        : items,
    ]),
  );

  // คำนวณ URI อัตโนมัติสำหรับ mongodb เมื่อฟิลด์อื่นที่ไม่ใช่ uri มีการเปลี่ยนแปลง
  if (category.startsWith("mongodb") && key !== "uri") {
    const items = nextConfigMap[category] ?? [];
    const host = items.find((i) => i.key === "host")?.value.trim() ?? "";
    const port = items.find((i) => i.key === "port")?.value.trim() ?? "";
    const user = items.find((i) => i.key === "username")?.value.trim() ?? "";
    const pass = items.find((i) => i.key === "password")?.value.trim() ?? "";
    const db = items.find((i) => i.key === "database")?.value.trim() ?? "";

    if (host) {
      let authStr = "";
      if (user) {
        authStr = pass ? `${encodeURIComponent(user)}:${encodeURIComponent(pass)}@` : `${encodeURIComponent(user)}@`;
      }
      const dbStr = db ? `/${db}` : "";

      let generatedUri = "";
      if (host.toLowerCase().includes("mongodb.net")) {
        // สำหรับ MongoDB Atlas (Cloud)
        generatedUri = `mongodb+srv://${authStr}${host}${dbStr}?retryWrites=true&w=majority`;
      } else {
        // สำหรับ Local / Server ทั่วไป
        const portStr = port ? `:${port}` : "";
        generatedUri = `mongodb://${authStr}${host}${portStr}${dbStr}`;
      }

      return Object.fromEntries(
        Object.entries(nextConfigMap).map(([categoryId, items]) => [
          categoryId,
          categoryId === category
            ? items.map((item) => (item.key === "uri" ? { ...item, value: generatedUri } : item))
            : items,
        ]),
      );
    }
  }

  return nextConfigMap;
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

      if (item.key === "port" || item.key === "serviceport") {
        const port = Number(value);
        if (!Number.isInteger(port) || port < 1 || port > 65535) {
          errors.push(`${category}.${item.key}: port ไม่ถูกต้อง (${value})`);
        }
      }

      if ((item.key === "uri" || item.key === "serverurl") && !looksLikeUriOrHostPort(value)) {
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
      case "databasename":
      case "dbname":
        payload.database = value;
        break;
      case "serverurl": {
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
    isSecret: item.issecret === true || isSecretKey(key),
    description: typeof item.description === "string" ? item.description : "",
  };
}

function remapLegacyConfigItem(item: ConfigItem): ConfigItem {
  if (item.category === "mongodb" && item.key === "databasename") {
    return { ...item, key: "database" };
  }
  return item;
}

function normalizeHostPort(configMap: ConfigMap) {
  for (const items of Object.values(configMap)) {
    const hostItem = items.find((item) => item.key === "host");
    const portItem = items.find((item) => item.key === "port");
    const serverUrlItem = items.find((item) => item.key === "serverurl");

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
    normalized.includes("apikey") ||
    normalized.includes("accountkey") ||
    normalized.includes("accesskey")
  );
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}
