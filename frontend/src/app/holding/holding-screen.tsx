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
import { isValidHoldingCode, normalizeHoldingCode } from "@/lib/holding-code";
import { normalizeLanguage, type LanguageCode } from "@/lib/i18n";
import {
  notifyWorkspaceChanged,
  shopDisplayName,
  type AuthSession,
  type ShopListItem,
  workspaceStorageKeys,
} from "@/lib/workspace-models";
import { AppHeaderControls } from "../app-header-controls";

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

const holdingTextEn: Record<HoldingTextKey, string> = {
  addHolding: "Add business group",
  available: "Available business groups",
  authRequired: "Please sign in before selecting a business group.",
  cancel: "Cancel",
  choose: "Choose",
  confirmCode: "Confirmation code",
  confirmCodeHelp: "Type the 4-digit code shown here to confirm real business group creation.",
  create: "Create business group",
  createFailed: "Could not create business group.",
  createHoldingDescription: "Use a unique lowercase business group code. This code cannot be changed after creation.",
  createHoldingTitle: "Create business group",
  createSuccess: "Business group created. Opening company selection.",
  creating: "Creating business group",
  description: "Select the business group you want to work in, then continue to company selection.",
  duplicateHolding: "This business group code already exists in your business group list.",
  edit: "Edit",
  emailLoginRequired: "Sign in with email authentication before creating a business group.",
  emptyDescription: "This account has no business group access yet. Create or assign a business group after Google sign-in.",
  emptyTitle: "No business group found",
  eyebrow: "Step 2",
  holdingCode: "business group code",
  holdingCodeInvalid: "Business group code must use a-z and 0-9 only, be 3-30 characters, and start with a-z. Do not use _ or symbols.",
  holdingCodePlaceholder: "bcdemo01",
  holdingName: "Business group name",
  holdingNamePlaceholder: "Example business group",
  holdingNameRequired: "Please enter a business group name.",
  loading: "Loading business groups",
  logout: "Log out",
  requestFailed: "Request failed.",
  save: "Save",
  saving: "Saving",
  searchPlaceholder: "Search business group",
  selectFailed: "Could not select business group.",
  selected: "Business group selected. Opening company selection.",
  signedInAs: "Signed in as",
  title: "Select business group",
  updateFailed: "Could not update business group.",
  updateSuccess: "Business group name updated.",
};

const holdingText: Record<LanguageCode, Record<HoldingTextKey, string>> = {
  th: {
    addHolding: "เพิ่มกลุ่มกิจการ",
    available: "กลุ่มกิจการที่เข้าได้",
    authRequired: "กรุณาเข้าสู่ระบบก่อนเลือกกลุ่มกิจการ",
    cancel: "ยกเลิก",
    choose: "เลือก",
    confirmCode: "รหัสยืนยัน",
    confirmCodeHelp: "พิมพ์รหัส 4 ตัวที่แสดง เพื่อยืนยันว่าต้องการสร้างกลุ่มกิจการจริง",
    create: "สร้างกลุ่มกิจการ",
    createFailed: "สร้างกลุ่มกิจการไม่สำเร็จ",
    createHoldingDescription: "ใช้รหัสกลุ่มกิจการเป็นตัวเล็กไม่ซ้ำ รหัสนี้เปลี่ยนไม่ได้หลังสร้าง",
    createHoldingTitle: "สร้างกลุ่มกิจการ",
    createSuccess: "สร้างกลุ่มกิจการแล้ว กำลังเปิดหน้าเลือกบริษัท",
    creating: "กำลังสร้างกลุ่มกิจการ",
    description: "เลือกกลุ่มกิจการที่ต้องการใช้งาน แล้วไปหน้าเลือกบริษัท",
    duplicateHolding: "รหัสกลุ่มกิจการนี้มีอยู่แล้วในรายการกลุ่มกิจการของบัญชีนี้",
    edit: "แก้ไข",
    emailLoginRequired: "ต้องเข้าสู่ระบบด้วยการยืนยันตัวตนผ่านอีเมลก่อนสร้างกลุ่มกิจการ",
    emptyDescription: "บัญชีนี้ยังไม่มีสิทธิ์เข้ากลุ่มกิจการ ให้สร้างหรือกำหนดกลุ่มกิจการหลังเข้าสู่ระบบด้วย Google",
    emptyTitle: "ยังไม่มีกลุ่มกิจการ",
    eyebrow: "ขั้นตอนที่ 2",
    holdingCode: "รหัสกลุ่มกิจการ",
    holdingCodeInvalid: "รหัสกลุ่มกิจการต้องใช้ a-z และ 0-9 เท่านั้น ยาว 3-30 ตัว และขึ้นต้นด้วย a-z ห้ามใช้ _ หรือสัญลักษณ์",
    holdingCodePlaceholder: "bcdemo01",
    holdingName: "ชื่อกลุ่มกิจการ",
    holdingNamePlaceholder: "เช่น กลุ่มกิจการตัวอย่าง",
    holdingNameRequired: "กรุณากรอกชื่อกลุ่มกิจการ",
    loading: "กำลังโหลดกลุ่มกิจการ",
    logout: "ออกจากระบบ",
    requestFailed: "เรียกข้อมูลไม่สำเร็จ",
    save: "บันทึก",
    saving: "กำลังบันทึก",
    searchPlaceholder: "ค้นหากลุ่มกิจการ",
    selectFailed: "เลือกกลุ่มกิจการไม่สำเร็จ",
    selected: "เลือกกลุ่มกิจการแล้ว กำลังเปิดหน้าเลือกบริษัท",
    signedInAs: "เข้าสู่ระบบเป็น",
    title: "เลือกกลุ่มกิจการ",
    updateFailed: "บันทึกชื่อกลุ่มกิจการไม่สำเร็จ",
    updateSuccess: "บันทึกชื่อกลุ่มกิจการแล้ว",
  },
  en: holdingTextEn,
  cn: {
    addHolding: "新增集团",
    available: "可访问的集团",
    authRequired: "请选择集团前先登录。",
    cancel: "取消",
    choose: "选择",
    confirmCode: "确认码",
    confirmCodeHelp: "输入这里显示的 4 位数字，以确认真的要创建集团。",
    create: "创建集团",
    createFailed: "无法创建集团。",
    createHoldingDescription: "使用唯一的小写集团代码。创建后不能更改。",
    createHoldingTitle: "创建集团",
    createSuccess: "集团已创建。正在打开公司选择。",
    creating: "正在创建集团",
    description: "选择要使用的集团，然后进入公司选择。",
    duplicateHolding: "此集团代码已存在于你的集团列表中。",
    edit: "编辑",
    emailLoginRequired: "请先使用邮箱验证登录后再创建集团。",
    emptyDescription: "此账号尚无集团权限。请在 Google 登录后创建或分配集团。",
    emptyTitle: "未找到集团",
    eyebrow: "第 2 步",
    holdingCode: "集团代码",
    holdingCodeInvalid: "集团代码只能使用 a-z 和 0-9，长度 3-30，并以 a-z 开头。请勿使用 _ 或符号。",
    holdingCodePlaceholder: "bcdemo01",
    holdingName: "集团名称",
    holdingNamePlaceholder: "示例集团",
    holdingNameRequired: "请输入集团名称。",
    loading: "正在加载集团",
    logout: "退出登录",
    requestFailed: "请求失败。",
    save: "保存",
    saving: "正在保存",
    searchPlaceholder: "搜索集团",
    selectFailed: "无法选择集团。",
    selected: "已选择集团，正在打开公司选择。",
    signedInAs: "登录身份",
    title: "选择集团",
    updateFailed: "无法更新集团。",
    updateSuccess: "集团名称已更新。",
  },
  ja: {
    addHolding: "企業グループを追加",
    available: "利用可能な企業グループ",
    authRequired: "企業グループを選択する前にログインしてください。",
    cancel: "キャンセル",
    choose: "選択",
    confirmCode: "確認コード",
    confirmCodeHelp: "ここに表示された4桁のコードを入力して、企業グループ作成を確認してください。",
    create: "企業グループを作成",
    createFailed: "企業グループを作成できません。",
    createHoldingDescription: "一意の小文字 企業グループコード を使用してください。作成後は変更できません。",
    createHoldingTitle: "企業グループ作成",
    createSuccess: "企業グループを作成しました。会社選択を開きます。",
    creating: "企業グループを作成中",
    description: "利用する企業グループを選択して、会社選択へ進みます。",
    duplicateHolding: "この 企業グループコード は 企業グループ一覧に既に存在します。",
    edit: "編集",
    emailLoginRequired: "企業グループを作成する前にメール認証でログインしてください。",
    emptyDescription: "このアカウントには 企業グループ権限がありません。Google ログイン後に作成または割り当ててください。",
    emptyTitle: "企業グループがありません",
    eyebrow: "ステップ 2",
    holdingCode: "企業グループコード",
    holdingCodeInvalid: "企業グループコードは a-z と 0-9 のみ、3-30文字、先頭は a-z です。_ や記号は使えません。",
    holdingCodePlaceholder: "bcdemo01",
    holdingName: "企業グループ名",
    holdingNamePlaceholder: "サンプル企業グループ",
    holdingNameRequired: "企業グループ名を入力してください。",
    loading: "企業グループを読み込み中",
    logout: "ログアウト",
    requestFailed: "リクエストに失敗しました。",
    save: "保存",
    saving: "保存中",
    searchPlaceholder: "企業グループを検索",
    selectFailed: "企業グループを選択できません。",
    selected: "企業グループを選択しました。会社選択を開きます。",
    signedInAs: "ログイン中",
    title: "企業グループを選択",
    updateFailed: "企業グループを更新できません。",
    updateSuccess: "企業グループ名を更新しました。",
  },
  ko: {
    addHolding: "사업 그룹 추가",
    available: "사용 가능한 사업 그룹",
    authRequired: "사업 그룹을 선택하기 전에 로그인하세요.",
    cancel: "취소",
    choose: "선택",
    confirmCode: "확인 코드",
    confirmCodeHelp: "표시된 4자리 코드를 입력해 사업 그룹 생성을 확인하세요.",
    create: "사업 그룹 생성",
    createFailed: "사업 그룹을 생성할 수 없습니다.",
    createHoldingDescription: "중복 없는 소문자 사업 그룹 코드를 사용하세요. 생성 후 변경할 수 없습니다.",
    createHoldingTitle: "사업 그룹 생성",
    createSuccess: "사업 그룹을 생성했습니다. 회사 선택을 엽니다.",
    creating: "사업 그룹 생성 중",
    description: "사용할 사업 그룹을 선택한 뒤 회사 선택으로 이동합니다.",
    duplicateHolding: "이 사업 그룹 코드는 사업 그룹 목록에 이미 있습니다.",
    edit: "수정",
    emailLoginRequired: "사업 그룹을 생성하려면 이메일 인증으로 로그인하세요.",
    emptyDescription: "이 계정에는 사업 그룹 권한이 없습니다. Google 로그인 후 생성하거나 배정하세요.",
    emptyTitle: "사업 그룹 없음",
    eyebrow: "2단계",
    holdingCode: "사업 그룹 코드",
    holdingCodeInvalid: "사업 그룹 코드는 a-z와 0-9만 사용하고 3-30자이며 a-z로 시작해야 합니다. _ 또는 기호는 사용할 수 없습니다.",
    holdingCodePlaceholder: "bcdemo01",
    holdingName: "사업 그룹 이름",
    holdingNamePlaceholder: "예시 사업 그룹",
    holdingNameRequired: "사업 그룹 이름을 입력하세요.",
    loading: "사업 그룹 불러오는 중",
    logout: "로그아웃",
    requestFailed: "요청 실패",
    save: "저장",
    saving: "저장 중",
    searchPlaceholder: "사업 그룹 검색",
    selectFailed: "사업 그룹을 선택할 수 없습니다.",
    selected: "사업 그룹을 선택했습니다. 회사 선택을 엽니다.",
    signedInAs: "로그인 계정",
    title: "사업 그룹 선택",
    updateFailed: "사업 그룹을 업데이트할 수 없습니다.",
    updateSuccess: "사업 그룹 이름을 업데이트했습니다.",
  },
  lo: {
    addHolding: "ເພີ່ມກຸ່ມທຸລະກິດ",
    available: "ກຸ່ມທຸລະກິດທີ່ເຂົ້າໄດ້",
    authRequired: "ກະລຸນາ login ກ່ອນເລືອກກຸ່ມທຸລະກິດ.",
    cancel: "ຍົກເລີກ",
    choose: "ເລືອກ",
    confirmCode: "ລະຫັດຢືນຢັນ",
    confirmCodeHelp: "ພິມລະຫັດ 4 ຕົວທີ່ສະແດງເພື່ອຢືນຢັນການສ້າງກຸ່ມທຸລະກິດ.",
    create: "ສ້າງກຸ່ມທຸລະກິດ",
    createFailed: "ສ້າງກຸ່ມທຸລະກິດບໍ່ສຳເລັດ.",
    createHoldingDescription: "ໃຊ້ລະຫັດກຸ່ມທຸລະກິດຕົວພິມນ້ອຍທີ່ບໍ່ຊ້ຳ. ສ້າງແລ້ວປ່ຽນບໍ່ໄດ້.",
    createHoldingTitle: "ສ້າງກຸ່ມທຸລະກິດ",
    createSuccess: "ສ້າງກຸ່ມທຸລະກິດແລ້ວ ກຳລັງເປີດໜ້າເລືອກບໍ່ລິສັດ.",
    creating: "ກຳລັງສ້າງກຸ່ມທຸລະກິດ",
    description: "ເລືອກກຸ່ມທຸລະກິດທີ່ຈະໃຊ້ ແລ້ວໄປໜ້າເລືອກບໍ່ລິສັດ.",
    duplicateHolding: "ລະຫັດກຸ່ມທຸລະກິດນີ້ມີຢູ່ໃນລາຍການກຸ່ມທຸລະກິດແລ້ວ.",
    edit: "ແກ້ໄຂ",
    emailLoginRequired: "ຕ້ອງ login ດ້ວຍການຢືນຢັນອີເມວກ່ອນສ້າງກຸ່ມທຸລະກິດ.",
    emptyDescription: "ບັນຊີນີ້ຍັງບໍ່ມີສິດກຸ່ມທຸລະກິດ. ສ້າງ ຫຼື ກຳນົດກຸ່ມທຸລະກິດຫຼັງ login ດ້ວຍ Google.",
    emptyTitle: "ບໍ່ພົບກຸ່ມທຸລະກິດ",
    eyebrow: "ຂັ້ນຕອນ 2",
    holdingCode: "ລະຫັດກຸ່ມທຸລະກິດ",
    holdingCodeInvalid: "ລະຫັດກຸ່ມທຸລະກິດໃຊ້ໄດ້ a-z ແລະ 0-9 ເທົ່ານັ້ນ ຍາວ 3-30 ຕົວ ແລະຂຶ້ນຕົ້ນດ້ວຍ a-z. ຫ້າມໃຊ້ _ ຫຼືສັນຍາລັກ.",
    holdingCodePlaceholder: "bcdemo01",
    holdingName: "ຊື່ກຸ່ມທຸລະກິດ",
    holdingNamePlaceholder: "ກຸ່ມທຸລະກິດຕົວຢ່າງ",
    holdingNameRequired: "ກະລຸນາໃສ່ຊື່ກຸ່ມທຸລະກິດ.",
    loading: "ກຳລັງໂຫຼດກຸ່ມທຸລະກິດ",
    logout: "ອອກຈາກລະບົບ",
    requestFailed: "ຄຳຂໍລົ້ມເຫຼວ.",
    save: "ບັນທຶກ",
    saving: "ກຳລັງບັນທຶກ",
    searchPlaceholder: "ຄົນຫາກຸ່ມທຸລະກິດ",
    selectFailed: "ເລືອກກຸ່ມທຸລະກິດບໍ່ສຳເລັດ.",
    selected: "ເລືອກກຸ່ມທຸລະກິດແລ້ວ ກຳລັງເປີດໜ້າເລືອກບໍ່ລິສັດ.",
    signedInAs: "login ເປັນ",
    title: "ເລືອກກຸ່ມທຸລະກິດ",
    updateFailed: "ອັບເດດກຸ່ມທຸລະກິດບໍ່ສຳເລັດ.",
    updateSuccess: "ອັບເດດຊື່ກຸ່ມທຸລະກິດແລ້ວ.",
  },
  my: {
    addHolding: "စီးပွားရေးလုပ်ငန်းအုပ်စု အသစ်ထည့်ရန်",
    available: "အသုံးပြုနိုင်သော စီးပွားရေးလုပ်ငန်းအုပ်စု",
    authRequired: "စီးပွားရေးလုပ်ငန်းအုပ်စု မရွေးခင် login ဝင်ပါ။",
    cancel: "မလုပ်တော့ပါ",
    choose: "ရွေးပါ",
    confirmCode: "အတည်ပြုကုဒ်",
    confirmCodeHelp: "စီးပွားရေးလုပ်ငန်းအုပ်စု ဖန်တီးရန် သေချာကြောင်း အတည်ပြုရန် ပြထားသော 4 လုံးကုဒ်ကို ရိုက်ပါ။",
    create: "စီးပွားရေးလုပ်ငန်းအုပ်စု ဖန်တီးရန်",
    createFailed: "စီးပွားရေးလုပ်ငန်းအုပ်စု ဖန်တီး၍ မရပါ။",
    createHoldingDescription: "မတူညီသော lowercase စီးပွားရေးလုပ်ငန်းအုပ်စုကုဒ်ကိုသုံးပါ။ ဖန်တီးပြီးနောက် ပြောင်း၍မရပါ။",
    createHoldingTitle: "စီးပွားရေးလုပ်ငန်းအုပ်စု ဖန်တီးရန်",
    createSuccess: "စီးပွားရေးလုပ်ငန်းအုပ်စု ဖန်တီးပြီးပါပြီ။ ကုမ္ပဏီရွေးချယ်မှုကို ဖွင့်နေသည်။",
    creating: "စီးပွားရေးလုပ်ငန်းအုပ်စု ဖန်တီးနေသည်",
    description: "အသုံးပြုမည့် စီးပွားရေးလုပ်ငန်းအုပ်စုကိုရွေးပြီး ကုမ္ပဏီရွေးချယ်မှုသို့ ဆက်သွားပါ။",
    duplicateHolding: "ဤ စီးပွားရေးလုပ်ငန်းအုပ်စုကုဒ် သည် သင့် စီးပွားရေးလုပ်ငန်းအုပ်စု စာရင်းတွင် ရှိပြီးသားဖြစ်သည်။",
    edit: "ပြင်ရန်",
    emailLoginRequired: "စီးပွားရေးလုပ်ငန်းအုပ်စု မဖန်တီးမီ အီးမေးလ်အတည်ပြု login ဖြင့် ဝင်ပါ။",
    emptyDescription: "ဤအကောင့်တွင် စီးပွားရေးလုပ်ငန်းအုပ်စု အသုံးပြုခွင့် မရှိသေးပါ။ Google login ပြီးနောက် ဖန်တီး/သတ်မှတ်ပါ။",
    emptyTitle: "စီးပွားရေးလုပ်ငန်းအုပ်စု မတွေ့ပါ",
    eyebrow: "အဆင့် 2",
    holdingCode: "စီးပွားရေးလုပ်ငန်းအုပ်စုကုဒ်",
    holdingCodeInvalid: "စီးပွားရေးလုပ်ငန်းအုပ်စုကုဒ် သည် a-z နှင့် 0-9 ကိုသာသုံးပြီး 3-30 လုံးရှိရမည်၊ a-z ဖြင့်စတင်ရမည်။ _ သို့မဟုတ် သင်္ကေတ မသုံးရပါ။",
    holdingCodePlaceholder: "bcdemo01",
    holdingName: "စီးပွားရေးလုပ်ငန်းအုပ်စုအမည်",
    holdingNamePlaceholder: "ဥပမာ စီးပွားရေးလုပ်ငန်းအုပ်စု",
    holdingNameRequired: "စီးပွားရေးလုပ်ငန်းအုပ်စုအမည် ထည့်ပါ။",
    loading: "စီးပွားရေးလုပ်ငန်းအုပ်စုများကို ဖွင့်နေသည်",
    logout: "ထွက်မည်",
    requestFailed: "တောင်းဆိုမှု မအောင်မြင်ပါ။",
    save: "သိမ်းမည်",
    saving: "သိမ်းနေသည်",
    searchPlaceholder: "စီးပွားရေးလုပ်ငန်းအုပ်စု ရှာရန်",
    selectFailed: "စီးပွားရေးလုပ်ငန်းအုပ်စု ရွေးမရပါ။",
    selected: "စီးပွားရေးလုပ်ငန်းအုပ်စု ရွေးပြီးပါပြီ။ ကုမ္ပဏီရွေးချယ်မှုကို ဖွင့်နေသည်။",
    signedInAs: "ဝင်ထားသောအကောင့်",
    title: "စီးပွားရေးလုပ်ငန်းအုပ်စု ရွေးချယ်ပါ",
    updateFailed: "စီးပွားရေးလုပ်ငန်းအုပ်စုကို update မလုပ်နိုင်ပါ။",
    updateSuccess: "စီးပွားရေးလုပ်ငန်းအုပ်စုအမည် update ပြီးပါပြီ။",
  },
  km: {
    addHolding: "បន្ថែមក្រុមអាជីវកម្ម",
    available: "ក្រុមអាជីវកម្មដែលអាចចូលបាន",
    authRequired: "សូម login មុនពេលជ្រើសក្រុមអាជីវកម្ម។",
    cancel: "បោះបង់",
    choose: "ជ្រើស",
    confirmCode: "លេខកូដបញ្ជាក់",
    confirmCodeHelp: "វាយលេខកូដ 4 ខ្ទង់ដែលបង្ហាញនៅទីនេះ ដើម្បីបញ្ជាក់ការបង្កើតក្រុមអាជីវកម្ម។",
    create: "បង្កើតក្រុមអាជីវកម្ម",
    createFailed: "មិនអាចបង្កើតក្រុមអាជីវកម្មបាន។",
    createHoldingDescription: "ប្រើលេខកូដក្រុមអាជីវកម្មអក្សរតូចដែលមិនស្ទួន។ បង្កើតរួចមិនអាចកែបាន។",
    createHoldingTitle: "បង្កើតក្រុមអាជីវកម្ម",
    createSuccess: "បានបង្កើតក្រុមអាជីវកម្ម។ កំពុងបើកជម្រើសក្រុមហ៊ុន។",
    creating: "កំពុងបង្កើតក្រុមអាជីវកម្ម",
    description: "ជ្រើសក្រុមអាជីវកម្មដែលត្រូវប្រើ បន្ទាប់មកបន្តទៅជ្រើសក្រុមហ៊ុន។",
    duplicateHolding: "លេខកូដក្រុមអាជីវកម្មនេះមានរួចហើយក្នុងបញ្ជីក្រុមអាជីវកម្មរបស់អ្នក។",
    edit: "កែសម្រួល",
    emailLoginRequired: "សូម login ដោយផ្ទៀងផ្ទាត់អ៊ីមែល មុនពេលបង្កើតក្រុមអាជីវកម្ម។",
    emptyDescription: "គណនីនេះមិនទាន់មានសិទ្ធិក្រុមអាជីវកម្មទេ។ សូមបង្កើត ឬកំណត់បន្ទាប់ពី login ដោយ Google។",
    emptyTitle: "រកមិនឃើញក្រុមអាជីវកម្ម",
    eyebrow: "ជំហាន 2",
    holdingCode: "លេខកូដក្រុមអាជីវកម្ម",
    holdingCodeInvalid: "លេខកូដក្រុមអាជីវកម្មត្រូវប្រើតែ a-z និង 0-9 ប្រវែង 3-30 តួ និងចាប់ផ្តើមដោយ a-z។ ហាមប្រើ _ ឬនិមិត្តសញ្ញា។",
    holdingCodePlaceholder: "bcdemo01",
    holdingName: "ឈ្មោះក្រុមអាជីវកម្ម",
    holdingNamePlaceholder: "ក្រុមអាជីវកម្មគំរូ",
    holdingNameRequired: "សូមបញ្ចូលឈ្មោះក្រុមអាជីវកម្ម។",
    loading: "កំពុងផ្ទុកក្រុមអាជីវកម្ម",
    logout: "ចេញពីប្រព័ន្ធ",
    requestFailed: "សំណើបរាជ័យ។",
    save: "រក្សាទុក",
    saving: "កំពុងរក្សាទុក",
    searchPlaceholder: "ស្វែងរកក្រុមអាជីវកម្ម",
    selectFailed: "មិនអាចជ្រើសក្រុមអាជីវកម្មបាន។",
    selected: "បានជ្រើសក្រុមអាជីវកម្ម កំពុងបើកជម្រើសក្រុមហ៊ុន។",
    signedInAs: "ចូលប្រើជា",
    title: "ជ្រើសក្រុមអាជីវកម្ម",
    updateFailed: "មិនអាចធ្វើបច្ចុប្បន្នភាពក្រុមអាជីវកម្មបាន។",
    updateSuccess: "បានធ្វើបច្ចុប្បន្នភាពឈ្មោះក្រុមអាជីវកម្ម។",
  },
  vi: {
    addHolding: "Thêm nhóm doanh nghiệp",
    available: "Nhóm doanh nghiệp có thể truy cập",
    authRequired: "Vui lòng đăng nhập trước khi chọn nhóm doanh nghiệp.",
    cancel: "Hủy",
    choose: "Chọn",
    confirmCode: "Mã xác nhận",
    confirmCodeHelp: "Nhập mã 4 chữ số hiển thị ở đây để xác nhận tạo nhóm doanh nghiệp.",
    create: "Tạo nhóm doanh nghiệp",
    createFailed: "Không thể tạo nhóm doanh nghiệp.",
    createHoldingDescription: "Dùng mã nhóm doanh nghiệp chữ thường và không trùng. Mã này không thể đổi sau khi tạo.",
    createHoldingTitle: "Tạo nhóm doanh nghiệp",
    createSuccess: "Đã tạo nhóm doanh nghiệp. Đang mở trang chọn công ty.",
    creating: "Đang tạo nhóm doanh nghiệp",
    description: "Chọn nhóm doanh nghiệp cần làm việc, sau đó chuyển sang chọn công ty.",
    duplicateHolding: "Mã nhóm doanh nghiệp này đã có trong danh sách nhóm doanh nghiệp của bạn.",
    edit: "Sửa",
    emailLoginRequired: "Hãy đăng nhập bằng xác thực email trước khi tạo nhóm doanh nghiệp.",
    emptyDescription: "Tài khoản này chưa có quyền nhóm doanh nghiệp. Hãy tạo hoặc gán nhóm doanh nghiệp sau khi đăng nhập Google.",
    emptyTitle: "Không có nhóm doanh nghiệp",
    eyebrow: "Bước 2",
    holdingCode: "mã nhóm doanh nghiệp",
    holdingCodeInvalid: "Mã nhóm doanh nghiệp chỉ dùng a-z và 0-9, dài 3-30 ký tự và bắt đầu bằng a-z. Không dùng _ hoặc ký hiệu.",
    holdingCodePlaceholder: "bcdemo01",
    holdingName: "Tên nhóm doanh nghiệp",
    holdingNamePlaceholder: "Nhóm doanh nghiệp mẫu",
    holdingNameRequired: "Vui lòng nhập tên nhóm doanh nghiệp.",
    loading: "Đang tải nhóm doanh nghiệp",
    logout: "Đăng xuất",
    requestFailed: "Yêu cầu thất bại.",
    save: "Lưu",
    saving: "Đang lưu",
    searchPlaceholder: "Tìm nhóm doanh nghiệp",
    selectFailed: "Không thể chọn nhóm doanh nghiệp.",
    selected: "Đã chọn nhóm doanh nghiệp. Đang mở trang chọn công ty.",
    signedInAs: "Đăng nhập bằng",
    title: "Chọn nhóm doanh nghiệp",
    updateFailed: "Không thể cập nhật nhóm doanh nghiệp.",
    updateSuccess: "Đã cập nhật tên nhóm doanh nghiệp.",
  },
  ms: {
    addHolding: "Tambah kumpulan perniagaan",
    available: "Kumpulan perniagaan yang boleh diakses",
    authRequired: "Sila log masuk sebelum memilih kumpulan perniagaan.",
    cancel: "Batal",
    choose: "Pilih",
    confirmCode: "Kod pengesahan",
    confirmCodeHelp: "Taip kod 4 digit yang dipaparkan di sini untuk mengesahkan ciptaan kumpulan perniagaan.",
    create: "Cipta kumpulan perniagaan",
    createFailed: "Tidak dapat mencipta kumpulan perniagaan.",
    createHoldingDescription: "Gunakan kod kumpulan perniagaan huruf kecil yang unik. Kod ini tidak boleh diubah selepas dicipta.",
    createHoldingTitle: "Cipta kumpulan perniagaan",
    createSuccess: "Kumpulan perniagaan dicipta. Membuka pilihan syarikat.",
    creating: "Mencipta kumpulan perniagaan",
    description: "Pilih kumpulan perniagaan untuk digunakan, kemudian teruskan ke pilihan syarikat.",
    duplicateHolding: "Kod kumpulan perniagaan ini sudah wujud dalam senarai kumpulan perniagaan anda.",
    edit: "Edit",
    emailLoginRequired: "Log masuk dengan pengesahan e-mel sebelum mencipta kumpulan perniagaan.",
    emptyDescription: "Akaun ini belum mempunyai akses kumpulan perniagaan. Cipta atau tetapkan kumpulan perniagaan selepas log masuk Google.",
    emptyTitle: "Tiada kumpulan perniagaan",
    eyebrow: "Langkah 2",
    holdingCode: "kod kumpulan perniagaan",
    holdingCodeInvalid: "Kod kumpulan perniagaan mesti menggunakan a-z dan 0-9 sahaja, 3-30 aksara dan bermula dengan a-z. Jangan guna _ atau simbol.",
    holdingCodePlaceholder: "bcdemo01",
    holdingName: "Nama kumpulan perniagaan",
    holdingNamePlaceholder: "Kumpulan perniagaan contoh",
    holdingNameRequired: "Sila masukkan nama kumpulan perniagaan.",
    loading: "Memuatkan kumpulan perniagaan",
    logout: "Log keluar",
    requestFailed: "Permintaan gagal.",
    save: "Simpan",
    saving: "Menyimpan",
    searchPlaceholder: "Cari kumpulan perniagaan",
    selectFailed: "Tidak dapat memilih kumpulan perniagaan.",
    selected: "Kumpulan perniagaan dipilih. Membuka pilihan syarikat.",
    signedInAs: "Log masuk sebagai",
    title: "Pilih kumpulan perniagaan",
    updateFailed: "Tidak dapat mengemas kini kumpulan perniagaan.",
    updateSuccess: "Nama kumpulan perniagaan dikemas kini.",
  },
  id: {
    addHolding: "Tambah grup bisnis",
    available: "Grup bisnis yang dapat diakses",
    authRequired: "Silakan masuk sebelum memilih grup bisnis.",
    cancel: "Batal",
    choose: "Pilih",
    confirmCode: "Kode konfirmasi",
    confirmCodeHelp: "Ketik kode 4 digit yang ditampilkan di sini untuk mengonfirmasi pembuatan grup bisnis.",
    create: "Buat grup bisnis",
    createFailed: "Tidak dapat membuat grup bisnis.",
    createHoldingDescription: "Gunakan kode grup bisnis huruf kecil yang unik. Kode ini tidak dapat diubah setelah dibuat.",
    createHoldingTitle: "Buat grup bisnis",
    createSuccess: "Grup bisnis dibuat. Membuka pemilihan perusahaan.",
    creating: "Membuat grup bisnis",
    description: "Pilih grup bisnis yang akan digunakan, lalu lanjut ke pemilihan perusahaan.",
    duplicateHolding: "Kode grup bisnis ini sudah ada di daftar grup bisnis Anda.",
    edit: "Edit",
    emailLoginRequired: "Masuk dengan verifikasi email sebelum membuat grup bisnis.",
    emptyDescription: "Akun ini belum memiliki akses grup bisnis. Buat atau tetapkan grup bisnis setelah masuk Google.",
    emptyTitle: "Grup bisnis tidak ditemukan",
    eyebrow: "Langkah 2",
    holdingCode: "kode grup bisnis",
    holdingCodeInvalid: "Kode grup bisnis harus memakai a-z dan 0-9 saja, 3-30 karakter dan diawali a-z. Jangan memakai _ atau simbol.",
    holdingCodePlaceholder: "bcdemo01",
    holdingName: "Nama grup bisnis",
    holdingNamePlaceholder: "Grup bisnis contoh",
    holdingNameRequired: "Masukkan nama grup bisnis.",
    loading: "Memuat grup bisnis",
    logout: "Keluar",
    requestFailed: "Permintaan gagal.",
    save: "Simpan",
    saving: "Menyimpan",
    searchPlaceholder: "Cari grup bisnis",
    selectFailed: "Tidak dapat memilih grup bisnis.",
    selected: "Grup bisnis dipilih. Membuka pemilihan perusahaan.",
    signedInAs: "Masuk sebagai",
    title: "Pilih grup bisnis",
    updateFailed: "Tidak dapat memperbarui grup bisnis.",
    updateSuccess: "Nama grup bisnis diperbarui.",
  },
  fil: {
    addHolding: "Magdagdag ng grupo ng negosyo",
    available: "Mga grupo ng negosyo na puwedeng ma-access",
    authRequired: "Mag-login muna bago pumili ng grupo ng negosyo.",
    cancel: "Kanselahin",
    choose: "Piliin",
    confirmCode: "Confirmation code",
    confirmCodeHelp: "I-type ang 4-digit code na ipinapakita dito para kumpirmahin ang paggawa ng grupo ng negosyo.",
    create: "Gumawa ng grupo ng negosyo",
    createFailed: "Hindi magawa ang grupo ng negosyo.",
    createHoldingDescription: "Gumamit ng unique lowercase na kodigo ng grupo ng negosyo. Hindi na ito mababago pagkatapos malikha.",
    createHoldingTitle: "Gumawa ng grupo ng negosyo",
    createSuccess: "Nagawa ang grupo ng negosyo. Binubuksan ang pagpili ng kumpanya.",
    creating: "Ginagawa ang grupo ng negosyo",
    description: "Piliin ang grupo ng negosyo na gagamitin, pagkatapos ay magpatuloy sa pagpili ng kumpanya.",
    duplicateHolding: "Mayroon na ang kodigo ng grupo ng negosyo na ito sa iyong listahan ng grupo ng negosyo.",
    edit: "I-edit",
    emailLoginRequired: "Mag-login gamit ang email authentication bago gumawa ng grupo ng negosyo.",
    emptyDescription: "Wala pang access sa grupo ng negosyo ang account na ito. Gumawa o magtalaga ng grupo ng negosyo pagkatapos ng Google login.",
    emptyTitle: "Walang grupo ng negosyo",
    eyebrow: "Hakbang 2",
    holdingCode: "kodigo ng grupo ng negosyo",
    holdingCodeInvalid: "Ang kodigo ng grupo ng negosyo ay dapat gumamit lang ng a-z at 0-9, 3-30 character, at magsimula sa a-z. Huwag gumamit ng _ o simbolo.",
    holdingCodePlaceholder: "bcdemo01",
    holdingName: "Pangalan ng grupo ng negosyo",
    holdingNamePlaceholder: "Halimbawang grupo ng negosyo",
    holdingNameRequired: "Ilagay ang pangalan ng grupo ng negosyo.",
    loading: "Nilo-load ang grupo ng negosyo",
    logout: "Mag-log out",
    requestFailed: "Nabigo ang request.",
    save: "I-save",
    saving: "Sine-save",
    searchPlaceholder: "Hanapin ang grupo ng negosyo",
    selectFailed: "Hindi mapili ang grupo ng negosyo.",
    selected: "Napili ang grupo ng negosyo. Binubuksan ang pagpili ng kumpanya.",
    signedInAs: "Naka-login bilang",
    title: "Piliin ang grupo ng negosyo",
    updateFailed: "Hindi ma-update ang grupo ng negosyo.",
    updateSuccess: "Na-update ang pangalan ng grupo ng negosyo.",
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
    if (!holdingCode || !isValidHoldingCode(holdingCode)) {
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
    if (!holdingCode || !isValidHoldingCode(holdingCode)) {
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
            <h1>{ht(language, "holdingName")}</h1>
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
              <AppHeaderControls language={language} onLanguageChange={setLanguage} />
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
      languageconfigs: [{ code: "th", codetranslator: "th", name: "ภาษาไทย", isuse: true, isdefault: true }],
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
