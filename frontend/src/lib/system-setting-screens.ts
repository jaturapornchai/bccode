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
    | "holding-scope-rules"
    | "language-configs"
    | "language-list"
    | "master-picker"
    | "names"
    | "number"
    | "radio"
    | "select"
    | "string-list"
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
  /** Override the file input `accept` attribute. Defaults to common image types. Use "image/png" for logos. */
  acceptTypes?: string;
  /**
   * Marks a user-facing business/reference code field (e.g. permissioncode, groupcode,
   * approvalcode). On save it is normalized to uppercase + no whitespace, gets a duplicate
   * guard, and is shown as the record's code in the detail header. See business-code rules.
   */
  businessCode?: boolean;
  /**
   * Marks an identity code field that must be unique per holding but is NOT a business code
   * (e.g. permissionlink `employeecode` = email/username). It is NOT uppercased (emails are
   * exempt) — only trimmed — but still gets a case-insensitive duplicate guard.
   */
  uniqueCode?: boolean;
  /**
   * When set, the image upload keeps the original file untouched in `key` and also
   * generates a small WebP thumbnail stored under this field key (e.g. "avatarthumb").
   * Lists/previews read the thumbnail; detail/download read the full original.
   */
  thumbnailKey?: string;
};

export type SystemSettingKind =
  | "ai-provider"
  | "atlas"
  | "company"
  | "copy-uat"
  | "goapi-crud"
  | "main-crud"
  | "report"
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
    slug: "businesstypescreen",
    route: "/businesstypescreen",
    manual: "businesstypescreen",
    kind: "main-crud",
    icon: "tag",
    basePath: "/organization/business-type",
    listPath: "/organization/business-type/list",
    idField: "guidfixed",
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
    slug: "activelanguages",
    route: "/activelanguages",
    manual: "activelanguages",
    kind: "company",
    icon: "building",
    idField: "guidfixed",
    title: { th: "ภาษาที่ใช้งาน", en: "Active Languages" },
    subtitle: {
      th: "กำหนดก่อนข้อมูลอื่น ลำดับแรกคือภาษาแรกของบริษัทและช่องชื่อหลายภาษา",
      en: "Set this before other data. The first row is the primary language for multilingual names.",
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
    idField: "guidfixed",
    title: { th: "ข้อมูลบริษัท", en: "Company Profile" },
    subtitle: {
      th: "แก้ไขข้อมูลบริษัทปัจจุบันที่สร้างไว้แล้ว",
      en: "Edit the selected company profile that already exists.",
    },
    fields: [
      imageUploadField("logouri", "โลโก้บริษัท", "Company logo", "image/png"),
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
    idField: "guidfixed",
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
      imageUploadField("logouri", "โลโก้สาขา", "Branch logo", "image/png"),
      namesField(
        "companynames",
        "ชื่อบริษัทบนเอกสาร",
        "Company names on documents",
      ),
      namesField("names", "ชื่อสาขา", "Branch names"),
      namesField("contact.address", "ที่อยู่สาขา", "Branch address", true),
      comboField(
        "contact.countrycode",
        "ประเทศ",
        "Country",
        countryOptions(),
        false,
        "countries",
      ),
      textField("contact.provincecode", "จังหวัด", "Province"),
      textField("contact.districtcode", "อำเภอ/เขต", "District"),
      textField("contact.subdistrictcode", "ตำบล/แขวง", "Subdistrict"),
      textField("contact.zipcode", "รหัสไปรษณีย์", "Zip code"),
      numberField("contact.latitude", "ละติจูด", "Latitude"),
      numberField("contact.longitude", "ลองจิจูด", "Longitude"),
      textField("contact.phonenumber", "เบอร์โทรสาขา", "Branch phone"),
      masterPickerField(
        "businesstype",
        "รหัสประเภทธุรกิจ ~ ชื่อประเภทธุรกิจ",
        "Business type",
        "businesstype",
      ),
      textField(
        "companyregistrationno",
        "เลขทะเบียนบริษัท",
        "Company registration no.",
      ),
      comboField(
        "basecurrency",
        "สกุลเงินหลัก",
        "Base currency",
        currencyOptions(),
        false,
        "currency",
      ),
      comboField("timezone", "Timezone", "Timezone", [], false, "timezones"),
      selectField(
        "dateformat",
        "รูปแบบวันที่",
        "Date format",
        dateFormatOptions(),
      ),
      selectField("yeartype", "ประเภทปี", "Year type", [
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
      numberField("decimalquantity", "ทศนิยมจำนวน", "Quantity decimals"),
      numberField("decimalprice", "ทศนิยมราคา", "Price decimals"),
      numberField("decimaldocument", "ทศนิยมเอกสาร", "Document decimals"),
      radioField(
        "isvatregistered",
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
      textField("pos.taxid", "เลขที่ผู้เสียภาษี", "Tax ID"),
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
      checkboxField("isrestaurant", "ร้านอาหาร", "Restaurant"),
      checkboxField("istire", "ยางรถ", "Tire"),
      checkboxField("isagriculture", "เกษตร", "Agriculture"),
      checkboxField("ispharmacy", "ร้านยา", "Pharmacy"),
      checkboxField("isretail", "ค้าปลีก", "Retail"),
      checkboxField("isservice", "บริการ", "Service"),
      checkboxField("iswholesale", "ค้าส่ง", "Wholesale"),
      checkboxField("ismanufacturing", "ผลิต", "Manufacturing"),
      checkboxField("isimportexport", "นำเข้า/ส่งออก", "Import/Export"),
      checkboxField("iscontractor", "ผู้รับเหมา", "Contractor"),
      checkboxField("isrental", "ให้เช่า", "Rental"),
      checkboxField("isecommerce", "อีคอมเมิร์ซ", "E-commerce"),
      checkboxField("islogistics", "ขนส่ง", "Logistics"),
      checkboxField("iseducation", "การศึกษา", "Education"),
      checkboxField("ishotel", "โรงแรม", "Hotel"),
      checkboxField("isbeauty", "ความงาม", "Beauty"),
      checkboxField("isgoldshop", "ร้านทอง", "Gold shop"),
      checkboxField("isaccountingfirm", "สำนักงานบัญชี", "Accounting firm"),
      checkboxField("isconstruction", "วัสดุก่อสร้าง", "Construction"),
      checkboxField("iselectronics", "อิเล็กทรอนิกส์", "Electronics"),
      checkboxField("ismobileshop", "ร้านมือถือ", "Mobile shop"),
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
    slug: "workdayscreen",
    route: "/workdayscreen",
    manual: "workdayscreen",
    kind: "restaurant-setting",
    icon: "calendar",
    code: "workDay",
    idField: "guidfixed",
    title: { th: "วันทำงาน", en: "Work Day" },
    subtitle: {
      th: "ตั้งค่าวันทำงานและช่วงเวลาทำงานตามระบบเดิม",
      en: "Configure weekly work days and working time ranges.",
    },
    fields: [jsonField("body", "ข้อมูลวันทำงาน", "Work day JSON")],
  },
  {
    slug: "holidayscreen",
    route: "/holidayscreen",
    manual: "holidayscreen",
    kind: "restaurant-setting",
    icon: "calendar-check",
    code: "Holiday",
    idField: "guidfixed",
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
    basePath: "/holding/employee",
    listPath: "/holding/employee/list",
    idField: "guidfixed",
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
      holdingScopeRulesField("accessscopes", "บริษัท/สาขาที่เข้าใช้งานได้", "Company and branch access"),
    ],
  },
  {
    slug: "user",
    route: "/user",
    manual: "user",
    kind: "main-crud",
    icon: "users",
    basePath: "/holding/permission",
    listPath: "/holding/users",
    idField: "username",
    title: { th: "ผู้ใช้งาน", en: "User" },
    subtitle: {
      th: "จัดการผู้ใช้งาน บทบาท แผนก LINE และสิทธิ์อนุมัติ",
      en: "Manage users, role, department, LINE profile, and approval permissions.",
    },
    fields: [
      {
        ...imageUploadField("avatar", "รูปผู้ใช้งาน", "User avatar", undefined, "avatarthumb"),
        helper: {
          th: "รองรับไฟล์ PNG พื้นหลังโปร่งใสได้ แสดงในระบบและเอกสาร",
          en: "PNG supported. Transparent background OK. Shown across the app and documents.",
        },
      },
      {
        ...textField("username", "ชื่อเข้าสู่ระบบ", "Login username", true),
        placeholder: "เช่น somchai หรือ somchai@email.com",
        helper: {
          th: "ใช้ภาษาอังกฤษหรือตัวเลข ใช้สำหรับเข้าสู่ระบบ ตั้งครั้งเดียวแล้วไม่ควรเปลี่ยนบ่อย",
          en: "Letters or numbers. Used to sign in. Set once and avoid changing.",
        },
      },
      {
        ...textField("userprofilename", "ชื่อ-นามสกุล", "Full name"),
        placeholder: "เช่น สมชาย ใจดี",
        helper: {
          th: "ชื่อที่แสดงในระบบและเอกสาร ใช้ภาษาไทยได้",
          en: "Display name shown across the app and documents. Thai is allowed.",
        },
      },
      {
        ...textField("email", registeredEmailLabel.th, registeredEmailLabel.en),
        label: registeredEmailLabel,
        type: "text",
        placeholder: "เช่น somchai@email.com",
        helper: {
          th: "อีเมลสำหรับรับการแจ้งเตือนและกู้คืนรหัสผ่าน ถ้าเหมือนชื่อเข้าสู่ระบบก็กรอกซ้ำได้",
          en: "Email for notifications and password recovery. May match the login username.",
        },
      },
      {
        ...radioField(
          "role",
          "ระดับสิทธิ์",
          "Access level",
          [
            { value: "0", label: roleUserLabel.en, labels: roleUserLabel },
            { value: "2", label: roleOwnerLabel.en, labels: roleOwnerLabel },
            { value: "1", label: roleAdminLabel.en, labels: roleAdminLabel },
          ],
          true,
          "number",
        ),
        helper: {
          th: "ผู้ใช้งาน = ใช้งานได้ตามที่กำหนดให้ · แอดมิน = จัดการผู้ใช้และตั้งค่าทั้งหมด · เจ้าของร้าน = สิทธิสูงสุด เป็นผู้สร้างกลุ่มกิจการ",
          en: "User = access as granted · Admin = manage users and all settings · Owner = highest rights, the business group creator.",
        },
      },
      {
        ...radioField(
          "isaccessdisabled",
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
        helper: {
          th: "พักการเข้าใช้งานชั่วคราวได้โดยไม่ต้องลบผู้ใช้ เปลี่ยนกลับได้ตลอด",
          en: "Temporarily suspend access without deleting the user. Can be turned back on anytime.",
        },
      },
      {
        ...dateField("accessexpirydate", "วันหมดอายุการเข้าใช้งาน", "Access expiry date"),
        helper: {
          th: "ตั้งวันที่ปิดการเข้าใช้งานอัตโนมัติ (เช่น วันสุดท้ายของพนักงาน) เว้นว่างถ้าไม่มีกำหนด",
          en: "Auto-disable access on this date (e.g. employee's last working day). Leave empty for no expiry.",
        },
      },
      textField("position", "ตำแหน่ง", "Position"),
      textField("department", "แผนก", "Department"),
      {
        ...holdingScopeRulesField("accessscopes", "บริษัทและสาขาที่เข้าถึงได้", "Companies and branches this user can enter", true),
        helper: {
          th: "เลือกบริษัทและสาขาที่ผู้ใช้คนนี้เข้าใช้งานได้ เลือกทั้งกลุ่มกิจการได้ถ้าต้องการเข้าถึงทุกบริษัท",
          en: "Pick the companies and branches this user may enter. Choose the whole business group to allow all.",
        },
      },
      {
        ...textField("lineuserid", "LINE User ID", "LINE User ID"),
        readOnly: true,
        helper: {
          th: "ผูกกับบัญชี LINE อัตโนมัติเมื่อเชื่อมจากหน้าเข้าสู่ระบบด้วย LINE",
          en: "Auto-filled from LINE after linking via the LINE sign-in page.",
        },
      },
      {
        ...textField("linedisplayname", "ชื่อ LINE", "LINE display name"),
        readOnly: true,
        helper: {
          th: "ชื่อบน LINE ของผู้ใช้คนนี้ แสดงหลังเชื่อมบัญชี LINE เรียบร้อย",
          en: "This user's LINE display name, shown after LINE is linked.",
        },
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
    idField: "guidfixed",
    title: { th: "ออกแบบฟอร์ม", en: "Form Design" },
    subtitle: {
      th: "จัดการ template ฟอร์มเอกสารและข้อมูลแบบ JSON",
      en: "Manage document form templates and template JSON data.",
    },
    fields: [
      textField("code", "รหัสฟอร์ม", "Form code", true),
      namesField("names", "ชื่อฟอร์ม", "Form names"),
      textField("doctype", "ประเภทเอกสาร", "Document type"),
      checkboxField("isdefault", "ค่าเริ่มต้น", "Default"),
      jsonField("templatedata", "Template Data", "Template data"),
    ],
  },
  {
    slug: "linenotify",
    route: "/linenotify",
    manual: "linenotify",
    kind: "main-crud",
    icon: "bell",
    basePath: "/notify",
    listPath: "/notify/list",
    idField: "guidfixed",
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
    slug: "approvalsetting",
    route: "/approvalsetting",
    manual: "approvalsetting",
    kind: "atlas",
    icon: "shield",
    collection: "approvalsettings",
    idField: "guidfixed",
    deleteKey: "guidfixed",
    title: { th: "สิทธิ์การอนุมัติ", en: "Approval Permission" },
    subtitle: {
      th: "กำหนดสิทธิ์และวงเงินอนุมัติภายใต้กลุ่มกิจการ และกำหนดขอบเขตบริษัท/สาขา",
      en: "Configure business group approval roles, limits, and company/branch scope.",
    },
    fields: [
      businessCodeField("approvalcode", "รหัส", "Code", true),
      textField("approvalname", "ชื่อ", "Name", true),
      textareaField("description", "คำอธิบาย", "Description"),
      checkboxField("isactive", "เปิดใช้งาน", "Active"),
      holdingScopeRulesField("approvalrules", "ขอบเขตที่ใช้สิทธิ์อนุมัติ", "Approval scope"),
      jsonField("approvals", "สิทธิ์การอนุมัติ", "Approval permission JSON"),
    ],
  },
  {
    slug: "useraccessaudit",
    route: "/useraccessaudit",
    manual: "useraccessaudit",
    kind: "report",
    icon: "shield",
    title: { th: "ตรวจสอบสถานะผู้ใช้งาน", en: "User Access Audit" },
    subtitle: {
      th: "ตรวจสอบว่าผู้ใช้งานเข้าอะไรได้บ้าง ทำอะไรได้บ้าง และส่งออกเป็นรายงาน PDF",
      en: "Review what each user can access, what they can do, and export the result as a PDF report.",
    },
    editable: false,
    fields: [],
  },
  {
    slug: "permissiondefinition",
    route: "/permissiondefinition",
    manual: "permissiondefinition",
    kind: "atlas",
    icon: "shield",
    collection: "permissiondefinitions",
    idField: "guidfixed",
    deleteKey: "guidfixed",
    title: { th: "กำหนดสิทธิ์หน้าจอ", en: "Permission Definition" },
    subtitle: {
      th: "สร้างรหัสสิทธิ์หน้าจอภายใต้กลุ่มกิจการ และกำหนดขอบเขตบริษัท/สาขา",
      en: "Create business group screen permission codes and company/branch scope.",
    },
    fields: [
      businessCodeField("permissioncode", "รหัสสิทธิ์", "Permission code", true),
      textField("permissionname", "ชื่อสิทธิ์", "Permission name", true),
      textareaField("description", "คำอธิบาย", "Description"),
      checkboxField("isactive", "เปิดใช้งาน", "Active"),
      holdingScopeRulesField("scoperules", "ขอบเขตที่ใช้สิทธิ์นี้", "Permission scope"),
      jsonField("accessrules", "สิทธิ์หน้าจอ/action ตามสาขา", "Branch screen/action permissions"),
    ],
  },
  {
    slug: "permissiongroup",
    route: "/permissiongroup",
    manual: "permissiongroup",
    kind: "atlas",
    icon: "users",
    collection: "permissiongroups",
    idField: "guidfixed",
    deleteKey: "guidfixed",
    title: { th: "กำหนดสิทธิ์ตามกลุ่ม", en: "Permission Group" },
    subtitle: {
      th: "สร้างกลุ่มสิทธิ์การใช้งานหน้าจอสำหรับตำแหน่งงาน",
      en: "Create permission groups for job roles.",
    },
    fields: [
      businessCodeField("groupcode", "รหัสกลุ่มสิทธิ์", "Group code", true),
      textField("groupname", "ชื่อกลุ่มสิทธิ์", "Group name", true),
      textareaField("description", "คำอธิบาย", "Description"),
      checkboxField("isactive", "เปิดใช้งาน", "Active"),
      holdingScopeRulesField("scoperules", "ขอบเขตที่ใช้กลุ่มสิทธิ์นี้", "Permission group scope"),
      jsonField("permissioncodes", "สิทธิ์", "Permissions"),
    ],
  },
  {
    slug: "permissionlink",
    route: "/permissionlink",
    manual: "permissionlink",
    kind: "atlas",
    icon: "link",
    collection: "employeepermissions",
    idField: "guidfixed",
    deleteKey: "guidfixed",
    title: { th: "กำหนดสิทธิ์ผู้ใช้งาน", en: "User Permission" },
    subtitle: {
      th: "ผูกผู้ใช้กับกลุ่มสิทธิ์/สิทธิ์หน้าจอภายใต้กลุ่มกิจการ และกำหนดขอบเขตบริษัท/สาขา",
      en: "Link users to business group permissions and company/branch scope.",
    },
    fields: [
      uniqueCodeField(
        "employeecode",
        "รหัสผู้ใช้/พนักงาน",
        "User/employee code",
        true,
      ),
      textField("employeename", "ชื่อ", "Name"),
      textField("groupcode", "กลุ่มสิทธิ์", "Permission group", false),
      holdingScopeRulesField("scoperules", "ขอบเขตที่ผู้ใช้นี้ใช้สิทธิ์ได้", "User permission scope"),
      jsonField("permissioncodes", "สิทธิ์", "Permissions"),
      jsonField("approvalcodes", "สิทธิ์การอนุมัติ", "Approval permissions"),
    ],
  },
  {
    slug: "aiprovider",
    route: "/aiprovider",
    manual: "aiprovider",
    kind: "ai-provider",
    icon: "bot",
    idField: "providername",
    title: { th: "AI Provider", en: "AI Provider" },
    subtitle: {
      th: "จัดการ provider, model, API key และลำดับใช้งานของ AI",
      en: "Manage AI providers, models, API keys, and priority.",
    },
    fields: [
      selectField(
        "providername",
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
      textField("apikey", "API Key", "API Key"),
      textField("baseurl", "Base URL", "Base URL"),
      numberField("priority", "ลำดับ", "Priority"),
      checkboxField("isactive", "เปิดใช้งาน", "Active"),
    ],
  },
  {
    slug: "copyuattodev",
    route: "/copyuattodev",
    manual: "copyuattodev",
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
    idField: "guidfixed",
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

  const atlasMasterConfig = (
    slug: string,
    route: string,
    icon: string,
    collection: string,
    titleTh: string,
    titleEn: string,
    subtitleTh: string,
    subtitleEn: string,
    fields: SystemSettingField[],
  ): SystemSettingConfig => ({
    slug,
    route,
    manual: slug,
    kind: "atlas",
    icon,
    collection,
    idField: "guidfixed",
    title: { th: titleTh, en: titleEn },
    subtitle: { th: subtitleTh, en: subtitleEn },
    fields,
  });

  const sizeSystemOptions: SystemSettingOption[] = [
    { value: "intl", label: "International", labels: { th: "สากล", en: "International" } },
    { value: "th", label: "Thai", labels: { th: "ไทย", en: "Thai" } },
    { value: "us", label: "US", labels: { th: "US", en: "US" } },
    { value: "uk", label: "UK", labels: { th: "UK", en: "UK" } },
    { value: "eu", label: "EU", labels: { th: "EU", en: "EU" } },
    { value: "jp", label: "JP", labels: { th: "JP", en: "JP" } },
  ];

  const sizeTypeOptions: SystemSettingOption[] = [
    { value: "regular", label: "Regular", labels: { th: "ปกติ", en: "Regular" } },
    { value: "petite", label: "Petite", labels: { th: "ตัวเล็ก", en: "Petite" } },
    { value: "plus", label: "Plus", labels: { th: "พลัสไซซ์", en: "Plus size" } },
    { value: "tall", label: "Tall", labels: { th: "ตัวสูง", en: "Tall" } },
    { value: "kids", label: "Kids", labels: { th: "เด็ก", en: "Kids" } },
    { value: "freesize", label: "Free size", labels: { th: "ฟรีไซซ์", en: "Free size" } },
  ];

  const variantMatrixTypeOptions: SystemSettingOption[] = [
    { value: "general", label: "General", labels: { th: "สินค้าทั่วไป", en: "General" } },
    { value: "apparel", label: "Apparel", labels: { th: "เสื้อผ้า/แฟชั่น", en: "Apparel" } },
    { value: "mobilephone", label: "Mobile phone", labels: { th: "มือถือ/โทรศัพท์", en: "Mobile phone" } },
    { value: "sim", label: "SIM", labels: { th: "ซิมการ์ด", en: "SIM card" } },
    { value: "computer", label: "Computer", labels: { th: "คอมพิวเตอร์", en: "Computer" } },
    { value: "electronics", label: "Electronics", labels: { th: "อิเล็กทรอนิกส์", en: "Electronics" } },
  ];

  const serialTrackingModeOptions: SystemSettingOption[] = [
    { value: "none", label: "None", labels: { th: "ไม่คุมเลขเครื่อง", en: "None" } },
    { value: "serialno", label: "Serial No.", labels: { th: "Serial No.", en: "Serial No." } },
    { value: "imei", label: "IMEI", labels: { th: "IMEI", en: "IMEI" } },
    { value: "iccid", label: "ICCID", labels: { th: "ICCID ซิม", en: "SIM ICCID" } },
    { value: "macaddress", label: "MAC address", labels: { th: "MAC address", en: "MAC address" } },
    { value: "multiple", label: "Multiple", labels: { th: "หลายเลขต่อชิ้น", en: "Multiple identifiers" } },
  ];

  const serialRegistryStatusOptions: SystemSettingOption[] = [
    { value: "instock", label: "In stock", labels: { th: "อยู่ในสต๊อก", en: "In stock" } },
    { value: "reserved", label: "Reserved", labels: { th: "จองแล้ว", en: "Reserved" } },
    { value: "sold", label: "Sold", labels: { th: "ขายแล้ว", en: "Sold" } },
    { value: "repair", label: "Repair", labels: { th: "ส่งซ่อม", en: "Repair" } },
    { value: "claim", label: "Claim", labels: { th: "เคลม", en: "Claim" } },
    { value: "blocked", label: "Blocked", labels: { th: "ห้ามขาย", en: "Blocked" } },
  ];

  const channelPriceStatusOptions: SystemSettingOption[] = [
    { value: "active", label: "Active", labels: { th: "ใช้งาน", en: "Active" } },
    { value: "scheduled", label: "Scheduled", labels: { th: "ตั้งเวลาล่วงหน้า", en: "Scheduled" } },
    { value: "ended", label: "Ended", labels: { th: "สิ้นสุดแล้ว", en: "Ended" } },
  ];

  return [
    {
      slug: "productunit",
      route: "/productunit",
      manual: "productunit",
      kind: "main-crud",
      icon: "ruler",
      basePath: "/unit",
      listPath: "/unit/list",
      idField: "guidfixed",
      title: { th: "หน่วยนับสินค้า", en: "Product Unit" },
      subtitle: {
        th: "จัดการรหัสหน่วยนับและชื่อหน่วยนับสินค้า",
        en: "Manage product unit codes and names.",
      },
      fields: [
        textField("unitcode", "รหัสหน่วยนับ", "Unit code", true),
        namesField("names", "ชื่อหน่วยนับ", "Unit names"),
        companyMultiSelectField("businesscodes", "สิทธิ์การเข้าถึงบริษัท", "Active companies"),
      ],
    },
    {
      slug: "productgroup",
      route: "/productgroup",
      manual: "productgroup",
      kind: "main-crud",
      icon: "group",
      basePath: "/product/group",
      listPath: "/product/group/list",
      idField: "guidfixed",
      title: { th: "กลุ่มสินค้า", en: "Product Group" },
      subtitle: {
        th: "จัดกลุ่มสินค้าแบบลำดับชั้น กลุ่มหลักและกลุ่มย่อย ลากวางจัดลำดับได้",
        en: "Organize product groups into main groups and subgroups. Drag to reorder.",
      },
      fields: [
        { ...textField("parentguid", "กลุ่มแม่", "Parent group"), readOnly: true },
        namesField("names", "ชื่อกลุ่มสินค้า", "Product group names"),
        checkboxField("isdisabled", "ปิดใช้งาน", "Disabled"),
      ],
    },
    {
      slug: "productcategorygroupselectscreen",
      route: "/productcategorygroupselectscreen",
      manual: "productcategorygroupselectscreen",
      kind: "main-crud",
      icon: "folder",
      basePath: "/product/category",
      listPath: "/product/category/list",
      idField: "guidfixed",
      title: { th: "จัดหมวดสินค้า", en: "Product Categories" },
      subtitle: {
        th: "จัดการหมวดสินค้าและหมวดย่อย แบบลำดับชั้น",
        en: "Manage product categories and subcategories in a hierarchy.",
      },
      fields: [
        { ...numberField("groupnumber", "ลำดับกลุ่ม", "Group number"), readOnly: true },
        { ...textField("parentguid", "หมวดแม่", "Parent category"), readOnly: true },
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
      idField: "guidfixed",
      title: { th: "สินค้าในหมวด", en: "Products in Category" },
      subtitle: {
        th: "เลือกสินค้าที่จะแสดงในแต่ละหมวด",
        en: "Choose products to show in each category.",
      },
      fields: [
        namesField("names", "ชื่อหมวดสินค้า", "Product category names"),
        {
          ...stringListField("codelist", "รายการสินค้า", "Product list"),
          placeholder: "รหัสสินค้า1, รหัสสินค้า2",
          helper: {
            th: "พิมพ์รหัสสินค้าที่จะแสดงในหมวดนี้ คั่นด้วยลูกน้ำ",
            en: "Type the product codes to show in this category, separated by commas.",
          },
        },
      ],
    },
    {
      slug: "productwarehousescreen",
      route: "/productwarehousescreen",
      manual: "productwarehousescreen",
      kind: "main-crud",
      icon: "warehouse",
      basePath: "/warehouse",
      idField: "guidfixed",
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
    atlasMasterConfig(
      "productcolor",
      "/productcolor",
      "color",
      "productcolors",
      "สีสินค้า",
      "Product Color",
      "กำหนดรหัสสี ชื่อสี และชื่อเรียกอื่นสำหรับใช้สร้าง SKU",
      "Define color codes, names, and aliases for SKU generation.",
      [
        businessCodeField("code", "รหัสสี", "Color code", true),
        namesField("names", "ชื่อสี", "Color names"),
        textField("hexcolor", "ค่าสี (HEX)", "Color HEX"),
        textField("colorfamily", "กลุ่มสี", "Color family"),
        {
          ...stringListField("aliases", "ชื่อเรียกอื่น", "Aliases"),
          placeholder: "ดำ, black, สีดำ",
          helper: {
            th: "เพิ่มชื่อที่ระบบควรจับคู่เป็นสีเดียวกัน เช่น ดำ, black, สีดำ",
            en: "Add names that should map to the same color, such as black, dark, BK.",
          },
        },
        companyMultiSelectField("businesscodes", "สิทธิ์การเข้าถึงบริษัท", "Active companies"),
        checkboxField("isdisabled", "ปิดใช้งาน", "Disabled"),
      ],
    ),
    atlasMasterConfig(
      "productsize",
      "/productsize",
      "ruler",
      "productsizes",
      "ไซซ์สินค้า",
      "Product Size",
      "กำหนดรหัสไซซ์ ระบบไซซ์ และชื่อเรียกอื่น",
      "Define size codes, size systems, and aliases.",
      [
        businessCodeField("code", "รหัสไซซ์", "Size code", true),
        namesField("names", "ชื่อไซซ์", "Size names"),
        selectField("sizesystem", "ระบบไซซ์", "Size system", sizeSystemOptions),
        selectField("sizetype", "ประเภทไซซ์", "Size type", sizeTypeOptions),
        numberField("sortorder", "ลำดับ", "Sort order"),
        {
          ...stringListField("aliases", "ชื่อเรียกอื่น", "Aliases"),
          placeholder: "XL, XL(60-70kg), 36-37",
          helper: {
            th: "เพิ่มชื่อไซซ์ที่ระบบควรจับคู่เป็นไซซ์เดียวกัน เช่น XL, XL(60-70kg), 36-37",
            en: "Add imported size labels that should map to the same size.",
          },
        },
        companyMultiSelectField("businesscodes", "สิทธิ์การเข้าถึงบริษัท", "Active companies"),
        checkboxField("isdisabled", "ปิดใช้งาน", "Disabled"),
      ],
    ),
    atlasMasterConfig(
      "productvariantmatrix",
      "/productvariantmatrix",
      "grid",
      "productvariantmatrices",
      "ชุดตัวเลือกสินค้า",
      "Product Option Sets",
      "กำหนดชุดสี ไซซ์ หรือคุณสมบัติที่ใช้สร้างตัวเลือกสินค้าและรหัสขาย",
      "Define color, size, or attribute option sets used to create sellable product choices.",
      [
        businessCodeField("code", "รหัสชุดตัวเลือก", "Option set code", true),
        namesField("names", "ชื่อชุดตัวเลือก", "Option set names"),
        textareaField("description", "คำอธิบายเบื้องต้น", "Basic description"),
        selectField("matrixtype", "ประเภทธุรกิจสินค้า", "Product business type", variantMatrixTypeOptions),
        selectField("serialtrackingmode", "การคุมเลขเครื่อง", "Serial tracking mode", serialTrackingModeOptions),
        {
          ...jsonField("optiontiers", "แกนตัวเลือก", "Option tiers"),
          placeholder: `[{"tierno":1,"optioncode":"COLOR","name":"สี/Color"},{"tierno":2,"optioncode":"SIZE","name":"ขนาด/Size"}]`,
          helper: {
            th: "รองรับหลายแกน เช่น สี, ไซซ์, ความจุ, รุ่น, เครือข่าย, วัสดุ, รสชาติ, แพ็ก",
            en: "Supports multiple tiers such as color, size, storage, model, network, material, flavor, and pack.",
          },
        },
        {
          ...jsonField("skucombinations", "รายการ SKU/ตัวเลือก", "SKU combinations"),
          placeholder: `[{"sellersku":"SKU-001","barcode":"885000000001","gtin":"885000000001","optionvalues":["BLACK","128GB"],"saleprice":0,"cost":0,"openingstock":0,"packageweight":0}]`,
          helper: {
            th: "ใช้เก็บ SKU, barcode/GTIN, ราคา, ต้นทุน, สต๊อกเริ่มต้น, น้ำหนัก/ขนาด และเลขเครื่องต่อ SKU",
            en: "Stores SKU, barcode/GTIN, price, cost, opening stock, package size/weight, and serial policy per SKU.",
          },
        },
        {
          ...jsonField("mediaassets", "รูปภาพและวิดีโอสินค้า", "Product media"),
          placeholder: `[{"kind":"main","uri":"images/products/example-main.webp","sortorder":1},{"kind":"sku","optioncode":"COLOR","optionvalue":"BLACK","uri":"images/products/example-black.webp"}]`,
          helper: {
            th: "รองรับรูปหลัก รูปเพิ่มเติม รูปตามตัวเลือก วิดีโอ ตารางไซซ์ และรูปในรายละเอียดสินค้า",
            en: "Supports main images, gallery images, SKU option images, video, size chart, and detail-page media.",
          },
        },
        {
          ...jsonField("specificationgroups", "ข้อมูลจำเพาะสินค้า", "Product specifications"),
          placeholder: `[{"groupcode":"GENERAL","groupname":"ข้อมูลทั่วไป","attributes":[{"attributecode":"MATERIAL","attributename":"วัสดุ","inputtype":"multiselect","scope":"product","values":[{"valuecode":"COTTON","valuetext":"Cotton"}]}]}]`,
          helper: {
            th: "ใช้เก็บ attribute/specification จากหลายช่องทางขายแล้วแปลงเป็นสเปกกลางของระบบ",
            en: "Stores attributes/specifications from multiple channels as the system's neutral product spec structure.",
          },
        },
        {
          ...jsonField("importattributemaps", "จับคู่ชื่อจากไฟล์นำเข้า", "Import name matching"),
          placeholder: `[{"sourcename":"Color","targetoptioncode":"COLOR"},{"sourcename":"Storage","targetoptioncode":"STORAGE"}]`,
          helper: {
            th: "ใช้แมพชื่อ attribute จากข้อมูลนำเข้าให้เข้ากับแกนตัวเลือกของระบบ โดยไม่ต้องแสดงแหล่งที่มา",
            en: "Maps imported attribute names into system option tiers without exposing the source channel.",
          },
        },
        {
          ...jsonField("integrationprofiles", "การเชื่อมต่อช่องทางขาย", "Sales channel connections"),
          placeholder: `[{"channel":"external","skufields":["sellersku","barcode","price","stock"]}]`,
          helper: {
            th: "เก็บรายละเอียดสำหรับ import/sync ภายนอกไว้ให้ระบบใช้ ไม่ใช้เป็นข้อความแสดงที่มาของสินค้า",
            en: "Stores external import/sync details for the system without showing them as product-source labels.",
          },
        },
        {
          ...jsonField("payloadexamples", "ตัวอย่างข้อมูลนำเข้า/ส่งออก", "Import/export examples"),
          placeholder: `[{"direction":"import","usecase":"productdetail","payload":{"title":"Example","images":[],"attributes":[],"skus":[]}}]`,
          helper: {
            th: "เก็บตัวอย่างข้อมูลสำหรับทีมพัฒนา/ตัวนำเข้า ไม่ใช้เป็นข้อความแสดงในหน้าขายปกติ",
            en: "Stores sample data for developers/importers, not as normal product display text.",
          },
        },
        companyMultiSelectField("businesscodes", "สิทธิ์การเข้าถึงบริษัท", "Active companies"),
        checkboxField("isdisabled", "ปิดใช้งาน", "Disabled"),
      ],
    ),
    atlasMasterConfig(
      "productserialregistry",
      "/productserialregistry",
      "qr",
      "productserialregistries",
      "ทะเบียนเลขเครื่อง",
      "Serial Registry",
      "จัดการ Serial No., IMEI, ICCID, MAC address และสถานะเครื่องต่อสินค้า",
      "Manage Serial No., IMEI, ICCID, MAC address, and item status per product.",
      [
        businessCodeField("serialno", "เลขเครื่อง", "Serial / identifier", true),
        selectField("identifiertype", "ประเภทเลขเครื่อง", "Identifier type", serialTrackingModeOptions),
        selectField("status", "สถานะเลขเครื่อง", "Serial status", serialRegistryStatusOptions),
        textField("itemcode", "รหัสสินค้า", "Product code", true),
        textField("barcode", "บาร์โค้ด/SKU", "Barcode / SKU"),
        textField("skucode", "รหัส SKU", "SKU code"),
        textField("unitcode", "หน่วยนับ", "Unit code"),
        textField("warehousecode", "คลัง", "Warehouse"),
        textField("locationcode", "โซน/ชั้นวาง", "Location / shelf"),
        textField("suppliercode", "ผู้จำหน่าย", "Supplier"),
        textField("customercode", "ลูกค้า", "Customer"),
        textField("purchasedocno", "เอกสารซื้อ", "Purchase document"),
        textField("saledocno", "เอกสารขาย", "Sale document"),
        dateField("warrantystartdate", "เริ่มประกัน", "Warranty start"),
        dateField("warrantyenddate", "หมดประกัน", "Warranty end"),
        textareaField("note", "หมายเหตุ", "Note"),
        companyMultiSelectField("businesscodes", "สิทธิ์การเข้าถึงบริษัท", "Active companies"),
        checkboxField("isdisabled", "ปิดใช้งาน", "Disabled"),
      ],
    ),
    atlasMasterConfig(
      "channelprice",
      "/channelprice",
      "money",
      "productchannelprices",
      "ราคาตามช่องทางขาย",
      "Channel Prices",
      "กำหนดราคาขายตามช่องทาง เช่น หน้าร้าน Shopee Lazada TikTok และ Marketplace อื่น",
      "Set selling prices by channel such as POS, Shopee, Lazada, TikTok, and other marketplaces.",
      [
        businessCodeField("pricecode", "รหัสราคา", "Price code", true),
        namesField("names", "ชื่อราคา", "Price names"),
        textField("channelcode", "ช่องทางขาย", "Sales channel", true),
        textField("itemcode", "รหัสสินค้า", "Product code", true),
        textField("barcode", "บาร์โค้ด/SKU", "Barcode / SKU"),
        textField("dimensionkey", "ตัวเลือกสินค้า", "Product option key"),
        textField("pricelevel", "ระดับราคา", "Price level"),
        textField("currency", "สกุลเงิน", "Currency"),
        numberField("saleprice", "ราคาขาย", "Sale price"),
        numberField("compareatprice", "ราคาเต็ม/ก่อนลด", "Compare-at price"),
        dateField("startdate", "เริ่มใช้ราคา", "Start date"),
        dateField("enddate", "สิ้นสุดราคา", "End date"),
        selectField("status", "สถานะราคา", "Price status", channelPriceStatusOptions),
        checkboxField("syncprice", "ซิงค์ราคากับช่องทางขาย", "Sync channel price"),
        companyMultiSelectField("businesscodes", "สิทธิ์การเข้าถึงบริษัท", "Active companies"),
        checkboxField("isdisabled", "ปิดใช้งาน", "Disabled"),
      ],
    ),
    {
      slug: "productdimension",
      route: "/productdimension",
      manual: "productdimension",
      kind: "main-crud",
      icon: "design",
      basePath: "/dimension",
      listPath: "/dimension/list",
      idField: "guidfixed",
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
      slug: "productbom",
      route: "/productbom",
      manual: "productbom",
      kind: "main-crud",
      icon: "network",
      basePath: "/product/bom",
      listPath: "/product/bom/list",
      idField: "guidfixed",
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
        textField("itemunitcode", "หน่วยสูตร", "Recipe unit"),
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
      "/promotionscreen",
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
      "/masterbrandscreen",
      "settings",
      "brand",
      "ยี่ห้อ",
      "Brand",
    ),
    aicloudConfig(
      "master_category_screen",
      "/mastercategoryscreen",
      "category",
      "category",
      "หมวดคุณลักษณะสินค้า",
      "Product Attribute Category",
    ),
    aicloudConfig(
      "master_class_screen",
      "/masterclassscreen",
      "category",
      "class",
      "Class",
      "Class",
    ),
    aicloudConfig(
      "master_design_screen",
      "/masterdesignscreen",
      "design",
      "design",
      "Design",
      "Design",
    ),
    aicloudConfig(
      "master_grade_screen",
      "/mastergradescreen",
      "badge",
      "grade",
      "Grade",
      "Grade",
    ),
    aicloudConfig(
      "master_model_screen",
      "/mastermodelscreen",
      "activity",
      "model",
      "Model",
      "Model",
    ),
    aicloudConfig(
      "master_pattern_screen",
      "/masterpatternscreen",
      "grid",
      "pattern",
      "Pattern",
      "Pattern",
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

// businessCodeField — a text field that holds a user-facing business/reference code.
// Normalized to uppercase + no whitespace on save, duplicate-guarded, and shown as the
// record code in the detail header.
function businessCodeField(
  key: string,
  th: string,
  en: string,
  required = false,
): SystemSettingField {
  return { key, label: { th, en }, type: "text", required, businessCode: true };
}

// uniqueCodeField — an identity code that must be unique per holding but is NOT uppercased
// (e.g. employeecode = email/username). Trimmed + case-insensitive duplicate-guarded.
function uniqueCodeField(
  key: string,
  th: string,
  en: string,
  required = false,
): SystemSettingField {
  return { key, label: { th, en }, type: "text", required, uniqueCode: true };
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
  acceptTypes?: string,
  thumbnailKey?: string,
): SystemSettingField {
  const field: SystemSettingField = { key, label: { th, en }, type: "image-upload" };
  if (acceptTypes) field.acceptTypes = acceptTypes;
  if (thumbnailKey) field.thumbnailKey = thumbnailKey;
  return field;
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

function stringListField(key: string, th: string, en: string): SystemSettingField {
  return { key, label: { th, en }, type: "string-list" };
}

function holdingScopeRulesField(
  key: string,
  th: string,
  en: string,
  required = false,
): SystemSettingField {
  return { key, label: { th, en }, type: "holding-scope-rules", required };
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
