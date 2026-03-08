const fs = require('fs');

// Read existing language.json
const langData = JSON.parse(fs.readFileSync('assets/language.json', 'utf8'));

// Define new translation keys needed
const newTranslations = {
  'add_payment_entry': {
    th: 'เพิ่มรายการรับชำระ',
    lo: 'ເພີ່ມລາຍການຮັບຊໍາລະ',
    en: 'Add Payment Entry',
    cn: '添加付款条目',
    zh: '添加付款條目',
    ja: '支払い入力の追加',
    ko: '지불 항목 추가',
    my: 'ပေးချေမှုအသစ်ထည့်ရန်',
    km: 'បន្ថែមការទូទាត់',
    vi: 'Thêm mục thanh toán'
  },
  'sub_pattern_name': {
    th: 'ชื่อรูปแบบย่อย',
    lo: 'ຊື່ຮູບແບບຍ່ອຍ',
    en: 'Sub-Pattern Name',
    cn: '子模式名称',
    zh: '子模式名稱',
    ja: 'サブパターン名',
    ko: '하위 패턴 이름',
    my: 'ခွဲပုံစံအမည်',
    km: 'ឈ្មោះលំនាំរង',
    vi: 'Tên mẫu phụ'
  },
  'add_sub_item': {
    th: 'เพิ่มรายการย่อย',
    lo: 'ເພີ່ມລາຍການຍ່ອຍ',
    en: 'Add Sub Item',
    cn: '添加子项',
    zh: '添加子項',
    ja: 'サブアイテムを追加',
    ko: '하위 항목 추가',
    my: 'ခွဲစာရင်းထည့်ရန်',
    km: 'បន្ថែមធាតុរង',
    vi: 'Thêm mục phụ'
  },
  'import_success': {
    th: 'Import สำเร็จ',
    lo: 'Import ສໍາເລັດ',
    en: 'Import Successful',
    cn: '导入成功',
    zh: '導入成功',
    ja: 'インポート成功',
    ko: '가져오기 성공',
    my: 'Import အောင်မြင်သည်',
    km: 'នាំចូលជោគជ័យ',
    vi: 'Nhập thành công'
  },
  'items': {
    th: 'รายการ',
    lo: 'ລາຍການ',
    en: 'Items',
    cn: '项目',
    zh: '項目',
    ja: '項目',
    ko: '항목',
    my: 'စာရင်းများ',
    km: 'ធាតុ',
    vi: 'Mục'
  },
  'no_data_found_in_excel': {
    th: 'ไม่พบข้อมูลในไฟล์ Excel',
    lo: 'ບໍ່ພົບຂໍ້ມູນໃນໄຟລ໌ Excel',
    en: 'No Data Found in Excel File',
    cn: 'Excel文件中未找到数据',
    zh: 'Excel文件中未找到數據',
    ja: 'Excelファイルにデータが見つかりません',
    ko: 'Excel 파일에 데이터가 없습니다',
    my: 'Excel ဖိုင်တွင် အချက်အလက်မရှိပါ',
    km: 'រកមិនឃើញទិន្នន័យក្នុងឯកសារ Excel',
    vi: 'Không tìm thấy dữ liệu trong file Excel'
  },
  'import_error': {
    th: 'เกิดข้อผิดพลาดในการ Import',
    lo: 'ເກີດຂໍ້ຜິດພາດໃນການ Import',
    en: 'Import Error',
    cn: '导入错误',
    zh: '導入錯誤',
    ja: 'インポートエラー',
    ko: '가져오기 오류',
    my: 'Import အမှားအယွင်း',
    km: 'កំហុសក្នុងការនាំចូល',
    vi: 'Lỗi nhập'
  },
  'no_product_selected': {
    th: 'ยังไม่เลือกสินค้า',
    lo: 'ຍັງບໍ່ໄດ້ເລືອກສິນຄ້າ',
    en: 'No Product Selected',
    cn: '未选择产品',
    zh: '未選擇產品',
    ja: '商品が選択されていません',
    ko: '선택된 제품이 없습니다',
    my: 'ထုတ်ကုန်မရွေးထားပါ',
    km: 'មិនបានជ្រើសផលិតផល',
    vi: 'Chưa chọn sản phẩm'
  },
  'invalid_format': {
    th: 'รูปแบบไม่ถูกต้อง',
    lo: 'ຮູບແບບບໍ່ຖືກຕ້ອງ',
    en: 'Invalid Format',
    cn: '格式无效',
    zh: '格式無效',
    ja: '無効な形式',
    ko: '잘못된 형식',
    my: 'မှားယွင်းသော ဖော်မတ်',
    km: 'ទម្រង់មិនត្រឹមត្រូវ',
    vi: 'Định dạng không hợp lệ'
  },
  'cannot_save': {
    th: 'ไม่สามารถบันทึกได้',
    lo: 'ບໍ່ສາມາດບັນທຶກໄດ້',
    en: 'Cannot Save',
    cn: '无法保存',
    zh: '無法儲存',
    ja: '保存できません',
    ko: '저장할 수 없습니다',
    my: 'သိမ်းဆည်းမရပါ',
    km: 'មិនអាចរក្សាទុកបាន',
    vi: 'Không thể lưu'
  },
  'please_check_time_format': {
    th: 'โปรดตรวจสอบรูปแบบเวลา เช่น เวลาแต่ละช่วงทับซ้อนกัน เวลาเปิดและปิดตรงกัน',
    lo: 'ກະລຸນາກວດສອບຮູບແບບເວລາ ເຊັ່ນ: เวลาแต่ละช่วงทับซ้อนກัน ເວລາເປີດແລະປິດຕົງກັນ',
    en: 'Please check time format, e.g. overlapping time ranges, identical start and end times',
    cn: '请检查时间格式，例如时间范围重叠，开始和结束时间相同',
    zh: '請檢查時間格式，例如時間範圍重疊，開始和結束時間相同',
    ja: '時間形式を確認してください。例：時間範囲の重複、開始時刻と終了時刻が同じ',
    ko: '시간 형식을 확인하세요. 예: 겹치는 시간 범위, 동일한 시작 및 종료 시간',
    my: 'အချိန်ပုံစံကို စစ်ဆေးပါ။ ဥပမာ: အချိန်အပိုင်းအခြားများ ထပ်နေခြင်း၊ စတင်ချိန်နှင့် အဆုံးချိန် တူညီနေခြင်း',
    km: 'សូមពិនិត្យទម្រង់ពេលវេលា ឧទាហរណ៍៖ ពេលវេលាជាន់គ្នា ពេលវេលាបើកនិងបិទដូចគ្នា',
    vi: 'Vui lòng kiểm tra định dạng thời gian, ví dụ: khoảng thời gian trùng lặp, thời gian bắt đầu và kết thúc giống nhau'
  },
  'ok': {
    th: 'OK',
    lo: 'ຕົກລົງ',
    en: 'OK',
    cn: '确定',
    zh: '確定',
    ja: 'OK',
    ko: '확인',
    my: 'OK',
    km: 'យល់ព្រម',
    vi: 'Đồng ý'
  },
  'no_valid_data_in_excel': {
    th: 'ไม่พบข้อมูลที่ถูกต้องในไฟล์ Excel',
    lo: 'ບໍ່ພົບຂໍ້ມູນທີ່ຖືກຕ້ອງໃນໄຟລ໌ Excel',
    en: 'No Valid Data Found in Excel File',
    cn: 'Excel文件中未找到有效数据',
    zh: 'Excel文件中未找到有效數據',
    ja: 'Excelファイルに有効なデータが見つかりません',
    ko: 'Excel 파일에 유효한 데이터가 없습니다',
    my: 'Excel ဖိုင်တွင် မှန်ကန်သောအချက်အလက်မရှိပါ',
    km: 'រកមិនឃើញទិន្នន័យត្រឹមត្រូវក្នុងឯកសារ Excel',
    vi: 'Không tìm thấy dữ liệu hợp lệ trong file Excel'
  },
  'daily_data': {
    th: 'ข้อมูลรายวัน',
    lo: 'ຂໍ້ມູນລາຍວັນ',
    en: 'Daily Data',
    cn: '每日数据',
    zh: '每日數據',
    ja: '日次データ',
    ko: '일일 데이터',
    my: 'နေ့စဉ်အချက်အလက်',
    km: 'ទិន្នន័យប្រចាំថ្ងៃ',
    vi: 'Dữ liệu hàng ngày'
  },
  'shopping_cart_system': {
    th: 'ระบบตะกร้าสินค้า',
    lo: 'ລະບົບກະຕ່າສິນຄ້າ',
    en: 'Shopping Cart System',
    cn: '购物车系统',
    zh: '購物車系統',
    ja: 'ショッピングカートシステム',
    ko: '장바구니 시스템',
    my: 'စျေးဝယ်ခြင်းစနစ်',
    km: 'ប្រព័ន្ធរទេះទិញទំនិញ',
    vi: 'Hệ thống giỏ hàng'
  },
  'refresh': {
    th: 'รีเฟรช',
    lo: 'ໂຫຼດໃຫມ່',
    en: 'Refresh',
    cn: '刷新',
    zh: '刷新',
    ja: 'リフレッシュ',
    ko: '새로고침',
    my: 'ပြန်ဖွင့်ရန်',
    km: 'ផ្ទុកឡើងវិញ',
    vi: 'Làm mới'
  },
  'doc_type': {
    th: 'ประเภทเอกสาร',
    lo: 'ປະເພດເອກະສານ',
    en: 'Document Type',
    cn: '文档类型',
    zh: '文檔類型',
    ja: '文書タイプ',
    ko: '문서 유형',
    my: 'စာရွက်စာတမ်းအမျိုးအစား',
    km: 'ប្រភេទឯកសារ',
    vi: 'Loại tài liệu'
  },
  'quantity': {
    th: 'จำนวน',
    lo: 'ຈໍານວນ',
    en: 'Quantity',
    cn: '数量',
    zh: '數量',
    ja: '数量',
    ko: '수량',
    my: 'အရေအတွက်',
    km: 'បរិមាណ',
    vi: 'Số lượng'
  },
  'latest_number': {
    th: 'เลขที่ล่าสุด',
    lo: 'ເລກທີ່ຫຼ້າສຸດ',
    en: 'Latest Number',
    cn: '最新编号',
    zh: '最新編號',
    ja: '最新番号',
    ko: '최신 번호',
    my: 'နောက်ဆုံးနံပါတ်',
    km: 'លេខចុងក្រោយបំផុត',
    vi: 'Số mới nhất'
  },
  'latest_date': {
    th: 'วันที่ล่าสุด',
    lo: 'ວັນທີຫຼ້າສຸດ',
    en: 'Latest Date',
    cn: '最新日期',
    zh: '最新日期',
    ja: '最新日付',
    ko: '최신 날짜',
    my: 'နောက်ဆုံးရက်စွဲ',
    km: 'កាលបរិច្ឆេទចុងក្រោយបំផុត',
    vi: 'Ngày mới nhất'
  },
  'document_value': {
    th: 'มูลค่าเอกสาร',
    lo: 'ມູນຄ່າເອກະສານ',
    en: 'Document Value',
    cn: '文档价值',
    zh: '文檔價值',
    ja: '文書金額',
    ko: '문서 금액',
    my: 'စာရွက်စာတမ်းတန်ဖိုး',
    km: 'តម្លៃឯកសារ',
    vi: 'Giá trị tài liệu'
  }
};

// Add new entries to language data
let addedCount = 0;
for (const [code, translations] of Object.entries(newTranslations)) {
  // Check if key already exists
  const exists = langData.some(entry => entry.code === code);
  if (!exists) {
    const newEntry = {
      code: code,
      langs: []
    };

    for (const [langCode, text] of Object.entries(translations)) {
      newEntry.langs.push({
        code: langCode,
        text: text
      });
    }

    langData.push(newEntry);
    addedCount++;
  }
}

// Save updated language.json
fs.writeFileSync('assets/language.json', JSON.stringify(langData, null, 4), 'utf8');

console.log('Added ' + addedCount + ' new translation entries to language.json');
console.log('Total entries: ' + langData.length);
