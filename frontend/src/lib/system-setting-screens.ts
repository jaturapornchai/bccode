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
    | "bank-accounts"
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
    | "master-multi-picker"
    | "names"
    | "number"
    | "radio"
    | "select"
    | "string-list"
    | "text"
    | "thai-address"
    | "time-sale-list"
    | "textarea";
  required?: boolean;
  readOnly?: boolean;
  options?: SystemSettingOption[];
  optionSource?: "countries" | "timezones";
  master?: "businesstype" | "creditorgroup" | "debtorgroup";
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
   * (e.g. an employee code that is an email/username). It is NOT uppercased (emails are
   * exempt) — only trimmed — but still gets a case-insensitive duplicate guard.
   */
  uniqueCode?: boolean;
  /**
   * When set, the image upload keeps the original file untouched in `key` and also
   * generates a small WebP thumbnail stored under this field key (e.g. "avatarthumb").
   * Lists/previews read the thumbnail; detail/download read the full original.
   */
  thumbnailKey?: string;
  /**
   * thai-address fields only: renders a "copy from billing address" button that copies the
   * 5 cascade subkeys plus the `.address` string list from this other thai-address field's
   * prefix into this one (live form state only, not a save).
   */
  copyFromPrefix?: string;
};

export type SystemSettingKind =
  | "company"
  | "goapi-crud"
  | "main-crud"
  | "permission-catalog"
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
      en: "Manage branches, language, timezone, and business flags.",
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
      thaiAddressField(
        "contact",
        "ที่อยู่สาขา (จังหวัด/อำเภอ/ตำบล/รหัสไปรษณีย์)",
        "Branch address (province/district/subdistrict/zipcode)",
      ),
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
      comboField("timezone", "เขตเวลา", "Timezone", [], false, "timezones"),
      selectField(
        "dateformat",
        "รูปแบบวันที่",
        "Date format",
        dateFormatOptions(),
      ),
      selectField("yeartype", "ปีศักราชที่ใช้", "Year type", [
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
      numberField("decimalquantity", "ทศนิยมของปริมาณ", "Quantity decimals"),
      numberField("decimalprice", "ทศนิยมของราคา/ต้นทุน", "Price decimals"),
      numberField("decimaldocument", "ทศนิยมของมูลค่า", "Document decimals"),
      radioField(
        "isvatregistered",
        "การจดทะเบียนภาษี",
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
      textField("pos.taxid", "เลขประจำตัวผู้เสียภาษี", "Tax ID"),
      numberField("pos.vatrate", "อัตราภาษีมูลค่าเพิ่ม", "VAT rate (%)"),
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
        "ประเภทการซื้อ",
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
        "ประเภทการขาย",
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
    title: { th: "ทะเบียนพนักงาน", en: "Employee" },
    subtitle: {
      th: "จัดการพนักงาน อีเมล สถานะ POS และ PIN",
      en: "Manage employees, email, POS access, and PIN.",
    },
    fields: [
      {
        ...imageUploadField("profilepicture", "รูปพนักงาน", "Employee photo", undefined, "profilepicturethumb"),
        helper: {
          th: "รองรับ JPG/PNG — ระบบย่อรูปก่อนอัปโหลด และเก็บไฟล์ในระบบจัดเก็บภาพ (S3) พร้อมรูปย่ออัตโนมัติ",
          en: "JPG/PNG supported — resized before upload, stored in S3 with an automatic thumbnail.",
        },
      },
      businessCodeField("code", "รหัสพนักงาน", "Employee code", true),
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
    idField: "useruid",
    title: { th: "ผู้ใช้งานระบบ", en: "User" },
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
        ...textField("username", "รหัสผู้ใช้ (usercode)", "User code", true),
        placeholder: "เช่น somchai01",
        helper: {
          th: "ใช้เข้าสู่ระบบด้วยรหัสผ่านได้หลัง User ตั้งรหัสผ่านผ่านลิงก์แบบใช้ครั้งเดียว",
          en: "Can be used for password login after the user sets a password through a one-time link.",
        },
      },
      {
		...textField("userprofilename", "ชื่อผู้ใช้ (username)", "User name"),
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
		  th: "ไม่บังคับ สามารถเชื่อมอีเมลภายหลังเพื่อรับการแจ้งเตือนและกู้คืนรหัสผ่านได้",
		  en: "Optional. An email can be linked later for notifications and password recovery.",
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
      jsonField("permissionsets", "สิทธิ์การใช้งานเพิ่มเติม", "Additional permission sets"),
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
    kind: "permission-catalog",
    icon: "shield",
    idField: "permissioncode",
    title: { th: "รายการสิทธิ์หน้าจอ", en: "Screen Permission Catalog" },
    subtitle: {
      th: "รายการหน้าจอที่ระบบรองรับ ใช้เลือกรายการสิทธิ์ให้แต่ละบทบาท",
      en: "Read-only screen catalog used to assign permissions to each role.",
    },
    editable: false,
    fields: [
      businessCodeField("permissioncode", "รหัสสิทธิ์", "Permission code", true),
      textField("permissionname", "ชื่อสิทธิ์", "Permission name", true),
      textField("description", "เส้นทางหน้าจอ", "Screen route"),
      textField("category", "หมวด", "Category"),
    ],
  },
  {
    slug: "permissiongroup",
    route: "/permissiongroup",
    manual: "permissiongroup",
    kind: "main-crud",
    icon: "users",
    basePath: "/organization/role-permission",
    listPath: "/organization/role-permission",
    idField: "_id",
    deleteKey: "_id",
    title: { th: "กลุ่มสิทธิ์การใช้งาน", en: "Permission sets" },
    subtitle: {
      th: "ตั้งชุดสิทธิ์การใช้งานสำเร็จรูป (เช่น บัญชี ขาย คลัง) แล้วให้คนในองค์กรเลือกได้หลายชุด · USER/ADMIN/OWNER คือชุดมาตรฐานตามระดับสิทธิ์ ข้ามขั้นนี้ได้ถ้าใช้แค่ชุดมาตรฐาน",
      en: "Define reusable permission sets (e.g. Accounting, Sales, Stock) that people can combine · USER/ADMIN/OWNER are the access-level defaults.",
    },
    fields: [
      {
        ...businessCodeField("rolecode", "รหัสชุดสิทธิ์", "Permission set code", true),
        placeholder: "เช่น ACCOUNTING, SALES, STOCK",
        helper: {
          th: "ตัวพิมพ์ใหญ่ A-Z 0-9 _ - ยาว 2-30 ตัว · USER / ADMIN / OWNER = ชุดมาตรฐานของระดับสิทธิ์ (ทุกคนในระดับนั้นได้อัตโนมัติ)",
          en: "Uppercase A-Z 0-9 _ - (2-30 chars) · USER / ADMIN / OWNER are the access-level defaults everyone at that level gets.",
        },
      },
      namesField("names", "ชื่อชุดสิทธิ์", "Permission set name"),
      checkboxField("isactive", "เปิดใช้งาน", "Active"),
      jsonField("permissions", "หน้าจอที่เข้าใช้งานได้", "Accessible screens"),
    ],
  },
  {
    slug: "creditorgroup",
    route: "/creditorgroup",
    manual: "creditorgroup",
    kind: "main-crud",
    icon: "users",
    basePath: "/debtaccount/creditor-group",
    listPath: "/debtaccount/creditor-group/list",
    idField: "guidfixed",
    title: { th: "กลุ่มเจ้าหนี้", en: "Vendor Group" },
    subtitle: {
      th: "จัดการรหัสกลุ่มเจ้าหนี้และชื่อตามภาษาที่เลือก",
      en: "Manage vendor group codes and names for the selected language.",
    },
    fields: [
      businessCodeField("groupcode", "รหัสกลุ่มเจ้าหนี้", "Vendor group code", true),
      namesField("names", "ชื่อกลุ่มเจ้าหนี้", "Vendor group names"),
    ],
  },
  {
    slug: "debtorgroup",
    route: "/debtorgroup",
    manual: "debtorgroup",
    kind: "main-crud",
    icon: "users",
    basePath: "/debtaccount/debtor-group",
    listPath: "/debtaccount/debtor-group/list",
    idField: "guidfixed",
    title: { th: "กลุ่มลูกหนี้", en: "Customer Group" },
    subtitle: {
      th: "จัดการรหัสกลุ่มลูกค้าและชื่อตามภาษาที่เลือก",
      en: "Manage customer group codes and names for the selected language.",
    },
    fields: [
      businessCodeField("groupcode", "รหัสกลุ่มลูกค้า", "Customer group code", true),
      namesField("names", "ชื่อกลุ่มลูกค้า", "Customer group names"),
    ],
  },
  {
    slug: "bookbankscreen",
    route: "/bookbankscreen",
    manual: "bookbank",
    kind: "main-crud",
    icon: "bank",
    basePath: "/payment/bookbank",
    listPath: "/payment/bookbank/list",
    idField: "guidfixed",
    title: { th: "บันทึกสมุดเงินฝากธนาคาร", en: "Bank Accounts" },
    subtitle: {
      th: "จัดการสมุดบัญชีธนาคาร สาขา เลขที่บัญชี ชื่อบัญชี และผังบัญชีที่เชื่อมโยง",
      en: "Manage bank accounts, branch, account number, and linked ledger accounts.",
    },
    fields: [
      businessCodeField("bookcode", "รหัสสมุดบัญชี", "Book bank code", true),
      namesField("names", "ชื่อสมุดบัญชี", "Book bank names"),
      textField("passbook", "เลขที่บัญชี", "Account number", true),
      textField("bankbranch", "สาขาธนาคาร", "Bank branch"),
      textField("accountname", "ชื่อบัญชี", "Account name"),
      textField("bankcode", "รหัสธนาคาร", "Bank code"),
      namesField("banknames", "ชื่อธนาคาร", "Bank names"),
      textField("accountcode", "รหัสผังบัญชี", "GL Account code"),
      {
        ...imageUploadField("logo", "โลโก้ธนาคาร / รูปสมุดบัญชี", "Bank logo / Passbook image"),
        helper: {
          th: "รองรับ PNG/JPG/WebP — เก็บไฟล์ในระบบจัดเก็บภาพ (S3)",
          en: "PNG/JPG/WebP supported — stored in S3 object storage.",
        },
      },
    ],
  },
  {
    slug: "creditor",
    route: "/creditor",
    manual: "creditor",
    kind: "main-crud",
    icon: "user",
    basePath: "/debtaccount/creditor",
    listPath: "/debtaccount/creditor/list",
    idField: "guidfixed",
    title: { th: "รายละเอียดเจ้าหนี้", en: "Vendor (Creditor)" },
    subtitle: {
      th: "จัดการข้อมูลเจ้าหนี้ ภาษี เงื่อนไขการค้า และที่อยู่",
      en: "Manage vendor master data, tax, trade terms, and addresses.",
    },
    fields: [
      businessCodeField("code", "รหัสเจ้าหนี้", "Vendor code", true),
      namesField("names", "ชื่อเจ้าหนี้", "Vendor name"),
      radioField(
        "personaltype",
        "ประเภทบุคคล",
        "Contact type",
        [
          { value: "1", label: "บุคคลธรรมดา", labels: { th: "บุคคลธรรมดา", en: "Individual" } },
          { value: "2", label: "นิติบุคคล", labels: { th: "นิติบุคคล", en: "Company" } },
        ],
        false,
        "number",
      ),
      masterMultiPickerField("groups", "กลุ่มเจ้าหนี้", "Vendor groups", "creditorgroup"),
      checkboxField("isdisabled", "ปิดใช้งาน", "Disabled"),
      {
        ...textField("taxid", "เลขประจำตัวผู้เสียภาษี", "Tax ID"),
        placeholder: "1234567890123",
      },
      {
        ...textField("branchnumber", "รหัสสาขาภาษี", "Tax branch number"),
        placeholder: "00000",
        helper: {
          th: "ต้องระบุเมื่อคู่ค้าจด VAT: สำนักงานใหญ่ = 00000, สาขาที่ 1 = 00001 (ตาม ภ.พ.20 ของคู่ค้า)",
          en: "Required when the partner is VAT-registered: head office = 00000, branch 1 = 00001 (per the partner's Por Por 20).",
        },
      },
      checkboxField("whtenabled", "หักภาษี ณ ที่จ่าย", "Withholding tax"),
      numberField("whtrate", "อัตราภาษี ณ ที่จ่าย", "Withholding tax rate (%)"),
      numberField("creditday", "เครดิต(วัน)", "Credit term (days)"),
      numberField("creditlimitbaht", "วงเงินเครดิต", "Credit limit (THB)"),
      textField("email", "อีเมล", "Email"),
      textField("addressforbilling.phoneprimary", "เบอร์โทรหลัก", "Primary phone"),
      textField("addressforbilling.phonesecondary", "เบอร์โทรสำรอง", "Secondary phone"),
      stringListField("addressforbilling.address", "ที่อยู่ออกบิล", "Tax invoice address lines"),
      thaiAddressField(
        "addressforbilling",
        "ที่อยู่ออกบิล (จังหวัด/อำเภอ/ตำบล/รหัสไปรษณีย์)",
        "Tax invoice address (province/district/subdistrict/zipcode)",
      ),
      stringListField("addressforactual.address", "ที่อยู่จริงที่ดำเนินกิจการ", "Actual/operating address lines"),
      {
        ...thaiAddressField(
          "addressforactual",
          "ที่อยู่จริงที่ดำเนินกิจการ (จังหวัด/อำเภอ/ตำบล/รหัสไปรษณีย์)",
          "Actual address (province/district/subdistrict/zipcode)",
        ),
        copyFromPrefix: "addressforbilling",
      },
      bankAccountsField("bankaccounts", "บัญชีธนาคาร", "Bank accounts"),
      imageGalleryField("images", "รูปภาพ", "Images"),
    ],
  },
  {
    slug: "debtor",
    route: "/debtor",
    manual: "debtor",
    kind: "main-crud",
    icon: "user",
    basePath: "/debtaccount/debtor",
    listPath: "/debtaccount/debtor/list",
    idField: "guidfixed",
    title: { th: "รายละเอียดลูกหนี้", en: "Customer (Debtor)" },
    subtitle: {
      th: "จัดการข้อมูลลูกค้า ระดับราคา เครดิต และสมาชิก",
      en: "Manage customer master data, price level, credit, and membership.",
    },
    fields: [
      businessCodeField("code", "รหัสลูกค้า", "Customer code", true),
      namesField("names", "ชื่อลูกค้า", "Customer name"),
      radioField(
        "personaltype",
        "ประเภทบุคคล",
        "Contact type",
        [
          { value: "1", label: "บุคคลธรรมดา", labels: { th: "บุคคลธรรมดา", en: "Individual" } },
          { value: "2", label: "นิติบุคคล", labels: { th: "นิติบุคคล", en: "Company" } },
        ],
        false,
        "number",
      ),
      masterMultiPickerField("groups", "กลุ่มลูกค้า", "Customer groups", "debtorgroup"),
      checkboxField("isdisabled", "ปิดใช้งาน", "Disabled"),
      textField("taxid", "เลขประจำตัวผู้เสียภาษี", "Tax ID"),
      {
        ...textField("branchnumber", "รหัสสาขาภาษี", "Tax branch number"),
        placeholder: "00000",
        helper: {
          th: "ต้องระบุเมื่อคู่ค้าจด VAT: สำนักงานใหญ่ = 00000, สาขาที่ 1 = 00001 (ตาม ภ.พ.20 ของคู่ค้า)",
          en: "Required when the partner is VAT-registered: head office = 00000, branch 1 = 00001 (per the partner's Por Por 20).",
        },
      },
      numberField("creditday", "เครดิต(วัน)", "Credit term (days)"),
      numberField("creditlimitbaht", "วงเงินเครดิต", "Credit limit (THB)"),
      textField("pricelevel", "ระดับราคาขายที่", "Price level"),
      textField("email", "อีเมล", "Email"),
      textField("addressforbilling.phoneprimary", "เบอร์โทรหลัก", "Primary phone"),
      textField("addressforbilling.phonesecondary", "เบอร์โทรสำรอง", "Secondary phone"),
      stringListField("addressforbilling.address", "ที่อยู่ออกบิล", "Tax invoice address lines"),
      thaiAddressField(
        "addressforbilling",
        "ที่อยู่ออกบิล (จังหวัด/อำเภอ/ตำบล/รหัสไปรษณีย์)",
        "Tax invoice address (province/district/subdistrict/zipcode)",
      ),
      stringListField("addressforactual.address", "ที่อยู่จริงที่ดำเนินกิจการ", "Actual/operating address lines"),
      {
        ...thaiAddressField(
          "addressforactual",
          "ที่อยู่จริงที่ดำเนินกิจการ (จังหวัด/อำเภอ/ตำบล/รหัสไปรษณีย์)",
          "Actual address (province/district/subdistrict/zipcode)",
        ),
        copyFromPrefix: "addressforbilling",
      },
      checkboxField("ismember", "เป็นสมาชิก", "Member"),
      textField("pointscode", "รหัสสะสมแต้ม", "Points code"),
      { ...numberField("pointbalance", "แต้มคงเหลือ", "Point balance"), readOnly: true },
      bankAccountsField("bankaccounts", "บัญชีธนาคารลูกค้า (สำหรับโอนเงินคืน/คืนมัดจำ)", "Customer bank accounts (for refunds/deposit returns)"),
      imageGalleryField("images", "รูปภาพ", "Images"),
    ],
  },
  {
    slug: "salechannelscreen",
    route: "/salechannelscreen",
    manual: "salechannelscreen",
    kind: "main-crud",
    icon: "network",
    basePath: "/sale-channel",
    listPath: "/sale-channel/list",
    idField: "guidfixed",
    title: { th: "ช่องทางขาย", en: "Sale Channel" },
    subtitle: {
      th: "จัดการช่องทางขาย เช่น หน้าร้าน Shopee Lazada TikTok พร้อม GP และระดับราคา",
      en: "Manage sales channels such as POS, Shopee, Lazada, TikTok with GP and price level.",
    },
    fields: [
      businessCodeField("code", "รหัสช่องทางขาย", "Sale channel code", true),
      textField("name", "ชื่อช่องทางขาย", "Sale channel name", true),
      numberField("gp", "GP (ค่าธรรมเนียมช่องทาง)", "GP (channel fee)"),
      radioField(
        "gptype",
        "ประเภท GP",
        "GP type",
        [
          { value: "0", label: "%", labels: { th: "%", en: "%" } },
          { value: "1", label: "บาท", labels: { th: "บาท", en: "THB" } },
        ],
        false,
        "number",
      ),
      numberField("price", "ระดับราคาขายที่", "Price level number"),
      textField("imageuri", "ลิงก์รูปภาพ", "Image URL"),
    ],
  },
  {
    slug: "transportchannelscreen",
    route: "/transportchannelscreen",
    manual: "transportchannelscreen",
    kind: "main-crud",
    icon: "activity",
    basePath: "/transport-channel",
    listPath: "/transport-channel/list",
    idField: "guidfixed",
    title: { th: "ช่องทางขนส่ง", en: "Transport Channel" },
    subtitle: {
      th: "จัดการช่องทางขนส่ง เช่น Kerry, Flash, ไปรษณีย์ไทย, EMS",
      en: "Manage transport channels such as Kerry, Flash, Thai Post, EMS.",
    },
    fields: [
      businessCodeField("code", "รหัสช่องทางขนส่ง", "Transport channel code", true),
      textField("name", "ชื่อช่องทางขนส่ง", "Transport channel name", true),
      textField("imageuri", "ลิงก์รูปภาพ", "Image URL"),
    ],
  },
];

function productMasterConfigs(): SystemSettingConfig[] {
  const codeNameFields = (
    codeLabelTh = "รหัส",
    codeLabelEn = "Code",
    nameLabelTh = "ชื่อ",
    nameLabelEn = "Name",
  ) => [
    businessCodeField("code", codeLabelTh, codeLabelEn, true),
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
        businessCodeField("unitcode", "รหัสหน่วยนับ", "Unit code", true),
        namesField("names", "ชื่อหน่วยนับ", "Unit names"),
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
        th: "เลือกชุดหมวด แล้วจัดหมวดหลักและหมวดย่อยให้เหมาะกับแต่ละช่องทางใช้งาน",
        en: "Choose a category set, then organize root categories and subcategories for each usage channel.",
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
      title: { th: "คลัง", en: "Warehouse → Location" },
      subtitle: {
        th: "จัดการคลังสินค้าและที่เก็บสินค้า",
        en: "Manage warehouses and storage locations.",
      },
      fields: [
        businessCodeField("code", "รหัสคลังสินค้า", "Warehouse code", true),
        namesField("names", "ชื่อคลังสินค้า", "Warehouse names"),
        {
          key: "location",
          label: { th: "ที่เก็บสินค้า", en: "Storage Locations" },
          type: "json",
        },
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
      "กำหนดราคาขาย/โปรโมชั่นสินค้า",
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
      "ยี่ห้อสินค้า",
      "Brand",
    ),
  ];
}

export const SYSTEM_SETTING_SLUGS = SYSTEM_SETTING_CONFIGS.map(
  (item) => item.slug,
);

export function getSystemSettingConfig(
  routeOrSlug: string,
): SystemSettingConfig | undefined {
  if (routeOrSlug === "bank" || routeOrSlug === "/bank") {
    return SYSTEM_SETTING_CONFIGS.find((item) => item.slug === "bookbankscreen");
  }
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

// thaiAddressField — key is the address object prefix (e.g. "addressforbilling", "contact"),
// not a leaf path. Renders the province/district/subdistrict/zipcode cascade editor scoped to
// that prefix's subkeys (see ThailandAddressFieldEditor's `prefix` prop).
function thaiAddressField(key: string, th: string, en: string): SystemSettingField {
  return { key, label: { th, en }, type: "thai-address" };
}

// bankAccountsField — repeatable bankcode/accountnumber/accountname rows (see BankAccountsEditor).
// key is the array field on the backend record (e.g. "bankaccounts").
function bankAccountsField(key: string, th: string, en: string): SystemSettingField {
  return { key, label: { th, en }, type: "bank-accounts" };
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

// masterMultiPickerField — chip multi-select over a master-data list (e.g. creditorgroup,
// debtorgroup). Stores an array of { guidfixed, code, names } in the form; buildPayload maps it
// to an array of guids for the backend's top-level `groups` request key.
function masterMultiPickerField(
  key: string,
  th: string,
  en: string,
  master: "creditorgroup" | "debtorgroup",
): SystemSettingField {
  return { key, label: { th, en }, type: "master-multi-picker", master };
}

function comboField(
  key: string,
  th: string,
  en: string,
  options: SystemSettingOption[] = [],
  required = false,
  optionSource?: "countries" | "timezones",
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
