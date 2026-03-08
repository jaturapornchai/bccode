import json

# Read existing language.json
with open(r'c:\git\bcaiclouderp\frontend\assets\language.json', 'r', encoding='utf-8') as f:
    data = json.load(f)

# Keys to add for import_product_detail_screen.dart
import_product_detail_entries = [
    {
        'code': 'import_product_detail.no_data',
        'langs': [
            {'code': 'th', 'text': 'ไม่มีข้อมูล'},
            {'code': 'lo', 'text': 'ບໍ່ມີຂໍ້ມູນ'},
            {'code': 'en', 'text': 'No data'},
            {'code': 'cn', 'text': '无数据'},
            {'code': 'zh', 'text': '無數據'},
            {'code': 'ja', 'text': 'データなし'},
            {'code': 'ko', 'text': '데이터 없음'},
            {'code': 'my', 'text': 'ဒေတာမရှိပါ'},
            {'code': 'km', 'text': 'គ្មានទិន្នន័យ'},
            {'code': 'vi', 'text': 'Không có dữ liệu'}
        ]
    },
    {
        'code': 'import_product_detail.product_import_details',
        'langs': [
            {'code': 'th', 'text': 'รายละเอียดการนำเข้าสินค้า'},
            {'code': 'lo', 'text': 'ລາຍລະອຽດການນໍາເຂົ້າສິນຄ້າ'},
            {'code': 'en', 'text': 'Product Import Details'},
            {'code': 'cn', 'text': '产品导入详细信息'},
            {'code': 'zh', 'text': '產品導入詳細信息'},
            {'code': 'ja', 'text': '製品インポートの詳細'},
            {'code': 'ko', 'text': '제품 가져오기 세부정보'},
            {'code': 'my', 'text': 'ထုတ်ကုန်တင်သွင်းခြင်းအသေးစိတ်'},
            {'code': 'km', 'text': 'ព័ត៌មានលម្អិតនៃការនាំចូលផលិតផល'},
            {'code': 'vi', 'text': 'Chi tiết nhập khẩu sản phẩm'}
        ]
    },
    {
        'code': 'import_product_detail.total',
        'langs': [
            {'code': 'th', 'text': 'ทั้งหมด'},
            {'code': 'lo', 'text': 'ທັງໝົດ'},
            {'code': 'en', 'text': 'Total'},
            {'code': 'cn', 'text': '全部'},
            {'code': 'zh', 'text': '全部'},
            {'code': 'ja', 'text': '合計'},
            {'code': 'ko', 'text': '전체'},
            {'code': 'my', 'text': 'စုစုပေါင်း'},
            {'code': 'km', 'text': 'សរុប'},
            {'code': 'vi', 'text': 'Tổng cộng'}
        ]
    },
    {
        'code': 'import_product_detail.items',
        'langs': [
            {'code': 'th', 'text': 'รายการ'},
            {'code': 'lo', 'text': 'ລາຍການ'},
            {'code': 'en', 'text': 'Items'},
            {'code': 'cn', 'text': '项目'},
            {'code': 'zh', 'text': '項目'},
            {'code': 'ja', 'text': '項目'},
            {'code': 'ko', 'text': '항목'},
            {'code': 'my', 'text': 'ပစ္စည်းများ'},
            {'code': 'km', 'text': 'ធាតុ'},
            {'code': 'vi', 'text': 'Mục'}
        ]
    },
    {
        'code': 'import_product_detail.showing',
        'langs': [
            {'code': 'th', 'text': 'แสดง'},
            {'code': 'lo', 'text': 'ສະແດງ'},
            {'code': 'en', 'text': 'Showing'},
            {'code': 'cn', 'text': '显示'},
            {'code': 'zh', 'text': '顯示'},
            {'code': 'ja', 'text': '表示'},
            {'code': 'ko', 'text': '표시'},
            {'code': 'my', 'text': 'ပြသခြင်း'},
            {'code': 'km', 'text': 'កំពុងបង្ហាញ'},
            {'code': 'vi', 'text': 'Hiển thị'}
        ]
    },
    {
        'code': 'import_product_detail.search',
        'langs': [
            {'code': 'th', 'text': 'ค้นหา'},
            {'code': 'lo', 'text': 'ຄົ້ນຫາ'},
            {'code': 'en', 'text': 'Search'},
            {'code': 'cn', 'text': '搜索'},
            {'code': 'zh', 'text': '搜尋'},
            {'code': 'ja', 'text': '検索'},
            {'code': 'ko', 'text': '검색'},
            {'code': 'my', 'text': 'ရှာဖွေရန်'},
            {'code': 'km', 'text': 'ស្វែងរក'},
            {'code': 'vi', 'text': 'Tìm kiếm'}
        ]
    },
    {
        'code': 'import_product_detail.search_hint',
        'langs': [
            {'code': 'th', 'text': 'ค้นหาจาก Barcode, รหัสสินค้า, ชื่อสินค้า...'},
            {'code': 'lo', 'text': 'ຄົ້ນຫາຈາກ Barcode, ລະຫັດສິນຄ້າ, ຊື່ສິນຄ້າ...'},
            {'code': 'en', 'text': 'Search by Barcode, Product Code, Product Name...'},
            {'code': 'cn', 'text': '通过条码、产品代码、产品名称搜索...'},
            {'code': 'zh', 'text': '通過條碼、產品代碼、產品名稱搜尋...'},
            {'code': 'ja', 'text': 'バーコード、製品コード、製品名で検索...'},
            {'code': 'ko', 'text': '바코드, 제품 코드, 제품명으로 검색...'},
            {'code': 'my', 'text': 'Barcode၊ ထုတ်ကုန်ကုဒ်၊ ထုတ်ကုန်အမည်ဖြင့် ရှာဖွေရန်...'},
            {'code': 'km', 'text': 'ស្វែងរកតាម Barcode, លេខកូដផលិតផល, ឈ្មោះផលិតផល...'},
            {'code': 'vi', 'text': 'Tìm kiếm theo Barcode, Mã sản phẩm, Tên sản phẩm...'}
        ]
    },
    {
        'code': 'import_product_detail.all',
        'langs': [
            {'code': 'th', 'text': 'ทั้งหมด'},
            {'code': 'lo', 'text': 'ທັງໝົດ'},
            {'code': 'en', 'text': 'All'},
            {'code': 'cn', 'text': '全部'},
            {'code': 'zh', 'text': '全部'},
            {'code': 'ja', 'text': 'すべて'},
            {'code': 'ko', 'text': '전체'},
            {'code': 'my', 'text': 'အားလုံး'},
            {'code': 'km', 'text': 'ទាំងអស់'},
            {'code': 'vi', 'text': 'Tất cả'}
        ]
    },
    {
        'code': 'import_product_detail.unchanged',
        'langs': [
            {'code': 'th', 'text': 'เหมือนกัน'},
            {'code': 'lo', 'text': 'ເໝືອນກັນ'},
            {'code': 'en', 'text': 'Unchanged'},
            {'code': 'cn', 'text': '相同'},
            {'code': 'zh', 'text': '相同'},
            {'code': 'ja', 'text': '変更なし'},
            {'code': 'ko', 'text': '동일'},
            {'code': 'my', 'text': 'တူညီသည်'},
            {'code': 'km', 'text': 'ដូចគ្នា'},
            {'code': 'vi', 'text': 'Giống nhau'}
        ]
    },
    {
        'code': 'import_product_detail.no_data_found',
        'langs': [
            {'code': 'th', 'text': 'ไม่พบข้อมูลที่ค้นหา'},
            {'code': 'lo', 'text': 'ບໍ່ພົບຂໍ້ມູນທີ່ຄົ້ນຫາ'},
            {'code': 'en', 'text': 'No search results found'},
            {'code': 'cn', 'text': '未找到搜索结果'},
            {'code': 'zh', 'text': '未找到搜尋結果'},
            {'code': 'ja', 'text': '検索結果が見つかりません'},
            {'code': 'ko', 'text': '검색 결과를 찾을 수 없습니다'},
            {'code': 'my', 'text': 'ရှာဖွေမှုရလဒ်မတွေ့ပါ'},
            {'code': 'km', 'text': 'រកមិនឃើញលទ្ធផលស្វែងរក'},
            {'code': 'vi', 'text': 'Không tìm thấy kết quả tìm kiếm'}
        ]
    },
    {
        'code': 'import_product_detail.product_code',
        'langs': [
            {'code': 'th', 'text': 'รหัสสินค้า'},
            {'code': 'lo', 'text': 'ລະຫັດສິນຄ້າ'},
            {'code': 'en', 'text': 'Product Code'},
            {'code': 'cn', 'text': '产品代码'},
            {'code': 'zh', 'text': '產品代碼'},
            {'code': 'ja', 'text': '製品コード'},
            {'code': 'ko', 'text': '제품 코드'},
            {'code': 'my', 'text': 'ထုတ်ကုန်ကုဒ်'},
            {'code': 'km', 'text': 'លេខកូដផលិតផល'},
            {'code': 'vi', 'text': 'Mã sản phẩm'}
        ]
    },
    {
        'code': 'import_product_detail.product_name',
        'langs': [
            {'code': 'th', 'text': 'ชื่อสินค้า'},
            {'code': 'lo', 'text': 'ຊື່ສິນຄ້າ'},
            {'code': 'en', 'text': 'Product Name'},
            {'code': 'cn', 'text': '产品名称'},
            {'code': 'zh', 'text': '產品名稱'},
            {'code': 'ja', 'text': '製品名'},
            {'code': 'ko', 'text': '제품명'},
            {'code': 'my', 'text': 'ထုတ်ကုန်အမည်'},
            {'code': 'km', 'text': 'ឈ្មោះផលិតផល'},
            {'code': 'vi', 'text': 'Tên sản phẩm'}
        ]
    },
    {
        'code': 'import_product_detail.unit',
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
    }
]

# Keys to add for database_master_info.dart
database_master_info_entries = [
    {
        'code': 'database_master_info.data_by_type',
        'langs': [
            {'code': 'th', 'text': 'ข้อมูลตามประเภท'},
            {'code': 'lo', 'text': 'ຂໍ້ມູນຕາມປະເພດ'},
            {'code': 'en', 'text': 'Data by Type'},
            {'code': 'cn', 'text': '按类型分类数据'},
            {'code': 'zh', 'text': '按類型分類數據'},
            {'code': 'ja', 'text': 'タイプ別データ'},
            {'code': 'ko', 'text': '유형별 데이터'},
            {'code': 'my', 'text': 'အမျိုးအစားအလိုက်ဒေတာ'},
            {'code': 'km', 'text': 'ទិន្នន័យតាមប្រភេទ'},
            {'code': 'vi', 'text': 'Dữ liệu theo loại'}
        ]
    },
    {
        'code': 'database_master_info.refresh',
        'langs': [
            {'code': 'th', 'text': 'รีเฟรช'},
            {'code': 'lo', 'text': 'ໂຫຼດໃໝ່'},
            {'code': 'en', 'text': 'Refresh'},
            {'code': 'cn', 'text': '刷新'},
            {'code': 'zh', 'text': '刷新'},
            {'code': 'ja', 'text': '更新'},
            {'code': 'ko', 'text': '새로고침'},
            {'code': 'my', 'text': 'ပြန်လည်စတင်ရန်'},
            {'code': 'km', 'text': 'ផ្ទុកឡើងវិញ'},
            {'code': 'vi', 'text': 'Làm mới'}
        ]
    },
    {
        'code': 'database_master_info.data_type',
        'langs': [
            {'code': 'th', 'text': 'ประเภทข้อมูล'},
            {'code': 'lo', 'text': 'ປະເພດຂໍ້ມູນ'},
            {'code': 'en', 'text': 'Data Type'},
            {'code': 'cn', 'text': '数据类型'},
            {'code': 'zh', 'text': '數據類型'},
            {'code': 'ja', 'text': 'データ型'},
            {'code': 'ko', 'text': '데이터 유형'},
            {'code': 'my', 'text': 'ဒေတာအမျိုးအစား'},
            {'code': 'km', 'text': 'ប្រភេទទិន្នន័យ'},
            {'code': 'vi', 'text': 'Loại dữ liệu'}
        ]
    },
    {
        'code': 'database_master_info.item_count',
        'langs': [
            {'code': 'th', 'text': 'จำนวนรายการ'},
            {'code': 'lo', 'text': 'ຈໍານວນລາຍການ'},
            {'code': 'en', 'text': 'Item Count'},
            {'code': 'cn', 'text': '项目数'},
            {'code': 'zh', 'text': '項目數'},
            {'code': 'ja', 'text': '項目数'},
            {'code': 'ko', 'text': '항목 수'},
            {'code': 'my', 'text': 'ပစ္စည်းအရေအတွက်'},
            {'code': 'km', 'text': 'ចំនួនធាតុ'},
            {'code': 'vi', 'text': 'Số lượng mục'}
        ]
    },
    {
        'code': 'database_master_info.product',
        'langs': [
            {'code': 'th', 'text': 'สินค้า (product)'},
            {'code': 'lo', 'text': 'ສິນຄ້າ (product)'},
            {'code': 'en', 'text': 'Product'},
            {'code': 'cn', 'text': '产品 (product)'},
            {'code': 'zh', 'text': '產品 (product)'},
            {'code': 'ja', 'text': '製品 (product)'},
            {'code': 'ko', 'text': '제품 (product)'},
            {'code': 'my', 'text': 'ထုတ်ကုန် (product)'},
            {'code': 'km', 'text': 'ផលិតផល (product)'},
            {'code': 'vi', 'text': 'Sản phẩm (product)'}
        ]
    },
    {
        'code': 'database_master_info.productbarcode',
        'langs': [
            {'code': 'th', 'text': 'บาร์โค้ดสินค้า (productbarcode)'},
            {'code': 'lo', 'text': 'ບາໂຄດສິນຄ້າ (productbarcode)'},
            {'code': 'en', 'text': 'Product Barcode'},
            {'code': 'cn', 'text': '产品条码 (productbarcode)'},
            {'code': 'zh', 'text': '產品條碼 (productbarcode)'},
            {'code': 'ja', 'text': '製品バーコード (productbarcode)'},
            {'code': 'ko', 'text': '제품 바코드 (productbarcode)'},
            {'code': 'my', 'text': 'ထုတ်ကုန်ဘားကုဒ် (productbarcode)'},
            {'code': 'km', 'text': 'បាកូដផលិតផល (productbarcode)'},
            {'code': 'vi', 'text': 'Mã vạch sản phẩm (productbarcode)'}
        ]
    },
    {
        'code': 'database_master_info.customer',
        'langs': [
            {'code': 'th', 'text': 'ลูกค้า (customer)'},
            {'code': 'lo', 'text': 'ລູກຄ້າ (customer)'},
            {'code': 'en', 'text': 'Customer'},
            {'code': 'cn', 'text': '客户 (customer)'},
            {'code': 'zh', 'text': '客戶 (customer)'},
            {'code': 'ja', 'text': '顧客 (customer)'},
            {'code': 'ko', 'text': '고객 (customer)'},
            {'code': 'my', 'text': 'ဖောက်သည် (customer)'},
            {'code': 'km', 'text': 'អតិថិជន (customer)'},
            {'code': 'vi', 'text': 'Khách hàng (customer)'}
        ]
    },
    {
        'code': 'database_master_info.debtor',
        'langs': [
            {'code': 'th', 'text': 'ลูกหนี้ (debtor)'},
            {'code': 'lo', 'text': 'ລູກໜີ້ (debtor)'},
            {'code': 'en', 'text': 'Debtor'},
            {'code': 'cn', 'text': '债务人 (debtor)'},
            {'code': 'zh', 'text': '債務人 (debtor)'},
            {'code': 'ja', 'text': '債務者 (debtor)'},
            {'code': 'ko', 'text': '채무자 (debtor)'},
            {'code': 'my', 'text': 'ကြွေးရှင် (debtor)'},
            {'code': 'km', 'text': 'កម្ចីប្រាក់ (debtor)'},
            {'code': 'vi', 'text': 'Người nợ (debtor)'}
        ]
    },
    {
        'code': 'database_master_info.creditor',
        'langs': [
            {'code': 'th', 'text': 'เจ้าหนี้ (creditor)'},
            {'code': 'lo', 'text': 'ເຈົ້າໜີ້ (creditor)'},
            {'code': 'en', 'text': 'Creditor'},
            {'code': 'cn', 'text': '债权人 (creditor)'},
            {'code': 'zh', 'text': '債權人 (creditor)'},
            {'code': 'ja', 'text': '債権者 (creditor)'},
            {'code': 'ko', 'text': '채권자 (creditor)'},
            {'code': 'my', 'text': 'ကြွေးရှင် (creditor)'},
            {'code': 'km', 'text': 'ម្ចាស់បំណុល (creditor)'},
            {'code': 'vi', 'text': 'Chủ nợ (creditor)'}
        ]
    },
    {
        'code': 'database_master_info.ic_warehouse',
        'langs': [
            {'code': 'th', 'text': 'คลังสินค้า (ic_warehouse)'},
            {'code': 'lo', 'text': 'ຄັງສິນຄ້າ (ic_warehouse)'},
            {'code': 'en', 'text': 'Warehouse'},
            {'code': 'cn', 'text': '仓库 (ic_warehouse)'},
            {'code': 'zh', 'text': '倉庫 (ic_warehouse)'},
            {'code': 'ja', 'text': '倉庫 (ic_warehouse)'},
            {'code': 'ko', 'text': '창고 (ic_warehouse)'},
            {'code': 'my', 'text': 'ဂိုဒေါင် (ic_warehouse)'},
            {'code': 'km', 'text': 'ឃ្លាំង (ic_warehouse)'},
            {'code': 'vi', 'text': 'Kho hàng (ic_warehouse)'}
        ]
    },
    {
        'code': 'database_master_info.ic_shelf',
        'langs': [
            {'code': 'th', 'text': 'ชั้นวางสินค้า (ic_shelf)'},
            {'code': 'lo', 'text': 'ຊັ້ນວາງສິນຄ້າ (ic_shelf)'},
            {'code': 'en', 'text': 'Shelf'},
            {'code': 'cn', 'text': '货架 (ic_shelf)'},
            {'code': 'zh', 'text': '貨架 (ic_shelf)'},
            {'code': 'ja', 'text': '棚 (ic_shelf)'},
            {'code': 'ko', 'text': '선반 (ic_shelf)'},
            {'code': 'my', 'text': 'စင် (ic_shelf)'},
            {'code': 'km', 'text': 'ធ្នើ (ic_shelf)'},
            {'code': 'vi', 'text': 'Kệ hàng (ic_shelf)'}
        ]
    },
    {
        'code': 'database_master_info.doc',
        'langs': [
            {'code': 'th', 'text': 'เอกสาร (doc)'},
            {'code': 'lo', 'text': 'ເອກະສານ (doc)'},
            {'code': 'en', 'text': 'Document'},
            {'code': 'cn', 'text': '文档 (doc)'},
            {'code': 'zh', 'text': '文檔 (doc)'},
            {'code': 'ja', 'text': '文書 (doc)'},
            {'code': 'ko', 'text': '문서 (doc)'},
            {'code': 'my', 'text': 'စာရွက်စာတမ်း (doc)'},
            {'code': 'km', 'text': 'ឯកសារ (doc)'},
            {'code': 'vi', 'text': 'Tài liệu (doc)'}
        ]
    },
    {
        'code': 'database_master_info.docdetail',
        'langs': [
            {'code': 'th', 'text': 'รายละเอียดเอกสาร (docdetail)'},
            {'code': 'lo', 'text': 'ລາຍລະອຽດເອກະສານ (docdetail)'},
            {'code': 'en', 'text': 'Document Detail'},
            {'code': 'cn', 'text': '文档详细信息 (docdetail)'},
            {'code': 'zh', 'text': '文檔詳細信息 (docdetail)'},
            {'code': 'ja', 'text': '文書詳細 (docdetail)'},
            {'code': 'ko', 'text': '문서 세부정보 (docdetail)'},
            {'code': 'my', 'text': 'စာရွက်စာတမ်းအသေးစိတ် (docdetail)'},
            {'code': 'km', 'text': 'ព័ត៌មានលម្អិតឯកសារ (docdetail)'},
            {'code': 'vi', 'text': 'Chi tiết tài liệu (docdetail)'}
        ]
    }
]

# Add all new entries
data.extend(import_product_detail_entries)
data.extend(database_master_info_entries)

# Write back
with open(r'c:\git\bcaiclouderp\frontend\assets\language.json', 'w', encoding='utf-8') as f:
    json.dump(data, f, ensure_ascii=False, indent=4)

print(f'Added {len(import_product_detail_entries)} entries for import_product_detail_screen.dart')
print(f'Added {len(database_master_info_entries)} entries for database_master_info.dart')
print(f'Total entries now: {len(data)}')
