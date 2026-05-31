import { currencySymbolPresets } from "@/lib/currency-presets";
import { LANGUAGES, type LanguageCode } from "@/lib/i18n";

export type SystemSettingText = Partial<Record<LanguageCode, string>> & {
  th: string;
  en: string;
};
export type SystemSettingOption = {
  value: string;
  label: string;
  labels?: SystemSettingText;
};

export type SystemSettingField = {
  key: string;
  label: SystemSettingText;
  type:
    | "branch-multi-select"
    | "company-multi-select"
    | "checkbox"
    | "combo"
    | "date"
    | "image-upload"
    | "image-gallery"
    | "json"
    | "language-configs"
    | "language-list"
    | "master-picker"
    | "names"
    | "number"
    | "radio"
    | "select"
    | "text"
    | "time-sale-list"
    | "textarea";
  required?: boolean;
  readOnly?: boolean;
  options?: SystemSettingOption[];
  optionSource?: "countries" | "currency" | "timezones";
  master?: "businesstype";
  placeholder?: string;
  helper?: SystemSettingText;
  multiline?: boolean;
  valueType?: "boolean" | "number" | "string";
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

const roleUserLabel: SystemSettingText = {
  th: "ระดับผู้ใช้งาน",
  en: "User Level",
  cn: "用户级别",
  ja: "ユーザーレベル",
  ko: "사용자 레벨",
  lo: "ລະດັບຜູ້ໃຊ້ງານ",
  my: "အသုံးပြုသူအဆင့်",
  km: "កម្រិតអ្នកប្រើ",
  vi: "Cấp người dùng",
  ms: "User Level",
  id: "User Level",
  fil: "User Level",
};

const roleAdminLabel: SystemSettingText = {
  th: "ระดับแอดมิน",
  en: "Admin Level",
  cn: "管理员级别",
  ja: "管理者レベル",
  ko: "관리자 레벨",
  lo: "ລະດັບແອດມິນ",
  my: "အက်မင်အဆင့်",
  km: "កម្រិតអ្នកគ្រប់គ្រង",
  vi: "Cấp quản trị",
  ms: "Admin Level",
  id: "Admin Level",
  fil: "Admin Level",
};

const roleOwnerLabel: SystemSettingText = {
  th: "ระดับเจ้าของร้าน",
  en: "Owner Level",
  cn: "店主级别",
  ja: "オーナーレベル",
  ko: "점주 레벨",
  lo: "ລະດັບເຈົ້າຂອງຮ້ານ",
  my: "ပိုင်ရှင်အဆင့်",
  km: "កម្រិតម្ចាស់ហាង",
  vi: "Cấp chủ cửa hàng",
  ms: "Owner Level",
  id: "Owner Level",
  fil: "Owner Level",
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
  ...productMasterConfigs(),
  {
    slug: "business_type_screen",
    route: "/business_type_screen",
    manual: "business_type_screen",
    kind: "main-crud",
    icon: "tag",
    basePath: "/organization/business-type",
    listPath: "/organization/business-type/list",
    idField: "guid_fixed",
    title: { th: "ประเภทธุรกิจ", en: "Business Type" },
    subtitle: {
      th: "จัดการรหัสประเภทธุรกิจและชื่อตามภาษาที่เลือก",
      en: "Manage business type codes and names for the selected language.",
    },
    fields: [
      textField("code", "รหัส", "Code", true),
      namesField("names", "ชื่อประเภทธุรกิจ", "Business type names"),
      checkboxField("isdefault", "ค่าเริ่มต้น", "Default"),
    ],
  },
  {
    slug: "active_languages",
    route: "/active_languages",
    manual: "active_languages",
    kind: "company",
    icon: "building",
    idField: "guid_fixed",
    title: { th: "ภาษาที่ใช้งาน", en: "Active Languages" },
    subtitle: {
      th: "จัดการภาษาที่ใช้งาน ลำดับแรกคือภาษาแรกของบริษัท",
      en: "Manage active languages. The first one is the company's primary language.",
    },
    fields: [
      languageConfigsField(
        "settings.languageconfigs",
        "ภาษาที่ใช้งาน",
        "Active languages",
      ),
    ],
  },
  {
    slug: "company",
    route: "/company",
    manual: "company",
    kind: "company",
    icon: "building",
    idField: "guid_fixed",
    title: { th: "ข้อมูลบริษัท", en: "Company Profile" },
    subtitle: {
      th: "แก้ไขข้อมูลบริษัทปัจจุบันที่สร้างไว้แล้ว",
      en: "Edit the selected company profile that already exists.",
    },
    fields: [
      namesField("names", "ชื่อบริษัท", "Company names"),
      namesField("address", "ที่อยู่", "Address", true),
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
    idField: "guid_fixed",
    title: { th: "สาขา", en: "Branch" },
    subtitle: {
      th: "จัดการสาขา สกุลเงิน ภาษา timezone และคุณสมบัติธุรกิจ",
      en: "Manage branches, base currency, language, timezone, and business flags.",
    },
    fields: [
      languageListField("languages", "ภาษาที่ใช้งาน", "Active languages"),
      {
        ...textField("code", "รหัสสาขาภาษี", "Tax branch code", true),
        helper: {
          th: "ใช้รหัส 5 หลักตามภาษีไทย: สำนักงานใหญ่ = 00000, สาขาที่ 1 = 00001",
          en: "Use the Thai 5-digit tax branch code: head office = 00000, branch 1 = 00001.",
        },
        placeholder: "00000",
      },
      namesField(
        "companynames",
        "ชื่อบริษัทบนเอกสาร",
        "Company names on documents",
      ),
      namesField("names", "ชื่อสาขา", "Branch names"),
      namesField("contact.address", "ที่อยู่สาขา", "Branch address", true),
      comboField(
        "contact.country_code",
        "ประเทศ",
        "Country",
        countryOptions(),
        false,
        "countries",
      ),
      textField("contact.province_code", "จังหวัด", "Province"),
      textField("contact.district_code", "อำเภอ/เขต", "District"),
      textField("contact.sub_district_code", "ตำบล/แขวง", "Subdistrict"),
      textField("contact.zip_code", "รหัสไปรษณีย์", "Zip code"),
      numberField("contact.latitude", "ละติจูด", "Latitude"),
      numberField("contact.longitude", "ลองจิจูด", "Longitude"),
      textField("contact.phone_number", "เบอร์โทรสาขา", "Branch phone"),
      masterPickerField(
        "businesstype",
        "รหัสประเภทธุรกิจ ~ ชื่อประเภทธุรกิจ",
        "Business type",
        "businesstype",
      ),
      textField(
        "company_registration_no",
        "เลขทะเบียนบริษัท",
        "Company registration no.",
      ),
      comboField(
        "base_currency",
        "สกุลเงินหลัก",
        "Base currency",
        currencyOptions(),
        false,
        "currency",
      ),
      comboField("timezone", "Timezone", "Timezone", [], false, "timezones"),
      selectField(
        "date_format",
        "รูปแบบวันที่",
        "Date format",
        dateFormatOptions(),
      ),
      selectField("year_type", "ประเภทปี", "Year type", [
        {
          value: "buddhist",
          label: buddhistEraLabel.th,
          labels: buddhistEraLabel,
        },
        {
          value: "christian",
          label: christianEraLabel.th,
          labels: christianEraLabel,
        },
      ]),
      numberField("decimal_quantity", "ทศนิยมจำนวน", "Quantity decimals"),
      numberField("decimal_price", "ทศนิยมราคา", "Price decimals"),
      numberField("decimal_document", "ทศนิยมเอกสาร", "Document decimals"),
      radioField(
        "is_vat_registered",
        "สถานะ VAT",
        "VAT status",
        [
          {
            value: "false",
            label: vatNotRegisteredLabel.th,
            labels: vatNotRegisteredLabel,
          },
          {
            value: "true",
            label: vatRegisteredLabel.th,
            labels: vatRegisteredLabel,
          },
        ],
        false,
        "boolean",
      ),
      textField("pos.tax_id", "เลขที่ผู้เสียภาษี", "Tax ID"),
      numberField("pos.vatrate", "อัตราภาษี (%)", "VAT rate (%)"),
      checkboxField(
        "pos.isbom",
        "ตัดสต็อกตามสูตรผลิต (BOM)",
        "Cut stock by BOM",
      ),
      radioField(
        "pos.vattypepurchase",
        "ประเภทภาษีซื้อ",
        "Purchase VAT type",
        vatTypeOptions(),
        false,
        "number",
      ),
      radioField(
        "pos.inquirytypepurchase",
        "ประเภทรายการซื้อ",
        "Purchase inquiry type",
        inquiryTypeOptions(),
        false,
        "number",
      ),
      radioField(
        "pos.vattypesale",
        "ประเภทภาษีขาย",
        "Sale VAT type",
        vatTypeOptions(),
        false,
        "number",
      ),
      radioField(
        "pos.inquirytypesale",
        "ประเภทรายการขาย",
        "Sale inquiry type",
        inquiryTypeOptions(),
        false,
        "number",
      ),
      textareaField(
        "pos.headerreceiptpos",
        "ข้อความหัวใบเสร็จ",
        "Receipt header",
      ),
      textareaField(
        "pos.footerreceiptpos",
        "ข้อความท้ายใบเสร็จ",
        "Receipt footer",
      ),
      radioField(
        "couponusetype",
        "ประเภทการใช้งานคูปอง",
        "Coupon usage type",
        couponUseTypeOptions(),
        false,
        "number",
      ),
      jsonField("paymentrounding", "การปัดเศษ", "Payment rounding"),
      jsonField("pointconfig", "ตั้งค่าแต้ม", "Point config"),
      numberField("machinetype", "ประเภทเครื่อง", "Machine type"),
      imageGalleryField("imageuris", "รูปภาพ", "Images"),
      imageUploadField("logouri", "โลโก้ร้าน", "Logo"),
      checkboxField("is_restaurant", "ร้านอาหาร", "Restaurant"),
      checkboxField("is_tire", "ยางรถ", "Tire"),
      checkboxField("is_agriculture", "เกษตร", "Agriculture"),
      checkboxField("is_pharmacy", "ร้านยา", "Pharmacy"),
      checkboxField("is_retail", "ค้าปลีก", "Retail"),
      checkboxField("is_service", "บริการ", "Service"),
      checkboxField("is_wholesale", "ค้าส่ง", "Wholesale"),
      checkboxField("is_manufacturing", "ผลิต", "Manufacturing"),
      checkboxField("is_import_export", "นำเข้า/ส่งออก", "Import/Export"),
      checkboxField("is_contractor", "ผู้รับเหมา", "Contractor"),
      checkboxField("is_rental", "ให้เช่า", "Rental"),
      checkboxField("is_ecommerce", "อีคอมเมิร์ซ", "E-commerce"),
      checkboxField("is_logistics", "ขนส่ง", "Logistics"),
      checkboxField("is_education", "การศึกษา", "Education"),
      checkboxField("is_hotel", "โรงแรม", "Hotel"),
      checkboxField("is_beauty", "ความงาม", "Beauty"),
      checkboxField("is_gold_shop", "ร้านทอง", "Gold shop"),
      checkboxField("is_accounting_firm", "สำนักงานบัญชี", "Accounting firm"),
      checkboxField("is_construction", "วัสดุก่อสร้าง", "Construction"),
      checkboxField("is_electronics", "อิเล็กทรอนิกส์", "Electronics"),
      checkboxField("is_mobile_shop", "ร้านมือถือ", "Mobile shop"),
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
    idField: "guid_fixed",
    title: { th: "แผนก", en: "Department" },
    subtitle: {
      th: "จัดการรหัสแผนกและชื่อตามภาษาที่เลือก",
      en: "Manage department codes and names for the selected language.",
    },
    fields: [
      textField("code", "รหัสแผนก", "Department code", true),
      namesField("names", "ชื่อแผนก", "Department names"),
    ],
  },
  {
    slug: "work_day_screen",
    route: "/work_day_screen",
    manual: "work_day_screen",
    kind: "restaurant-setting",
    icon: "calendar",
    code: "workDay",
    idField: "guid_fixed",
    title: { th: "วันทำงาน", en: "Work Day" },
    subtitle: {
      th: "ตั้งค่าวันทำงานและช่วงเวลาทำงานตามระบบเดิม",
      en: "Configure weekly work days and working time ranges.",
    },
    fields: [jsonField("body", "ข้อมูลวันทำงาน", "Work day JSON")],
  },
  {
    slug: "holiday_screen",
    route: "/holiday_screen",
    manual: "holiday_screen",
    kind: "restaurant-setting",
    icon: "calendar-check",
    code: "Holiday",
    idField: "guid_fixed",
    title: { th: "วันหยุด", en: "Holiday" },
    subtitle: {
      th: "จัดการวันหยุดและคำอธิบายตามภาษาที่เลือก",
      en: "Manage holidays and descriptions for the selected language.",
    },
    fields: [
      dateField("date", "วันที่", "Date", true),
      namesField("desc", "คำอธิบาย", "Description"),
    ],
  },
  {
    slug: "employee",
    route: "/employee",
    manual: "employee",
    kind: "main-crud",
    icon: "user-round",
    basePath: "/shop/employee",
    listPath: "/shop/employee/list",
    idField: "guid_fixed",
    title: { th: "พนักงาน", en: "Employee" },
    subtitle: {
      th: "จัดการพนักงาน อีเมล สถานะ POS และ PIN",
      en: "Manage employees, email, POS access, and PIN.",
    },
    fields: [
      textField("code", "รหัสพนักงาน", "Employee code", true),
      textField("name", "ชื่อพนักงาน", "Employee name", true),
      textField("email", "อีเมล", "Email"),
      textField("pincode", "PIN", "PIN"),
      checkboxField("isenabled", "เปิดใช้งาน", "Enabled"),
      checkboxField("isusepos", "ใช้งาน POS", "Use POS"),
      companyMultiSelectField("company_guids", "สิทธิ์การเข้าถึงบริษัท", "Active companies"),
      branchMultiSelectField("branches", "สิทธิ์การเข้าถึงสาขา", "Active branches"),
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
    subtitle: {
      th: "จัดการผู้ใช้งาน บทบาท แผนก LINE และสิทธิ์อนุมัติ",
      en: "Manage users, role, department, LINE profile, and approval permissions.",
    },
    fields: [
      { key: "uid", label: userIdGuidLabel, type: "text", readOnly: true },
      textField(
        "username",
        "รหัสผู้ใช้ หรือ email",
        "User code or email",
        true,
      ),
      textField("user_profile_name", "ชื่อผู้ใช้งาน", "User name"),
      {
        ...textField("email", registeredEmailLabel.th, registeredEmailLabel.en),
        label: registeredEmailLabel,
      },
      radioField(
        "role",
        "สิทธิ์ผู้ใช้งาน",
        "User Role",
        [
          { value: "0", label: roleUserLabel.en, labels: roleUserLabel },
          { value: "2", label: roleOwnerLabel.en, labels: roleOwnerLabel },
          { value: "1", label: roleAdminLabel.en, labels: roleAdminLabel },
        ],
        true,
        "number",
      ),
      radioField(
        "is_access_disabled",
        accessStatusLabel.th,
        accessStatusLabel.en,
        [
          {
            value: "false",
            label: accessEnabledLabel.en,
            labels: accessEnabledLabel,
          },
          {
            value: "true",
            label: accessTemporarilyDisabledLabel.en,
            labels: accessTemporarilyDisabledLabel,
          },
        ],
        true,
        "boolean",
      ),
      textField("position", "ตำแหน่ง", "Position"),
      textField("department", "แผนก", "Department"),
      companyMultiSelectField("company_guids", "สิทธิ์การเข้าถึงบริษัท", "Active companies"),
      branchMultiSelectField("branches", "สิทธิ์การเข้าถึงสาขา", "Active branches"),
      {
        ...textField("line_user_id", "LINE User ID", "LINE User ID"),
        readOnly: true,
      },
      {
        ...textField("line_display_name", "ชื่อ LINE", "LINE display name"),
        readOnly: true,
      },
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
    idField: "guid_fixed",
    title: { th: "ออกแบบฟอร์ม", en: "Form Design" },
    subtitle: {
      th: "จัดการ template ฟอร์มเอกสารและข้อมูลแบบ JSON",
      en: "Manage document form templates and template JSON data.",
    },
    fields: [
      textField("code", "รหัสฟอร์ม", "Form code", true),
      namesField("names", "ชื่อฟอร์ม", "Form names"),
      textField("doc_type", "ประเภทเอกสาร", "Document type"),
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
    idField: "guid_fixed",
    title: { th: "LINE Notify", en: "LINE Notify" },
    subtitle: {
      th: "จัดการ token และ event แจ้งเตือนของสาขา",
      en: "Manage notification tokens and branch events.",
    },
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
    subtitle: {
      th: "กำหนดสิทธิ์และวงเงินอนุมัติแยกตามบริษัท",
      en: "Configure approval roles and limits by company.",
    },
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
    title: { th: "กำหนดสิทธิ์หน้าจอ", en: "Permission Definition" },
    subtitle: {
      th: "สร้างรหัสสิทธิ์และกำหนดสิทธิ์แยกตามสาขา/หน้าจอ",
      en: "Create permission codes and define screen permissions per branch.",
    },
    fields: [
      textField("permissionCode", "รหัสสิทธิ์", "Permission code", true),
      textField("permissionName", "ชื่อสิทธิ์", "Permission name", true),
      textareaField("description", "คำอธิบาย", "Description"),
      checkboxField("isActive", "เปิดใช้งาน", "Active"),
      jsonField("branches", "สิทธิ์ตามสาขา", "Branch permissions JSON"),
    ],
  },
  {
    slug: "permission_group",
    route: "/permission_group",
    manual: "permission_group",
    kind: "atlas",
    icon: "users",
    collection: "permission_groups",
    idField: "groupCode",
    deleteKey: "groupCode",
    title: { th: "กำหนดสิทธิ์ตามกลุ่ม", en: "Permission Group" },
    subtitle: {
      th: "สร้างกลุ่มสิทธิ์การใช้งานหน้าจอสำหรับตำแหน่งงาน",
      en: "Create permission groups for job roles.",
    },
    fields: [
      textField("groupCode", "รหัสกลุ่มสิทธิ์", "Group code", true),
      textField("groupName", "ชื่อกลุ่มสิทธิ์", "Group name", true),
      textareaField("description", "คำอธิบาย", "Description"),
      checkboxField("isActive", "เปิดใช้งาน", "Active"),
      jsonField("permissionCodes", "สิทธิ์", "Permissions"),
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
    title: { th: "กำหนดสิทธิ์ผู้ใช้งาน", en: "User Permission" },
    subtitle: {
      th: "กำหนดสิทธิ์การเข้าถึงบริษัทและสาขาของผู้ใช้งาน",
      en: "Configure company and branch access permissions for users.",
    },
    fields: [
      textField(
        "employeeCode",
        "รหัสผู้ใช้/พนักงาน",
        "User/employee code",
        true,
      ),
      textField("employeeName", "ชื่อ", "Name"),
      textField("groupCode", "กลุ่มสิทธิ์", "Permission group", false),
      companyMultiSelectField("company_guids", "สิทธิ์การเข้าถึงบริษัท", "Active companies"),
      branchMultiSelectField("branches", "สิทธิ์การเข้าถึงสาขา", "Active branches"),
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
    subtitle: {
      th: "จัดการ API key สำหรับ MCP และ export config",
      en: "Manage MCP API keys and export client configuration.",
    },
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
    subtitle: {
      th: "จัดการ provider, model, API key และลำดับใช้งานของ AI",
      en: "Manage AI providers, models, API keys, and priority.",
    },
    fields: [
      selectField(
        "provider_name",
        "Provider",
        "Provider",
        [
          { value: "ollama", label: "Ollama" },
          { value: "groq", label: "Groq" },
          { value: "openrouter", label: "OpenRouter" },
          { value: "deepseek", label: "DeepSeek" },
          { value: "gemini", label: "Google Gemini" },
          { value: "custom", label: "Custom" },
        ],
        true,
      ),
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
    subtitle: {
      th: "เลือก shop ต้นทาง ดู preview แล้วสั่งโอน MongoDB ไปยัง shop ปัจจุบัน",
      en: "Select a source shop, preview counts, then copy MongoDB data into the current shop.",
    },
    fields: [],
  },
];

function productMasterConfigs(): SystemSettingConfig[] {
  const codeNameFields = (
    codeLabelTh = "รหัส",
    codeLabelEn = "Code",
    nameLabelTh = "ชื่อ",
    nameLabelEn = "Name",
  ) => [
    textField("code", codeLabelTh, codeLabelEn, true),
    namesField("names", nameLabelTh, nameLabelEn),
  ];

  const codeNameConfig = (
    slug: string,
    route: string,
    icon: string,
    basePath: string,
    listPath: string,
    titleTh: string,
    titleEn: string,
    codeLabelTh = "รหัส",
    codeLabelEn = "Code",
    nameLabelTh = "ชื่อ",
    nameLabelEn = "Name",
  ): SystemSettingConfig => ({
    slug,
    route,
    manual: slug,
    kind: "main-crud",
    icon,
    basePath,
    listPath,
    idField: "guid_fixed",
    title: { th: titleTh, en: titleEn },
    subtitle: {
      th: `จัดการ${titleTh}ตามรูปแบบหน้าจอ Flutter เดิม`,
      en: `Manage ${titleEn.toLowerCase()} using the legacy Flutter workflow.`,
    },
    fields: codeNameFields(codeLabelTh, codeLabelEn, nameLabelTh, nameLabelEn),
  });

  const aicloudConfig = (
    slug: string,
    route: string,
    icon: string,
    path: string,
    titleTh: string,
    titleEn: string,
  ) =>
    codeNameConfig(
      slug,
      route,
      icon,
      `/aicloud/${path}`,
      `/aicloud/${path}/list`,
      titleTh,
      titleEn,
    );

  return [
    {
      slug: "productunit",
      route: "/productunit",
      manual: "productunit",
      kind: "main-crud",
      icon: "ruler",
      basePath: "/unit",
      listPath: "/unit/list",
      idField: "guid_fixed",
      title: { th: "หน่วยนับสินค้า", en: "Product Unit" },
      subtitle: {
        th: "จัดการรหัสหน่วยนับและชื่อหน่วยนับสินค้า",
        en: "Manage product unit codes and names.",
      },
      fields: [
        textField("unitcode", "รหัสหน่วยนับ", "Unit code", true),
        namesField("names", "ชื่อหน่วยนับ", "Unit names"),
        companyMultiSelectField("company_guids", "บริษัทที่ใช้งาน", "Active companies"),
      ],
    },
    codeNameConfig(
      "productgroup",
      "/productgroup",
      "group",
      "/product/group",
      "/product/group/list",
      "กลุ่มสินค้า",
      "Product Group",
      "รหัสกลุ่ม",
      "Group code",
      "ชื่อกลุ่มสินค้า",
      "Product group names",
    ),
    {
      slug: "product_category_group_select_screen",
      route: "/product_category_group_select_screen",
      manual: "product_category_group_select_screen",
      kind: "main-crud",
      icon: "folder",
      basePath: "/product/category",
      listPath: "/product/category/list",
      idField: "guid_fixed",
      title: { th: "โครงสร้างหมวดสินค้า", en: "Product Category Structure" },
      subtitle: {
        th: "จัดการโครงสร้างหมวดสินค้า กลุ่มหมวด และสินค้าในหมวด",
        en: "Manage product category structure, category groups, and products in categories.",
      },
      fields: [
        { ...numberField("group_number", "ลำดับกลุ่ม", "Group number"), readOnly: true },
        { ...textField("parent_guid", "หมวดแม่", "Parent category GUID"), readOnly: true },
        namesField("names", "ชื่อหมวดสินค้า", "Product category names"),
        checkboxField("isdisabled", "ปิดใช้งาน", "Disabled"),
        timeSaleListField("timeforsales", "เวลาการขาย", "Time for sale"),
        radioField(
          "useimageorcolor",
          "การแสดงผลหมวด",
          "Category display mode",
          [
            {
              value: "false",
              label: "Use image",
              labels: { th: "ใช้รูปภาพ", en: "Use image" },
            },
            {
              value: "true",
              label: "Use color",
              labels: { th: "ใช้สี", en: "Use color" },
            },
          ],
          false,
          "boolean",
        ),
        textField("colorselecthex", "สี (Hex)", "Color (Hex)"),
        imageUploadField("imageuri", "รูปภาพ", "Image"),
        imageUploadField("coveruri", "รูปหน้าปก", "Cover image"),
      ],
    },
    {
      slug: "productcategorylist",
      route: "/productcategorylist",
      manual: "productcategorylist",
      kind: "main-crud",
      icon: "list",
      basePath: "/product/category",
      listPath: "/product/category/list",
      idField: "guid_fixed",
      title: { th: "สินค้าในหมวด", en: "Products in Category" },
      subtitle: {
        th: "จัดการรายการสินค้าที่ผูกกับหมวดสินค้า",
        en: "Manage product items linked to product categories.",
      },
      fields: [
        namesField("names", "ชื่อหมวดสินค้า", "Product category names"),
        jsonField("codelist", "รายการสินค้า", "Product list JSON"),
      ],
    },
    {
      slug: "product_warehouse_screen",
      route: "/product_warehouse_screen",
      manual: "product_warehouse_screen",
      kind: "main-crud",
      icon: "warehouse",
      basePath: "/warehouse",
      idField: "guid_fixed",
      title: { th: "คลังสินค้า → โซนเก็บสินค้า → ชั้นวาง", en: "Warehouse → Storage Zone → Shelf" },
      subtitle: {
        th: "จัดการคลังสินค้า โซนเก็บสินค้า และชั้นวางสินค้า",
        en: "Manage warehouses, storage zones, and shelves.",
      },
      fields: [
        textField("code", "รหัสคลังสินค้า", "Warehouse code", true),
        namesField("names", "ชื่อคลังสินค้า", "Warehouse names"),
        {
          key: "location",
          label: { th: "โซนเก็บสินค้าและชั้นวาง", en: "Storage Zones & Shelves" },
          type: "json",
        },
      ],
    },
    codeNameConfig(
      "product_type_screen",
      "/product_type_screen",
      "tag",
      "/product/type",
      "/product/type/list",
      "ประเภทสินค้า",
      "Product Type",
      "รหัสประเภท",
      "Type code",
      "ชื่อประเภทสินค้า",
      "Product type names",
    ),
    {
      slug: "product_dimension",
      route: "/product_dimension",
      manual: "product_dimension",
      kind: "main-crud",
      icon: "design",
      basePath: "/dimension",
      listPath: "/dimension/list",
      idField: "guid_fixed",
      title: { th: "มิติสินค้า", en: "Product Dimension" },
      subtitle: {
        th: "จัดการมิติสินค้าและรายการย่อย",
        en: "Manage product dimensions and dimension items.",
      },
      fields: [
        namesField("names", "ชื่อมิติสินค้า", "Product dimension names"),
        checkboxField("isdisabled", "ปิดใช้งาน", "Disabled"),
        jsonField("items", "รายการมิติย่อย", "Dimension items JSON"),
      ],
    },
    {
      slug: "product_bom",
      route: "/product_bom",
      manual: "product_bom",
      kind: "main-crud",
      icon: "network",
      basePath: "/product/bom",
      listPath: "/product/bom/list",
      idField: "guid_fixed",
      editable: true,
      title: { th: "สูตรผลิต", en: "Product BOM" },
      subtitle: {
        th: "จัดการรหัสสูตรผลิต วัตถุดิบ และสูตรย่อยจากฐานข้อมูลจริง",
        en: "Manage recipe codes, ingredients, and sub-recipes from the real database.",
      },
      fields: [
        textField("barcode", "รหัสสูตรผลิต", "Recipe code"),
        {
          ...namesField("names", "ชื่อสูตรผลิต", "Recipe names"),
        },
        textField("item_unit_code", "หน่วยสูตร", "Recipe unit"),
        {
          ...namesField("itemunitnames", "ชื่อหน่วยนับ", "Unit names"),
        },
        {
          ...jsonField("bom", "รายการสูตรผลิต", "BOM items JSON"),
        },
      ],
    },
    codeNameConfig(
      "promotion_screen",
      "/promotion_screen",
      "gift",
      "/product/promotion",
      "/product/promotion/list",
      "โปรโมชั่น",
      "Promotion",
      "รหัสโปรโมชั่น",
      "Promotion code",
      "ชื่อโปรโมชั่น",
      "Promotion names",
    ),
    aicloudConfig(
      "master_brand_screen",
      "/master_brand_screen",
      "settings",
      "brand",
      "ยี่ห้อ",
      "Brand",
    ),
    aicloudConfig(
      "master_category_screen",
      "/master_category_screen",
      "category",
      "category",
      "หมวดจำแนกสินค้า",
      "Product Attribute Category",
    ),
    aicloudConfig(
      "master_class_screen",
      "/master_class_screen",
      "category",
      "class",
      "Class",
      "Class",
    ),
    aicloudConfig(
      "master_design_screen",
      "/master_design_screen",
      "design",
      "design",
      "Design",
      "Design",
    ),
    aicloudConfig(
      "master_grade_screen",
      "/master_grade_screen",
      "badge",
      "grade",
      "Grade",
      "Grade",
    ),
    aicloudConfig(
      "master_model_screen",
      "/master_model_screen",
      "activity",
      "model",
      "Model",
      "Model",
    ),
    aicloudConfig(
      "master_pattern_screen",
      "/master_pattern_screen",
      "grid",
      "pattern",
      "Pattern",
      "Pattern",
    ),
    aicloudConfig(
      "master_group_screen",
      "/master_group_screen",
      "group",
      "group",
      "กลุ่มหลัก",
      "Main Group",
    ),
    aicloudConfig(
      "master_group_sub1_screen",
      "/master_group_sub1_screen",
      "group",
      "groupsubone",
      "กลุ่มย่อย 1",
      "Sub Group 1",
    ),
    aicloudConfig(
      "master_group_sub2_screen",
      "/master_group_sub2_screen",
      "group",
      "groupsubtwo",
      "กลุ่มย่อย 2",
      "Sub Group 2",
    ),
  ];
}

export const SYSTEM_SETTING_SLUGS = SYSTEM_SETTING_CONFIGS.map(
  (item) => item.slug,
);

export function getSystemSettingConfig(
  routeOrSlug: string,
): SystemSettingConfig | undefined {
  const normalized = routeOrSlug.startsWith("/")
    ? routeOrSlug
    : `/${routeOrSlug}`;
  return SYSTEM_SETTING_CONFIGS.find(
    (item) => item.route === normalized || item.slug === routeOrSlug,
  );
}

export function systemSettingLabel(
  config: SystemSettingConfig,
  language: LanguageCode,
): string {
  return config.title[language] ?? config.title.en ?? config.title.th;
}

function textField(
  key: string,
  th: string,
  en: string,
  required = false,
): SystemSettingField {
  return { key, label: { th, en }, type: "text", required };
}

function textareaField(
  key: string,
  th: string,
  en: string,
): SystemSettingField {
  return { key, label: { th, en }, type: "textarea" };
}

function numberField(key: string, th: string, en: string): SystemSettingField {
  return { key, label: { th, en }, type: "number" };
}

function dateField(
  key: string,
  th: string,
  en: string,
  required = false,
): SystemSettingField {
  return { key, label: { th, en }, type: "date", required };
}

function imageUploadField(
  key: string,
  th: string,
  en: string,
): SystemSettingField {
  return { key, label: { th, en }, type: "image-upload" };
}

function imageGalleryField(
  key: string,
  th: string,
  en: string,
): SystemSettingField {
  return { key, label: { th, en }, type: "image-gallery" };
}

function timeSaleListField(
  key: string,
  th: string,
  en: string,
): SystemSettingField {
  return { key, label: { th, en }, type: "time-sale-list" };
}

function branchMultiSelectField(
  key: string,
  th: string,
  en: string,
): SystemSettingField {
  return { key, label: { th, en }, type: "branch-multi-select" };
}

function companyMultiSelectField(
  key: string,
  th: string,
  en: string,
): SystemSettingField {
  return { key, label: { th, en }, type: "company-multi-select" };
}

function checkboxField(
  key: string,
  th: string,
  en: string,
): SystemSettingField {
  return { key, label: { th, en }, type: "checkbox" };
}

function namesField(
  key: string,
  th: string,
  en: string,
  multiline = false,
): SystemSettingField {
  return { key, label: { th, en }, type: "names", required: true, multiline };
}

function jsonField(key: string, th: string, en: string): SystemSettingField {
  return { key, label: { th, en }, type: "json" };
}

function languageListField(
  key: string,
  th: string,
  en: string,
): SystemSettingField {
  return { key, label: { th, en }, type: "language-list" };
}

function masterPickerField(
  key: string,
  th: string,
  en: string,
  master: "businesstype",
): SystemSettingField {
  return { key, label: { th, en }, type: "master-picker", master };
}

function comboField(
  key: string,
  th: string,
  en: string,
  options: SystemSettingOption[] = [],
  required = false,
  optionSource?: "countries" | "currency" | "timezones",
): SystemSettingField {
  return {
    key,
    label: { th, en },
    type: "combo",
    options,
    required,
    optionSource,
  };
}

function radioField(
  key: string,
  th: string,
  en: string,
  options: SystemSettingOption[],
  required = false,
  valueType: "boolean" | "number" | "string" = "string",
): SystemSettingField {
  return {
    key,
    label: { th, en },
    type: "radio",
    options,
    required,
    valueType,
  };
}

function selectField(
  key: string,
  th: string,
  en: string,
  options: SystemSettingOption[],
  required = false,
  valueType: "number" | "string" = "string",
): SystemSettingField {
  return {
    key,
    label: { th, en },
    type: "select",
    options,
    required,
    valueType,
  };
}

function languageConfigsField(
  key: string,
  th: string,
  en: string,
): SystemSettingField {
  return { key, label: { th, en }, type: "language-configs" };
}

function vatTypeOptions(): SystemSettingOption[] {
  return [
    {
      value: "0",
      label: "ราคาไม่รวมภาษี",
      labels: { th: "ราคาไม่รวมภาษี", en: "VAT excluded" },
    },
    {
      value: "1",
      label: "ราคารวมภาษี",
      labels: { th: "ราคารวมภาษี", en: "VAT included" },
    },
    {
      value: "2",
      label: "ภาษีอัตราศูนย์",
      labels: { th: "ภาษีอัตราศูนย์", en: "Zero-rated VAT" },
    },
    {
      value: "3",
      label: "ไม่กระทบภาษี",
      labels: { th: "ไม่กระทบภาษี", en: "No VAT" },
    },
  ];
}

function inquiryTypeOptions(): SystemSettingOption[] {
  return [
    { value: "0", label: "เครดิต", labels: { th: "เครดิต", en: "Credit" } },
    { value: "1", label: "เงินสด", labels: { th: "เงินสด", en: "Cash" } },
  ];
}

function couponUseTypeOptions(): SystemSettingOption[] {
  return [
    {
      value: "0",
      label: "ใช้ได้หลายใบ",
      labels: { th: "ใช้ได้หลายใบ", en: "Multiple coupons per bill" },
    },
    {
      value: "1",
      label: "ใช้ได้ใบเดียว",
      labels: { th: "ใช้ได้ใบเดียว", en: "Single coupon per bill" },
    },
  ];
}

function currencyOptions() {
  return currencySymbolPresets.map((item) => ({
    value: item.code,
    label: `${item.code} - ${item.name}`,
  }));
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
    "dd/MM/yyyy",
    "dd-MM-yyyy",
    "yyyy-MM-dd",
    "MM/dd/yyyy",
    "dd MMM yy",
    "dd MMM yyyy",
    "dd MMMM yy",
    "dd MMMM yyyy",
    "MMM dd, yy",
    "MMM dd, yyyy",
    "MMMM dd, yy",
    "MMMM dd, yyyy",
    "yyyy MMM dd",
    "yyyy MMMM dd",
  ].map((pattern) => ({
    value: pattern,
    label: `${pattern} (${formatDatePatternExample(pattern, "th")})`,
    labels: Object.fromEntries(
      LANGUAGES.map((item) => [
        item.code,
        `${pattern} (${formatDatePatternExample(pattern, item.code)})`,
      ]),
    ) as SystemSettingText,
  }));
}

function formatDatePatternExample(
  pattern: string,
  language: LanguageCode,
): string {
  const date = new Date(Date.UTC(2026, 4, 22, 12, 0, 0));
  const parts = datePartsForLanguage(date, language);
  return pattern
    .replace(/yyyy/g, parts.year4)
    .replace(/yy/g, parts.year2)
    .replace(/MMMM/g, parts.monthLong)
    .replace(/MMM/g, parts.monthShort)
    .replace(/MM/g, parts.month2)
    .replace(/dd/g, parts.day2);
}

function datePartsForLanguage(date: Date, language: LanguageCode) {
  const locale = localeOf(language);
  const year4 = datePart(locale, date, { year: "numeric" }, "year");
  return {
    day2: datePart(locale, date, { day: "2-digit" }, "day"),
    month2: datePart(locale, date, { month: "2-digit" }, "month"),
    monthShort: datePart(locale, date, { month: "short" }, "month"),
    monthLong: datePart(locale, date, { month: "long" }, "month"),
    year2: datePart(locale, date, { year: "2-digit" }, "year"),
    year4,
  };
}

function datePart(
  locale: string,
  date: Date,
  options: Intl.DateTimeFormatOptions,
  partType: Intl.DateTimeFormatPartTypes,
): string {
  const parts = new Intl.DateTimeFormat(locale, {
    ...options,
    timeZone: "UTC",
  }).formatToParts(date);
  return parts.find((part) => part.type === partType)?.value ?? "";
}

function localeOf(language: LanguageCode): string {
  const map: Record<LanguageCode, string> = {
    th: "th-TH",
    en: "en-US",
    cn: "zh-CN",
    ja: "ja-JP",
    ko: "ko-KR",
    lo: "lo-LA",
    my: "my-MM",
    km: "km-KH",
    vi: "vi-VN",
    ms: "ms-MY",
    id: "id-ID",
    fil: "fil-PH",
  };
  return map[language];
}
