#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Script to add purchase partial report language entries to language.json
"""

import json
import sys

# New language entries to add
new_entries = [
    {
        "code": "purchase_partial_report_title",
        "langs": [
            {"code": "th", "text": "รายงานรับสินค้าแบบทยอยรับ"},
            {"code": "lo", "text": "ລາຍງານຮັບສິນຄ້າແບບທີ່ຍອຍຮັບ"},
            {"code": "en", "text": "Gradual Purchase Receipt Report"},
            {"code": "cn", "text": "逐步进货报告"},
            {"code": "ja", "text": "段階的仕入レポート"},
            {"code": "ko", "text": "점진적 구매 수령 보고서"},
            {"code": "my", "text": "အဆင့်ဆင့်ဝယ်ယူမှုလက်ခံအစီရင်ခံစာ"},
            {"code": "km", "text": "របាយការណ៍ទទួលទំនិញជាជំហាន"},
            {"code": "vi", "text": "Báo cáo nhận hàng mua dần"}
        ]
    },
    {
        "code": "purchase_partial_report_dedebi_title",
        "langs": [
            {"code": "th", "text": "รายงานรับสินค้าแบบทยอยรับ Dedebi"},
            {"code": "lo", "text": "ລາຍງານຮັບສິນຄ້າແບບທີ່ຍອຍຮັບ Dedebi"},
            {"code": "en", "text": "Gradual Purchase Receipt Report Dedebi"},
            {"code": "cn", "text": "逐步进货报告 Dedebi"},
            {"code": "ja", "text": "段階的仕入レポート Dedebi"},
            {"code": "ko", "text": "점진적 구매 수령 보고서 Dedebi"},
            {"code": "my", "text": "အဆင့်ဆင့်ဝယ်ယူမှုလက်ခံအစီရင်ခံစာ Dedebi"},
            {"code": "km", "text": "របាយការណ៍ទទួលទំនិញជាជំហាន Dedebi"},
            {"code": "vi", "text": "Báo cáo nhận hàng mua dần Dedebi"}
        ]
    },
    {
        "code": "hide_filter_panel",
        "langs": [
            {"code": "th", "text": "ซ่อนแผงตัวกรอง"},
            {"code": "lo", "text": "ເຊື່ອງແຜງຕົວກອງ"},
            {"code": "en", "text": "Hide Filter Panel"},
            {"code": "cn", "text": "隐藏筛选面板"},
            {"code": "ja", "text": "フィルターパネルを非表示"},
            {"code": "ko", "text": "필터 패널 숨기기"},
            {"code": "my", "text": "စစ်ထုတ်မှုပြားကွက်ဖျောက်ရန်"},
            {"code": "km", "text": "លាក់បន្ទះតម្រង"},
            {"code": "vi", "text": "Ẩn bảng lọc"}
        ]
    },
    {
        "code": "show_filter_panel",
        "langs": [
            {"code": "th", "text": "แสดงแผงตัวกรอง"},
            {"code": "lo", "text": "ສະແດງແຜງຕົວກອງ"},
            {"code": "en", "text": "Show Filter Panel"},
            {"code": "cn", "text": "显示筛选面板"},
            {"code": "ja", "text": "フィルターパネルを表示"},
            {"code": "ko", "text": "필터 패널 표시"},
            {"code": "my", "text": "စစ်ထုတ်မှုပြားကွက်ပြရန်"},
            {"code": "km", "text": "បង្ហាញបន្ទះតម្រង"},
            {"code": "vi", "text": "Hiển thị bảng lọc"}
        ]
    },
    {
        "code": "set_search_conditions",
        "langs": [
            {"code": "th", "text": "ตั้งค่าเงื่อนไขการค้นหา"},
            {"code": "lo", "text": "ຕັ້ງຄ່າເງື່ອນໄຂການຄົ້ນຫາ"},
            {"code": "en", "text": "Set Search Conditions"},
            {"code": "cn", "text": "设置搜索条件"},
            {"code": "ja", "text": "検索条件を設定"},
            {"code": "ko", "text": "검색 조건 설정"},
            {"code": "my", "text": "ရှာဖွေမှုအခြေအနေများသတ်မှတ်ရန်"},
            {"code": "km", "text": "កំណត់លក្ខខណ្ឌស្វែងរក"},
            {"code": "vi", "text": "Đặt điều kiện tìm kiếm"}
        ]
    },
    {
        "code": "refresh_data",
        "langs": [
            {"code": "th", "text": "รีเฟรชข้อมูล"},
            {"code": "lo", "text": "ໂຫຼດຂໍ້ມູນໃໝ່"},
            {"code": "en", "text": "Refresh Data"},
            {"code": "cn", "text": "刷新数据"},
            {"code": "ja", "text": "データを更新"},
            {"code": "ko", "text": "데이터 새로고침"},
            {"code": "my", "text": "ဒေတာပြန်လည်ဖွင့်ရန်"},
            {"code": "km", "text": "ធ្វើឱ្យទិន្នន័យស្រស់"},
            {"code": "vi", "text": "Làm mới dữ liệu"}
        ]
    },
    {
        "code": "export_report",
        "langs": [
            {"code": "th", "text": "ส่งออกรายงาน"},
            {"code": "lo", "text": "ສົ່ງອອກລາຍງານ"},
            {"code": "en", "text": "Export Report"},
            {"code": "cn", "text": "导出报告"},
            {"code": "ja", "text": "レポートをエクスポート"},
            {"code": "ko", "text": "보고서 내보내기"},
            {"code": "my", "text": "အစီရင်ခံစာထုတ်ယူရန်"},
            {"code": "km", "text": "នាំចេញរបាយការណ៍"},
            {"code": "vi", "text": "Xuất báo cáo"}
        ]
    },
    {
        "code": "generating_report",
        "langs": [
            {"code": "th", "text": "กำลังเริ่มสร้างรายงาน..."},
            {"code": "lo", "text": "ກຳລັງເລີ່ມສ້າງລາຍງານ..."},
            {"code": "en", "text": "Generating report..."},
            {"code": "cn", "text": "正在生成报告..."},
            {"code": "ja", "text": "レポートを生成中..."},
            {"code": "ko", "text": "보고서 생성 중..."},
            {"code": "my", "text": "အစီရင်ခံစာဖန်တီးနေသည်..."},
            {"code": "km", "text": "កំពុងបង្កើតរបាយការណ៍..."},
            {"code": "vi", "text": "Đang tạo báo cáo..."}
        ]
    },
    {
        "code": "loading_data",
        "langs": [
            {"code": "th", "text": "กำลังโหลดข้อมูล..."},
            {"code": "lo", "text": "ກຳລັງໂຫຼດຂໍ້ມູນ..."},
            {"code": "en", "text": "Loading data..."},
            {"code": "cn", "text": "正在加载数据..."},
            {"code": "ja", "text": "データを読み込み中..."},
            {"code": "ko", "text": "데이터 로드 중..."},
            {"code": "my", "text": "ဒေတာတင်နေသည်..."},
            {"code": "km", "text": "កំពុងផ្ទុកទិន្នន័យ..."},
            {"code": "vi", "text": "Đang tải dữ liệu..."}
        ]
    },
    {
        "code": "loading_new_page",
        "langs": [
            {"code": "th", "text": "กำလังโหลดหน้าใหม่..."},
            {"code": "lo", "text": "ກຳລັງໂຫຼດໜ້າໃໝ່..."},
            {"code": "en", "text": "Loading new page..."},
            {"code": "cn", "text": "正在加载新页面..."},
            {"code": "ja", "text": "新しいページを読み込み中..."},
            {"code": "ko", "text": "새 페이지 로드 중..."},
            {"code": "my", "text": "စာမျက်နှာအသစ်တင်နေသည်..."},
            {"code": "km", "text": "កំពុងផ្ទុកទំព័រថ្មី..."},
            {"code": "vi", "text": "Đang tải trang mới..."}
        ]
    },
    {
        "code": "try_again",
        "langs": [
            {"code": "th", "text": "ลองใหม่"},
            {"code": "lo", "text": "ລອງໃໝ່"},
            {"code": "en", "text": "Try Again"},
            {"code": "cn", "text": "重试"},
            {"code": "ja", "text": "再試行"},
            {"code": "ko", "text": "다시 시도"},
            {"code": "my", "text": "ထပ်စမ်းကြည့်ပါ"},
            {"code": "km", "text": "ព្យាយាមម្តងទៀត"},
            {"code": "vi", "text": "Thử lại"}
        ]
    },
    {
        "code": "set_search_conditions_to_view_data",
        "langs": [
            {"code": "th", "text": "กำหนดเงื่อนไขการค้นหาเพื่อดูข้อมูล"},
            {"code": "lo", "text": "ກຳນົດເງື່ອນໄຂການຄົ້ນຫາເພື່ອເບິ່ງຂໍ້ມູນ"},
            {"code": "en", "text": "Set search conditions to view data"},
            {"code": "cn", "text": "设置搜索条件以查看数据"},
            {"code": "ja", "text": "データを表示するための検索条件を設定"},
            {"code": "ko", "text": "데이터를 보려면 검색 조건 설정"},
            {"code": "my", "text": "ဒေတာကြည့်ရှုရန် ရှာဖွေမှုအခြေအနေများသတ်မှတ်ပါ"},
            {"code": "km", "text": "កំណត់លក្ខខណ្ឌស្វែងរកដើម្បីមើលទិន្នន័យ"},
            {"code": "vi", "text": "Đặt điều kiện tìm kiếm để xem dữ liệu"}
        ]
    },
    {
        "code": "please_wait",
        "langs": [
            {"code": "th", "text": "กรุณารอสักครู่..."},
            {"code": "lo", "text": "ກະລຸນາລໍຖ້າສັກຄູ່..."},
            {"code": "en", "text": "Please wait..."},
            {"code": "cn", "text": "请稍候..."},
            {"code": "ja", "text": "お待ちください..."},
            {"code": "ko", "text": "잠시만 기다려주세요..."},
            {"code": "my", "text": "ခဏစောင့်ဆိုင်းပါ..."},
            {"code": "km", "text": "សូមរង់ចាំបន្តិច..."},
            {"code": "vi", "text": "Vui lòng đợi..."}
        ]
    },
    {
        "code": "showing_items",
        "langs": [
            {"code": "th", "text": "แสดง"},
            {"code": "lo", "text": "ສະແດງ"},
            {"code": "en", "text": "Showing"},
            {"code": "cn", "text": "显示"},
            {"code": "ja", "text": "表示"},
            {"code": "ko", "text": "표시"},
            {"code": "my", "text": "ပြသခြင်း"},
            {"code": "km", "text": "បង្ហាញ"},
            {"code": "vi", "text": "Hiển thị"}
        ]
    },
    {
        "code": "items_label",
        "langs": [
            {"code": "th", "text": "รายการ"},
            {"code": "lo", "text": "ລາຍການ"},
            {"code": "en", "text": "items"},
            {"code": "cn", "text": "项"},
            {"code": "ja", "text": "件"},
            {"code": "ko", "text": "항목"},
            {"code": "my", "text": "ပစ္စည်းများ"},
            {"code": "km", "text": "ធាតុ"},
            {"code": "vi", "text": "mục"}
        ]
    },
    {
        "code": "first_page",
        "langs": [
            {"code": "th", "text": "หน้าแรก"},
            {"code": "lo", "text": "ໜ້າທຳອິດ"},
            {"code": "en", "text": "First Page"},
            {"code": "cn", "text": "第一页"},
            {"code": "ja", "text": "最初のページ"},
            {"code": "ko", "text": "첫 페이지"},
            {"code": "my", "text": "ပထမစာမျက်နှာ"},
            {"code": "km", "text": "ទំព័រដំបូង"},
            {"code": "vi", "text": "Trang đầu"}
        ]
    },
    {
        "code": "previous_page",
        "langs": [
            {"code": "th", "text": "หน้าก่อนหน้า"},
            {"code": "lo", "text": "ໜ້າກ່ອນໜ້າ"},
            {"code": "en", "text": "Previous Page"},
            {"code": "cn", "text": "上一页"},
            {"code": "ja", "text": "前のページ"},
            {"code": "ko", "text": "이전 페이지"},
            {"code": "my", "text": "ယခင်စာမျက်နှာ"},
            {"code": "km", "text": "ទំព័រមុន"},
            {"code": "vi", "text": "Trang trước"}
        ]
    },
    {
        "code": "next_page",
        "langs": [
            {"code": "th", "text": "หน้าถัดไป"},
            {"code": "lo", "text": "ໜ້າຕໍ່ໄປ"},
            {"code": "en", "text": "Next Page"},
            {"code": "cn", "text": "下一页"},
            {"code": "ja", "text": "次のページ"},
            {"code": "ko", "text": "다음 페이지"},
            {"code": "my", "text": "နောက်စာမျက်နှာ"},
            {"code": "km", "text": "ទំព័របន្ទាប់"},
            {"code": "vi", "text": "Trang tiếp"}
        ]
    },
    {
        "code": "last_page",
        "langs": [
            {"code": "th", "text": "หน้าสุดท้าย"},
            {"code": "lo", "text": "ໜ້າສຸດທ້າຍ"},
            {"code": "en", "text": "Last Page"},
            {"code": "cn", "text": "最后一页"},
            {"code": "ja", "text": "最後のページ"},
            {"code": "ko", "text": "마지막 페이지"},
            {"code": "my", "text": "နောက်ဆုံးစာမျက်နှာ"},
            {"code": "km", "text": "ទំព័រចុងក្រោយ"},
            {"code": "vi", "text": "Trang cuối"}
        ]
    },
    {
        "code": "error_occurred",
        "langs": [
            {"code": "th", "text": "เกิดข้อผิดพลาด"},
            {"code": "lo", "text": "ເກີດຂໍ້ຜິດພາດ"},
            {"code": "en", "text": "Error Occurred"},
            {"code": "cn", "text": "发生错误"},
            {"code": "ja", "text": "エラーが発生しました"},
            {"code": "ko", "text": "오류 발생"},
            {"code": "my", "text": "အမှားအယွင်းဖြစ်ပွားခဲ့သည်"},
            {"code": "km", "text": "មានកំហុសកើតឡើង"},
            {"code": "vi", "text": "Đã xảy ra lỗi"}
        ]
    },
    {
        "code": "change_conditions",
        "langs": [
            {"code": "th", "text": "เปลี่ยนเงื่อนไข"},
            {"code": "lo", "text": "ປ່ຽນເງື່ອນໄຂ"},
            {"code": "en", "text": "Change Conditions"},
            {"code": "cn", "text": "更改条件"},
            {"code": "ja", "text": "条件を変更"},
            {"code": "ko", "text": "조건 변경"},
            {"code": "my", "text": "အခြေအနေပြောင်းရန်"},
            {"code": "km", "text": "ផ្លាស់ប្តូរលក្ខខណ្ឌ"},
            {"code": "vi", "text": "Thay đổi điều kiện"}
        ]
    },
    {
        "code": "clear_data",
        "langs": [
            {"code": "th", "text": "ล้างข้อมูล"},
            {"code": "lo", "text": "ລົບຂໍ້ມູນ"},
            {"code": "en", "text": "Clear Data"},
            {"code": "cn", "text": "清除数据"},
            {"code": "ja", "text": "データをクリア"},
            {"code": "ko", "text": "데이터 지우기"},
            {"code": "my", "text": "ဒေတာရှင်းလင်းရန်"},
            {"code": "km", "text": "លុបទិន្នន័យ"},
            {"code": "vi", "text": "Xóa dữ liệu"}
        ]
    },
    {
        "code": "please_search_data_before_export",
        "langs": [
            {"code": "th", "text": "กรุณาค้นหาข้อมูลก่อนส่งออก"},
            {"code": "lo", "text": "ກະລຸນາຄົ້ນຫາຂໍ້ມູນກ່ອນສົ່ງອອກ"},
            {"code": "en", "text": "Please search data before export"},
            {"code": "cn", "text": "请在导出前搜索数据"},
            {"code": "ja", "text": "エクスポート前にデータを検索してください"},
            {"code": "ko", "text": "내보내기 전에 데이터를 검색하세요"},
            {"code": "my", "text": "ထုတ်ယူခြင်းမပြုမီ ဒေတာရှာဖွေပါ"},
            {"code": "km", "text": "សូមស្វែងរកទិន្នន័យមុននឹងនាំចេញ"},
            {"code": "vi", "text": "Vui lòng tìm kiếm dữ liệu trước khi xuất"}
        ]
    },
    {
        "code": "please_select_date_range",
        "langs": [
            {"code": "th", "text": "กรุณาเลือกช่วงวันที่"},
            {"code": "lo", "text": "ກະລຸນາເລືອກຊ່ວງວັນທີ່"},
            {"code": "en", "text": "Please select date range"},
            {"code": "cn", "text": "请选择日期范围"},
            {"code": "ja", "text": "日付範囲を選択してください"},
            {"code": "ko", "text": "날짜 범위를 선택하세요"},
            {"code": "my", "text": "ရက်စွဲအပိုင်းအခြားရွေးချယ်ပါ"},
            {"code": "km", "text": "សូមជ្រើសរើសចន្លោះកាលបរិច្ឆេទ"},
            {"code": "vi", "text": "Vui lòng chọn khoảng ngày"}
        ]
    },
    {
        "code": "start_date_must_not_exceed_end_date",
        "langs": [
            {"code": "th", "text": "วันที่เริ่มต้นต้องไม่เกินวันที่สิ้นสุด"},
            {"code": "lo", "text": "ວັນທີ່ເລີ່ມຕົ້ນຕ້ອງບໍ່ເກີນວັນທີ່ສິ້ນສຸດ"},
            {"code": "en", "text": "Start date must not exceed end date"},
            {"code": "cn", "text": "开始日期不能超过结束日期"},
            {"code": "ja", "text": "開始日は終了日を超えてはいけません"},
            {"code": "ko", "text": "시작 날짜는 종료 날짜를 초과할 수 없습니다"},
            {"code": "my", "text": "စတင်ရက်သည် ပြီးဆုံးရက်ထက် မကျော်လွန်ရပါ"},
            {"code": "km", "text": "កាលបរិច្ឆេទចាប់ផ្តើមមិនត្រូវលើសពីកាលបរិច្ឆេទបញ្ចប់ទេ"},
            {"code": "vi", "text": "Ngày bắt đầu không được vượt quá ngày kết thúc"}
        ]
    },
    {
        "code": "date_range_must_not_exceed_1_year",
        "langs": [
            {"code": "th", "text": "ช่วงวันที่ต้องไม่เกิน 1 ปี"},
            {"code": "lo", "text": "ຊ່ວງວັນທີ່ຕ້ອງບໍ່ເກີນ 1 ປີ"},
            {"code": "en", "text": "Date range must not exceed 1 year"},
            {"code": "cn", "text": "日期范围不得超过1年"},
            {"code": "ja", "text": "日付範囲は1年を超えてはいけません"},
            {"code": "ko", "text": "날짜 범위는 1년을 초과할 수 없습니다"},
            {"code": "my", "text": "ရက်စွဲအပိုင်းအခြားသည် ၁ နှစ်ထက် မကျော်လွန်ရပါ"},
            {"code": "km", "text": "ចន្លោះកាលបរិច្ឆេទមិនត្រូវលើសពី 1 ឆ្នាំទេ"},
            {"code": "vi", "text": "Khoảng ngày không được vượt quá 1 năm"}
        ]
    },
    {
        "code": "token_not_found_please_login_again",
        "langs": [
            {"code": "th", "text": "ไม่พบ Token การเชื่อมต่อ กรุณาเข้าสู่ระบบใหม่"},
            {"code": "lo", "text": "ບໍ່ພົບ Token ການເຊື່ອມຕໍ່ ກະລຸນາເຂົ້າສູ່ລະບົບໃໝ່"},
            {"code": "en", "text": "Connection token not found, please login again"},
            {"code": "cn", "text": "未找到连接令牌，请重新登录"},
            {"code": "ja", "text": "接続トークンが見つかりません。再度ログインしてください"},
            {"code": "ko", "text": "연결 토큰을 찾을 수 없습니다. 다시 로그인하세요"},
            {"code": "my", "text": "ချိတ်ဆက်မှု Token ရှာမတွေ့ပါ၊ ပြန်လည် login ဝင်ပါ"},
            {"code": "km", "text": "រកមិនឃើញ Token ការតភ្ជាប់ សូមចូលប្រើប្រាស់ម្តងទៀត"},
            {"code": "vi", "text": "Không tìm thấy mã kết nối, vui lòng đăng nhập lại"}
        ]
    },
    {
        "code": "preparing_report_data",
        "langs": [
            {"code": "th", "text": "กำลังเตรียมข้อมูลรายงาน..."},
            {"code": "lo", "text": "ກຳລັງກຽມຂໍ້ມູນລາຍງານ..."},
            {"code": "en", "text": "Preparing report data..."},
            {"code": "cn", "text": "正在准备报告数据..."},
            {"code": "ja", "text": "レポートデータを準備中..."},
            {"code": "ko", "text": "보고서 데이터 준비 중..."},
            {"code": "my", "text": "အစီရင်ခံစာဒေတာပြင်ဆင်နေသည်..."},
            {"code": "km", "text": "កំពុងរៀបចំទិន្នន័យរបាយការណ៍..."},
            {"code": "vi", "text": "Đang chuẩn bị dữ liệu báo cáo..."}
        ]
    }
]

def main():
    lang_file = r"c:\git\bcaiclouderp\frontend\assets\language.json"

    try:
        # Read the current language.json
        with open(lang_file, 'r', encoding='utf-8') as f:
            data = json.load(f)

        # Check if entries already exist
        existing_codes = {entry['code'] for entry in data}
        new_codes_to_add = [entry for entry in new_entries if entry['code'] not in existing_codes]

        if not new_codes_to_add:
            print("All entries already exist in language.json")
            return

        # Append new entries
        data.extend(new_codes_to_add)

        # Write back to file with proper formatting
        with open(lang_file, 'w', encoding='utf-8') as f:
            json.dump(data, f, ensure_ascii=False, indent=4)

        print(f"Successfully added {len(new_codes_to_add)} new language entries")
        for entry in new_codes_to_add:
            print(f"  - {entry['code']}")

    except Exception as e:
        print(f"Error: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()
