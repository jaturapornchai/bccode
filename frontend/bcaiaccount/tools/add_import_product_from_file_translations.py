#!/usr/bin/env python3
# -*- coding: utf-8 -*-
import json
import sys

# Reconfigure stdout for UTF-8
sys.stdout.reconfigure(encoding='utf-8')

# Define all new translations
new_translations = [
    {
        "code": "import_product_file_confirm_import",
        "th": "ยืนยันการนำเข้าข้อมูล",
        "lo": "ຢືນຢັນການນຳເຂົ້າຂໍ້ມູນ",
        "en": "Confirm Import",
        "cn": "确认导入",
        "ja": "インポートを確認",
        "ko": "가져오기 확인",
        "my": "တင်သွင်းမှုအတည်ပြုပါ",
        "km": "បញ្ជាក់ការនាំចូល",
        "vi": "Xác nhận nhập"
    },
    {
        "code": "import_product_file_cancel",
        "th": "ยกเลิก",
        "lo": "ຍົກເລີກ",
        "en": "Cancel",
        "cn": "取消",
        "ja": "キャンセル",
        "ko": "취소",
        "my": "ပယ်ဖျက်",
        "km": "បោះបង់",
        "vi": "Hủy"
    },
    {
        "code": "import_product_file_confirm",
        "th": "ยืนยัน",
        "lo": "ຢືນຢັນ",
        "en": "Confirm",
        "cn": "确认",
        "ja": "確認",
        "ko": "확인",
        "my": "အတည်ပြု",
        "km": "បញ្ជាក់",
        "vi": "Xác nhận"
    },
    {
        "code": "import_product_file_close",
        "th": "ปิด",
        "lo": "ປິດ",
        "en": "Close",
        "cn": "关闭",
        "ja": "閉じる",
        "ko": "닫기",
        "my": "ပိတ်",
        "km": "បិទ",
        "vi": "Đóng"
    },
    {
        "code": "import_product_file_title",
        "th": "นำเข้าข้อมูลสินค้าจากไฟล์",
        "lo": "ນຳເຂົ້າຂໍ້ມູນສິນຄ້າຈາກໄຟລ໌",
        "en": "Import Products from File",
        "cn": "从文件导入产品",
        "ja": "ファイルから製品をインポート",
        "ko": "파일에서 제품 가져오기",
        "my": "ဖိုင်မှထုတ်ကုန်များတင်သွင်းပါ",
        "km": "នាំចូលផលិតផលពីឯកសារ",
        "vi": "Nhập sản phẩm từ tệp"
    },
    {
        "code": "import_product_file_select_file",
        "th": "เลือกไฟล์",
        "lo": "ເລືອກໄຟລ໌",
        "en": "Select File",
        "cn": "选择文件",
        "ja": "ファイルを選択",
        "ko": "파일 선택",
        "my": "ဖိုင်ရွေးချယ်ပါ",
        "km": "ជ្រើសរើសឯកសារ",
        "vi": "Chọn tệp"
    },
    {
        "code": "import_product_file_back_to_main",
        "th": "กลับหน้าหลัก",
        "lo": "ກັບໄປໜ້າຫຼັກ",
        "en": "Back to Main",
        "cn": "返回主页",
        "ja": "メインに戻る",
        "ko": "메인으로 돌아가기",
        "my": "ပင်မစာမျက်နှာသို့ပြန်သွားပါ",
        "km": "ត្រឡប់ទៅទំព័រដើម",
        "vi": "Quay lại trang chính"
    },
    {
        "code": "import_product_file_excel_only",
        "th": "รูปแบบไฟล์ Excel (.xlsx) เท่านั้น",
        "lo": "ໄຟລ໌ Excel (.xlsx) ເທົ່ານັ້ນ",
        "en": "Excel (.xlsx) files only",
        "cn": "仅限Excel（.xlsx）文件",
        "ja": "Excel（.xlsx）ファイルのみ",
        "ko": "Excel(.xlsx) 파일만",
        "my": "Excel (.xlsx) ဖိုင်များသာ",
        "km": "ឯកសារ Excel (.xlsx) តែប៉ុណ្ណោះ",
        "vi": "Chỉ tệp Excel (.xlsx)"
    },
    {
        "code": "import_product_file_max_size",
        "th": "ขนาดไฟล์สูงสุด: 500 MB",
        "lo": "ຂະໜາດໄຟລ໌ສູງສຸດ: 500 MB",
        "en": "Maximum file size: 500 MB",
        "cn": "最大文件大小：500 MB",
        "ja": "最大ファイルサイズ：500 MB",
        "ko": "최대 파일 크기: 500 MB",
        "my": "အများဆုံးဖိုင်အရွယ်အစား: 500 MB",
        "km": "ទំហំឯកសារអតិបរមា: 500 MB",
        "vi": "Kích thước tệp tối đa: 500 MB"
    },
    {
        "code": "import_product_file_latest_file",
        "th": "ไฟล์ล่าสุด",
        "lo": "ໄຟລ໌ລ່າສຸດ",
        "en": "Latest File",
        "cn": "最新文件",
        "ja": "最新ファイル",
        "ko": "최신 파일",
        "my": "နောက်ဆုံးဖိုင်",
        "km": "ឯកសារចុងក្រោយ",
        "vi": "Tệp mới nhất"
    },
    {
        "code": "import_product_file_upload_complete",
        "th": "อัปโหลดเรียบร้อยแล้ว",
        "lo": "ອັບໂຫຼດສຳເລັດແລ້ວ",
        "en": "Upload complete",
        "cn": "上传完成",
        "ja": "アップロード完了",
        "ko": "업로드 완료",
        "my": "အပ်လုဒ်ပြီးပါပြီ",
        "km": "ការបញ្ជូនរួចរាល់",
        "vi": "Tải lên hoàn tất"
    },
    {
        "code": "import_product_file_importing",
        "th": "กำลังนำเข้าข้อมูล...",
        "lo": "ກຳລັງນຳເຂົ້າຂໍ້ມູນ...",
        "en": "Importing data...",
        "cn": "正在导入数据...",
        "ja": "データをインポート中...",
        "ko": "데이터 가져오는 중...",
        "my": "ဒေတာတင်သွင်းနေသည်...",
        "km": "កំពុងនាំចូលទិន្នន័យ...",
        "vi": "Đang nhập dữ liệu..."
    },
    {
        "code": "import_product_file_uploading",
        "th": "กำลังอัปโหลด...",
        "lo": "ກຳລັງອັບໂຫຼດ...",
        "en": "Uploading...",
        "cn": "正在上传...",
        "ja": "アップロード中...",
        "ko": "업로드 중...",
        "my": "အပ်လုဒ်လုပ်နေသည်...",
        "km": "កំពុងបញ្ជូន...",
        "vi": "Đang tải lên..."
    },
    {
        "code": "import_product_file_upload_success",
        "th": "อัปโหลดสำเร็จ!",
        "lo": "อັບໂຫຼດສຳເລັດ!",
        "en": "Upload successful!",
        "cn": "上传成功！",
        "ja": "アップロード成功！",
        "ko": "업로드 성공!",
        "my": "အပ်လုဒ်အောင်မြင်ပါသည်!",
        "km": "ការបញ្ជូនជោគជ័យ!",
        "vi": "Tải lên thành công!"
    },
    {
        "code": "import_product_file_import_product_barcode",
        "th": "นำเข้าสินค้า/Barcode",
        "lo": "ນຳເຂົ້າສິນຄ້າ/Barcode",
        "en": "Import Product/Barcode",
        "cn": "导入产品/条形码",
        "ja": "製品/バーコードをインポート",
        "ko": "제품/바코드 가져오기",
        "my": "ထုတ်ကုန်/ဘားကုဒ်တင်သွင်းပါ",
        "km": "នាំចូលផលិតផល/បារកូដ",
        "vi": "Nhập sản phẩm/Mã vạch"
    },
    {
        "code": "import_product_file_duplicate_barcode_found",
        "th": "พบ Barcode ซ้ำ",
        "lo": "ພົບ Barcode ຊ້ຳກັນ",
        "en": "Duplicate Barcode Found",
        "cn": "发现重复条形码",
        "ja": "重複バーコードが見つかりました",
        "ko": "중복 바코드 발견",
        "my": "ထပ်နေသော Barcode တွေ့ရှိပါသည်",
        "km": "រកឃើញបារកូដស្ទួន",
        "vi": "Tìm thấy mã vạch trùng lặp"
    },
    {
        "code": "import_product_file_duplicate_barcode_detail",
        "th": "รายละเอียด Barcode ซ้ำ:",
        "lo": "ລາຍລະອຽດ Barcode ຊ້ຳກັນ:",
        "en": "Duplicate Barcode Details:",
        "cn": "重复条形码详情：",
        "ja": "重複バーコード詳細：",
        "ko": "중복 바코드 세부 정보:",
        "my": "ထပ်နေသော Barcode အသေးစိတ်:",
        "km": "ព័ត៌មានលម្អិតបារកូដស្ទួន:",
        "vi": "Chi tiết mã vạch trùng lặp:"
    },
    {
        "code": "import_product_file_duplicate_count",
        "th": "ซ้ำ",
        "lo": "ຊ້ຳ",
        "en": "Duplicate",
        "cn": "重复",
        "ja": "重複",
        "ko": "중복",
        "my": "ထပ်",
        "km": "ស្ទួន",
        "vi": "Trùng lặp"
    },
    {
        "code": "import_product_file_row_found",
        "th": "แถวที่พบ",
        "lo": "ແຖວທີ່ພົບ",
        "en": "Row Found",
        "cn": "找到行",
        "ja": "見つかった行",
        "ko": "찾은 행",
        "my": "တွေ့ရှိသောအတန်း",
        "km": "ជួរដែលរកឃើញ",
        "vi": "Hàng tìm thấy"
    },
    {
        "code": "import_product_file_fix_duplicates_msg",
        "th": "กรุณาแก้ไขไฟล์ Excel ให้ barcode ไม่ซ้ำกันก่อนนำเข้าอีกครั้ง",
        "lo": "ກະລຸນາແກ້ໄຂໄຟລ໌ Excel ໃຫ້ barcode ບໍ່ຊ້ຳກັນກ່ອນນຳເຂົ້າອີກຄັ້ງ",
        "en": "Please fix the Excel file to ensure barcodes are unique before importing again",
        "cn": "请修复Excel文件以确保条形码唯一后再次导入",
        "ja": "再度インポートする前に、Excelファイルを修正してバーコードが一意であることを確認してください",
        "ko": "다시 가져오기 전에 Excel 파일을 수정하여 바코드가 고유한지 확인하십시오",
        "my": "ထပ်မံတင်သွင်းမီ barcode များထူးခြားစေရန် Excel ဖိုင်ကိုပြင်ဆင်ပါ",
        "km": "សូមជួសជុលឯកសារ Excel ឱ្យបារកូដមិនស្ទួនមុនពេលនាំចូលម្តងទៀត",
        "vi": "Vui lòng sửa tệp Excel để đảm bảo mã vạch là duy nhất trước khi nhập lại"
    },
    {
        "code": "import_product_file_error_found",
        "th": "พบข้อผิดพลาด",
        "lo": "ພົບຂໍ້ຜິດພາດ",
        "en": "Error Found",
        "cn": "发现错误",
        "ja": "エラーが見つかりました",
        "ko": "오류 발견",
        "my": "အမှားတွေ့ရှိပါသည်",
        "km": "រកឃើញកំហុស",
        "vi": "Phát hiện lỗi"
    },
    {
        "code": "import_product_file_import_result",
        "th": "ผลการนำเข้าข้อมูล",
        "lo": "ຜົນການນຳເຂົ້າຂໍ້ມູນ",
        "en": "Import Results",
        "cn": "导入结果",
        "ja": "インポート結果",
        "ko": "가져오기 결과",
        "my": "တင်သွင်းမှုရလဒ်များ",
        "km": "លទ្ធផលការនាំចូល",
        "vi": "Kết quả nhập"
    },
    {
        "code": "import_product_file_file",
        "th": "ไฟล์",
        "lo": "ໄຟລ໌",
        "en": "File",
        "cn": "文件",
        "ja": "ファイル",
        "ko": "파일",
        "my": "ဖိုင်",
        "km": "ឯកសារ",
        "vi": "Tệp"
    },
    {
        "code": "import_product_file_time_used",
        "th": "เวลาที่ใช้",
        "lo": "ເວລາທີ່ໃຊ້",
        "en": "Time Used",
        "cn": "使用时间",
        "ja": "使用時間",
        "ko": "사용 시간",
        "my": "အသုံးပြုချိန်",
        "km": "ពេលវេលាប្រើប្រាស់",
        "vi": "Thời gian sử dụng"
    },
    {
        "code": "import_product_file_excel_rows",
        "th": "แถวใน Excel",
        "lo": "ແຖວໃນ Excel",
        "en": "Rows in Excel",
        "cn": "Excel中的行",
        "ja": "Excelの行",
        "ko": "Excel의 행",
        "my": "Excel ရှိ အတန်းများ",
        "km": "ជួរក្នុង Excel",
        "vi": "Hàng trong Excel"
    },
    {
        "code": "import_product_file_products_in_system",
        "th": "สินค้าในระบบ",
        "lo": "ສິນຄ້າໃນລະບົບ",
        "en": "Products in System",
        "cn": "系统中的产品",
        "ja": "システム内の製品",
        "ko": "시스템의 제품",
        "my": "စနစ်ထဲရှိထုတ်ကုန်များ",
        "km": "ផលិតផលក្នុងប្រព័ន្ធ",
        "vi": "Sản phẩm trong hệ thống"
    },
    {
        "code": "import_product_file_new_products",
        "th": "สินค้าใหม่",
        "lo": "ສິນຄ້າໃໝ່",
        "en": "New Products",
        "cn": "新产品",
        "ja": "新製品",
        "ko": "새 제품",
        "my": "ထုတ်ကုန်အသစ်များ",
        "km": "ផលិតផលថ្មី",
        "vi": "Sản phẩm mới"
    },
    {
        "code": "import_product_file_updated",
        "th": "อัปเดต",
        "lo": "ອັບເດດ",
        "en": "Updated",
        "cn": "已更新",
        "ja": "更新済み",
        "ko": "업데이트됨",
        "my": "အပ်ဒိတ်လုပ်ပြီး",
        "km": "បានធ្វើបច្ចុប្បន្នភាព",
        "vi": "Đã cập nhật"
    },
    {
        "code": "import_product_file_unchanged",
        "th": "ไม่เปลี่ยนแปลง",
        "lo": "ບໍ່ມີການປ່ຽນແປງ",
        "en": "Unchanged",
        "cn": "未更改",
        "ja": "変更なし",
        "ko": "변경 없음",
        "my": "ပြောင်းလဲမှုမရှိ",
        "km": "មិនមានការផ្លាស់ប្តូរ",
        "vi": "Không thay đổi"
    },
    {
        "code": "import_product_file_success",
        "th": "สำเร็จ",
        "lo": "ສຳເລັດ",
        "en": "Success",
        "cn": "成功",
        "ja": "成功",
        "ko": "성공",
        "my": "အောင်မြင်",
        "km": "ជោគជ័យ",
        "vi": "Thành công"
    },
    {
        "code": "import_product_file_error",
        "th": "ข้อผิดพลาด",
        "lo": "ຂໍ້ຜິດພາດ",
        "en": "Error",
        "cn": "错误",
        "ja": "エラー",
        "ko": "오류",
        "my": "အမှား",
        "km": "កំហុស",
        "vi": "Lỗi"
    },
    {
        "code": "import_product_file_import_detail",
        "th": "รายละเอียดการนำเข้า",
        "lo": "ລາຍລະອຽດການນຳເຂົ້າ",
        "en": "Import Details",
        "cn": "导入详情",
        "ja": "インポート詳細",
        "ko": "가져오기 세부 정보",
        "my": "တင်သွင်းမှုအသေးစိတ်",
        "km": "ព័ត៌មានលម្អិតការនាំចូល",
        "vi": "Chi tiết nhập"
    },
    {
        "code": "import_product_file_view_all_details",
        "th": "ดูรายละเอียดทั้งหมด",
        "lo": "ເບິ່ງລາຍລະອຽດທັງໝົດ",
        "en": "View All Details",
        "cn": "查看所有详情",
        "ja": "すべての詳細を表示",
        "ko": "모든 세부 정보 보기",
        "my": "အသေးစိတ်အားလုံးကြည့်ပါ",
        "km": "មើលព័ត៌មានលម្អិតទាំងអស់",
        "vi": "Xem tất cả chi tiết"
    },
    {
        "code": "import_product_file_view_duplicate_list",
        "th": "ดูรายการ Barcode ซ้ำ",
        "lo": "ເບິ່ງລາຍການ Barcode ຊ້ຳກັນ",
        "en": "View Duplicate Barcode List",
        "cn": "查看重复条形码列表",
        "ja": "重複バーコードリストを表示",
        "ko": "중복 바코드 목록 보기",
        "my": "ထပ်နေသော Barcode စာရင်းကြည့်ပါ",
        "km": "មើលបញ្ជីបារកូដស្ទួន",
        "vi": "Xem danh sách mã vạch trùng lặp"
    },
    {
        "code": "import_product_file_size",
        "th": "ขนาด",
        "lo": "ຂະໜາດ",
        "en": "Size",
        "cn": "大小",
        "ja": "サイズ",
        "ko": "크기",
        "my": "အရွယ်အစား",
        "km": "ទំហំ",
        "vi": "Kích thước"
    }
]

# Load existing language.json
with open(r'c:\git\bcaiclouderp\frontend\assets\language.json', 'r', encoding='utf-8') as f:
    language_data = json.load(f)

# Convert and add new entries
for entry in new_translations:
    code = entry["code"]
    new_entry = {
        "code": code,
        "langs": [
            {"code": "th", "text": entry["th"]},
            {"code": "lo", "text": entry["lo"]},
            {"code": "en", "text": entry["en"]},
            {"code": "cn", "text": entry["cn"]},
            {"code": "ja", "text": entry["ja"]},
            {"code": "ko", "text": entry["ko"]},
            {"code": "my", "text": entry["my"]},
            {"code": "km", "text": entry["km"]},
            {"code": "vi", "text": entry["vi"]}
        ]
    }
    language_data.append(new_entry)

# Write back to file
with open(r'c:\git\bcaiclouderp\frontend\assets\language.json', 'w', encoding='utf-8') as f:
    json.dump(language_data, f, ensure_ascii=False, indent=4)

print(f"✓ Added {len(new_translations)} new language entries")
print(f"✓ Total entries in language.json: {len(language_data)}")
