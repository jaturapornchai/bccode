import json

# Read existing language.json
with open(r'c:\git\bcaiclouderp\frontend\assets\language.json', 'r', encoding='utf-8') as f:
    data = json.load(f)

# Keys to add for point_transaction_screen.dart
point_transaction_entries = [
    {
        'code': 'point_transaction.received_points',
        'langs': [
            {'code': 'th', 'text': 'ได้รับแต้ม'},
            {'code': 'lo', 'text': 'ໄດ້ຮັບແຕ້ມ'},
            {'code': 'en', 'text': 'Received Points'},
            {'code': 'cn', 'text': '获得积分'},
            {'code': 'zh', 'text': '獲得積分'},
            {'code': 'ja', 'text': 'ポイント獲得'},
            {'code': 'ko', 'text': '포인트 획득'},
            {'code': 'my', 'text': 'အမှတ်ရရှိသည်'},
            {'code': 'km', 'text': 'ទទួលបានពិន្ទុ'},
            {'code': 'vi', 'text': 'Nhận điểm'}
        ]
    },
    {
        'code': 'point_transaction.used_points',
        'langs': [
            {'code': 'th', 'text': 'ใช้แต้ม'},
            {'code': 'lo', 'text': 'ໃຊ້ແຕ້ມ'},
            {'code': 'en', 'text': 'Used Points'},
            {'code': 'cn', 'text': '使用积分'},
            {'code': 'zh', 'text': '使用積分'},
            {'code': 'ja', 'text': 'ポイント使用'},
            {'code': 'ko', 'text': '포인트 사용'},
            {'code': 'my', 'text': 'အမှတ်သုံးသည်'},
            {'code': 'km', 'text': 'ប្រើប្រាស់ពិន្ទុ'},
            {'code': 'vi', 'text': 'Sử dụng điểm'}
        ]
    },
    {
        'code': 'point_transaction.unknown',
        'langs': [
            {'code': 'th', 'text': 'ไม่ทราบ'},
            {'code': 'lo', 'text': 'ບໍ່ຮູ້'},
            {'code': 'en', 'text': 'Unknown'},
            {'code': 'cn', 'text': '未知'},
            {'code': 'zh', 'text': '未知'},
            {'code': 'ja', 'text': '不明'},
            {'code': 'ko', 'text': '알 수 없음'},
            {'code': 'my', 'text': 'မသိရှိပါ'},
            {'code': 'km', 'text': 'មិនស្គាល់'},
            {'code': 'vi', 'text': 'Không rõ'}
        ]
    },
    {
        'code': 'point_transaction.points_code',
        'langs': [
            {'code': 'th', 'text': 'รหัสแต้ม'},
            {'code': 'lo', 'text': 'ລະຫັດແຕ້ມ'},
            {'code': 'en', 'text': 'Points Code'},
            {'code': 'cn', 'text': '积分代码'},
            {'code': 'zh', 'text': '積分代碼'},
            {'code': 'ja', 'text': 'ポイントコード'},
            {'code': 'ko', 'text': '포인트 코드'},
            {'code': 'my', 'text': 'အမှတ်ကုဒ်'},
            {'code': 'km', 'text': 'លេខកូដពិន្ទុ'},
            {'code': 'vi', 'text': 'Mã điểm'}
        ]
    },
    {
        'code': 'point_transaction.balance_before',
        'langs': [
            {'code': 'th', 'text': 'ยอดก่อน'},
            {'code': 'lo', 'text': 'ຍອດກ່ອນ'},
            {'code': 'en', 'text': 'Balance Before'},
            {'code': 'cn', 'text': '之前余额'},
            {'code': 'zh', 'text': '之前餘額'},
            {'code': 'ja', 'text': '前残高'},
            {'code': 'ko', 'text': '이전 잔액'},
            {'code': 'my', 'text': 'ယခင်လက်ကျန်'},
            {'code': 'km', 'text': 'សមតុល្យមុន'},
            {'code': 'vi', 'text': 'Số dư trước'}
        ]
    },
    {
        'code': 'point_transaction.balance_after',
        'langs': [
            {'code': 'th', 'text': 'ยอดหลัง'},
            {'code': 'lo', 'text': 'ຍອດຫຼັງ'},
            {'code': 'en', 'text': 'Balance After'},
            {'code': 'cn', 'text': '之后余额'},
            {'code': 'zh', 'text': '之後餘額'},
            {'code': 'ja', 'text': '後残高'},
            {'code': 'ko', 'text': '이후 잔액'},
            {'code': 'my', 'text': 'နောက်လက်ကျန်'},
            {'code': 'km', 'text': 'សមតុល្យក្រោយ'},
            {'code': 'vi', 'text': 'Số dư sau'}
        ]
    },
    {
        'code': 'point_transaction.points_history',
        'langs': [
            {'code': 'th', 'text': 'ประวัติการใช้แต้ม'},
            {'code': 'lo', 'text': 'ປະຫວັດການໃຊ້ແຕ້ມ'},
            {'code': 'en', 'text': 'Points History'},
            {'code': 'cn', 'text': '积分历史'},
            {'code': 'zh', 'text': '積分歷史'},
            {'code': 'ja', 'text': 'ポイント履歴'},
            {'code': 'ko', 'text': '포인트 내역'},
            {'code': 'my', 'text': 'အမှတ်သမိုင်း'},
            {'code': 'km', 'text': 'ប្រវត្តិពិន្ទុ'},
            {'code': 'vi', 'text': 'Lịch sử điểm'}
        ]
    },
    {
        'code': 'point_transaction.error_occurred',
        'langs': [
            {'code': 'th', 'text': 'เกิดข้อผิดพลาด'},
            {'code': 'lo', 'text': 'ເກີດຄວາມຜິດພາດ'},
            {'code': 'en', 'text': 'An error occurred'},
            {'code': 'cn', 'text': '发生错误'},
            {'code': 'zh', 'text': '發生錯誤'},
            {'code': 'ja', 'text': 'エラーが発生しました'},
            {'code': 'ko', 'text': '오류 발생'},
            {'code': 'my', 'text': 'အမှားအယွင်းဖြစ်ပွားသည်'},
            {'code': 'km', 'text': 'មានកំហុសកើតឡើង'},
            {'code': 'vi', 'text': 'Đã xảy ra lỗi'}
        ]
    },
    {
        'code': 'point_transaction.no_points_history_found',
        'langs': [
            {'code': 'th', 'text': 'ไม่พบประวัติการใช้แต้ม'},
            {'code': 'lo', 'text': 'ບໍ່ພົບປະຫວັດການໃຊ້ແຕ້ມ'},
            {'code': 'en', 'text': 'No points history found'},
            {'code': 'cn', 'text': '未找到积分历史'},
            {'code': 'zh', 'text': '未找到積分歷史'},
            {'code': 'ja', 'text': 'ポイント履歴が見つかりません'},
            {'code': 'ko', 'text': '포인트 내역을 찾을 수 없습니다'},
            {'code': 'my', 'text': 'အမှတ်သမိုင်းမတွေ့ပါ'},
            {'code': 'km', 'text': 'រកមិនឃើញប្រវត្តិពិន្ទុ'},
            {'code': 'vi', 'text': 'Không tìm thấy lịch sử điểm'}
        ]
    }
]

# Add all new entries
data.extend(point_transaction_entries)

# Write back
with open(r'c:\git\bcaiclouderp\frontend\assets\language.json', 'w', encoding='utf-8') as f:
    json.dump(data, f, ensure_ascii=False, indent=4)

print(f'Added {len(point_transaction_entries)} entries for point_transaction_screen.dart')
print(f'Total entries now: {len(data)}')
print('\n=== CONVERSION SUMMARY ===')
print('Files converted:')
print('1. option_detail.dart - 17 keys added')
print('2. import_product_detail_screen.dart - 13 keys added')
print('3. database_master_info.dart - 13 keys added')
print('4. point_transaction_screen.dart - 9 keys added')
print(f'\nTotal new translation keys: {17 + 13 + 13 + 9} keys')
print(f'Total entries in language.json: {len(data)}')
print('\nAll translations include 10 languages: th, lo, en, cn, zh, ja, ko, my, km, vi')
