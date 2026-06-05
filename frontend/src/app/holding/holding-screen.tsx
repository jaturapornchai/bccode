"use client";

import {
  AlertCircle,
  ArrowRight,
  Building2,
  CheckCircle2,
  GitBranch,
  KeyRound,
  Loader2,
  LogOut,
  Pencil,
  Plus,
  RefreshCcw,
  Save,
  Search,
  ShieldCheck,
  X,
} from "lucide-react";
import { useRouter } from "next/navigation";
import { FormEvent, useEffect, useMemo, useState } from "react";
import { normalizeLanguage, type LanguageCode } from "@/lib/i18n";
import {
  notifyWorkspaceChanged,
  shopDisplayName,
  type AuthSession,
  type ShopListItem,
  workspaceStorageKeys,
} from "@/lib/workspace-models";
import { LanguageDialog } from "../language-dialog";
import { ThemeToggle } from "../theme-toggle";

type HoldingListItem = ShopListItem & {
  companies?: unknown[];
  branches?: unknown[];
};

type CreateHoldingForm = {
  holdingcode: string;
  name: string;
  confirm_code: string;
  challenge_code: string;
};
type EditHoldingForm = {
  holdingcode: string;
  name: string;
};

type Notice = { type: "success" | "error" | "info"; text: string } | null;
type HoldingTextKey =
  | "addHolding"
  | "available"
  | "authRequired"
  | "cancel"
  | "choose"
  | "confirmCode"
  | "confirmCodeHelp"
  | "create"
  | "createFailed"
  | "createHoldingDescription"
  | "createHoldingTitle"
  | "createSuccess"
  | "creating"
  | "description"
  | "duplicateHolding"
  | "edit"
  | "emailLoginRequired"
  | "emptyDescription"
  | "emptyTitle"
  | "eyebrow"
  | "holdingCode"
  | "holdingCodeInvalid"
  | "holdingCodePlaceholder"
  | "holdingName"
  | "holdingNamePlaceholder"
  | "holdingNameRequired"
  | "loading"
  | "logout"
  | "requestFailed"
  | "save"
  | "saving"
  | "searchPlaceholder"
  | "selectFailed"
  | "selected"
  | "signedInAs"
  | "title"
  | "updateFailed"
  | "updateSuccess";

const emptyCreateHoldingForm: CreateHoldingForm = {
  holdingcode: "",
  name: "",
  confirm_code: "",
  challenge_code: "",
};
const holdingCodePattern = /^[a-z][a-z0-9_]{2,29}$/;

const holdingTextEn: Record<HoldingTextKey, string> = {
  addHolding: "Add Holding",
  available: "Available Holdings",
  authRequired: "Please sign in before selecting a Holding.",
  cancel: "Cancel",
  choose: "Choose",
  confirmCode: "Confirmation code",
  confirmCodeHelp: "Type the 4-digit code shown here to confirm real Holding creation.",
  create: "Create Holding",
  createFailed: "Could not create Holding.",
  createHoldingDescription: "Use a unique lowercase holdingcode. This code cannot be changed after creation.",
  createHoldingTitle: "Create Holding",
  createSuccess: "Holding created. Opening company selection.",
  creating: "Creating Holding",
  description: "Select the Holding you want to work in, then continue to company selection.",
  duplicateHolding: "This holdingcode already exists in your Holding list.",
  edit: "Edit",
  emailLoginRequired: "Sign in with email authentication before creating a Holding.",
  emptyDescription: "This account has no Holding access yet. Create or assign a Holding after Google sign-in.",
  emptyTitle: "No Holding found",
  eyebrow: "Step 2",
  holdingCode: "holdingcode",
  holdingCodeInvalid: "holdingcode must use a-z, 0-9, _, be 3-30 characters, and start with a-z.",
  holdingCodePlaceholder: "bc_demo",
  holdingName: "Holding name",
  holdingNamePlaceholder: "Example Holding",
  holdingNameRequired: "Please enter a Holding name.",
  loading: "Loading Holdings",
  logout: "Log out",
  requestFailed: "Request failed.",
  save: "Save",
  saving: "Saving",
  searchPlaceholder: "Search Holding",
  selectFailed: "Could not select Holding.",
  selected: "Holding selected. Opening company selection.",
  signedInAs: "Signed in as",
  title: "Select Holding",
  updateFailed: "Could not update Holding.",
  updateSuccess: "Holding name updated.",
};

const holdingText: Record<LanguageCode, Record<HoldingTextKey, string>> = {
  th: {
    addHolding: "เพิ่ม Holding",
    available: "Holding ที่เข้าได้",
    authRequired: "กรุณาเข้าสู่ระบบก่อนเลือก Holding",
    cancel: "ยกเลิก",
    choose: "เลือก",
    confirmCode: "รหัสยืนยัน",
    confirmCodeHelp: "พิมพ์รหัส 4 ตัวที่แสดง เพื่อยืนยันว่าต้องการสร้าง Holding จริง",
    create: "สร้าง Holding",
    createFailed: "สร้าง Holding ไม่สำเร็จ",
    createHoldingDescription: "ใช้ holdingcode ตัวเล็กไม่ซ้ำ รหัสนี้เปลี่ยนไม่ได้หลังสร้าง",
    createHoldingTitle: "สร้าง Holding",
    createSuccess: "สร้าง Holding แล้ว กำลังเปิดหน้าเลือกบริษัท",
    creating: "กำลังสร้าง Holding",
    description: "เลือก Holding ที่ต้องการใช้งาน แล้วไปหน้าเลือกบริษัท",
    duplicateHolding: "holdingcode นี้มีอยู่แล้วในรายการ Holding ของบัญชีนี้",
    edit: "แก้ไข",
    emailLoginRequired: "ต้องเข้าสู่ระบบด้วยการยืนยันตัวตนผ่านอีเมลก่อนสร้าง Holding",
    emptyDescription: "บัญชีนี้ยังไม่มีสิทธิ์เข้า Holding ให้สร้างหรือกำหนด Holding หลังเข้าสู่ระบบด้วย Google",
    emptyTitle: "ยังไม่มี Holding",
    eyebrow: "ขั้นตอนที่ 2",
    holdingCode: "holdingcode",
    holdingCodeInvalid: "holdingcode ต้องเป็น a-z, 0-9, _ ยาว 3-30 ตัว และขึ้นต้นด้วย a-z",
    holdingCodePlaceholder: "bc_demo",
    holdingName: "ชื่อ Holding",
    holdingNamePlaceholder: "เช่น Holding ตัวอย่าง",
    holdingNameRequired: "กรุณากรอกชื่อ Holding",
    loading: "กำลังโหลด Holding",
    logout: "ออกจากระบบ",
    requestFailed: "เรียกข้อมูลไม่สำเร็จ",
    save: "บันทึก",
    saving: "กำลังบันทึก",
    searchPlaceholder: "ค้นหา Holding",
    selectFailed: "เลือก Holding ไม่สำเร็จ",
    selected: "เลือก Holding แล้ว กำลังเปิดหน้าเลือกบริษัท",
    signedInAs: "เข้าสู่ระบบเป็น",
    title: "เลือก Holding",
    updateFailed: "บันทึกชื่อ Holding ไม่สำเร็จ",
    updateSuccess: "บันทึกชื่อ Holding แล้ว",
  },
  en: holdingTextEn,
  cn: {
    addHolding: "新增 Holding",
    available: "可访问的 Holding",
    authRequired: "请选择 Holding 前先登录。",
    cancel: "取消",
    choose: "选择",
    confirmCode: "确认码",
    confirmCodeHelp: "输入这里显示的 4 位数字，以确认真的要创建 Holding。",
    create: "创建 Holding",
    createFailed: "无法创建 Holding。",
    createHoldingDescription: "使用唯一的小写 holdingcode。创建后不能更改。",
    createHoldingTitle: "创建 Holding",
    createSuccess: "Holding 已创建。正在打开公司选择。",
    creating: "正在创建 Holding",
    description: "选择要使用的 Holding，然后进入公司选择。",
    duplicateHolding: "此 holdingcode 已存在于你的 Holding 列表中。",
    edit: "编辑",
    emailLoginRequired: "请先使用邮箱验证登录后再创建 Holding。",
    emptyDescription: "此账号尚无 Holding 权限。请在 Google 登录后创建或分配 Holding。",
    emptyTitle: "未找到 Holding",
    eyebrow: "第 2 步",
    holdingCode: "holdingcode",
    holdingCodeInvalid: "holdingcode 只能使用 a-z、0-9、_，长度 3-30，并以 a-z 开头。",
    holdingCodePlaceholder: "bc_demo",
    holdingName: "Holding 名称",
    holdingNamePlaceholder: "示例 Holding",
    holdingNameRequired: "请输入 Holding 名称。",
    loading: "正在加载 Holding",
    logout: "退出登录",
    requestFailed: "请求失败。",
    save: "保存",
    saving: "正在保存",
    searchPlaceholder: "搜索 Holding",
    selectFailed: "无法选择 Holding。",
    selected: "已选择 Holding，正在打开公司选择。",
    signedInAs: "登录身份",
    title: "选择 Holding",
    updateFailed: "无法更新 Holding。",
    updateSuccess: "Holding 名称已更新。",
  },
  ja: {
    addHolding: "Holding を追加",
    available: "利用可能な Holding",
    authRequired: "Holding を選択する前にログインしてください。",
    cancel: "キャンセル",
    choose: "選択",
    confirmCode: "確認コード",
    confirmCodeHelp: "ここに表示された4桁のコードを入力して、Holding 作成を確認してください。",
    create: "Holding を作成",
    createFailed: "Holding を作成できません。",
    createHoldingDescription: "一意の小文字 holdingcode を使用してください。作成後は変更できません。",
    createHoldingTitle: "Holding 作成",
    createSuccess: "Holding を作成しました。会社選択を開きます。",
    creating: "Holding を作成中",
    description: "利用する Holding を選択して、会社選択へ進みます。",
    duplicateHolding: "この holdingcode は Holding 一覧に既に存在します。",
    edit: "編集",
    emailLoginRequired: "Holding を作成する前にメール認証でログインしてください。",
    emptyDescription: "このアカウントには Holding 権限がありません。Google ログイン後に作成または割り当ててください。",
    emptyTitle: "Holding がありません",
    eyebrow: "ステップ 2",
    holdingCode: "holdingcode",
    holdingCodeInvalid: "holdingcode は a-z、0-9、_ のみ、3-30文字、先頭は a-z です。",
    holdingCodePlaceholder: "bc_demo",
    holdingName: "Holding 名",
    holdingNamePlaceholder: "サンプル Holding",
    holdingNameRequired: "Holding 名を入力してください。",
    loading: "Holding を読み込み中",
    logout: "ログアウト",
    requestFailed: "リクエストに失敗しました。",
    save: "保存",
    saving: "保存中",
    searchPlaceholder: "Holding を検索",
    selectFailed: "Holding を選択できません。",
    selected: "Holding を選択しました。会社選択を開きます。",
    signedInAs: "ログイン中",
    title: "Holding を選択",
    updateFailed: "Holding を更新できません。",
    updateSuccess: "Holding 名を更新しました。",
  },
  ko: {
    addHolding: "Holding 추가",
    available: "사용 가능한 Holding",
    authRequired: "Holding을 선택하기 전에 로그인하세요.",
    cancel: "취소",
    choose: "선택",
    confirmCode: "확인 코드",
    confirmCodeHelp: "표시된 4자리 코드를 입력해 Holding 생성을 확인하세요.",
    create: "Holding 생성",
    createFailed: "Holding을 생성할 수 없습니다.",
    createHoldingDescription: "중복 없는 소문자 holdingcode를 사용하세요. 생성 후 변경할 수 없습니다.",
    createHoldingTitle: "Holding 생성",
    createSuccess: "Holding을 생성했습니다. 회사 선택을 엽니다.",
    creating: "Holding 생성 중",
    description: "사용할 Holding을 선택한 뒤 회사 선택으로 이동합니다.",
    duplicateHolding: "이 holdingcode는 Holding 목록에 이미 있습니다.",
    edit: "수정",
    emailLoginRequired: "Holding을 생성하려면 이메일 인증으로 로그인하세요.",
    emptyDescription: "이 계정에는 Holding 권한이 없습니다. Google 로그인 후 생성하거나 배정하세요.",
    emptyTitle: "Holding 없음",
    eyebrow: "2단계",
    holdingCode: "holdingcode",
    holdingCodeInvalid: "holdingcode는 a-z, 0-9, _ 만 사용하고 3-30자이며 a-z로 시작해야 합니다.",
    holdingCodePlaceholder: "bc_demo",
    holdingName: "Holding 이름",
    holdingNamePlaceholder: "예시 Holding",
    holdingNameRequired: "Holding 이름을 입력하세요.",
    loading: "Holding 불러오는 중",
    logout: "로그아웃",
    requestFailed: "요청 실패",
    save: "저장",
    saving: "저장 중",
    searchPlaceholder: "Holding 검색",
    selectFailed: "Holding을 선택할 수 없습니다.",
    selected: "Holding을 선택했습니다. 회사 선택을 엽니다.",
    signedInAs: "로그인 계정",
    title: "Holding 선택",
    updateFailed: "Holding을 업데이트할 수 없습니다.",
    updateSuccess: "Holding 이름을 업데이트했습니다.",
  },
  lo: {
    addHolding: "ເພີ່ມ Holding",
    available: "Holding ທີ່ເຂົ້າໄດ້",
    authRequired: "ກະລຸນາ login ກ່ອນເລືອກ Holding.",
    cancel: "ຍົກເລີກ",
    choose: "ເລືອກ",
    confirmCode: "ລະຫັດຢືນຢັນ",
    confirmCodeHelp: "ພິມລະຫັດ 4 ຕົວທີ່ສະແດງເພື່ອຢືນຢັນການສ້າງ Holding.",
    create: "ສ້າງ Holding",
    createFailed: "ສ້າງ Holding ບໍ່ສຳເລັດ.",
    createHoldingDescription: "ໃຊ້ holdingcode ຕົວພິມນ້ອຍທີ່ບໍ່ຊ້ຳ. ສ້າງແລ້ວປ່ຽນບໍ່ໄດ້.",
    createHoldingTitle: "ສ້າງ Holding",
    createSuccess: "ສ້າງ Holding ແລ້ວ ກຳລັງເປີດໜ້າເລືອກບໍລິສັດ.",
    creating: "ກຳລັງສ້າງ Holding",
    description: "ເລືອກ Holding ທີ່ຈະໃຊ້ ແລ້ວໄປໜ້າເລືອກບໍລິສັດ.",
    duplicateHolding: "holdingcode ນີ້ມີຢູ່ໃນລາຍການ Holding ແລ້ວ.",
    edit: "ແກ້ໄຂ",
    emailLoginRequired: "ຕ້ອງ login ດ້ວຍການຢືນຢັນອີເມວກ່ອນສ້າງ Holding.",
    emptyDescription: "ບັນຊີນີ້ຍັງບໍ່ມີສິດ Holding. ສ້າງ ຫຼື ກຳນົດ Holding ຫຼັງ login ດ້ວຍ Google.",
    emptyTitle: "ບໍ່ພົບ Holding",
    eyebrow: "ຂັ້ນຕອນ 2",
    holdingCode: "holdingcode",
    holdingCodeInvalid: "holdingcode ໃຊ້ໄດ້ a-z, 0-9, _ ຍາວ 3-30 ຕົວ ແລະຂຶ້ນຕົ້ນດ້ວຍ a-z.",
    holdingCodePlaceholder: "bc_demo",
    holdingName: "ຊື່ Holding",
    holdingNamePlaceholder: "Holding ຕົວຢ່າງ",
    holdingNameRequired: "ກະລຸນາໃສ່ຊື່ Holding.",
    loading: "ກຳລັງໂຫຼດ Holding",
    logout: "ອອກຈາກລະບົບ",
    requestFailed: "ຄຳຂໍລົ້ມເຫຼວ.",
    save: "ບັນທຶກ",
    saving: "ກຳລັງບັນທຶກ",
    searchPlaceholder: "ຄົ້ນຫາ Holding",
    selectFailed: "ເລືອກ Holding ບໍ່ສຳເລັດ.",
    selected: "ເລືອກ Holding ແລ້ວ ກຳລັງເປີດໜ້າເລືອກບໍລິສັດ.",
    signedInAs: "login ເປັນ",
    title: "ເລືອກ Holding",
    updateFailed: "ອັບເດດ Holding ບໍ່ສຳເລັດ.",
    updateSuccess: "ອັບເດດຊື່ Holding ແລ້ວ.",
  },
  my: {
    addHolding: "Holding အသစ်ထည့်ရန်",
    available: "အသုံးပြုနိုင်သော Holding",
    authRequired: "Holding မရွေးခင် login ဝင်ပါ။",
    cancel: "မလုပ်တော့ပါ",
    choose: "ရွေးပါ",
    confirmCode: "အတည်ပြုကုဒ်",
    confirmCodeHelp: "Holding ဖန်တီးရန် သေချာကြောင်း အတည်ပြုရန် ပြထားသော 4 လုံးကုဒ်ကို ရိုက်ပါ။",
    create: "Holding ဖန်တီးရန်",
    createFailed: "Holding ဖန်တီး၍ မရပါ။",
    createHoldingDescription: "မတူညီသော lowercase holdingcode ကိုသုံးပါ။ ဖန်တီးပြီးနောက် ပြောင်း၍မရပါ။",
    createHoldingTitle: "Holding ဖန်တီးရန်",
    createSuccess: "Holding ဖန်တီးပြီးပါပြီ။ ကုမ္ပဏီရွေးချယ်မှုကို ဖွင့်နေသည်။",
    creating: "Holding ဖန်တီးနေသည်",
    description: "အသုံးပြုမည့် Holding ကိုရွေးပြီး ကုမ္ပဏီရွေးချယ်မှုသို့ ဆက်သွားပါ။",
    duplicateHolding: "ဤ holdingcode သည် သင့် Holding စာရင်းတွင် ရှိပြီးသားဖြစ်သည်။",
    edit: "ပြင်ရန်",
    emailLoginRequired: "Holding မဖန်တီးမီ အီးမေးလ်အတည်ပြု login ဖြင့် ဝင်ပါ။",
    emptyDescription: "ဤအကောင့်တွင် Holding အသုံးပြုခွင့် မရှိသေးပါ။ Google login ပြီးနောက် ဖန်တီး/သတ်မှတ်ပါ။",
    emptyTitle: "Holding မတွေ့ပါ",
    eyebrow: "အဆင့် 2",
    holdingCode: "holdingcode",
    holdingCodeInvalid: "holdingcode သည် a-z, 0-9, _ ကိုသာသုံးပြီး 3-30 လုံးရှိရမည်၊ a-z ဖြင့်စတင်ရမည်။",
    holdingCodePlaceholder: "bc_demo",
    holdingName: "Holding အမည်",
    holdingNamePlaceholder: "ဥပမာ Holding",
    holdingNameRequired: "Holding အမည် ထည့်ပါ။",
    loading: "Holding များကို ဖွင့်နေသည်",
    logout: "ထွက်မည်",
    requestFailed: "တောင်းဆိုမှု မအောင်မြင်ပါ။",
    save: "သိမ်းမည်",
    saving: "သိမ်းနေသည်",
    searchPlaceholder: "Holding ရှာရန်",
    selectFailed: "Holding ရွေးမရပါ။",
    selected: "Holding ရွေးပြီးပါပြီ။ ကုမ္ပဏီရွေးချယ်မှုကို ဖွင့်နေသည်။",
    signedInAs: "ဝင်ထားသောအကောင့်",
    title: "Holding ရွေးချယ်ပါ",
    updateFailed: "Holding ကို update မလုပ်နိုင်ပါ။",
    updateSuccess: "Holding အမည် update ပြီးပါပြီ။",
  },
  km: {
    addHolding: "បន្ថែម Holding",
    available: "Holding ដែលអាចចូលបាន",
    authRequired: "សូម login មុនពេលជ្រើស Holding។",
    cancel: "បោះបង់",
    choose: "ជ្រើស",
    confirmCode: "លេខកូដបញ្ជាក់",
    confirmCodeHelp: "វាយលេខកូដ 4 ខ្ទង់ដែលបង្ហាញនៅទីនេះ ដើម្បីបញ្ជាក់ការបង្កើត Holding។",
    create: "បង្កើត Holding",
    createFailed: "មិនអាចបង្កើត Holding បាន។",
    createHoldingDescription: "ប្រើ holdingcode អក្សរតូចដែលមិនស្ទួន។ បង្កើតរួចមិនអាចកែបាន។",
    createHoldingTitle: "បង្កើត Holding",
    createSuccess: "បានបង្កើត Holding។ កំពុងបើកជម្រើសក្រុមហ៊ុន។",
    creating: "កំពុងបង្កើត Holding",
    description: "ជ្រើស Holding ដែលត្រូវប្រើ បន្ទាប់មកបន្តទៅជ្រើសក្រុមហ៊ុន។",
    duplicateHolding: "holdingcode នេះមានរួចហើយក្នុងបញ្ជី Holding របស់អ្នក។",
    edit: "កែសម្រួល",
    emailLoginRequired: "សូម login ដោយផ្ទៀងផ្ទាត់អ៊ីមែល មុនពេលបង្កើត Holding។",
    emptyDescription: "គណនីនេះមិនទាន់មានសិទ្ធិ Holding ទេ។ សូមបង្កើត ឬកំណត់បន្ទាប់ពី login ដោយ Google។",
    emptyTitle: "រកមិនឃើញ Holding",
    eyebrow: "ជំហាន 2",
    holdingCode: "holdingcode",
    holdingCodeInvalid: "holdingcode ត្រូវប្រើ a-z, 0-9, _ ប្រវែង 3-30 តួ និងចាប់ផ្តើមដោយ a-z។",
    holdingCodePlaceholder: "bc_demo",
    holdingName: "ឈ្មោះ Holding",
    holdingNamePlaceholder: "Holding គំរូ",
    holdingNameRequired: "សូមបញ្ចូលឈ្មោះ Holding។",
    loading: "កំពុងផ្ទុក Holding",
    logout: "ចេញពីប្រព័ន្ធ",
    requestFailed: "សំណើបរាជ័យ។",
    save: "រក្សាទុក",
    saving: "កំពុងរក្សាទុក",
    searchPlaceholder: "ស្វែងរក Holding",
    selectFailed: "មិនអាចជ្រើស Holding បាន។",
    selected: "បានជ្រើស Holding កំពុងបើកជម្រើសក្រុមហ៊ុន។",
    signedInAs: "ចូលប្រើជា",
    title: "ជ្រើស Holding",
    updateFailed: "មិនអាចធ្វើបច្ចុប្បន្នភាព Holding បាន។",
    updateSuccess: "បានធ្វើបច្ចុប្បន្នភាពឈ្មោះ Holding។",
  },
  vi: {
    addHolding: "Thêm Holding",
    available: "Holding có thể truy cập",
    authRequired: "Vui lòng đăng nhập trước khi chọn Holding.",
    cancel: "Hủy",
    choose: "Chọn",
    confirmCode: "Mã xác nhận",
    confirmCodeHelp: "Nhập mã 4 chữ số hiển thị ở đây để xác nhận tạo Holding.",
    create: "Tạo Holding",
    createFailed: "Không thể tạo Holding.",
    createHoldingDescription: "Dùng holdingcode chữ thường và không trùng. Mã này không thể đổi sau khi tạo.",
    createHoldingTitle: "Tạo Holding",
    createSuccess: "Đã tạo Holding. Đang mở trang chọn công ty.",
    creating: "Đang tạo Holding",
    description: "Chọn Holding cần làm việc, sau đó chuyển sang chọn công ty.",
    duplicateHolding: "holdingcode này đã có trong danh sách Holding của bạn.",
    edit: "Sửa",
    emailLoginRequired: "Hãy đăng nhập bằng xác thực email trước khi tạo Holding.",
    emptyDescription: "Tài khoản này chưa có quyền Holding. Hãy tạo hoặc gán Holding sau khi đăng nhập Google.",
    emptyTitle: "Không có Holding",
    eyebrow: "Bước 2",
    holdingCode: "holdingcode",
    holdingCodeInvalid: "holdingcode chỉ dùng a-z, 0-9, _, dài 3-30 ký tự và bắt đầu bằng a-z.",
    holdingCodePlaceholder: "bc_demo",
    holdingName: "Tên Holding",
    holdingNamePlaceholder: "Holding mẫu",
    holdingNameRequired: "Vui lòng nhập tên Holding.",
    loading: "Đang tải Holding",
    logout: "Đăng xuất",
    requestFailed: "Yêu cầu thất bại.",
    save: "Lưu",
    saving: "Đang lưu",
    searchPlaceholder: "Tìm Holding",
    selectFailed: "Không thể chọn Holding.",
    selected: "Đã chọn Holding. Đang mở trang chọn công ty.",
    signedInAs: "Đăng nhập bằng",
    title: "Chọn Holding",
    updateFailed: "Không thể cập nhật Holding.",
    updateSuccess: "Đã cập nhật tên Holding.",
  },
  ms: {
    addHolding: "Tambah Holding",
    available: "Holding yang boleh diakses",
    authRequired: "Sila log masuk sebelum memilih Holding.",
    cancel: "Batal",
    choose: "Pilih",
    confirmCode: "Kod pengesahan",
    confirmCodeHelp: "Taip kod 4 digit yang dipaparkan di sini untuk mengesahkan ciptaan Holding.",
    create: "Cipta Holding",
    createFailed: "Tidak dapat mencipta Holding.",
    createHoldingDescription: "Gunakan holdingcode huruf kecil yang unik. Kod ini tidak boleh diubah selepas dicipta.",
    createHoldingTitle: "Cipta Holding",
    createSuccess: "Holding dicipta. Membuka pilihan syarikat.",
    creating: "Mencipta Holding",
    description: "Pilih Holding untuk digunakan, kemudian teruskan ke pilihan syarikat.",
    duplicateHolding: "holdingcode ini sudah wujud dalam senarai Holding anda.",
    edit: "Edit",
    emailLoginRequired: "Log masuk dengan pengesahan e-mel sebelum mencipta Holding.",
    emptyDescription: "Akaun ini belum mempunyai akses Holding. Cipta atau tetapkan Holding selepas log masuk Google.",
    emptyTitle: "Tiada Holding",
    eyebrow: "Langkah 2",
    holdingCode: "holdingcode",
    holdingCodeInvalid: "holdingcode mesti menggunakan a-z, 0-9, _, 3-30 aksara dan bermula dengan a-z.",
    holdingCodePlaceholder: "bc_demo",
    holdingName: "Nama Holding",
    holdingNamePlaceholder: "Holding contoh",
    holdingNameRequired: "Sila masukkan nama Holding.",
    loading: "Memuatkan Holding",
    logout: "Log keluar",
    requestFailed: "Permintaan gagal.",
    save: "Simpan",
    saving: "Menyimpan",
    searchPlaceholder: "Cari Holding",
    selectFailed: "Tidak dapat memilih Holding.",
    selected: "Holding dipilih. Membuka pilihan syarikat.",
    signedInAs: "Log masuk sebagai",
    title: "Pilih Holding",
    updateFailed: "Tidak dapat mengemas kini Holding.",
    updateSuccess: "Nama Holding dikemas kini.",
  },
  id: {
    addHolding: "Tambah Holding",
    available: "Holding yang dapat diakses",
    authRequired: "Silakan masuk sebelum memilih Holding.",
    cancel: "Batal",
    choose: "Pilih",
    confirmCode: "Kode konfirmasi",
    confirmCodeHelp: "Ketik kode 4 digit yang ditampilkan di sini untuk mengonfirmasi pembuatan Holding.",
    create: "Buat Holding",
    createFailed: "Tidak dapat membuat Holding.",
    createHoldingDescription: "Gunakan holdingcode huruf kecil yang unik. Kode ini tidak dapat diubah setelah dibuat.",
    createHoldingTitle: "Buat Holding",
    createSuccess: "Holding dibuat. Membuka pemilihan perusahaan.",
    creating: "Membuat Holding",
    description: "Pilih Holding yang akan digunakan, lalu lanjut ke pemilihan perusahaan.",
    duplicateHolding: "holdingcode ini sudah ada di daftar Holding Anda.",
    edit: "Edit",
    emailLoginRequired: "Masuk dengan verifikasi email sebelum membuat Holding.",
    emptyDescription: "Akun ini belum memiliki akses Holding. Buat atau tetapkan Holding setelah masuk Google.",
    emptyTitle: "Holding tidak ditemukan",
    eyebrow: "Langkah 2",
    holdingCode: "holdingcode",
    holdingCodeInvalid: "holdingcode harus memakai a-z, 0-9, _, 3-30 karakter dan diawali a-z.",
    holdingCodePlaceholder: "bc_demo",
    holdingName: "Nama Holding",
    holdingNamePlaceholder: "Holding contoh",
    holdingNameRequired: "Masukkan nama Holding.",
    loading: "Memuat Holding",
    logout: "Keluar",
    requestFailed: "Permintaan gagal.",
    save: "Simpan",
    saving: "Menyimpan",
    searchPlaceholder: "Cari Holding",
    selectFailed: "Tidak dapat memilih Holding.",
    selected: "Holding dipilih. Membuka pemilihan perusahaan.",
    signedInAs: "Masuk sebagai",
    title: "Pilih Holding",
    updateFailed: "Tidak dapat memperbarui Holding.",
    updateSuccess: "Nama Holding diperbarui.",
  },
  fil: {
    addHolding: "Magdagdag ng Holding",
    available: "Mga Holding na puwedeng ma-access",
    authRequired: "Mag-login muna bago pumili ng Holding.",
    cancel: "Kanselahin",
    choose: "Piliin",
    confirmCode: "Confirmation code",
    confirmCodeHelp: "I-type ang 4-digit code na ipinapakita dito para kumpirmahin ang paggawa ng Holding.",
    create: "Gumawa ng Holding",
    createFailed: "Hindi magawa ang Holding.",
    createHoldingDescription: "Gumamit ng unique lowercase holdingcode. Hindi na ito mababago pagkatapos malikha.",
    createHoldingTitle: "Gumawa ng Holding",
    createSuccess: "Nagawa ang Holding. Binubuksan ang pagpili ng kumpanya.",
    creating: "Ginagawa ang Holding",
    description: "Piliin ang Holding na gagamitin, pagkatapos ay magpatuloy sa pagpili ng kumpanya.",
    duplicateHolding: "Mayroon na ang holdingcode na ito sa iyong Holding list.",
    edit: "I-edit",
    emailLoginRequired: "Mag-login gamit ang email authentication bago gumawa ng Holding.",
    emptyDescription: "Wala pang Holding access ang account na ito. Gumawa o magtalaga ng Holding pagkatapos ng Google login.",
    emptyTitle: "Walang Holding",
    eyebrow: "Hakbang 2",
    holdingCode: "holdingcode",
    holdingCodeInvalid: "Ang holdingcode ay dapat gumamit ng a-z, 0-9, _, 3-30 character, at magsimula sa a-z.",
    holdingCodePlaceholder: "bc_demo",
    holdingName: "Pangalan ng Holding",
    holdingNamePlaceholder: "Halimbawang Holding",
    holdingNameRequired: "Ilagay ang pangalan ng Holding.",
    loading: "Nilo-load ang Holding",
    logout: "Mag-log out",
    requestFailed: "Nabigo ang request.",
    save: "I-save",
    saving: "Sine-save",
    searchPlaceholder: "Hanapin ang Holding",
    selectFailed: "Hindi mapili ang Holding.",
    selected: "Napili ang Holding. Binubuksan ang pagpili ng kumpanya.",
    signedInAs: "Naka-login bilang",
    title: "Piliin ang Holding",
    updateFailed: "Hindi ma-update ang Holding.",
    updateSuccess: "Na-update ang pangalan ng Holding.",
  },
};

export function HoldingScreen({ initialLanguage }: { initialLanguage: LanguageCode }) {
  const router = useRouter();
  const [mounted, setMounted] = useState(false);
  const [language, setLanguage] = useState<LanguageCode>(initialLanguage);
  const [auth, setAuth] = useState<AuthSession | null>(null);
  const [holdings, setHoldings] = useState<HoldingListItem[]>([]);
  const [query, setQuery] = useState("");
  const [busyHoldingCode, setBusyHoldingCode] = useState("");
  const [loading, setLoading] = useState(true);
  const [notice, setNotice] = useState<Notice>(null);
  const [createOpen, setCreateOpen] = useState(false);
  const [createForm, setCreateForm] = useState<CreateHoldingForm>(emptyCreateHoldingForm);
  const [creating, setCreating] = useState(false);
  const [editForm, setEditForm] = useState<EditHoldingForm | null>(null);
  const [savingHoldingCode, setSavingHoldingCode] = useState("");

  useEffect(() => {
    setMounted(true);
    const savedLanguage = normalizeLanguage(localStorage.getItem("user_language") ?? initialLanguage);
    setLanguage(savedLanguage);
    document.documentElement.lang = savedLanguage;

    const savedAuth = readAuth();
    if (!savedAuth) {
      setNotice({ type: "error", text: ht(savedLanguage, "authRequired") });
      setLoading(false);
      router.replace("/");
      return;
    }

    const selectedHoldingCode = activeHoldingCode(savedAuth);
    if (savedAuth.method === "password" && selectedHoldingCode) {
      router.replace("/workspace");
      return;
    }

    setAuth(savedAuth);
    void loadHoldings(savedAuth, savedLanguage);
  }, [initialLanguage, router]);

  useEffect(() => {
    document.documentElement.lang = language;
    localStorage.setItem("user_language", language);
  }, [language]);

  const filteredHoldings = useMemo(() => {
    const needle = query.trim().toLowerCase();
    if (!needle) return holdings;

    return holdings.filter((shop) => {
      const holdingCode = tenantCodeForShop(shop);
      return `${shopDisplayName(shop)} ${holdingCode}`.toLowerCase().includes(needle);
    });
  }, [holdings, query]);

  const signedInAs = auth?.profile?.email || auth?.username || "";
  const canCreateHolding = Boolean(
    auth && auth.method !== "password" && (auth.profile?.email || auth.username.includes("@")),
  );
  const normalizedCreateHoldingCode = normalizeHoldingCode(createForm.holdingcode);
  const createHoldingDuplicate = useMemo(() => {
    if (!normalizedCreateHoldingCode) return false;
    return holdings.some((shop) => tenantCodeForShop(shop).toLowerCase() === normalizedCreateHoldingCode);
  }, [holdings, normalizedCreateHoldingCode]);

  async function loadHoldings(currentAuth: AuthSession, activeLanguage: LanguageCode, options?: { preserveNotice?: boolean }) {
    setLoading(true);
    if (!options?.preserveNotice) setNotice(null);
    try {
      const payload = await callWorkspaceApi<{ data?: HoldingListItem[] }>(currentAuth, "holdings");
      const nextHoldings = dedupeHoldings(Array.isArray(payload.data) ? payload.data : []);
      setHoldings(nextHoldings);
      if (nextHoldings.length === 0 && !options?.preserveNotice) {
        setNotice({ type: "info", text: ht(activeLanguage, "emptyDescription") });
      }
    } catch (error) {
      setNotice({
        type: "error",
        text: error instanceof Error && error.message ? error.message : ht(activeLanguage, "requestFailed"),
      });
    } finally {
      setLoading(false);
    }
  }

  function openCreateHolding() {
    if (!canCreateHolding) {
      setNotice({ type: "error", text: ht(language, "emailLoginRequired") });
      return;
    }
    setCreateForm({ ...emptyCreateHoldingForm, challenge_code: generateConfirmationCode() });
    setCreateOpen(true);
    setEditForm(null);
    setNotice(null);
  }

  function closeCreateHolding() {
    setCreateOpen(false);
    setCreateForm(emptyCreateHoldingForm);
  }

  async function createHolding(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!auth) return;
    if (!canCreateHolding) {
      setNotice({ type: "error", text: ht(language, "emailLoginRequired") });
      return;
    }

    const holdingCode = normalizeHoldingCode(createForm.holdingcode);
    const name = normalizeDisplayName(createForm.name);
    if (!holdingCode || !holdingCodePattern.test(holdingCode)) {
      setNotice({ type: "error", text: ht(language, "holdingCodeInvalid") });
      return;
    }
    if (createHoldingDuplicate) {
      setNotice({ type: "error", text: ht(language, "duplicateHolding") });
      return;
    }
    if (!name) {
      setNotice({ type: "error", text: ht(language, "holdingNameRequired") });
      return;
    }
    if (createForm.confirm_code !== createForm.challenge_code) {
      setNotice({ type: "error", text: ht(language, "confirmCodeHelp") });
      return;
    }

    setCreating(true);
    setNotice(null);
    try {
      await callWorkspaceApi(auth, "create-holding", {
        method: "POST",
        body: createHoldingPayload(holdingCode, name, signedInAs),
      });
      closeCreateHolding();
      setNotice({ type: "success", text: ht(language, "createSuccess") });
      await activateHolding(holdingCode);
    } catch (error) {
      setNotice({
        type: "error",
        text: error instanceof Error && error.message ? error.message : ht(language, "createFailed"),
      });
    } finally {
      setCreating(false);
    }
  }

  function startEditHolding(shop: HoldingListItem) {
    if (!canEditHolding(shop, auth)) return;
    const holdingCode = tenantCodeForShop(shop);
    if (!holdingCode) return;
    setCreateOpen(false);
    setEditForm({
      holdingcode: holdingCode,
      name: displayNameForEdit(shop),
    });
    setNotice(null);
  }

  function closeEditHolding() {
    setEditForm(null);
    setSavingHoldingCode("");
  }

  async function updateHolding(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!auth || !editForm) return;

    const holdingCode = normalizeHoldingCode(editForm.holdingcode);
    const name = normalizeDisplayName(editForm.name);
    if (!holdingCode || !holdingCodePattern.test(holdingCode)) {
      setNotice({ type: "error", text: ht(language, "holdingCodeInvalid") });
      return;
    }
    if (!name) {
      setNotice({ type: "error", text: ht(language, "holdingNameRequired") });
      return;
    }

    setSavingHoldingCode(holdingCode);
    setNotice(null);
    try {
      await callWorkspaceApi(auth, "update-holding", {
        method: "POST",
        body: { holdingcode: holdingCode, name1: name },
      });
      setEditForm(null);
      await loadHoldings(auth, language, { preserveNotice: true });
      setNotice({ type: "success", text: ht(language, "updateSuccess") });
    } catch (error) {
      setNotice({
        type: "error",
        text: error instanceof Error && error.message ? error.message : ht(language, "updateFailed"),
      });
    } finally {
      setSavingHoldingCode("");
    }
  }

  async function selectHolding(shop: HoldingListItem) {
    if (!auth) return;
    const holdingCode = tenantCodeForShop(shop);
    if (!holdingCode) {
      setNotice({ type: "error", text: ht(language, "selectFailed") });
      return;
    }

    setBusyHoldingCode(holdingCode);
    setNotice(null);
    try {
      await activateHolding(holdingCode);
    } catch (error) {
      setNotice({
        type: "error",
        text: error instanceof Error && error.message ? error.message : ht(language, "selectFailed"),
      });
    } finally {
      setBusyHoldingCode("");
    }
  }

  async function activateHolding(holdingCode: string) {
    if (!auth) return;
    await callWorkspaceApi(auth, "select-holding", {
      method: "POST",
      body: { holdingcode: holdingCode },
    });
    const nextAuth: AuthSession = { ...auth, holdingcode: holdingCode };
    localStorage.setItem(workspaceStorageKeys.auth, JSON.stringify(nextAuth));
    localStorage.setItem(workspaceStorageKeys.holdingCode, holdingCode);
    localStorage.removeItem(workspaceStorageKeys.workspace);
    localStorage.removeItem(workspaceStorageKeys.shopInfo);
    localStorage.removeItem(workspaceStorageKeys.branch);
    notifyWorkspaceChanged();
    setAuth(nextAuth);
    setNotice({ type: "success", text: ht(language, "selected") });
    router.push("/workspace");
  }

  function logout() {
    localStorage.removeItem(workspaceStorageKeys.auth);
    localStorage.removeItem(workspaceStorageKeys.holdingCode);
    localStorage.removeItem(workspaceStorageKeys.workspace);
    localStorage.removeItem(workspaceStorageKeys.shopInfo);
    localStorage.removeItem(workspaceStorageKeys.branch);
    notifyWorkspaceChanged();
    router.replace("/");
  }

  return (
    <main className="login-shell">
      <section className="brand-panel" aria-label="BC Ai Account">
        <div className="brand-badge-row">
          <span className="brand-mark">BC</span>
          <div>
            <p className="eyebrow">BC Ai Account</p>
            <strong>{ht(language, "title")}</strong>
          </div>
        </div>
        <div className="brand-hero">
          <div className="brand-copy">
            <p className="eyebrow">{ht(language, "eyebrow")}</p>
            <h1>Holding</h1>
            <p>{ht(language, "description")}</p>
          </div>
        </div>
        <div className="brand-feature-grid">
          <div className="brand-feature-card">
            <ShieldCheck aria-hidden="true" size={18} />
            <div>
              <strong>{ht(language, "signedInAs")}</strong>
              <span>{signedInAs || "-"}</span>
            </div>
          </div>
          <div className="brand-feature-card">
            <Building2 aria-hidden="true" size={18} />
            <div>
              <strong>{ht(language, "available")}</strong>
              <span>{holdings.length}</span>
            </div>
          </div>
        </div>
      </section>

      <section className="form-panel">
        <section className="login-card" aria-label={ht(language, "title")}>
          <div className="card-header">
            <div>
              <p className="eyebrow">{ht(language, "eyebrow")}</p>
              <h2>{ht(language, "title")}</h2>
            </div>
            <div className="header-actions">
              {canCreateHolding ? (
                <button className="secondary-button holding-create-toggle" type="button" onClick={openCreateHolding}>
                  <Plus aria-hidden="true" size={16} />
                  <span>{ht(language, "addHolding")}</span>
                </button>
              ) : null}
              <ThemeToggle language={language} />
              <LanguageDialog language={language} onLanguageChange={setLanguage} />
              <button className="icon-button" type="button" onClick={logout} aria-label={ht(language, "logout")} title={ht(language, "logout")}>
                <LogOut aria-hidden="true" size={18} />
              </button>
            </div>
          </div>

          <label className="search-shell">
            <Search aria-hidden="true" size={17} />
            <input
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              placeholder={ht(language, "searchPlaceholder")}
            />
          </label>

          {createOpen ? (
            <form className="holding-create-panel" onSubmit={createHolding}>
              <div className="holding-create-head">
                <div>
                  <strong>{ht(language, "createHoldingTitle")}</strong>
                  <span>{ht(language, "createHoldingDescription")}</span>
                </div>
                <button className="icon-button" type="button" onClick={closeCreateHolding} aria-label={ht(language, "cancel")} title={ht(language, "cancel")}>
                  <X aria-hidden="true" size={18} />
                </button>
              </div>

              <label className="field-group">
                <span>{ht(language, "holdingCode")}</span>
                <div className="input-shell">
                  <Building2 aria-hidden="true" size={18} />
                  <input
                    autoComplete="off"
                    value={createForm.holdingcode}
                    onChange={(event) => {
                      setCreateForm((current) => ({ ...current, holdingcode: normalizeHoldingCode(event.target.value) }));
                    }}
                    placeholder={ht(language, "holdingCodePlaceholder")}
                  />
                </div>
                <small className={createHoldingDuplicate ? "field-help danger" : "field-help"}>
                  {createHoldingDuplicate ? ht(language, "duplicateHolding") : ht(language, "holdingCodeInvalid")}
                </small>
              </label>

              <label className="field-group">
                <span>{ht(language, "holdingName")}</span>
                <div className="input-shell">
                  <ShieldCheck aria-hidden="true" size={18} />
                  <input
                    autoComplete="organization"
                    value={createForm.name}
                    onChange={(event) => setCreateForm((current) => ({ ...current, name: event.target.value }))}
                    placeholder={ht(language, "holdingNamePlaceholder")}
                  />
                </div>
              </label>

              <div className="holding-confirm-box">
                <div>
                  <span>{ht(language, "confirmCode")}</span>
                  <strong>{createForm.challenge_code}</strong>
                  <small>{ht(language, "confirmCodeHelp")}</small>
                </div>
                <button
                  className="icon-button"
                  type="button"
                  onClick={() => setCreateForm((current) => ({
                    ...current,
                    confirm_code: "",
                    challenge_code: generateConfirmationCode(),
                  }))}
                  aria-label={ht(language, "confirmCode")}
                  title={ht(language, "confirmCode")}
                >
                  <RefreshCcw aria-hidden="true" size={17} />
                </button>
              </div>

              <label className="field-group">
                <span>{ht(language, "confirmCode")}</span>
                <div className="input-shell">
                  <KeyRound aria-hidden="true" size={18} />
                  <input
                    autoComplete="one-time-code"
                    inputMode="numeric"
                    maxLength={4}
                    value={createForm.confirm_code}
                    onChange={(event) => {
                      setCreateForm((current) => ({
                        ...current,
                        confirm_code: event.target.value.replace(/\D/g, "").slice(0, 4),
                      }));
                    }}
                    placeholder="0000"
                  />
                </div>
              </label>

              <div className="holding-create-actions">
                <button className="secondary-button" type="button" onClick={closeCreateHolding} disabled={creating}>
                  {ht(language, "cancel")}
                </button>
                <button
                  className="primary-button"
                  type="submit"
                  disabled={creating || createHoldingDuplicate}
                >
                  {creating ? <Loader2 className="spin" aria-hidden="true" size={18} /> : <Plus aria-hidden="true" size={18} />}
                  <span>{creating ? ht(language, "creating") : ht(language, "create")}</span>
                </button>
              </div>
            </form>
          ) : null}

          {editForm ? (
            <form className="holding-create-panel holding-edit-panel" onSubmit={updateHolding}>
              <div className="holding-create-head">
                <div>
                  <strong>{ht(language, "edit")}: {editForm.holdingcode}</strong>
                  <span>{ht(language, "holdingNameRequired")}</span>
                </div>
                <button className="icon-button" type="button" onClick={closeEditHolding} aria-label={ht(language, "cancel")} title={ht(language, "cancel")}>
                  <X aria-hidden="true" size={18} />
                </button>
              </div>

              <label className="field-group">
                <span>{ht(language, "holdingName")}</span>
                <div className="input-shell">
                  <ShieldCheck aria-hidden="true" size={18} />
                  <input
                    autoComplete="organization"
                    value={editForm.name}
                    onChange={(event) => setEditForm((current) => current ? ({ ...current, name: event.target.value }) : current)}
                    placeholder={ht(language, "holdingNamePlaceholder")}
                  />
                </div>
              </label>

              <div className="holding-create-actions">
                <button className="secondary-button" type="button" onClick={closeEditHolding} disabled={Boolean(savingHoldingCode)}>
                  {ht(language, "cancel")}
                </button>
                <button className="primary-button" type="submit" disabled={Boolean(savingHoldingCode)}>
                  {savingHoldingCode ? <Loader2 className="spin" aria-hidden="true" size={18} /> : <Save aria-hidden="true" size={18} />}
                  <span>{savingHoldingCode ? ht(language, "saving") : ht(language, "save")}</span>
                </button>
              </div>
            </form>
          ) : null}

          {notice ? (
            <div className={`message ${notice.type === "success" ? "success" : notice.type === "error" ? "error" : "info"}`}>
              {notice.type === "success" ? <CheckCircle2 aria-hidden="true" size={18} /> : <AlertCircle aria-hidden="true" size={18} />}
              <span>{notice.text}</span>
            </div>
          ) : null}

          {loading || !mounted ? (
            <div className="loading-state">
              <Loader2 className="spin" aria-hidden="true" size={24} />
              <span>{ht(language, "loading")}</span>
            </div>
          ) : filteredHoldings.length === 0 ? (
            <div className="workspace-empty-state">
              <Building2 aria-hidden="true" size={24} />
              <strong>{ht(language, "emptyTitle")}</strong>
              <span>{ht(language, "emptyDescription")}</span>
            </div>
          ) : (
            <div className="workspace-card-grid">
              {filteredHoldings.map((shop, index) => {
                const holdingCode = tenantCodeForShop(shop);
                const companyCount = Array.isArray(shop.companies) ? shop.companies.length : 0;
                const branchCount = Array.isArray(shop.branches) ? shop.branches.length : 0;
                const busy = busyHoldingCode === holdingCode;
                const canEdit = canEditHolding(shop, auth);

                return (
                  <div
                    className="group/card relative flex items-center justify-between p-4 rounded-xl border border-border/60 bg-card/75 backdrop-blur-md overflow-hidden shadow-sm hover:shadow-md hover:border-primary/30 hover:scale-[1.01] transition-all duration-300 gap-3"
                    key={holdingCode || index}
                  >
                    {/* Left color ribbon indicator */}
                    <div className={`absolute left-0 top-0 bottom-0 w-1 bg-gradient-to-b ${
                      index % 2 === 0 ? "from-primary to-primary/60" : "from-teal-600 to-teal-500/60"
                    }`} />

                    {/* Left side: Avatar + Holding details */}
                    <button
                      className="flex-1 flex items-center gap-3 text-left border-0 bg-transparent p-0 cursor-pointer disabled:cursor-wait"
                      disabled={Boolean(busyHoldingCode)}
                      onClick={() => void selectHolding(shop)}
                      type="button"
                    >
                      {/* Avatar */}
                      <span className={`w-10 h-10 rounded-lg flex items-center justify-center shrink-0 transition-transform duration-300 group-hover/card:scale-105 group-hover/card:rotate-2 ${
                        index % 2 === 0
                          ? "bg-primary/10 text-primary"
                          : "bg-teal-500/10 text-teal-600 dark:text-teal-400"
                      }`}>
                        {busy ? (
                          <Loader2 className="spin" aria-hidden="true" size={18} />
                        ) : (
                          <Building2 aria-hidden="true" size={18} />
                        )}
                      </span>

                      {/* Info & Badges */}
                      <div className="min-w-0 flex-1 space-y-1">
                        <div className="flex flex-wrap items-baseline gap-x-2">
                          <strong className="text-sm font-bold text-foreground tracking-tight leading-snug group-hover/card:text-primary transition-colors duration-200">
                            {shopDisplayName(shop)}
                          </strong>
                          <span className="font-mono text-[9px] px-1.5 py-0.2 bg-muted border border-border/50 text-muted-foreground rounded uppercase font-bold shrink-0">
                            {holdingCode}
                          </span>
                        </div>

                        {/* Horizontal stats */}
                        <div className="flex flex-wrap items-center gap-1.5 mt-0.5">
                          {companyCount > 0 ? (
                            <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-semibold bg-primary/5 text-primary border border-primary/10">
                              <Building2 size={10} />
                              <span>{companyCount} {language === "th" ? "บริษัท" : "Companies"}</span>
                            </span>
                          ) : null}
                          {branchCount > 0 ? (
                            <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-semibold bg-teal-500/5 text-teal-600 dark:text-teal-400 border border-teal-500/10">
                              <GitBranch size={10} />
                              <span>{branchCount} {language === "th" ? "สาขา" : "Branches"}</span>
                            </span>
                          ) : null}
                        </div>
                      </div>
                    </button>

                    {/* Right side: Edit & Choose actions */}
                    <div className="flex items-center gap-1.5 shrink-0 z-10">
                      {canEdit ? (
                        <button
                          className="w-8 h-8 rounded-lg flex items-center justify-center text-muted-foreground hover:text-primary hover:bg-primary/5 border border-transparent hover:border-primary/10 transition-all cursor-pointer"
                          disabled={Boolean(busyHoldingCode || savingHoldingCode)}
                          onClick={() => startEditHolding(shop)}
                          type="button"
                          aria-label={`${ht(language, "edit")} ${shopDisplayName(shop)}`}
                          title={ht(language, "edit")}
                        >
                          <Pencil aria-hidden="true" size={14} />
                        </button>
                      ) : null}

                      <button
                        className="h-8 px-3.5 rounded-lg flex items-center gap-1 text-[11px] font-bold bg-primary text-primary-foreground hover:brightness-110 shadow-sm transition-all cursor-pointer"
                        disabled={Boolean(busyHoldingCode)}
                        onClick={() => void selectHolding(shop)}
                        type="button"
                      >
                        <span>{ht(language, "choose")}</span>
                        <ArrowRight aria-hidden="true" size={12} className="transition-transform group-hover/card:translate-x-0.5" />
                      </button>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </section>
      </section>
    </main>
  );
}

function ht(language: LanguageCode, key: HoldingTextKey): string {
  return holdingText[language]?.[key] ?? holdingTextEn[key];
}

function normalizeHoldingCode(value: string): string {
  return value.trim().toLowerCase();
}

function normalizeDisplayName(value: string): string {
  return value
    .replace(/[\u0000-\u001F\u007F]/g, " ")
    .replace(/\s+/g, " ")
    .trim()
    .slice(0, 120);
}

function generateConfirmationCode(): string {
  if (typeof window !== "undefined" && window.crypto?.getRandomValues) {
    const values = new Uint32Array(1);
    window.crypto.getRandomValues(values);
    return String(1000 + (values[0] % 9000));
  }
  return String(Math.floor(1000 + Math.random() * 9000));
}

function readAuth(): AuthSession | null {
  try {
    const raw = localStorage.getItem(workspaceStorageKeys.auth);
    if (!raw) return null;
    const auth = JSON.parse(raw) as AuthSession;
    return auth.token && auth.backendUrl ? auth : null;
  } catch {
    return null;
  }
}

function activeHoldingCode(auth: AuthSession): string {
  return auth.holdingcode?.trim() || localStorage.getItem(workspaceStorageKeys.holdingCode)?.trim() || "";
}

async function callWorkspaceApi<T extends Record<string, unknown>>(
  auth: AuthSession,
  path: string,
  init?: { method?: "GET" | "POST"; body?: Record<string, unknown> },
): Promise<T> {
  const response = await fetch(`/api/workspace/${path}`, {
    method: init?.method ?? "GET",
    headers: {
      "Content-Type": "application/json",
      "x-bc-backend-url": auth.backendUrl,
      Authorization: `Bearer ${auth.token}`,
    },
    body: init?.method === "POST" ? JSON.stringify({ backendUrl: auth.backendUrl, ...(init.body ?? {}) }) : undefined,
    cache: "no-store",
  });
  const data = await response.json() as T & { success?: boolean; message?: string };
  if (!response.ok || data.success === false) {
    throw new Error(data.message ?? "");
  }
  return data;
}

function dedupeHoldings(shops: HoldingListItem[]): HoldingListItem[] {
  const seen = new Set<string>();
  const result: HoldingListItem[] = [];
  shops.forEach((shop) => {
    const holdingCode = tenantCodeForShop(shop);
    if (!holdingCode || seen.has(holdingCode)) return;
    seen.add(holdingCode);
    result.push(shop);
  });
  return result;
}

function tenantCodeForShop(shop: HoldingListItem | null | undefined): string {
  return shop?.holdingcode?.trim() || "";
}

function displayNameForEdit(shop: HoldingListItem): string {
  const holdingCode = tenantCodeForShop(shop);
  const displayName = shopDisplayName(shop).trim();
  return displayName && displayName !== holdingCode ? displayName : "";
}

function canEditHolding(shop: HoldingListItem, auth: AuthSession | null): boolean {
  if (!auth) return false;
  if (Number(shop.role) === 2 || shop.is_creator === true) return true;
  const createdBy = shop.createdby?.trim().toLowerCase();
  if (!createdBy) return false;
  return [auth.username, auth.profile?.email]
    .map((item) => item?.trim().toLowerCase() ?? "")
    .filter(Boolean)
    .includes(createdBy);
}

function createHoldingPayload(holdingCode: string, name: string, ownerEmail: string): Record<string, unknown> {
  return {
    address: [],
    branchcode: "",
    holdingcode: holdingCode,
    images: [],
    logo: "",
    name1: name,
    names: [{ code: "th", name }],
    profilepicture: "",
    settings: {
      emailowners: ownerEmail ? [ownerEmail] : [],
      emailstaffs: [],
      isusebranch: false,
      isusedepartment: false,
      language: "th",
      languageconfigs: [{ code: "th", codetranslator: "th", name: "ภาษาไทย", is_use: true, isdefault: true }],
      latitude: 0,
      longitude: 0,
      taxid: "",
      vatrate: 7,
      vattypesale: 0,
      vattypepurchase: 0,
      inquirytypesale: 0,
      inquirytypepurchase: 0,
    },
    telephone: "",
    businesstype: {},
  };
}
