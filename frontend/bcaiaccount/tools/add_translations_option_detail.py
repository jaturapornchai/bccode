import json

# Read existing language.json
with open(r'c:\git\bcaiclouderp\frontend\assets\language.json', 'r', encoding='utf-8') as f:
    data = json.load(f)

# Keys to add for option_detail.dart
new_entries = [
    {
        'code': 'option_detail.add_sub_option',
        'langs': [
            {'code': 'th', 'text': 'เพิ่มตัวเลือกย่อย'},
            {'code': 'lo', 'text': 'ເພີ່ມຕົວເລືອກຍ່ອຍ'},
            {'code': 'en', 'text': 'Add Sub-option'},
            {'code': 'cn', 'text': '添加子选项'},
            {'code': 'zh', 'text': '添加子選項'},
            {'code': 'ja', 'text': 'サブオプションを追加'},
            {'code': 'ko', 'text': '하위 옵션 추가'},
            {'code': 'my', 'text': 'ခွဲရွေးချယ်မှု ထည့်ရန်'},
            {'code': 'km', 'text': 'បន្ថែមជម្រើសរង'},
            {'code': 'vi', 'text': 'Thêm tùy chọn phụ'}
        ]
    },
    {
        'code': 'option_detail.edit_sub_option',
        'langs': [
            {'code': 'th', 'text': 'แก้ไขตัวเลือกย่อย'},
            {'code': 'lo', 'text': 'ແກ້ໄຂຕົວເລືອກຍ່ອຍ'},
            {'code': 'en', 'text': 'Edit Sub-option'},
            {'code': 'cn', 'text': '编辑子选项'},
            {'code': 'zh', 'text': '編輯子選項'},
            {'code': 'ja', 'text': 'サブオプションを編集'},
            {'code': 'ko', 'text': '하위 옵션 수정'},
            {'code': 'my', 'text': 'ခွဲရွေးချယ်မှု ပြင်ဆင်ရန်'},
            {'code': 'km', 'text': 'កែសម្រួលជម្រើសរង'},
            {'code': 'vi', 'text': 'Sửa tùy chọn phụ'}
        ]
    },
    {
        'code': 'option_detail.please_specify_barcode',
        'langs': [
            {'code': 'th', 'text': 'กรุณาระบุ Barcode'},
            {'code': 'lo', 'text': 'ກະລຸນາລະບຸ Barcode'},
            {'code': 'en', 'text': 'Please specify barcode'},
            {'code': 'cn', 'text': '请指定条码'},
            {'code': 'zh', 'text': '請指定條碼'},
            {'code': 'ja', 'text': 'バーコードを指定してください'},
            {'code': 'ko', 'text': '바코드를 지정해주세요'},
            {'code': 'my', 'text': 'ဘားကုဒ်ကို သတ်မှတ်ပေးပါ'},
            {'code': 'km', 'text': 'សូមបញ្ជាក់ Barcode'},
            {'code': 'vi', 'text': 'Vui lòng chỉ định mã vạch'}
        ]
    },
    {
        'code': 'option_detail.option_group_name_1',
        'langs': [
            {'code': 'th', 'text': 'ชื่อกลุ่มตัวเลือก 1'},
            {'code': 'lo', 'text': 'ຊື່ກຸ່ມຕົວເລືອກ 1'},
            {'code': 'en', 'text': 'Option Group Name 1'},
            {'code': 'cn', 'text': '选项组名称 1'},
            {'code': 'zh', 'text': '選項組名稱 1'},
            {'code': 'ja', 'text': 'オプショングループ名 1'},
            {'code': 'ko', 'text': '옵션 그룹명 1'},
            {'code': 'my', 'text': 'ရွေးချယ်မှုအုပ်စုအမည် 1'},
            {'code': 'km', 'text': 'ឈ្មោះក្រុមជម្រើស 1'},
            {'code': 'vi', 'text': 'Tên nhóm tùy chọn 1'}
        ]
    },
    {
        'code': 'option_detail.option_group_name_2',
        'langs': [
            {'code': 'th', 'text': 'ชื่อกลุ่มตัวเลือก 2'},
            {'code': 'lo', 'text': 'ຊື່ກຸ່ມຕົວເລືອກ 2'},
            {'code': 'en', 'text': 'Option Group Name 2'},
            {'code': 'cn', 'text': '选项组名称 2'},
            {'code': 'zh', 'text': '選項組名稱 2'},
            {'code': 'ja', 'text': 'オプショングループ名 2'},
            {'code': 'ko', 'text': '옵션 그룹명 2'},
            {'code': 'my', 'text': 'ရွေးချယ်မှုအုပ်စုအမည် 2'},
            {'code': 'km', 'text': 'ឈ្មោះក្រុមជម្រើស 2'},
            {'code': 'vi', 'text': 'Tên nhóm tùy chọn 2'}
        ]
    },
    {
        'code': 'option_detail.option_group_name_3',
        'langs': [
            {'code': 'th', 'text': 'ชื่อกลุ่มตัวเลือก 3'},
            {'code': 'lo', 'text': 'ຊື່ກຸ່ມຕົວເລືອກ 3'},
            {'code': 'en', 'text': 'Option Group Name 3'},
            {'code': 'cn', 'text': '选项组名称 3'},
            {'code': 'zh', 'text': '選項組名稱 3'},
            {'code': 'ja', 'text': 'オプショングループ名 3'},
            {'code': 'ko', 'text': '옵션 그룹명 3'},
            {'code': 'my', 'text': 'ရွေးချယ်မှုအုပ်စုအမည် 3'},
            {'code': 'km', 'text': 'ឈ្មោះក្រុមជម្រើស 3'},
            {'code': 'vi', 'text': 'Tên nhóm tùy chọn 3'}
        ]
    },
    {
        'code': 'option_detail.option_group_name_4',
        'langs': [
            {'code': 'th', 'text': 'ชื่อกลุ่มตัวเลือก 4'},
            {'code': 'lo', 'text': 'ຊື່ກຸ່ມຕົວເລືອກ 4'},
            {'code': 'en', 'text': 'Option Group Name 4'},
            {'code': 'cn', 'text': '选项组名称 4'},
            {'code': 'zh', 'text': '選項組名稱 4'},
            {'code': 'ja', 'text': 'オプショングループ名 4'},
            {'code': 'ko', 'text': '옵션 그룹명 4'},
            {'code': 'my', 'text': 'ရွေးချယ်မှုအုပ်စုအမည် 4'},
            {'code': 'km', 'text': 'ឈ្មោះក្រុមជម្រើស 4'},
            {'code': 'vi', 'text': 'Tên nhóm tùy chọn 4'}
        ]
    },
    {
        'code': 'option_detail.option_group_name_5',
        'langs': [
            {'code': 'th', 'text': 'ชื่อกลุ่มตัวเลือก 5'},
            {'code': 'lo', 'text': 'ຊື່ກຸ່ມຕົວເລືອກ 5'},
            {'code': 'en', 'text': 'Option Group Name 5'},
            {'code': 'cn', 'text': '选项组名称 5'},
            {'code': 'zh', 'text': '選項組名稱 5'},
            {'code': 'ja', 'text': 'オプショングループ名 5'},
            {'code': 'ko', 'text': '옵션 그룹명 5'},
            {'code': 'my', 'text': 'ရွေးချယ်မှုအုပ်စုအမည် 5'},
            {'code': 'km', 'text': 'ឈ្មោះក្រុមជម្រើស 5'},
            {'code': 'vi', 'text': 'Tên nhóm tùy chọn 5'}
        ]
    },
    {
        'code': 'option_detail.please_specify_option_group_name',
        'langs': [
            {'code': 'th', 'text': 'กรุณาระบุ ชื่อกลุ่มตัวเลือก'},
            {'code': 'lo', 'text': 'ກະລຸນາລະບຸ ຊື່ກຸ່ມຕົວເລືອກ'},
            {'code': 'en', 'text': 'Please specify option group name'},
            {'code': 'cn', 'text': '请指定选项组名称'},
            {'code': 'zh', 'text': '請指定選項組名稱'},
            {'code': 'ja', 'text': 'オプショングループ名を指定してください'},
            {'code': 'ko', 'text': '옵션 그룹명을 지정해주세요'},
            {'code': 'my', 'text': 'ရွေးချယ်မှုအုပ်စုအမည်ကို သတ်မှတ်ပေးပါ'},
            {'code': 'km', 'text': 'សូមបញ្ជាក់ឈ្មោះក្រុមជម្រើស'},
            {'code': 'vi', 'text': 'Vui lòng chỉ định tên nhóm tùy chọn'}
        ]
    },
    {
        'code': 'option_detail.price',
        'langs': [
            {'code': 'th', 'text': 'ราคา'},
            {'code': 'lo', 'text': 'ລາຄາ'},
            {'code': 'en', 'text': 'Price'},
            {'code': 'cn', 'text': '价格'},
            {'code': 'zh', 'text': '價格'},
            {'code': 'ja', 'text': '価格'},
            {'code': 'ko', 'text': '가격'},
            {'code': 'my', 'text': 'ဈေး'},
            {'code': 'km', 'text': 'តម្លៃ'},
            {'code': 'vi', 'text': 'Giá'}
        ]
    },
    {
        'code': 'option_detail.please_specify_price',
        'langs': [
            {'code': 'th', 'text': 'กรุณาระบุ ราคา'},
            {'code': 'lo', 'text': 'ກະລຸນາລະບຸ ລາຄາ'},
            {'code': 'en', 'text': 'Please specify price'},
            {'code': 'cn', 'text': '请指定价格'},
            {'code': 'zh', 'text': '請指定價格'},
            {'code': 'ja', 'text': '価格を指定してください'},
            {'code': 'ko', 'text': '가격을 지정해주세요'},
            {'code': 'my', 'text': 'ဈေးနှုန်းကို သတ်မှတ်ပေးပါ'},
            {'code': 'km', 'text': 'សូមបញ្ជាក់តម្លៃ'},
            {'code': 'vi', 'text': 'Vui lòng chỉ định giá'}
        ]
    },
    {
        'code': 'option_detail.quantity',
        'langs': [
            {'code': 'th', 'text': 'จำนวน'},
            {'code': 'lo', 'text': 'ຈໍານວນ'},
            {'code': 'en', 'text': 'Quantity'},
            {'code': 'cn', 'text': '数量'},
            {'code': 'zh', 'text': '數量'},
            {'code': 'ja', 'text': '数量'},
            {'code': 'ko', 'text': '수량'},
            {'code': 'my', 'text': 'အရေအတွက်'},
            {'code': 'km', 'text': 'បរិមាណ'},
            {'code': 'vi', 'text': 'Số lượng'}
        ]
    },
    {
        'code': 'option_detail.maximum_quantity',
        'langs': [
            {'code': 'th', 'text': 'จำนวนสูงสุด'},
            {'code': 'lo', 'text': 'ຈໍານວນສູງສຸດ'},
            {'code': 'en', 'text': 'Maximum Quantity'},
            {'code': 'cn', 'text': '最大数量'},
            {'code': 'zh', 'text': '最大數量'},
            {'code': 'ja', 'text': '最大数量'},
            {'code': 'ko', 'text': '최대 수량'},
            {'code': 'my', 'text': 'အများဆုံးအရေအတွက်'},
            {'code': 'km', 'text': 'បរិមាណអតិបរមា'},
            {'code': 'vi', 'text': 'Số lượng tối đa'}
        ]
    },
    {
        'code': 'option_detail.suggest_code',
        'langs': [
            {'code': 'th', 'text': 'รหัสแนะนำ'},
            {'code': 'lo', 'text': 'ລະຫັດແນະນຳ'},
            {'code': 'en', 'text': 'Suggested Code'},
            {'code': 'cn', 'text': '推荐代码'},
            {'code': 'zh', 'text': '推薦代碼'},
            {'code': 'ja', 'text': '推奨コード'},
            {'code': 'ko', 'text': '추천 코드'},
            {'code': 'my', 'text': 'အကြံပြုကုဒ်'},
            {'code': 'km', 'text': 'លេខកូដណែនាំ'},
            {'code': 'vi', 'text': 'Mã gợi ý'}
        ]
    },
    {
        'code': 'option_detail.unit',
        'langs': [
            {'code': 'th', 'text': 'หน่วยนับ'},
            {'code': 'lo', 'text': 'ຫົວໜ່ວຍນັບ'},
            {'code': 'en', 'text': 'Unit'},
            {'code': 'cn', 'text': '单位'},
            {'code': 'zh', 'text': '單位'},
            {'code': 'ja', 'text': '単位'},
            {'code': 'ko', 'text': '단위'},
            {'code': 'my', 'text': 'ယူနစ်'},
            {'code': 'km', 'text': 'ឯកតា'},
            {'code': 'vi', 'text': 'Đơn vị'}
        ]
    },
    {
        'code': 'option_detail.select_immediately',
        'langs': [
            {'code': 'th', 'text': 'เลือกให้เลย'},
            {'code': 'lo', 'text': 'ເລືອກໃຫ້ເລີຍ'},
            {'code': 'en', 'text': 'Select Immediately'},
            {'code': 'cn', 'text': '立即选择'},
            {'code': 'zh', 'text': '立即選擇'},
            {'code': 'ja', 'text': 'すぐに選択'},
            {'code': 'ko', 'text': '즉시 선택'},
            {'code': 'my', 'text': 'ချက်ချင်းရွေးရန်'},
            {'code': 'km', 'text': 'ជ្រើសរើសភ្លាម'},
            {'code': 'vi', 'text': 'Chọn ngay'}
        ]
    },
    {
        'code': 'option_detail.default_value',
        'langs': [
            {'code': 'th', 'text': 'ค่าเริ่มต้น'},
            {'code': 'lo', 'text': 'ຄ່າເລີ່ມຕົ້ນ'},
            {'code': 'en', 'text': 'Default Value'},
            {'code': 'cn', 'text': '默认值'},
            {'code': 'zh', 'text': '預設值'},
            {'code': 'ja', 'text': 'デフォルト値'},
            {'code': 'ko', 'text': '기본값'},
            {'code': 'my', 'text': 'မူလတန်ဖိုး'},
            {'code': 'km', 'text': 'តម្លៃលំនាំដើម'},
            {'code': 'vi', 'text': 'Giá trị mặc định'}
        ]
    }
]

# Add new entries
data.extend(new_entries)

# Write back
with open(r'c:\git\bcaiclouderp\frontend\assets\language.json', 'w', encoding='utf-8') as f:
    json.dump(data, f, ensure_ascii=False, indent=4)

print(f'Added {len(new_entries)} entries for option_detail.dart')
print(f'Total entries now: {len(data)}')
