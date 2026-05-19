import { currencySymbolPresets } from "@/lib/currency-presets";
import { LANGUAGES, type LanguageCode } from "@/lib/i18n";

export type SystemSettingText = Partial<Record<LanguageCode, string>> & { th: string; en: string };
export type SystemSettingOption = {
  value: string;
  label: string;
  labels?: SystemSettingText;
};

export type SystemSettingField = {
  key: string;
  label: SystemSettingText;
  type: "checkbox" | "combo" | "date" | "image-upload" | "json" | "names" | "number" | "radio" | "select" | "text" | "textarea";
  required?: boolean;
  readOnly?: boolean;
  options?: SystemSettingOption[];
  optionSource?: "countries" | "currency" | "timezones";
  placeholder?: string;
  multiline?: boolean;
  valueType?: "boolean" | "string";
};

export type SystemSettingKind =
  | "ai-provider"
  | "atlas"
  | "company"
  | "copy-uat"
  | "goapi-crud"
  | "main-crud"
  | "restaurant-setting";

export type SystemSettingConfig = {
  slug: string;
  route: string;
  manual: string;
  kind: SystemSettingKind;
  title: Partial<Record<LanguageCode, string>> & { th: string; en: string };
  subtitle: Partial<Record<LanguageCode, string>> & { th: string; en: string };
  icon: string;
  basePath?: string;
  listPath?: string;
  idField?: string;
  code?: string;
  collection?: string;
  editable?: boolean;
  deleteKey?: string;
  fields: SystemSettingField[];
};

const buddhistEraLabel: SystemSettingText = {
  th: "พุทธศักราช (พ.ศ.)",
  en: "Buddhist Era (B.E.)",
  cn: "佛历",
  ja: "仏暦",
  ko: "불력",
  lo: "ພຸດທະສັກກະຣາດ",
  my: "ဗုဒ္ဓသက္ကရာဇ်",
  km: "ពុទ្ធសករាជ",
  vi: "Phật lịch",
  ms: "Buddhist Era (B.E.)",
  id: "Buddhist Era (B.E.)",
  fil: "Buddhist Era (B.E.)",
};

const christianEraLabel: SystemSettingText = {
  th: "คริสต์ศักราช (ค.ศ.)",
  en: "Christian Era (A.D.)",
  cn: "公元",
  ja: "西暦",
  ko: "서력기원",
  lo: "ສາກົນ",
  my: "ခရစ်နှစ်",
  km: "គ្រិស្តសករាជ",
  vi: "Công nguyên",
  ms: "Christian Era (A.D.)",
  id: "Christian Era (A.D.)",
  fil: "Christian Era (A.D.)",
};

const accessStatusLabel: SystemSettingText = {
  th: "สถานะเข้าใช้งาน",
  en: "Access status",
  cn: "访问状态",
  ja: "アクセス状態",
  ko: "접속 상태",
  lo: "ສະຖານະການເຂົ້າໃຊ້ງານ",
  my: "ဝင်ရောက်အသုံးပြုမှုအခြေအနေ",
  km: "ស្ថានភាពចូលប្រើ",
  vi: "Trạng thái truy cập",
  ms: "Status akses",
  id: "Status akses",
  fil: "Katayuan ng access",
};

const accessEnabledLabel: SystemSettingText = {
  th: "เข้าใช้งานได้",
  en: "Can access",
  cn: "允许访问",
  ja: "アクセス可能",
  ko: "접속 가능",
  lo: "ເຂົ້າໃຊ້ງານໄດ້",
  my: "ဝင်ရောက်အသုံးပြုနိုင်သည်",
  km: "អាចចូលប្រើបាន",
  vi: "Có thể truy cập",
  ms: "Boleh akses",
  id: "Dapat mengakses",
  fil: "Maaaring maka-access",
};

const accessTemporarilyDisabledLabel: SystemSettingText = {
  th: "เข้าใช้งานไม่ได้ชั่วคราว",
  en: "Temporarily disabled",
  cn: "暂时禁止访问",
  ja: "一時的に無効",
  ko: "일시적으로 접속 중지",
  lo: "ປິດການເຂົ້າໃຊ້ງານຊົ່ວຄາວ",
  my: "ယာယီဝင်ရောက်ခွင့်ပိတ်ထားသည်",
  km: "បិទការចូលប្រើបណ្តោះអាសន្ន",
  vi: "Tạm thời bị khóa",
  ms: "Dilumpuhkan sementara",
  id: "Dinonaktifkan sementara",
  fil: "Pansamantalang hindi pinapayagan",
};

const userIdGuidLabel: SystemSettingText = {
  th: "รหัสผู้ใช้ (GUID)",
  en: "User ID (GUID)",
  cn: "用户 ID (GUID)",
  ja: "ユーザーID (GUID)",
  ko: "사용자 ID (GUID)",
  lo: "ລະຫັດຜູ້ໃຊ້ (GUID)",
  my: "အသုံးပြုသူ ID (GUID)",
  km: "លេខសម្គាល់អ្នកប្រើ (GUID)",
  vi: "ID người dùng (GUID)",
  ms: "ID pengguna (GUID)",
  id: "ID pengguna (GUID)",
  fil: "User ID (GUID)",
};

const loginUsernameLabel: SystemSettingText = {
  th: "ชื่อผู้ใช้สำหรับเข้าสู่ระบบ",
  en: "Login username",
  cn: "登录用户名",
  ja: "ログインユーザー名",
  ko: "로그인 사용자 이름",
  lo: "ຊື່ຜູ້ໃຊ້ສຳລັບເຂົ້າລະບົບ",
  my: "ဝင်ရောက်ရန် အသုံးပြုသူအမည်",
  km: "ឈ្មោះអ្នកប្រើសម្រាប់ចូលប្រព័ន្ធ",
  vi: "Tên đăng nhập",
  ms: "Nama pengguna log masuk",
  id: "Nama pengguna login",
  fil: "Username sa pag-login",
};

const registeredEmailLabel: SystemSettingText = {
  th: "อีเมลที่ลงทะเบียน",
  en: "Registered email",
  cn: "注册邮箱",
  ja: "登録メール",
  ko: "등록 이메일",
  lo: "ອີເມວທີ່ລົງທະບຽນ",
  my: "မှတ်ပုံတင်ထားသော အီးမေးလ်",
  km: "អ៊ីមែលដែលបានចុះឈ្មោះ",
  vi: "Email đã đăng ký",
  ms: "E-mel berdaftar",
  id: "Email terdaftar",
  fil: "Nakarehistrong email",
};

const vatNotRegisteredLabel: SystemSettingText = {
  th: "ไม่จดทะเบียนภาษีมูลค่าเพิ่ม",
  en: "Not Registered for VAT",
  cn: "未注册增值税",
  ja: "VAT未登録",
  ko: "부가세 미등록",
  lo: "ບໍ່ໄດ້ຈົດທະບຽນພາສີ",
  my: "အခွန်မမှတ်ပုံတင်ထားပါ",
  km: "មិនបានចុះបញ្ជីអាករ",
  vi: "Không đăng ký thuế GTGT",
  ms: "Not Registered for VAT",
  id: "Not Registered for VAT",
  fil: "Not Registered for VAT",
};

const vatRegisteredLabel: SystemSettingText = {
  th: "จดทะเบียนภาษีมูลค่าเพิ่ม",
  en: "Registered for VAT",
  cn: "注册增值税",
  ja: "VAT登録済み",
  ko: "부가세 등록",
  lo: "ຈົດທະບຽນພາສີ",
  my: "အခွန်မှတ်ပုံတင်ထားသည်",
  km: "ចុះបញ្ជីអាករ",
  vi: "Đăng ký thuế GTGT",
  ms: "Registered for VAT",
  id: "Registered for VAT",
  fil: "Registered for VAT",
};

export const SYSTEM_SETTING_CONFIGS: SystemSettingConfig[] = [
  {
    slug: "business_type_screen",
    route: "/business_type_screen",
    manual: "business_type_screen",
    kind: "main-crud",
    icon: "tag",
    basePath: "/organization/business-type",
    listPath: "/organization/business-type/list",
    idField: "guidfixed",
    title: { th: "ประเภทธุรกิจ", en: "Business Type" },
    subtitle: { th: "จัดการรหัสประเภทธุรกิจและชื่อตามภาษาที่เลือก", en: "Manage business type codes and names for the selected language." },
    fields: [
      textField("code", "รหัส", "Code", true),
      namesField("names", "ชื่อประเภทธุรกิจ", "Business type names"),
      checkboxField("isdefault", "ค่าเริ่มต้น", "Default"),
    ],
  },
  {
    slug: "company",
    route: "/company",
    manual: "company",
    kind: "company",
    icon: "building",
    idField: "guidfixed",
    title: { th: "บริษัท", en: "Company" },
    subtitle: { th: "แก้ไขข้อมูลกิจการปัจจุบัน ภาษา สกุลเงินหลัก และข้อมูลติดต่อ", en: "Edit the selected company, language settings, base currency, and contact data." },
    fields: [
      namesField("names", "ชื่อบริษัท", "Company names"),
      namesField("address", "ที่อยู่", "Address", true),
      textField("telephone", "โทรศัพท์", "Telephone"),
      imageUploadField("logo", "โลโก้บริษัท", "Company logo"),
      comboField("settings.countrycode", "ประเทศ", "Country", countryOptions(), false, "countries"),
      selectField("settings.language", "ภาษาหลัก", "Main language", languageOptions()),
      textField("settings.taxid", "เลขผู้เสียภาษี", "Tax ID"),
      textField("settings.company_registration_no", "เลขทะเบียนบริษัท", "Company registration no."),
      radioField("settings.is_vat_registered", "สถานะ VAT", "VAT status", [
        { value: "false", label: vatNotRegisteredLabel.th, labels: vatNotRegisteredLabel },
        { value: "true", label: vatRegisteredLabel.th, labels: vatRegisteredLabel },
      ], false, "boolean"),
      numberField("settings.vatrate", "อัตรา VAT เริ่มต้น", "Default VAT rate"),
      comboField("settings.base_currency", "สกุลเงินหลัก", "Base currency", currencyOptions(), false, "currency"),
      comboField("settings.timezone", "Timezone", "Timezone", [], false, "timezones"),
      selectField("settings.date_format", "รูปแบบวันที่", "Date format", dateFormatOptions()),
      numberField("settings.decimal_quantity", "ทศนิยมจำนวน", "Quantity decimals"),
      numberField("settings.decimal_price", "ทศนิยมราคา", "Price decimals"),
      numberField("settings.decimal_document", "ทศนิยมเอกสาร", "Document decimals"),
      checkboxField("settings.isusebranch", "ใช้ระบบสาขา", "Use branches"),
      checkboxField("settings.isusedepartment", "ใช้ระบบแผนก", "Use departments"),
      radioField("settings.usebuddhistcalendar", "รูปแบบปี", "Year type", [
        { value: "true", label: buddhistEraLabel.th, labels: buddhistEraLabel },
        { value: "false", label: christianEraLabel.th, labels: christianEraLabel },
      ], true, "boolean"),
    ],
  },
  {
    slug: "branch",
    route: "/branch",
    manual: "branch",
    kind: "main-crud",
    icon: "branch",
    basePath: "/organization/branch",
    listPath: "/organization/branch/list",
    idField: "guidfixed",
    title: { th: "สาขา", en: "Branch" },
    subtitle: { th: "จัดการสาขา สกุลเงิน ภาษา timezone และคุณสมบัติธุรกิจ", en: "Manage branches, base currency, language, timezone, and business flags." },
    fields: [
      textField("code", "รหัสสาขา", "Branch code", true),
      namesField("names", "ชื่อสาขา", "Branch names"),
      namesField("companynames", "ชื่อบริษัทบนเอกสาร", "Company names on documents"),
      comboField("base_currency", "สกุลเงินหลัก", "Base currency", currencyOptions(), false, "currency"),
      selectField("language", "ภาษาหลัก", "Main language", languageOptions()),
      comboField("timezone", "Timezone", "Timezone", [], false, "timezones"),
      textField("date_format", "รูปแบบวันที่", "Date format"),
      selectField("yeartype", "ประเภทปี", "Year type", [
        { value: "buddhist", label: buddhistEraLabel.th, labels: buddhistEraLabel },
        { value: "christian", label: christianEraLabel.th, labels: christianEraLabel },
      ]),
      numberField("decimal_quantity", "ทศนิยมจำนวน", "Quantity decimals"),
      numberField("decimal_price", "ทศนิยมราคา", "Price decimals"),
      numberField("decimal_document", "ทศนิยมเอกสาร", "Document decimals"),
      checkboxField("is_restaurant", "ร้านอาหาร", "Restaurant"),
      checkboxField("is_retail", "ค้าปลีก", "Retail"),
      checkboxField("is_service", "บริการ", "Service"),
      checkboxField("is_wholesale", "ค้าส่ง", "Wholesale"),
    ],
  },
  {
    slug: "department",
    route: "/department",
    manual: "department",
    kind: "main-crud",
    icon: "network",
    basePath: "/organization/department",
    listPath: "/organization/department/list",
    idField: "guidfixed",
    title: { th: "แผนก", en: "Department" },
    subtitle: { th: "จัดการรหัสแผนกและชื่อตามภาษาที่เลือก", en: "Manage department codes and names for the selected language." },
    fields: [textField("code", "รหัสแผนก", "Department code", true), namesField("names", "ชื่อแผนก", "Department names")],
  },
  {
    slug: "work_day_screen",
    route: "/work_day_screen",
    manual: "work_day_screen",
    kind: "restaurant-setting",
    icon: "calendar",
    code: "workDay",
    idField: "guidfixed",
    title: { th: "วันทำงาน", en: "Work Day" },
    subtitle: { th: "ตั้งค่าวันทำงานและช่วงเวลาทำงานตามระบบเดิม", en: "Configure weekly work days and working time ranges." },
    fields: [jsonField("body", "ข้อมูลวันทำงาน", "Work day JSON")],
  },
  {
    slug: "holiday_screen",
    route: "/holiday_screen",
    manual: "holiday_screen",
    kind: "restaurant-setting",
    icon: "calendar-check",
    code: "Holiday",
    idField: "guidfixed",
    title: { th: "วันหยุด", en: "Holiday" },
    subtitle: { th: "จัดการวันหยุดและคำอธิบายตามภาษาที่เลือก", en: "Manage holidays and descriptions for the selected language." },
    fields: [dateField("date", "วันที่", "Date", true), namesField("desc", "คำอธิบาย", "Description")],
  },
  {
    slug: "employee",
    route: "/employee",
    manual: "employee",
    kind: "main-crud",
    icon: "user-round",
    basePath: "/shop/employee",
    listPath: "/shop/employee/list",
    idField: "guidfixed",
    title: { th: "พนักงาน", en: "Employee" },
    subtitle: { th: "จัดการพนักงาน อีเมล สถานะ POS และ PIN", en: "Manage employees, email, POS access, and PIN." },
    fields: [
      textField("code", "รหัสพนักงาน", "Employee code", true),
      textField("name", "ชื่อพนักงาน", "Employee name", true),
      textField("email", "อีเมล", "Email"),
      textField("pincode", "PIN", "PIN"),
      checkboxField("isenabled", "เปิดใช้งาน", "Enabled"),
      checkboxField("isusepos", "ใช้งาน POS", "Use POS"),
      jsonField("branches", "สาขาที่ใช้งาน", "Branches JSON"),
    ],
  },
  {
    slug: "user",
    route: "/user",
    manual: "user",
    kind: "main-crud",
    icon: "users",
    basePath: "/shop/permission",
    listPath: "/shop/users",
    idField: "username",
    title: { th: "ผู้ใช้งาน", en: "User" },
    subtitle: { th: "จัดการผู้ใช้งาน บทบาท แผนก LINE และสิทธิ์อนุมัติ", en: "Manage users, role, department, LINE profile, and approval permissions." },
    fields: [
      { key: "uid", label: userIdGuidLabel, type: "text", readOnly: true },
      { ...textField("username", loginUsernameLabel.th, loginUsernameLabel.en, true), label: loginUsernameLabel },
      { ...textField("email", registeredEmailLabel.th, registeredEmailLabel.en), label: registeredEmailLabel, readOnly: true },
      numberField("role", "บทบาท", "Role"),
      radioField("isaccessdisabled", accessStatusLabel.th, accessStatusLabel.en, [
        { value: "false", label: accessEnabledLabel.en, labels: accessEnabledLabel },
        { value: "true", label: accessTemporarilyDisabledLabel.en, labels: accessTemporarilyDisabledLabel },
      ], true, "boolean"),
      textField("position", "ตำแหน่ง", "Position"),
      textField("department", "แผนก", "Department"),
      { ...textField("line_user_id", "LINE User ID", "LINE User ID"), readOnly: true },
      { ...textField("line_display_name", "ชื่อ LINE", "LINE display name"), readOnly: true },
    ],
  },
  {
    slug: "formdesign",
    route: "/formdesign",
    manual: "formdesign",
    kind: "main-crud",
    icon: "file-cog",
    basePath: "/form/template",
    listPath: "/form/template",
    idField: "guidfixed",
    title: { th: "ออกแบบฟอร์ม", en: "Form Design" },
    subtitle: { th: "จัดการ template ฟอร์มเอกสารและข้อมูลแบบ JSON", en: "Manage document form templates and template JSON data." },
    fields: [
      textField("code", "รหัสฟอร์ม", "Form code", true),
      namesField("names", "ชื่อฟอร์ม", "Form names"),
      textField("doctype", "ประเภทเอกสาร", "Document type"),
      checkboxField("isdefault", "ค่าเริ่มต้น", "Default"),
      jsonField("templatedata", "Template Data", "Template data"),
    ],
  },
  {
    slug: "line_notify",
    route: "/line_notify",
    manual: "line_notify",
    kind: "main-crud",
    icon: "bell",
    basePath: "/notify",
    listPath: "/notify/list",
    idField: "guidfixed",
    title: { th: "LINE Notify", en: "LINE Notify" },
    subtitle: { th: "จัดการ token และ event แจ้งเตือนของสาขา", en: "Manage notification tokens and branch events." },
    fields: [
      textField("name", "ชื่อ", "Name", true),
      textField("type", "ประเภท", "Type"),
      textField("token", "Token", "Token"),
      jsonField("branchevents", "Branch Events", "Branch events JSON"),
      jsonField("options", "Options", "Options JSON"),
    ],
  },
  {
    slug: "approval_setting",
    route: "/approval_setting",
    manual: "approval_setting",
    kind: "atlas",
    icon: "shield",
    collection: "approval_settings",
    idField: "approvalCode",
    deleteKey: "approvalCode",
    title: { th: "สิทธิ์การอนุมัติ", en: "Approval Permission" },
    subtitle: { th: "กำหนดสิทธิ์และวงเงินอนุมัติแยกตามบริษัท", en: "Configure approval roles and limits by company." },
    fields: [
      textField("approvalCode", "รหัส", "Code", true),
      textField("approvalName", "ชื่อ", "Name", true),
      textareaField("description", "คำอธิบาย", "Description"),
      checkboxField("isActive", "เปิดใช้งาน", "Active"),
      jsonField("approvals", "สิทธิ์การอนุมัติ", "Approval permission JSON"),
    ],
  },
  {
    slug: "permission_definition",
    route: "/permission_definition",
    manual: "permission_definition",
    kind: "atlas",
    icon: "shield",
    collection: "permission_definitions",
    idField: "permissionCode",
    deleteKey: "permissionCode",
    title: { th: "กำหนดสิทธิ์", en: "Permission Definition" },
    subtitle: { th: "สร้างรหัสสิทธิ์และกำหนดสิทธิ์แยกตามสาขา/หน้าจอ", en: "Create permission codes and define screen permissions per branch." },
    fields: [
      textField("permissionCode", "รหัสสิทธิ์", "Permission code", true),
      textField("permissionName", "ชื่อสิทธิ์", "Permission name", true),
      textareaField("description", "คำอธิบาย", "Description"),
      checkboxField("isActive", "เปิดใช้งาน", "Active"),
      jsonField("branches", "สิทธิ์ตามสาขา", "Branch permissions JSON"),
    ],
  },
  {
    slug: "permission_link",
    route: "/permission_link",
    manual: "permission_link",
    kind: "atlas",
    icon: "link",
    collection: "employee_permissions",
    idField: "employeeCode",
    deleteKey: "employeeCode",
    title: { th: "ผูกสิทธิ์", en: "Permission Link" },
    subtitle: { th: "เชื่อมผู้ใช้/พนักงานกับรหัสสิทธิ์", en: "Link users or employees to permission codes." },
    fields: [
      textField("employeeCode", "รหัสผู้ใช้/พนักงาน", "User/employee code", true),
      textField("employeeName", "ชื่อ", "Name"),
      jsonField("permissionCodes", "สิทธิ์", "Permissions"),
      jsonField("approvalCodes", "สิทธิ์การอนุมัติ", "Approval permissions"),
    ],
  },
  {
    slug: "mcp_apikey",
    route: "/mcp_apikey",
    manual: "mcp_apikey",
    kind: "goapi-crud",
    icon: "key",
    basePath: "/api/mcp/keys",
    idField: "id",
    title: { th: "MCP Token", en: "MCP Token" },
    subtitle: { th: "จัดการ API key สำหรับ MCP และ export config", en: "Manage MCP API keys and export client configuration." },
    fields: [
      textField("name", "ชื่อ Token", "Token name", true),
      textareaField("description", "คำอธิบาย", "Description"),
      jsonField("allowed_tools", "Allowed Tools", "Allowed tools JSON"),
      numberField("rate_limit_per_minute", "Rate/min", "Rate/min"),
      dateField("expires_at", "วันหมดอายุ", "Expires at"),
      checkboxField("is_active", "เปิดใช้งาน", "Active"),
    ],
  },
  {
    slug: "ai_provider",
    route: "/ai_provider",
    manual: "ai_provider",
    kind: "ai-provider",
    icon: "bot",
    idField: "provider_name",
    title: { th: "AI Provider", en: "AI Provider" },
    subtitle: { th: "จัดการ provider, model, API key และลำดับใช้งานของ AI", en: "Manage AI providers, models, API keys, and priority." },
    fields: [
      selectField("provider_name", "Provider", "Provider", [
        { value: "ollama", label: "Ollama" },
        { value: "groq", label: "Groq" },
        { value: "openrouter", label: "OpenRouter" },
        { value: "deepseek", label: "DeepSeek" },
        { value: "gemini", label: "Google Gemini" },
        { value: "custom", label: "Custom" },
      ], true),
      textField("model", "Model", "Model", true),
      textField("api_key", "API Key", "API Key"),
      textField("base_url", "Base URL", "Base URL"),
      numberField("priority", "ลำดับ", "Priority"),
      checkboxField("is_active", "เปิดใช้งาน", "Active"),
    ],
  },
  {
    slug: "copy_uat_to_dev",
    route: "/copy_uat_to_dev",
    manual: "copy_uat_to_dev",
    kind: "copy-uat",
    icon: "download-cloud",
    editable: false,
    title: { th: "โอนข้อมูล UAT ไป DEV", en: "Copy UAT to DEV" },
    subtitle: { th: "เลือก shop ต้นทาง ดู preview แล้วสั่งโอน MongoDB ไปยัง shop ปัจจุบัน", en: "Select a source shop, preview counts, then copy MongoDB data into the current shop." },
    fields: [],
  },
];

export const SYSTEM_SETTING_SLUGS = SYSTEM_SETTING_CONFIGS.map((item) => item.slug);

export function getSystemSettingConfig(routeOrSlug: string): SystemSettingConfig | undefined {
  const normalized = routeOrSlug.startsWith("/") ? routeOrSlug : `/${routeOrSlug}`;
  return SYSTEM_SETTING_CONFIGS.find((item) => item.route === normalized || item.slug === routeOrSlug);
}

export function systemSettingLabel(config: SystemSettingConfig, language: LanguageCode): string {
  return config.title[language] ?? config.title.en ?? config.title.th;
}

function textField(key: string, th: string, en: string, required = false): SystemSettingField {
  return { key, label: { th, en }, type: "text", required };
}

function textareaField(key: string, th: string, en: string): SystemSettingField {
  return { key, label: { th, en }, type: "textarea" };
}

function numberField(key: string, th: string, en: string): SystemSettingField {
  return { key, label: { th, en }, type: "number" };
}

function dateField(key: string, th: string, en: string, required = false): SystemSettingField {
  return { key, label: { th, en }, type: "date", required };
}

function imageUploadField(key: string, th: string, en: string): SystemSettingField {
  return { key, label: { th, en }, type: "image-upload" };
}

function checkboxField(key: string, th: string, en: string): SystemSettingField {
  return { key, label: { th, en }, type: "checkbox" };
}

function namesField(key: string, th: string, en: string, multiline = false): SystemSettingField {
  return { key, label: { th, en }, type: "names", required: true, multiline };
}

function jsonField(key: string, th: string, en: string): SystemSettingField {
  return { key, label: { th, en }, type: "json" };
}

function comboField(
  key: string,
  th: string,
  en: string,
  options: SystemSettingOption[] = [],
  required = false,
  optionSource?: "countries" | "currency" | "timezones",
): SystemSettingField {
  return { key, label: { th, en }, type: "combo", options, required, optionSource };
}

function radioField(
  key: string,
  th: string,
  en: string,
  options: SystemSettingOption[],
  required = false,
  valueType: "boolean" | "string" = "string",
): SystemSettingField {
  return { key, label: { th, en }, type: "radio", options, required, valueType };
}

function selectField(key: string, th: string, en: string, options: SystemSettingOption[], required = false): SystemSettingField {
  return { key, label: { th, en }, type: "select", options, required };
}

function languageOptions() {
  return LANGUAGES.map((item) => ({ value: item.code, label: item.name }));
}

function currencyOptions() {
  return currencySymbolPresets.map((item) => ({ value: item.code, label: `${item.code} - ${item.name}` }));
}

function countryOptions() {
  return [
    { value: "TH", label: "Thailand" },
    { value: "LA", label: "Laos" },
    { value: "MM", label: "Myanmar" },
    { value: "KH", label: "Cambodia" },
    { value: "VN", label: "Vietnam" },
    { value: "MY", label: "Malaysia" },
    { value: "ID", label: "Indonesia" },
    { value: "PH", label: "Philippines" },
    { value: "SG", label: "Singapore" },
    { value: "BN", label: "Brunei" },
    { value: "CN", label: "China" },
    { value: "JP", label: "Japan" },
    { value: "KR", label: "South Korea" },
    { value: "IN", label: "India" },
    { value: "HK", label: "Hong Kong" },
    { value: "TW", label: "Taiwan" },
    { value: "US", label: "United States" },
    { value: "GB", label: "United Kingdom" },
    { value: "AU", label: "Australia" },
    { value: "CA", label: "Canada" },
  ];
}

function dateFormatOptions() {
  return [
    { value: "dd/MM/yyyy", label: "dd/MM/yyyy" },
    { value: "yyyy-MM-dd", label: "yyyy-MM-dd" },
    { value: "MM/dd/yyyy", label: "MM/dd/yyyy" },
    { value: "dd-MM-yyyy", label: "dd-MM-yyyy" },
  ];
}
