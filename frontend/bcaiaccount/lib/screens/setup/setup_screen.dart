import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:dio/dio.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/utils/backend_url_manager.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';
import 'package:url_launcher/url_launcher.dart';

/// Multi-language text สำหรับหน้า Setup — hardcode 9 ภาษา
/// เพราะตอนเปิดหน้านี้อาจยังไม่มี backend URL ที่ถูกต้อง
String _t(String key) {
  final lang = global.userLanguage;
  return _setupTexts[key]?[lang] ?? _setupTexts[key]?['en'] ?? key;
}

const Map<String, Map<String, String>> _setupTexts = {
  'setup_config': {
    'en': 'Setup Config', 'th': 'ตั้งค่าระบบ', 'lo': 'ຕັ້ງຄ່າລະບົບ',
    'zh': '系统设置', 'ja': 'システム設定', 'ko': '시스템 설정',
    'my': 'စနစ်ထည့်သွင်းခြင်း', 'km': 'ការកំណត់ប្រព័ន្ធ', 'vi': 'Cài đặt hệ thống',
  },
  'backend_connection': {
    'en': 'Backend Connection', 'th': 'เชื่อมต่อ Backend', 'lo': 'ເຊື່ອມຕໍ່ Backend',
    'zh': '后端连接', 'ja': 'バックエンド接続', 'ko': '백엔드 연결',
    'my': 'Backend ချိတ်ဆက်ခြင်း', 'km': 'ការតភ្ជាប់ Backend', 'vi': 'Kết nối Backend',
  },
  'backend_desc': {
    'en': 'Enter backend server (goapi) URL and Setup password',
    'th': 'ระบุ URL ของ backend server (goapi) และรหัสผ่าน Setup',
    'lo': 'ລະບຸ URL ຂອງ backend server (goapi) ແລະລະຫັດຜ່ານ Setup',
    'zh': '输入后端服务器 (goapi) URL 和设置密码',
    'ja': 'バックエンドサーバー (goapi) URLとセットアップパスワードを入力',
    'ko': '백엔드 서버 (goapi) URL과 설정 비밀번호를 입력하세요',
    'my': 'Backend server (goapi) URL နှင့် Setup စကားဝှက်ထည့်ပါ',
    'km': 'បញ្ចូល URL សឺវើ backend (goapi) និងពាក្យសម្ងាត់ Setup',
    'vi': 'Nhập URL máy chủ backend (goapi) và mật khẩu Setup',
  },
  'setup_password': {
    'en': 'Setup Password', 'th': 'รหัสผ่าน Setup', 'lo': 'ລະຫັດຜ່ານ Setup',
    'zh': '设置密码', 'ja': 'セットアップパスワード', 'ko': '설정 비밀번호',
    'my': 'Setup စကားဝှက်', 'km': 'ពាក្យសម្ងាត់ Setup', 'vi': 'Mật khẩu Setup',
  },
  'login_setup': {
    'en': 'Login Setup', 'th': 'เข้าสู่ระบบ Setup', 'lo': 'ເຂົ້າສູ່ລະບົບ Setup',
    'zh': '登录设置', 'ja': 'セットアップにログイン', 'ko': '설정 로그인',
    'my': 'Setup ဝင်ရောက်', 'km': 'ចូល Setup', 'vi': 'Đăng nhập Setup',
  },
  'connected': {
    'en': 'Connected', 'th': 'เชื่อมต่อแล้ว', 'lo': 'ເຊື່ອມຕໍ່ແລ້ວ',
    'zh': '已连接', 'ja': '接続済み', 'ko': '연결됨',
    'my': 'ချိတ်ဆက်ပြီး', 'km': 'បានតភ្ជាប់', 'vi': 'Đã kết nối',
  },
  'logout': {
    'en': 'Logout', 'th': 'ออกจากระบบ', 'lo': 'ອອກຈາກລະບົບ',
    'zh': '退出', 'ja': 'ログアウト', 'ko': '로그아웃',
    'my': 'ထွက်ရန်', 'km': 'ចាកចេញ', 'vi': 'Đăng xuất',
  },
  'please_enter_url': {
    'en': 'Please enter Backend URL', 'th': 'กรุณาระบุ Backend URL', 'lo': 'ກະລຸນາລະບຸ Backend URL',
    'zh': '请输入后端URL', 'ja': 'バックエンドURLを入力してください', 'ko': '백엔드 URL을 입력하세요',
    'my': 'Backend URL ထည့်ပါ', 'km': 'សូមបញ្ចូល Backend URL', 'vi': 'Vui lòng nhập Backend URL',
  },
  'please_enter_password': {
    'en': 'Please enter password', 'th': 'กรุณาระบุรหัสผ่าน', 'lo': 'ກະລຸນາລະບຸລະຫັດຜ່ານ',
    'zh': '请输入密码', 'ja': 'パスワードを入力してください', 'ko': '비밀번호를 입력하세요',
    'my': 'စကားဝှက်ထည့်ပါ', 'km': 'សូមបញ្ចូលពាក្យសម្ងាត់', 'vi': 'Vui lòng nhập mật khẩu',
  },
  'verifying': {
    'en': 'Verifying password...', 'th': 'กำลังยืนยันรหัสผ่าน...', 'lo': 'ກຳລັງຢືນຢັນລະຫັດຜ່ານ...',
    'zh': '正在验证密码...', 'ja': 'パスワードを確認中...', 'ko': '비밀번호 확인 중...',
    'my': 'စကားဝှက်စစ်ဆေးနေသည်...', 'km': 'កំពុងផ្ទៀងផ្ទាត់ពាក្យសម្ងាត់...', 'vi': 'Đang xác minh mật khẩu...',
  },
  'verify_success': {
    'en': 'Password verified', 'th': 'ยืนยันรหัสผ่านสำเร็จ', 'lo': 'ຢືນຢັນລະຫັດຜ່ານສຳເລັດ',
    'zh': '密码验证成功', 'ja': 'パスワード確認完了', 'ko': '비밀번호 확인 완료',
    'my': 'စကားဝှက်အတည်ပြုပြီး', 'km': 'បានផ្ទៀងផ្ទាត់ពាក្យសម្ងាត់', 'vi': 'Đã xác minh mật khẩu',
  },
  'wrong_password': {
    'en': 'Wrong password', 'th': 'รหัสผ่านไม่ถูกต้อง', 'lo': 'ລະຫັດຜ່ານບໍ່ຖືກຕ້ອງ',
    'zh': '密码错误', 'ja': 'パスワードが正しくありません', 'ko': '비밀번호가 틀렸습니다',
    'my': 'စကားဝှက်မှားနေသည်', 'km': 'ពាក្យសម្ងាត់មិនត្រឹមត្រូវ', 'vi': 'Sai mật khẩu',
  },
  'connection_failed': {
    'en': 'Connection failed', 'th': 'เชื่อมต่อ backend ล้มเหลว', 'lo': 'ເຊື່ອມຕໍ່ backend ລົ້ມເຫລວ',
    'zh': '连接失败', 'ja': '接続に失敗しました', 'ko': '연결 실패',
    'my': 'ချိတ်ဆက်မှုမအောင်မြင်ပါ', 'km': 'ការតភ្ជាប់បរាជ័យ', 'vi': 'Kết nối thất bại',
  },
  'error_occurred': {
    'en': 'An error occurred', 'th': 'เกิดข้อผิดพลาด', 'lo': 'ເກີດຂໍ້ຜິດພາດ',
    'zh': '发生错误', 'ja': 'エラーが発生しました', 'ko': '오류가 발생했습니다',
    'my': 'အမှားတစ်ခုဖြစ်ပွားခဲ့သည်', 'km': 'មានកំហុសកើតឡើង', 'vi': 'Đã xảy ra lỗi',
  },
  'load_config_success': {
    'en': 'Config loaded', 'th': 'โหลด config สำเร็จ', 'lo': 'ໂຫລດ config ສຳເລັດ',
    'zh': '配置已加载', 'ja': '設定を読み込みました', 'ko': '설정 로드 완료',
    'my': 'Config ဖတ်ပြီး', 'km': 'បានផ្ទុក config', 'vi': 'Đã tải config',
  },
  'load_config_failed': {
    'en': 'Load config failed', 'th': 'โหลด config ล้มเหลว', 'lo': 'ໂຫລດ config ລົ້ມເຫລວ',
    'zh': '加载配置失败', 'ja': '設定の読み込みに失敗', 'ko': '설정 로드 실패',
    'my': 'Config ဖတ်မရပါ', 'km': 'ការផ្ទុក config បរាជ័យ', 'vi': 'Tải config thất bại',
  },
  'saving_config': {
    'en': 'Saving config...', 'th': 'กำลังบันทึก config...', 'lo': 'ກຳລັງບັນທຶກ config...',
    'zh': '正在保存配置...', 'ja': '設定を保存中...', 'ko': '설정 저장 중...',
    'my': 'Config သိမ်းနေသည်...', 'km': 'កំពុងរក្សាទុក config...', 'vi': 'Đang lưu config...',
  },
  'save_success': {
    'en': 'Saved successfully', 'th': 'บันทึกสำเร็จ', 'lo': 'ບັນທຶກສຳເລັດ',
    'zh': '保存成功', 'ja': '保存しました', 'ko': '저장 완료',
    'my': 'သိမ်းဆည်းပြီး', 'km': 'បានរក្សាទុក', 'vi': 'Đã lưu thành công',
  },
  'save_failed': {
    'en': 'Save failed', 'th': 'บันทึกล้มเหลว', 'lo': 'ບັນທຶກລົ້ມເຫລວ',
    'zh': '保存失败', 'ja': '保存に失敗', 'ko': '저장 실패',
    'my': 'သိမ်းဆည်းမှုမအောင်မြင်ပါ', 'km': 'ការរក្សាទុកបរាជ័យ', 'vi': 'Lưu thất bại',
  },
  'seeding': {
    'en': 'Seeding config from environment...', 'th': 'กำลัง seed config จาก environment...',
    'lo': 'ກຳລັງ seed config ຈາກ environment...', 'zh': '正在从环境变量导入配置...',
    'ja': '環境変数から設定をインポート中...', 'ko': '환경 변수에서 설정 가져오는 중...',
    'my': 'Environment မှ config seed လုပ်နေသည်...', 'km': 'កំពុង seed config ពី environment...',
    'vi': 'Đang seed config từ environment...',
  },
  'testing': {
    'en': 'Testing...', 'th': 'กำลังทดสอบ...', 'lo': 'ກຳລັງທົດສອບ...',
    'zh': '测试中...', 'ja': 'テスト中...', 'ko': '테스트 중...',
    'my': 'စမ်းသပ်နေသည်...', 'km': 'កំពុងសាកល្បង...', 'vi': 'Đang kiểm tra...',
  },
  'test_connection': {
    'en': 'Test Connection', 'th': 'ทดสอบการเชื่อมต่อ', 'lo': 'ທົດສອບການເຊື່ອມຕໍ່',
    'zh': '测试连接', 'ja': '接続テスト', 'ko': '연결 테스트',
    'my': 'ချိတ်ဆက်မှုစမ်းသပ်', 'km': 'សាកល្បងការតភ្ជាប់', 'vi': 'Kiểm tra kết nối',
  },
  'success': {
    'en': 'Success', 'th': 'สำเร็จ', 'lo': 'ສຳເລັດ',
    'zh': '成功', 'ja': '成功', 'ko': '성공',
    'my': 'အောင်မြင်', 'km': 'ជោគជ័យ', 'vi': 'Thành công',
  },
  'failed': {
    'en': 'Failed', 'th': 'ล้มเหลว', 'lo': 'ລົ້ມເຫລວ',
    'zh': '失败', 'ja': '失敗', 'ko': '실패',
    'my': 'မအောင်မြင်', 'km': 'បរាជ័យ', 'vi': 'Thất bại',
  },
  'change_password': {
    'en': 'Change Password', 'th': 'เปลี่ยนรหัสผ่าน', 'lo': 'ປ່ຽນລະຫັດຜ່ານ',
    'zh': '修改密码', 'ja': 'パスワード変更', 'ko': '비밀번호 변경',
    'my': 'စကားဝှက်ပြောင်း', 'km': 'ប្តូរពាក្យសម្ងាត់', 'vi': 'Đổi mật khẩu',
  },
  'change_password_setup': {
    'en': 'Change Setup Password', 'th': 'เปลี่ยนรหัสผ่าน Setup', 'lo': 'ປ່ຽນລະຫັດຜ່ານ Setup',
    'zh': '修改设置密码', 'ja': 'セットアップパスワード変更', 'ko': '설정 비밀번호 변경',
    'my': 'Setup စကားဝှက်ပြောင်း', 'km': 'ប្តូរពាក្យសម្ងាត់ Setup', 'vi': 'Đổi mật khẩu Setup',
  },
  'current_password': {
    'en': 'Current Password', 'th': 'รหัสผ่านเดิม', 'lo': 'ລະຫັດຜ່ານເກົ່າ',
    'zh': '当前密码', 'ja': '現在のパスワード', 'ko': '현재 비밀번호',
    'my': 'လက်ရှိစကားဝှက်', 'km': 'ពាក្យសម្ងាត់បច្ចុប្បន្ន', 'vi': 'Mật khẩu hiện tại',
  },
  'new_password': {
    'en': 'New Password', 'th': 'รหัสผ่านใหม่', 'lo': 'ລະຫັດຜ່ານໃໝ່',
    'zh': '新密码', 'ja': '新しいパスワード', 'ko': '새 비밀번호',
    'my': 'စကားဝှက်အသစ်', 'km': 'ពាក្យសម្ងាត់ថ្មី', 'vi': 'Mật khẩu mới',
  },
  'cancel': {
    'en': 'Cancel', 'th': 'ยกเลิก', 'lo': 'ຍົກເລີກ',
    'zh': '取消', 'ja': 'キャンセル', 'ko': '취소',
    'my': 'ပယ်ဖျက်', 'km': 'បោះបង់', 'vi': 'Hủy',
  },
  'change': {
    'en': 'Change', 'th': 'เปลี่ยน', 'lo': 'ປ່ຽນ',
    'zh': '修改', 'ja': '変更', 'ko': '변경',
    'my': 'ပြောင်း', 'km': 'ប្តូរ', 'vi': 'Đổi',
  },
  'password_changed': {
    'en': 'Password changed successfully', 'th': 'เปลี่ยนรหัสผ่านสำเร็จ', 'lo': 'ປ່ຽນລະຫັດຜ່ານສຳເລັດ',
    'zh': '密码已修改', 'ja': 'パスワードを変更しました', 'ko': '비밀번호 변경 완료',
    'my': 'စကားဝှက်ပြောင်းပြီး', 'km': 'បានប្តូរពាក្យសម្ងាត់', 'vi': 'Đã đổi mật khẩu',
  },
  'password_change_failed': {
    'en': 'Password change failed', 'th': 'เปลี่ยนรหัสผ่านล้มเหลว', 'lo': 'ປ່ຽນລະຫັດຜ່ານລົ້ມເຫລວ',
    'zh': '修改密码失败', 'ja': 'パスワード変更に失敗', 'ko': '비밀번호 변경 실패',
    'my': 'စကားဝှက်ပြောင်းမှုမအောင်မြင်', 'km': 'ការប្តូរពាក្យសម្ងាត់បរាជ័យ', 'vi': 'Đổi mật khẩu thất bại',
  },
  'seed_from_env': {
    'en': 'Seed from Environment', 'th': 'Seed จาก Environment', 'lo': 'Seed ຈາກ Environment',
    'zh': '从环境变量导入', 'ja': '環境変数からインポート', 'ko': '환경 변수에서 가져오기',
    'my': 'Environment မှ Seed', 'km': 'Seed ពី Environment', 'vi': 'Seed từ Environment',
  },
  'save_config': {
    'en': 'Save Config', 'th': 'บันทึก Config', 'lo': 'ບັນທຶກ Config',
    'zh': '保存配置', 'ja': '設定を保存', 'ko': '설정 저장',
    'my': 'Config သိမ်း', 'km': 'រក្សាទុក Config', 'vi': 'Lưu Config',
  },
  // Feature: Test All / Backup / Restore / Confirm / Validate
  'test_all': {
    'en': 'Test All', 'th': 'ทดสอบทั้งหมด', 'lo': 'ທົດສອບທັງໝົດ',
    'zh': '全部测试', 'ja': '全てテスト', 'ko': '전체 테스트',
    'my': 'အားလုံးစမ်းသပ်', 'km': 'សាកល្បងទាំងអស់', 'vi': 'Kiểm tra tất cả',
  },
  'testing_all': {
    'en': 'Testing all connections...', 'th': 'กำลังทดสอบการเชื่อมต่อทั้งหมด...', 'lo': 'ກຳລັງທົດສອບການເຊື່ອມຕໍ່ທັງໝົດ...',
    'zh': '正在测试所有连接...', 'ja': '全接続をテスト中...', 'ko': '모든 연결 테스트 중...',
    'my': 'ချိတ်ဆက်မှုအားလုံးစမ်းသပ်နေသည်...', 'km': 'កំពុងសាកល្បងការតភ្ជាប់ទាំងអស់...', 'vi': 'Đang kiểm tra tất cả kết nối...',
  },
  'test_all_done': {
    'en': 'All tests completed', 'th': 'ทดสอบเสร็จทั้งหมด', 'lo': 'ທົດສອບສຳເລັດທັງໝົດ',
    'zh': '全部测试完成', 'ja': '全テスト完了', 'ko': '전체 테스트 완료',
    'my': 'စမ်းသပ်မှုအားလုံးပြီး', 'km': 'ការសាកល្បងទាំងអស់បានបញ្ចប់', 'vi': 'Đã kiểm tra tất cả',
  },
  'export_config': {
    'en': 'Export Config', 'th': 'ส่งออก Config', 'lo': 'ສົ່ງອອກ Config',
    'zh': '导出配置', 'ja': '設定をエクスポート', 'ko': '설정 내보내기',
    'my': 'Config တင်ပို့', 'km': 'នាំចេញ Config', 'vi': 'Xuất Config',
  },
  'import_config': {
    'en': 'Import Config', 'th': 'นำเข้า Config', 'lo': 'ນຳເຂົ້າ Config',
    'zh': '导入配置', 'ja': '設定をインポート', 'ko': '설정 가져오기',
    'my': 'Config သွင်းယူ', 'km': 'នាំចូល Config', 'vi': 'Nhập Config',
  },
  'config_copied': {
    'en': 'Config copied to clipboard', 'th': 'คัดลอก Config แล้ว', 'lo': 'ຄັດລອກ Config ແລ້ວ',
    'zh': '配置已复制到剪贴板', 'ja': '設定をクリップボードにコピーしました', 'ko': '설정이 클립보드에 복사되었습니다',
    'my': 'Config ကူးယူပြီး', 'km': 'បានចម្លង Config', 'vi': 'Đã sao chép Config',
  },
  'config_imported': {
    'en': 'Config imported successfully', 'th': 'นำเข้า Config สำเร็จ', 'lo': 'ນຳເຂົ້າ Config ສຳເລັດ',
    'zh': '配置导入成功', 'ja': '設定のインポート完了', 'ko': '설정 가져오기 완료',
    'my': 'Config သွင်းယူပြီး', 'km': 'បាននាំចូល Config', 'vi': 'Đã nhập Config',
  },
  'import_failed': {
    'en': 'Import failed: invalid JSON', 'th': 'นำเข้าล้มเหลว: JSON ไม่ถูกต้อง', 'lo': 'ນຳເຂົ້າລົ້ມເຫລວ: JSON ບໍ່ຖືກຕ້ອງ',
    'zh': '导入失败：JSON格式错误', 'ja': 'インポート失敗：無効なJSON', 'ko': '가져오기 실패: 잘못된 JSON',
    'my': 'သွင်းယူမှုမအောင်မြင်: JSON မမှန်', 'km': 'នាំចូលបរាជ័យ: JSON មិនត្រឹមត្រូវ', 'vi': 'Nhập thất bại: JSON không hợp lệ',
  },
  'paste_json': {
    'en': 'Paste JSON config here', 'th': 'วาง JSON config ที่นี่', 'lo': 'ວາງ JSON config ທີ່ນີ້',
    'zh': '在此粘贴JSON配置', 'ja': 'JSON設定をここに貼り付け', 'ko': 'JSON 설정을 여기에 붙여넣으세요',
    'my': 'JSON config ဒီမှာထည့်ပါ', 'km': 'បិទភ្ជាប់ JSON config នៅទីនេះ', 'vi': 'Dán JSON config tại đây',
  },
  'copy_to_clipboard': {
    'en': 'Copy to Clipboard', 'th': 'คัดลอกไปคลิปบอร์ด', 'lo': 'ຄັດລອກໄປ Clipboard',
    'zh': '复制到剪贴板', 'ja': 'クリップボードにコピー', 'ko': '클립보드에 복사',
    'my': 'Clipboard သို့ကူးယူ', 'km': 'ចម្លងទៅ Clipboard', 'vi': 'Sao chép vào Clipboard',
  },
  'backend_url': {
    'en': 'Backend URL', 'th': 'Backend URL', 'lo': 'Backend URL',
    'zh': '后端URL', 'ja': 'バックエンドURL', 'ko': '백엔드 URL',
    'my': 'Backend URL', 'km': 'Backend URL', 'vi': 'Backend URL',
  },
  'scanning_in_progress': {
    'en': 'Scanning...', 'th': 'กำลังสแกน...', 'lo': 'ກຳລັງສະແກນ...',
    'zh': '扫描中...', 'ja': 'スキャン中...', 'ko': '스캔 중...',
    'my': 'စကန်ဖတ်နေသည်...', 'km': 'កំពុងស្កេន...', 'vi': 'Đang quét...',
  },
  'scan_latest': {
    'en': 'Latest Scan', 'th': 'Scan ล่าสุด', 'lo': 'Scan ລ່າສຸດ',
    'zh': '最新扫描', 'ja': '最新スキャン', 'ko': '최근 스캔',
    'my': 'နောက်ဆုံးစကန်', 'km': 'ស្កេនចុងក្រោយ', 'vi': 'Quét gần nhất',
  },
  'confirm_save': {
    'en': 'Confirm Save', 'th': 'ยืนยันการบันทึก', 'lo': 'ຢືນຢັນການບັນທຶກ',
    'zh': '确认保存', 'ja': '保存を確認', 'ko': '저장 확인',
    'my': 'သိမ်းဆည်းမှုအတည်ပြု', 'km': 'បញ្ជាក់ការរក្សាទុក', 'vi': 'Xác nhận lưu',
  },
  'confirm_save_msg': {
    'en': 'Save config to backend? This will overwrite existing settings.',
    'th': 'บันทึก config ไปที่ backend? จะเขียนทับค่าที่มีอยู่',
    'lo': 'ບັນທຶກ config ໄປ backend? ຈະຂຽນທັບຄ່າທີ່ມີຢູ່',
    'zh': '保存配置到后端？这将覆盖现有设置。',
    'ja': 'バックエンドに設定を保存しますか？既存の設定が上書きされます。',
    'ko': '백엔드에 설정을 저장하시겠습니까? 기존 설정을 덮어씁니다.',
    'my': 'Backend သို့ config သိမ်းမလား? လက်ရှိ setting ကို overwrite လုပ်ပါမည်။',
    'km': 'រក្សាទុក config ទៅ backend? នឹងសរសេរជាន់លើការកំណត់ដែលមាន។',
    'vi': 'Lưu config vào backend? Sẽ ghi đè cài đặt hiện tại.',
  },
  'save': {
    'en': 'Save', 'th': 'บันทึก', 'lo': 'ບັນທຶກ',
    'zh': '保存', 'ja': '保存', 'ko': '저장',
    'my': 'သိမ်းဆည်း', 'km': 'រក្សាទុក', 'vi': 'Lưu',
  },
  'validation_error': {
    'en': 'Validation Error', 'th': 'ข้อมูลไม่ถูกต้อง', 'lo': 'ຂໍ້ມູນບໍ່ຖືກຕ້ອງ',
    'zh': '验证错误', 'ja': 'バリデーションエラー', 'ko': '유효성 검사 오류',
    'my': 'အချက်အလက်မမှန်ကန်', 'km': 'កំហុសផ្ទៀងផ្ទាត់', 'vi': 'Lỗi xác thực',
  },
  'invalid_port': {
    'en': 'Invalid port number', 'th': 'หมายเลข port ไม่ถูกต้อง', 'lo': 'ໝາຍເລກ port ບໍ່ຖືກຕ້ອງ',
    'zh': '端口号无效', 'ja': 'ポート番号が無効です', 'ko': '잘못된 포트 번호',
    'my': 'Port နံပါတ်မမှန်', 'km': 'លេខ port មិនត្រឹមត្រូវ', 'vi': 'Số cổng không hợp lệ',
  },
  'invalid_url': {
    'en': 'Invalid URL format', 'th': 'รูปแบบ URL ไม่ถูกต้อง', 'lo': 'ຮູບແບບ URL ບໍ່ຖືກຕ້ອງ',
    'zh': 'URL格式无效', 'ja': 'URLの形式が無効です', 'ko': '잘못된 URL 형식',
    'my': 'URL ပုံစံမမှန်', 'km': 'ទម្រង់ URL មិនត្រឹមត្រូវ', 'vi': 'Định dạng URL không hợp lệ',
  },
  // Field labels — แปลชื่อ key ของ config fields
  'field_uri': {
    'en': 'URI', 'th': 'URI', 'lo': 'URI',
    'zh': 'URI', 'ja': 'URI', 'ko': 'URI',
    'my': 'URI', 'km': 'URI', 'vi': 'URI',
  },
  'field_host': {
    'en': 'Host', 'th': 'โฮสต์', 'lo': 'ໂຮສ',
    'zh': '主机', 'ja': 'ホスト', 'ko': '호스트',
    'my': 'Host', 'km': 'Host', 'vi': 'Host',
  },
  'field_port': {
    'en': 'Port', 'th': 'พอร์ต', 'lo': 'ພອດ',
    'zh': '端口', 'ja': 'ポート', 'ko': '포트',
    'my': 'Port', 'km': 'Port', 'vi': 'Cổng',
  },
  'field_user': {
    'en': 'User', 'th': 'ผู้ใช้', 'lo': 'ຜູ້ໃຊ້',
    'zh': '用户', 'ja': 'ユーザー', 'ko': '사용자',
    'my': 'အသုံးပြုသူ', 'km': 'អ្នកប្រើ', 'vi': 'Người dùng',
  },
  'field_password': {
    'en': 'Password', 'th': 'รหัสผ่าน', 'lo': 'ລະຫັດຜ່ານ',
    'zh': '密码', 'ja': 'パスワード', 'ko': '비밀번호',
    'my': 'စကားဝှက်', 'km': 'ពាក្យសម្ងាត់', 'vi': 'Mật khẩu',
  },
  'field_database_name': {
    'en': 'Database Name', 'th': 'ชื่อ Database', 'lo': 'ຊື່ Database',
    'zh': '数据库名', 'ja': 'データベース名', 'ko': '데이터베이스 이름',
    'my': 'Database အမည်', 'km': 'ឈ្មោះ Database', 'vi': 'Tên Database',
  },
  'field_database': {
    'en': 'Database Name', 'th': 'ชื่อ Database', 'lo': 'ຊື່ Database',
    'zh': '数据库名', 'ja': 'データベース名', 'ko': '데이터베이스 이름',
    'my': 'Database အမည်', 'km': 'ឈ្មោះ Database', 'vi': 'Tên Database',
  },
  'field_ssl_mode': {
    'en': 'SSL Mode', 'th': 'โหมด SSL', 'lo': 'ໂໝດ SSL',
    'zh': 'SSL 模式', 'ja': 'SSLモード', 'ko': 'SSL 모드',
    'my': 'SSL Mode', 'km': 'របៀប SSL', 'vi': 'Chế độ SSL',
  },
  // Integrations fields — AI Providers
  'field_ai_provider': {'en': 'AI Provider', 'th': 'AI Provider ที่ใช้งาน'},
  'field_openrouter_api_key': {'en': 'OpenRouter API Key', 'th': 'OpenRouter API Key'},
  'field_openrouter_model': {'en': 'OpenRouter Model', 'th': 'โมเดล OpenRouter'},
  'field_gemini_api_key': {'en': 'Gemini API Key', 'th': 'Gemini API Key'},
  'field_gemini_model': {'en': 'Gemini Model', 'th': 'โมเดล Gemini'},
  'field_groq_api_key': {'en': 'Groq API Key', 'th': 'Groq API Key'},
  'field_groq_model': {'en': 'Groq Model', 'th': 'โมเดล Groq'},
  'field_deepseek_api_key': {'en': 'DeepSeek API Key', 'th': 'DeepSeek API Key'},
  'field_deepseek_model': {'en': 'DeepSeek Model', 'th': 'โมเดล DeepSeek'},
  // Service fields
  'field_enable_kafka': {'en': 'Enable Kafka', 'th': 'เปิดใช้ Kafka'},
  'field_log_level': {'en': 'Log Level', 'th': 'ระดับ Log'},
  'field_jwt_secret_key': {'en': 'JWT Secret Key', 'th': 'JWT Secret Key'},
  'field_firebase_project_id': {'en': 'Firebase Project ID', 'th': 'Firebase Project ID'},
  // Description labels — คำอธิบาย field
  'desc_mongo_uri': {
    'en': 'MongoDB Connection URI', 'th': 'URI เชื่อมต่อ MongoDB', 'lo': 'URI ເຊື່ອມຕໍ່ MongoDB',
    'zh': 'MongoDB 连接 URI', 'ja': 'MongoDB 接続 URI', 'ko': 'MongoDB 연결 URI',
    'my': 'MongoDB ချိတ်ဆက် URI', 'km': 'URI តភ្ជាប់ MongoDB', 'vi': 'URI kết nối MongoDB',
  },
  'desc_mongo_db': {
    'en': 'MongoDB Database Name', 'th': 'ชื่อ Database ของ MongoDB', 'lo': 'ຊື່ Database ຂອງ MongoDB',
    'zh': 'MongoDB 数据库名', 'ja': 'MongoDB データベース名', 'ko': 'MongoDB 데이터베이스 이름',
    'my': 'MongoDB Database အမည်', 'km': 'ឈ្មោះ Database MongoDB', 'vi': 'Tên Database MongoDB',
  },
  'desc_pg_host': {
    'en': 'PostgreSQL Host', 'th': 'โฮสต์ PostgreSQL', 'lo': 'ໂຮສ PostgreSQL',
    'zh': 'PostgreSQL 主机', 'ja': 'PostgreSQL ホスト', 'ko': 'PostgreSQL 호스트',
    'my': 'PostgreSQL Host', 'km': 'Host PostgreSQL', 'vi': 'Host PostgreSQL',
  },
  'desc_pg_port': {
    'en': 'PostgreSQL Port', 'th': 'พอร์ต PostgreSQL', 'lo': 'ພອດ PostgreSQL',
    'zh': 'PostgreSQL 端口', 'ja': 'PostgreSQL ポート', 'ko': 'PostgreSQL 포트',
    'my': 'PostgreSQL Port', 'km': 'Port PostgreSQL', 'vi': 'Cổng PostgreSQL',
  },
  'desc_pg_user': {
    'en': 'PostgreSQL User', 'th': 'ผู้ใช้ PostgreSQL', 'lo': 'ຜູ້ໃຊ້ PostgreSQL',
    'zh': 'PostgreSQL 用户', 'ja': 'PostgreSQL ユーザー', 'ko': 'PostgreSQL 사용자',
    'my': 'PostgreSQL အသုံးပြုသူ', 'km': 'អ្នកប្រើ PostgreSQL', 'vi': 'Người dùng PostgreSQL',
  },
  'desc_pg_password': {
    'en': 'PostgreSQL Password', 'th': 'รหัสผ่าน PostgreSQL', 'lo': 'ລະຫັດຜ່ານ PostgreSQL',
    'zh': 'PostgreSQL 密码', 'ja': 'PostgreSQL パスワード', 'ko': 'PostgreSQL 비밀번호',
    'my': 'PostgreSQL စကားဝှက်', 'km': 'ពាក្យសម្ងាត់ PostgreSQL', 'vi': 'Mật khẩu PostgreSQL',
  },
  'desc_pg_ssl': {
    'en': 'PostgreSQL SSL mode', 'th': 'โหมด SSL ของ PostgreSQL', 'lo': 'ໂໝດ SSL ຂອງ PostgreSQL',
    'zh': 'PostgreSQL SSL 模式', 'ja': 'PostgreSQL SSLモード', 'ko': 'PostgreSQL SSL 모드',
    'my': 'PostgreSQL SSL Mode', 'km': 'របៀប SSL PostgreSQL', 'vi': 'Chế độ SSL PostgreSQL',
  },
  'desc_ch_host': {
    'en': 'ClickHouse Host', 'th': 'โฮสต์ ClickHouse', 'lo': 'ໂຮສ ClickHouse',
    'zh': 'ClickHouse 主机', 'ja': 'ClickHouse ホスト', 'ko': 'ClickHouse 호스트',
    'my': 'ClickHouse Host', 'km': 'Host ClickHouse', 'vi': 'Host ClickHouse',
  },
  'desc_ch_port': {
    'en': 'ClickHouse Port', 'th': 'พอร์ต ClickHouse', 'lo': 'ພອດ ClickHouse',
    'zh': 'ClickHouse 端口', 'ja': 'ClickHouse ポート', 'ko': 'ClickHouse 포트',
    'my': 'ClickHouse Port', 'km': 'Port ClickHouse', 'vi': 'Cổng ClickHouse',
  },
  'desc_ch_user': {
    'en': 'ClickHouse User', 'th': 'ผู้ใช้ ClickHouse', 'lo': 'ຜູ້ໃຊ້ ClickHouse',
    'zh': 'ClickHouse 用户', 'ja': 'ClickHouse ユーザー', 'ko': 'ClickHouse 사용자',
    'my': 'ClickHouse အသုံးပြုသူ', 'km': 'អ្នកប្រើ ClickHouse', 'vi': 'Người dùng ClickHouse',
  },
  'desc_ch_password': {
    'en': 'ClickHouse Password', 'th': 'รหัสผ่าน ClickHouse', 'lo': 'ລະຫັດຜ່ານ ClickHouse',
    'zh': 'ClickHouse 密码', 'ja': 'ClickHouse パスワード', 'ko': 'ClickHouse 비밀번호',
    'my': 'ClickHouse စကားဝှက်', 'km': 'ពាក្យសម្ងាត់ ClickHouse', 'vi': 'Mật khẩu ClickHouse',
  },
  'desc_ch_db': {
    'en': 'ClickHouse Database', 'th': 'ชื่อ Database ของ ClickHouse', 'lo': 'ຊື່ Database ຂອງ ClickHouse',
    'zh': 'ClickHouse 数据库', 'ja': 'ClickHouse データベース', 'ko': 'ClickHouse 데이터베이스',
    'my': 'ClickHouse Database', 'km': 'Database ClickHouse', 'vi': 'Database ClickHouse',
  },
  // Service descriptions
  'desc_enable_kafka': {'en': 'true/false', 'th': 'เปิด/ปิด Kafka (true/false)'},
  'desc_log_level': {'en': 'DEBUG/INFO/WARN/ERROR', 'th': 'ระดับ Log (DEBUG/INFO/WARN/ERROR)'},
  'desc_jwt_secret': {'en': 'JWT signing secret key', 'th': 'คีย์ลับสำหรับลงนาม JWT'},
  'desc_firebase_id': {'en': 'Firebase project ID for authentication', 'th': 'Firebase project ID สำหรับ authentication'},
  // Integrations descriptions — AI Providers
  'desc_ai_provider': {'en': 'Preferred AI provider (auto-fallback to others if fail)', 'th': 'Provider ที่ลองก่อน (auto-fallback ตัวอื่นถ้า fail)'},
  'desc_openrouter_key': {'en': 'OpenRouter API key (recommended, free models available)', 'th': 'API key ของ OpenRouter (แนะนำ, มีโมเดลฟรี)'},
  'desc_openrouter_model': {'en': 'e.g. google/gemini-2.0-flash-exp:free', 'th': 'เช่น google/gemini-2.0-flash-exp:free'},
  'desc_gemini_key': {'en': 'Google Gemini API key', 'th': 'API key ของ Google Gemini'},
  'desc_gemini_model': {'en': 'Gemini model name', 'th': 'ชื่อโมเดล Gemini'},
  'desc_groq_key': {'en': 'Groq API key (fast inference, free tier)', 'th': 'API key ของ Groq (เร็วมาก, มี free tier)'},
  'desc_groq_model': {'en': 'e.g. llama-3.3-70b-versatile', 'th': 'เช่น llama-3.3-70b-versatile'},
  'desc_deepseek_key': {'en': 'DeepSeek API key (cheap, good quality)', 'th': 'API key ของ DeepSeek (ราคาถูก, คุณภาพดี)'},
  'desc_deepseek_model': {'en': 'e.g. deepseek-chat', 'th': 'เช่น deepseek-chat'},
};

/// Setup Screen — จัดการ config ระบบผ่าน backend API
/// เก็บเฉพาะ backend URL ไว้ที่ SharedPreferences
/// config ทั้งหมดเก็บที่ backend (MongoDB)
class SetupScreen extends StatefulWidget {
  const SetupScreen({super.key});

  @override
  State<SetupScreen> createState() => _SetupScreenState();
}

class _SetupScreenState extends State<SetupScreen> with global.ThemeRefreshMixin {
  // Backend URL — ค่าเดียวที่เก็บใน local
  final _backendUrlController = TextEditingController();
  final _passwordController = TextEditingController();

  bool _isAuthenticated = false;
  bool _isLoading = false;
  bool _isPasswordVisible = false;
  String _statusMessage = '';
  bool _isError = false;

  // Config entries จาก backend (grouped by category)
  Map<String, List<_ConfigItem>> _configsByCategory = {};

  // Connection test results
  final Map<String, _TestResult> _testResults = {};

  // Scanned free models จาก provider APIs (dynamic)
  final Map<String, List<String>> _scannedFreeModels = {};
  bool _isScanning = false;

  @override
  void initState() {
    super.initState();
    // โหลด backend URL จาก BackendUrlManager (key: backend_url)
    // ใช้ key เดียวกับที่ appConfigInit() ใช้ — ไม่แยก key
    final savedUrl = BackendUrlManager.getCurrentUrl() ?? '';
    if (savedUrl.isNotEmpty) {
      _backendUrlController.text = savedUrl;
    } else {
      // ถ้ายังไม่มีค่า → ใช้ effective URL เป็น default
      // web release → Environment URL, native/debug → localhost default
      _backendUrlController.text = BackendUrlManager.getEffectiveUrl() ?? '';
    }
  }

  @override
  void dispose() {
    _backendUrlController.dispose();
    _passwordController.dispose();
    super.dispose();
  }

  String get _baseUrl {
    String url = _backendUrlController.text.trim();
    if (url.isEmpty) return '';
    if (!url.startsWith('http://') && !url.startsWith('https://')) {
      url = 'https://$url';
    }
    if (url.endsWith('/')) url = url.substring(0, url.length - 1);
    // Setup routes อยู่ใน GoAPI — ต้องผ่าน /goapi prefix (Caddy route)
    if (!url.contains('/goapi')) {
      url = '$url/goapi';
    }
    return url;
  }

  Dio _createDio() {
    return Dio(BaseOptions(
      baseUrl: _baseUrl,
      connectTimeout: const Duration(seconds: 10),
      receiveTimeout: const Duration(seconds: 15),
      headers: {'Content-Type': 'application/json'},
    ));
  }

  void _setStatus(String message, {bool isError = false}) {
    if (!mounted) return;
    setState(() {
      _statusMessage = message;
      _isError = isError;
    });
  }

  // ==========================================
  // ยืนยันรหัสผ่าน
  // ==========================================
  Future<void> _verifyPassword() async {
    if (_baseUrl.isEmpty) {
      _setStatus(_t('please_enter_url'), isError: true);
      return;
    }
    if (_passwordController.text.isEmpty) {
      _setStatus(_t('please_enter_password'), isError: true);
      return;
    }

    setState(() => _isLoading = true);
    _setStatus(_t('verifying'));

    try {
      final dio = _createDio();
      final response = await dio.post('/api/setup/verify-password', data: {
        'password': _passwordController.text,
      });

      if (response.data['success'] == true) {
        // บันทึก backend URL ผ่าน BackendUrlManager (key: backend_url)
        await BackendUrlManager.saveUrl(_baseUrl);

        setState(() => _isAuthenticated = true);
        _setStatus(_t('verify_success'));

        // โหลด config จาก backend
        await _loadConfig();
      } else {
        _setStatus(response.data['message'] ?? _t('wrong_password'), isError: true);
      }
    } on DioException catch (e) {
      AppLogger.error('[SetupScreen] verify password error: $e');
      _setStatus('${_t('connection_failed')}: ${e.message}', isError: true);
    } catch (e) {
      AppLogger.error('[SetupScreen] verify password error: $e');
      _setStatus('${_t('error_occurred')}: $e', isError: true);
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  // ==========================================
  // โหลด config จาก backend
  // ==========================================
  Future<void> _loadConfig() async {
    setState(() => _isLoading = true);

    try {
      final dio = _createDio();
      final response = await dio.post('/api/setup/config/get', data: {
        'password': _passwordController.text,
      });

      if (response.data['success'] == true) {
        final List<dynamic> data = response.data['data'] ?? [];

        // เริ่มจาก default categories ก่อน แล้ว merge ค่าจาก backend
        _initDefaultCategories();

        // สร้าง map สำหรับ lookup ค่าจาก backend
        final backendValues = <String, Map<String, dynamic>>{};
        for (final item in data) {
          final category = item['category'] ?? '';
          final key = item['key'] ?? '';
          if (category.isNotEmpty && key.isNotEmpty) {
            backendValues['$category/$key'] = item;
          }
        }

        // แปลง key ที่ backend ใช้ format เก่า (host:port รวมกัน) → แยกเป็น host + port
        _remapBackendKeys(backendValues);

        // merge ค่าจาก backend ลง default categories
        for (final entry in _configsByCategory.entries) {
          for (final item in entry.value) {
            final backendItem = backendValues['${item.category}/${item.key}'];
            if (backendItem != null) {
              item.controller.text = backendItem['value'] ?? '';
              item.value = backendItem['value'] ?? '';
              backendValues.remove('${item.category}/${item.key}');
            }
          }
        }

        // เพิ่ม fields ที่มีใน backend แต่ไม่มีใน default (เช่น integrations)
        for (final entry in backendValues.entries) {
          final item = entry.value;
          final category = item['category'] ?? 'other';
          final configItem = _ConfigItem(
            category: category,
            key: item['key'] ?? '',
            value: item['value'] ?? '',
            isSecret: item['is_secret'] ?? false,
            description: item['description'] ?? '',
            controller: TextEditingController(text: item['value'] ?? ''),
          );
          _configsByCategory.putIfAbsent(category, () => []);
          _configsByCategory[category]!.add(configItem);
        }

        _normalizeHostPort();
        _setStatus('${_t('load_config_success')} (${data.length})');
      } else {
        _setStatus(response.data['message'] ?? _t('load_config_failed'), isError: true);
      }
    } on DioException catch (e) {
      AppLogger.error('[SetupScreen] load config error: $e');
      _setStatus('${_t('load_config_failed')}: ${e.message}', isError: true);
    } catch (e) {
      AppLogger.error('[SetupScreen] load config error: $e');
      _setStatus('${_t('error_occurred')}: $e', isError: true);
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  void _initDefaultCategories() {
    _configsByCategory = {
      'mongodb': [
        _ConfigItem(category: 'mongodb', key: 'uri', value: '', isSecret: true, descKey: 'desc_mongo_uri', controller: TextEditingController()),
        _ConfigItem(category: 'mongodb', key: 'database', value: '', descKey: 'desc_mongo_db', controller: TextEditingController()),
      ],
      'postgresql': [
        _ConfigItem(category: 'postgresql', key: 'host', value: '', descKey: 'desc_pg_host', controller: TextEditingController()),
        _ConfigItem(category: 'postgresql', key: 'port', value: '', descKey: 'desc_pg_port', controller: TextEditingController()),
        _ConfigItem(category: 'postgresql', key: 'user', value: '', descKey: 'desc_pg_user', controller: TextEditingController()),
        _ConfigItem(category: 'postgresql', key: 'password', value: '', isSecret: true, descKey: 'desc_pg_password', controller: TextEditingController()),
        _ConfigItem(category: 'postgresql', key: 'ssl_mode', value: '', descKey: 'desc_pg_ssl', controller: TextEditingController()),
      ],
      'clickhouse': [
        _ConfigItem(category: 'clickhouse', key: 'host', value: '', descKey: 'desc_ch_host', controller: TextEditingController()),
        _ConfigItem(category: 'clickhouse', key: 'port', value: '', descKey: 'desc_ch_port', controller: TextEditingController()),
        _ConfigItem(category: 'clickhouse', key: 'user', value: '', descKey: 'desc_ch_user', controller: TextEditingController()),
        _ConfigItem(category: 'clickhouse', key: 'password', value: '', isSecret: true, descKey: 'desc_ch_password', controller: TextEditingController()),
        _ConfigItem(category: 'clickhouse', key: 'database_name', value: '', descKey: 'desc_ch_db', controller: TextEditingController()),
      ],
      'service': [
        _ConfigItem(category: 'service', key: 'log_level', value: '', descKey: 'desc_log_level', controller: TextEditingController()),
        _ConfigItem(category: 'service', key: 'jwt_secret_key', value: '', isSecret: true, descKey: 'desc_jwt_secret', controller: TextEditingController()),
        _ConfigItem(category: 'service', key: 'firebase_project_id', value: '', descKey: 'desc_firebase_id', controller: TextEditingController()),
        _ConfigItem(category: 'service', key: 'enable_kafka', value: '', descKey: 'desc_enable_kafka', controller: TextEditingController()),
      ],
      'integrations': [
        // AI Provider selector (ตาม backend GetProvider priority)
        _ConfigItem(category: 'integrations', key: 'ai_provider', value: '', descKey: 'desc_ai_provider', controller: TextEditingController()),
        // 1. OpenRouter (ฟรี + 200+ model) — backend priority #1
        _ConfigItem(category: 'integrations', key: 'openrouter_api_key', value: '', isSecret: true, descKey: 'desc_openrouter_key', controller: TextEditingController()),
        _ConfigItem(category: 'integrations', key: 'openrouter_model', value: '', descKey: 'desc_openrouter_model', controller: TextEditingController()),
        // 2. Groq (ฟรี + เร็วมาก) — backend priority #2
        _ConfigItem(category: 'integrations', key: 'groq_api_key', value: '', isSecret: true, descKey: 'desc_groq_key', controller: TextEditingController()),
        _ConfigItem(category: 'integrations', key: 'groq_model', value: '', descKey: 'desc_groq_model', controller: TextEditingController()),
        // 3. DeepSeek (ราคาถูก) — backend priority #3
        _ConfigItem(category: 'integrations', key: 'deepseek_api_key', value: '', isSecret: true, descKey: 'desc_deepseek_key', controller: TextEditingController()),
        _ConfigItem(category: 'integrations', key: 'deepseek_model', value: '', descKey: 'desc_deepseek_model', controller: TextEditingController()),
        // 4. Gemini (default fallback)
        _ConfigItem(category: 'integrations', key: 'gemini_api_key', value: '', isSecret: true, descKey: 'desc_gemini_key', controller: TextEditingController()),
        _ConfigItem(category: 'integrations', key: 'gemini_model', value: '', descKey: 'desc_gemini_model', controller: TextEditingController()),
      ],
    };
  }

  // ==========================================
  // บันทึก config ไปที่ backend
  // ==========================================
  Future<void> _saveConfig() async {
    setState(() => _isLoading = true);
    _setStatus(_t('saving_config'));

    try {
      final dio = _createDio();
      final List<Map<String, dynamic>> configs = [];

      for (final entry in _configsByCategory.entries) {
        for (final item in entry.value) {
          configs.add({
            'category': item.category,
            'key': item.key,
            'value': item.controller.text,
            'is_secret': item.isSecret,
            'description': item.description,
          });
        }
      }

      final response = await dio.post('/api/setup/config/save', data: {
        'password': _passwordController.text,
        'configs': configs,
      });

      if (response.data['success'] == true) {
        _setStatus(response.data['message'] ?? _t('save_success'));
      } else {
        _setStatus(response.data['message'] ?? _t('save_failed'), isError: true);
      }
    } on DioException catch (e) {
      AppLogger.error('[SetupScreen] save config error: $e');
      _setStatus('${_t('save_failed')}: ${e.message}', isError: true);
    } catch (e) {
      AppLogger.error('[SetupScreen] save config error: $e');
      _setStatus('${_t('error_occurred')}: $e', isError: true);
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  // ==========================================
  // Seed config จาก environment variables
  // ==========================================
  Future<void> _seedConfig() async {
    setState(() => _isLoading = true);
    _setStatus(_t('seeding'));

    try {
      final dio = _createDio();
      final response = await dio.post('/api/setup/config/seed', data: {
        'password': _passwordController.text,
      });

      if (response.data['success'] == true) {
        _setStatus(response.data['message'] ?? 'Seed ${_t('success')}');
        await _loadConfig();
      } else {
        _setStatus(response.data['message'] ?? 'Seed ${_t('failed')}', isError: true);
      }
    } on DioException catch (e) {
      AppLogger.error('[SetupScreen] seed config error: $e');
      _setStatus('Seed ${_t('failed')}: ${e.message}', isError: true);
    } catch (e) {
      _setStatus('${_t('error_occurred')}: $e', isError: true);
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  // ==========================================
  // ทดสอบ connection
  // ==========================================
  Future<void> _testConnection(String type) async {
    setState(() {
      _testResults[type] = _TestResult(status: 'testing', message: _t('testing'));
    });

    try {
      final dio = _createDio();
      final Map<String, dynamic> data = {
        'password': _passwordController.text,
        'type': type,
      };

      // ดึงค่าจาก config fields ตาม type
      final items = _configsByCategory[type] ?? [];
      for (final item in items) {
        final value = item.controller.text;
        switch (item.key) {
          case 'uri':
            data['uri'] = value;
            break;
          case 'host':
            data['host'] = value;
            break;
          case 'port':
            data['port'] = value;
            break;
          case 'user':
            data['user'] = value;
            break;
          case 'password':
            data['password2'] = value;
            break;
          case 'database_name':
          case 'database':
            data['database'] = value;
            break;
        }
      }

      final response = await dio.post('/api/setup/test-connection', data: data);

      if (mounted) {
        setState(() {
          final success = response.data['success'] == true;
          final latency = response.data['latency_ms'];
          _testResults[type] = _TestResult(
            status: success ? 'success' : 'failed',
            message: response.data['message'] ?? (success ? _t('success') : _t('failed')),
            latencyMs: latency is int ? latency : null,
            errorCode: response.data['error_code'],
            database: response.data['database'],
          );
        });
      }
    } on DioException catch (e) {
      if (mounted) {
        setState(() {
          _testResults[type] = _TestResult(
            status: 'failed',
            message: '${_t('connection_failed')}: ${e.message}',
          );
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _testResults[type] = _TestResult(
            status: 'failed',
            message: '${_t('error_occurred')}: $e',
          );
        });
      }
    }
  }

  // ==========================================
  // แปลง key จาก backend → key ที่ frontend ใช้
  // ==========================================
  void _remapBackendKeys(Map<String, Map<String, dynamic>> backendValues) {
    // remap key เก่าที่ยังอยู่ใน MongoDB system_config → key ใหม่ที่ตรงกับ predefined
    final remaps = {
      'mongodb/database_name': 'mongodb/database',
    };
    for (final entry in remaps.entries) {
      if (backendValues.containsKey(entry.key) && !backendValues.containsKey(entry.value)) {
        final item = backendValues.remove(entry.key)!;
        item['key'] = entry.value.split('/').last;
        backendValues[entry.value] = item;
      }
    }
  }

  // ==========================================
  // แยก host:port อัตโนมัติ + port เป็นตัวเลข
  // ==========================================
  void _normalizeHostPort() {
    for (final entry in _configsByCategory.entries) {
      final items = entry.value;

      // หา host + port + server_url items ใน category เดียวกัน
      _ConfigItem? hostItem;
      _ConfigItem? portItem;
      _ConfigItem? serverUrlItem;
      for (final item in items) {
        if (item.key == 'host') hostItem = item;
        if (item.key == 'port') portItem = item;
        if (item.key == 'server_url') serverUrlItem = item;
      }

      // กรณี backend ส่ง server_url มา (key เก่า เช่น kafka) → แยกเป็น host+port
      if (serverUrlItem != null && hostItem != null && portItem != null) {
        final urlValue = serverUrlItem.controller.text.trim();
        if (urlValue.isNotEmpty && hostItem.controller.text.trim().isEmpty) {
          if (urlValue.contains(':')) {
            final parts = urlValue.split(':');
            final host = parts.first;
            final port = parts.last;
            if (int.tryParse(port) != null) {
              hostItem.controller.text = host;
              portItem.controller.text = port;
              AppLogger.info('[SetupScreen] แปลง ${entry.key}.server_url "$urlValue" → host="$host", port="$port"');
            }
          } else {
            hostItem.controller.text = urlValue;
          }
        }
      }

      // ถ้ามี host ที่มี : อยู่ → แยก port ออก
      if (hostItem != null && portItem != null) {
        final hostValue = hostItem.controller.text.trim();
        if (hostValue.contains(':')) {
          final parts = hostValue.split(':');
          final host = parts.first;
          final port = parts.last;
          // ตรวจว่า port เป็นตัวเลขจริง
          if (int.tryParse(port) != null) {
            hostItem.controller.text = host;
            if (portItem.controller.text.trim().isEmpty) {
              portItem.controller.text = port;
            }
            AppLogger.info('[SetupScreen] แยก ${entry.key}.host "$hostValue" → host="$host", port="$port"');
          }
        }
      }

      // port ต้องเป็นตัวเลขเท่านั้น — ลบตัวอักษรที่ไม่ใช่ตัวเลขออก
      if (portItem != null) {
        final portValue = portItem.controller.text.trim();
        if (portValue.isNotEmpty) {
          final numericOnly = portValue.replaceAll(RegExp(r'[^0-9]'), '');
          if (numericOnly != portValue) {
            portItem.controller.text = numericOnly;
            AppLogger.info('[SetupScreen] แก้ ${entry.key}.port "$portValue" → "$numericOnly"');
          }
        }
      }
    }
  }

  // ==========================================
  // สร้าง ClickHouse Database
  // ==========================================
  Future<void> _createClickHouseDatabase(String database) async {
    // ดึง host/port/user/password จาก clickhouse config
    String host = '', port = '', user = '', password = '';
    for (final item in _configsByCategory['clickhouse'] ?? []) {
      switch (item.key) {
        case 'host': host = item.controller.text.trim(); break;
        case 'port': port = item.controller.text.trim(); break;
        case 'user': user = item.controller.text.trim(); break;
        case 'password': password = item.controller.text.trim(); break;
      }
    }

    setState(() {
      _testResults['clickhouse'] = _TestResult(status: 'testing', message: 'กำลังสร้าง database...');
    });

    try {
      final dio = _createDio();
      final response = await dio.post('/api/setup/create-clickhouse-database', data: {
        'password': _passwordController.text,
        'host': host,
        'port': port,
        'user': user,
        'password2': password,
        'database': database,
      });

      if (mounted) {
        setState(() {
          final success = response.data['success'] == true;
          _testResults['clickhouse'] = _TestResult(
            status: success ? 'success' : 'failed',
            message: response.data['message'] ?? (success ? 'สร้าง database สำเร็จ' : 'สร้าง database ล้มเหลว'),
          );
        });

        if (response.data['success'] == true) {
          // ทดสอบ connection อีกครั้งหลังสร้าง database
          await _testConnection('clickhouse');
        }
      }
    } on DioException catch (e) {
      if (mounted) {
        setState(() {
          _testResults['clickhouse'] = _TestResult(
            status: 'failed',
            message: 'สร้าง database ล้มเหลว: ${e.message}',
          );
        });
      }
    }
  }

  // ==========================================
  // ทดสอบ connection ทั้งหมด
  // ==========================================
  Future<void> _testAllConnections() async {
    final categoryNames = {
      'mongodb': 'MongoDB',
      'postgresql': 'PostgreSQL',
      'clickhouse': 'ClickHouse',
    };
    final testableCategories = ['mongodb', 'postgresql', 'clickhouse'];
    final categoriesToTest = testableCategories
        .where((c) => _configsByCategory.containsKey(c))
        .toList();

    if (categoriesToTest.isEmpty) return;

    _setStatus(_t('testing_all'));

    // ทดสอบทุก category พร้อมกัน
    await Future.wait(
      categoriesToTest.map((category) => _testConnection(category)),
    );

    if (!mounted) return;

    // สรุปผล
    final successCount = categoriesToTest
        .where((c) => _testResults[c]?.status == 'success')
        .length;
    final total = categoriesToTest.length;
    final allSuccess = successCount == total;

    _setStatus(
      '${_t('test_all_done')}: $successCount/$total ${_t('success')}',
      isError: !allSuccess,
    );

    // แสดง popup ผลลัพธ์
    if (!mounted) return;
    await showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Row(
          children: [
            Icon(
              allSuccess ? Icons.check_circle : Icons.warning_amber,
              color: allSuccess ? Colors.green : Colors.orange,
            ),
            const SizedBox(width: 8),
            Text('${_t('test_all_done')} ($successCount/$total)'),
          ],
        ),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: categoriesToTest.map((category) {
            final result = _testResults[category];
            final isOk = result?.status == 'success';
            return ListTile(
              dense: true,
              leading: Icon(
                isOk ? Icons.check_circle : Icons.cancel,
                color: isOk ? Colors.green : Colors.red,
              ),
              title: Text(categoryNames[category] ?? category),
              subtitle: Text(
                '${result?.message ?? '-'}${result?.latencyMs != null ? ' (${result!.latencyMs}ms)' : ''}',
                style: TextStyle(
                  fontSize: 12,
                  color: isOk ? Colors.green.shade700 : Colors.red.shade700,
                ),
              ),
            );
          }).toList(),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx),
            child: const Text('OK'),
          ),
        ],
      ),
    );
  }

  // ==========================================
  // Export Config — คัดลอก JSON ไปที่ clipboard
  // ==========================================
  Future<void> _exportConfig() async {
    final Map<String, Map<String, String>> configData = {};
    for (final entry in _configsByCategory.entries) {
      final items = <String, String>{};
      for (final item in entry.value) {
        items[item.key] = item.controller.text;
      }
      configData[entry.key] = items;
    }

    final jsonStr = const JsonEncoder.withIndent('  ').convert(configData);

    if (!mounted) return;
    await showDialog(
      context: context,
      builder: (ctx) {
        var copied = false;
        return StatefulBuilder(
          builder: (ctx, setDialogState) => AlertDialog(
            title: Row(
              children: [
                Icon(Icons.file_upload_outlined, color: global.theme.primaryColor),
                const SizedBox(width: 8),
                Text(_t('export_config')),
              ],
            ),
            content: SizedBox(
              width: 600,
              height: 400,
              child: TextField(
                controller: TextEditingController(text: jsonStr),
                readOnly: true,
                maxLines: null,
                expands: true,
                style: const TextStyle(fontSize: 12, fontFamily: 'monospace'),
                decoration: const InputDecoration(
                  border: OutlineInputBorder(),
                  isDense: true,
                ),
              ),
            ),
            actions: [
              TextButton(
                onPressed: () => Navigator.pop(ctx),
                child: Text(_t('cancel')),
              ),
              ElevatedButton.icon(
                onPressed: () async {
                  await Clipboard.setData(ClipboardData(text: jsonStr));
                  setDialogState(() => copied = true);
                },
                icon: Icon(copied ? Icons.check : Icons.copy, size: 18),
                label: Text(copied ? _t('config_copied') : _t('copy_to_clipboard')),
                style: ElevatedButton.styleFrom(
                  backgroundColor: copied ? Colors.green : global.theme.primaryColor,
                  foregroundColor: global.theme.onPrimaryColor,
                ),
              ),
            ],
          ),
        );
      },
    );
  }

  // ==========================================
  // Import Config — วาง JSON จาก clipboard
  // ==========================================
  Future<void> _showImportConfigDialog() async {
    final textController = TextEditingController();

    final result = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text(_t('import_config')),
        content: SizedBox(
          width: 500,
          child: TextField(
            controller: textController,
            maxLines: 10,
            decoration: InputDecoration(
              hintText: _t('paste_json'),
              border: const OutlineInputBorder(),
            ),
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: Text(_t('cancel')),
          ),
          ElevatedButton(
            onPressed: () => Navigator.pop(ctx, true),
            child: Text(_t('import_config')),
          ),
        ],
      ),
    );

    if (result != true || textController.text.isEmpty) return;

    try {
      final Map<String, dynamic> imported = json.decode(textController.text);
      for (final categoryEntry in imported.entries) {
        final category = categoryEntry.key;
        final values = categoryEntry.value;
        if (values is! Map) continue;

        final items = _configsByCategory[category] ?? [];
        for (final item in items) {
          if (values.containsKey(item.key)) {
            item.controller.text = values[item.key].toString();
          }
        }
      }
      _normalizeHostPort();
      setState(() {});
      _setStatus(_t('config_imported'));
    } catch (e) {
      AppLogger.error('[SetupScreen] import config error: $e');
      _setStatus(_t('import_failed'), isError: true);
    }
  }

  // ==========================================
  // Config Validation — ตรวจสอบค่าก่อนบันทึก
  // ==========================================
  List<String> _validateConfig() {
    final errors = <String>[];

    for (final entry in _configsByCategory.entries) {
      for (final item in entry.value) {
        final value = item.controller.text.trim();
        if (value.isEmpty) continue; // ว่างไม่ validate

        // ตรวจ port (ต้องเป็นตัวเลข 1-65535)
        if (item.key == 'port') {
          final port = int.tryParse(value);
          if (port == null || port < 1 || port > 65535) {
            errors.add('${entry.key}.${item.key}: ${_t('invalid_port')} ($value)');
          }
        }

        // ตรวจ URI format
        if (item.key == 'uri' || item.key == 'server_url') {
          final uri = Uri.tryParse(value);
          if (uri == null || (!uri.hasScheme && !value.contains(':'))) {
            errors.add('${entry.key}.${item.key}: ${_t('invalid_url')} ($value)');
          }
        }
      }
    }

    return errors;
  }

  // ==========================================
  // Confirmation Dialog + Validate ก่อนบันทึก
  // ==========================================
  Future<void> _confirmAndSave() async {
    // Validate ก่อน
    final errors = _validateConfig();
    if (errors.isNotEmpty) {
      if (!mounted) return;
      await showDialog(
        context: context,
        builder: (ctx) => AlertDialog(
          title: Row(
            children: [
              const Icon(Icons.warning_amber, color: Colors.orange),
              const SizedBox(width: 8),
              Text(_t('validation_error')),
            ],
          ),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: errors
                .map((e) => Padding(
                      padding: const EdgeInsets.only(bottom: 4),
                      child: Text('• $e', style: const TextStyle(color: Colors.red)),
                    ))
                .toList(),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(ctx),
              child: Text(_t('cancel')),
            ),
          ],
        ),
      );
      return;
    }

    // Confirm ก่อนบันทึก
    if (!mounted) return;
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text(_t('confirm_save')),
        content: Text(_t('confirm_save_msg')),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: Text(_t('cancel')),
          ),
          ElevatedButton(
            onPressed: () => Navigator.pop(ctx, true),
            child: Text(_t('save')),
          ),
        ],
      ),
    );

    if (confirmed == true) {
      await _saveConfig();
    }
  }

  // ==========================================
  // เปลี่ยนรหัสผ่าน
  // ==========================================
  Future<void> _showChangePasswordDialog() async {
    final currentPwdCtrl = TextEditingController();
    final newPwdCtrl = TextEditingController();

    final result = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text(_t('change_password_setup')),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(
              controller: currentPwdCtrl,
              obscureText: true,
              decoration: InputDecoration(
                labelText: _t('current_password'),
                border: const OutlineInputBorder(),
              ),
            ),
            const SizedBox(height: 12),
            TextField(
              controller: newPwdCtrl,
              obscureText: true,
              decoration: InputDecoration(
                labelText: _t('new_password'),
                border: const OutlineInputBorder(),
              ),
            ),
          ],
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx, false), child: Text(_t('cancel'))),
          ElevatedButton(onPressed: () => Navigator.pop(ctx, true), child: Text(_t('change'))),
        ],
      ),
    );

    if (result != true) return;

    try {
      final dio = _createDio();
      final response = await dio.post('/api/setup/change-password', data: {
        'current_password': currentPwdCtrl.text,
        'new_password': newPwdCtrl.text,
      });

      if (response.data['success'] == true) {
        _passwordController.text = newPwdCtrl.text;
        _setStatus(_t('password_changed'));
      } else {
        _setStatus(response.data['message'] ?? _t('password_change_failed'), isError: true);
      }
    } on DioException catch (e) {
      _setStatus('${_t('password_change_failed')}: ${e.message}', isError: true);
    }
  }

  // ==========================================
  // UI
  // ==========================================

  Widget _appBarButton(IconData icon, String label, VoidCallback? onPressed) {
    final enabled = onPressed != null;
    return Tooltip(
      message: label,
      child: InkWell(
        onTap: onPressed,
        borderRadius: BorderRadius.circular(8),
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(icon, size: 24, color: enabled ? global.theme.onPrimaryColor : global.theme.onPrimaryColor.withValues(alpha: 0.38)),
              const SizedBox(height: 2),
              Text(
                label,
                style: TextStyle(
                  fontSize: 9,
                  color: enabled ? global.theme.onPrimaryColor.withValues(alpha: 0.7) : global.theme.onPrimaryColor.withValues(alpha: 0.24),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: global.theme.surfaceColor,
      appBar: AppBar(
        title: Text(_t('setup_config')),
        backgroundColor: global.theme.appBarColor,
        foregroundColor: global.theme.onPrimaryColor,
        actions: [
          if (_isAuthenticated) ...[
            _appBarButton(Icons.wifi_tethering, _t('test_all'), _isLoading ? null : _testAllConnections),
            _appBarButton(Icons.upload, _t('export_config'), _isLoading ? null : _exportConfig),
            _appBarButton(Icons.folder_open, _t('import_config'), _isLoading ? null : _showImportConfigDialog),
            _appBarButton(Icons.lock, _t('change_password'), _showChangePasswordDialog),
            _appBarButton(Icons.cloud_download, _t('seed_from_env'), _isLoading ? null : _seedConfig),
            _appBarButton(Icons.save, _t('save_config'), _isLoading ? null : _confirmAndSave),
            const SizedBox(width: 8),
          ],
        ],
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : SingleChildScrollView(
              padding: const EdgeInsets.all(16),
              child: Center(
                child: ConstrainedBox(
                  constraints: const BoxConstraints(maxWidth: 800),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: [
                      // Status message
                      if (_statusMessage.isNotEmpty)
                        Container(
                          padding: const EdgeInsets.all(12),
                          margin: const EdgeInsets.only(bottom: 16),
                          decoration: BoxDecoration(
                            color: _isError ? Colors.red.shade50 : Colors.green.shade50,
                            borderRadius: BorderRadius.circular(8),
                            border: Border.all(
                              color: _isError ? Colors.red.shade200 : Colors.green.shade200,
                            ),
                          ),
                          child: Row(
                            children: [
                              Icon(
                                _isError ? Icons.error_outline : Icons.check_circle_outline,
                                color: _isError ? Colors.red : Colors.green,
                                size: 20,
                              ),
                              const SizedBox(width: 8),
                              Expanded(
                                child: Text(
                                  _statusMessage,
                                  style: TextStyle(
                                    color: _isError ? Colors.red.shade800 : Colors.green.shade800,
                                  ),
                                ),
                              ),
                            ],
                          ),
                        ),

                      // Backend URL + Password
                      _buildLoginCard(),

                      // Config sections
                      if (_isAuthenticated) ...[
                        const SizedBox(height: 16),
                        ..._buildConfigSections(),
                      ],
                    ],
                  ),
                ),
              ),
            ),
    );
  }

  Widget _buildLoginCard() {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Text(
              _t('backend_connection'),
              style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: global.theme.appBarColor),
            ),
            const SizedBox(height: 4),
            Text(
              _t('backend_desc'),
              style: TextStyle(fontSize: 13, color: global.theme.iconSecondaryColor),
            ),
            const SizedBox(height: 16),
            TextField(
              controller: _backendUrlController,
              enabled: !_isAuthenticated,
              decoration: InputDecoration(
                labelText: _t('backend_url'),
                hintText: 'https://dev-api.bcaicloud.com/goapi',
                border: OutlineInputBorder(),
                prefixIcon: Icon(Icons.link),
              ),
            ),
            const SizedBox(height: 12),
            TextField(
              controller: _passwordController,
              obscureText: !_isPasswordVisible,
              enabled: !_isAuthenticated,
              decoration: InputDecoration(
                labelText: _t('setup_password'),
                hintText: 'default: 12345',
                border: const OutlineInputBorder(),
                prefixIcon: const Icon(Icons.lock),
                suffixIcon: IconButton(
                  icon: Icon(_isPasswordVisible ? Icons.visibility_off : Icons.visibility),
                  onPressed: () => setState(() => _isPasswordVisible = !_isPasswordVisible),
                ),
              ),
              onSubmitted: (_) {
                if (!_isAuthenticated) _verifyPassword();
              },
            ),
            const SizedBox(height: 16),
            if (!_isAuthenticated)
              ElevatedButton.icon(
                onPressed: _isLoading ? null : _verifyPassword,
                icon: const Icon(Icons.login),
                label: Text(_t('login_setup')),
                style: ElevatedButton.styleFrom(
                  backgroundColor: global.theme.primaryColor,
                  foregroundColor: global.theme.onPrimaryColor,
                  padding: const EdgeInsets.symmetric(vertical: 14),
                ),
              )
            else
              Row(
                children: [
                  const Icon(Icons.check_circle, color: Colors.green, size: 20),
                  const SizedBox(width: 8),
                  Text('${_t('connected')}: $_baseUrl', style: const TextStyle(color: Colors.green)),
                  const Spacer(),
                  TextButton(
                    onPressed: () {
                      setState(() {
                        _isAuthenticated = false;
                        _configsByCategory = {};
                        _testResults.clear();
                        _statusMessage = '';
                      });
                    },
                    child: Text(_t('logout')),
                  ),
                ],
              ),
          ],
        ),
      ),
    );
  }

  List<Widget> _buildConfigSections() {
    final categoryNames = {
      'mongodb': 'MongoDB',
      'postgresql': 'PostgreSQL',
      'clickhouse': 'ClickHouse',
      'service': 'Service',
      'integrations': 'Integrations',
    };

    final categoryIcons = {
      'mongodb': Icons.storage,
      'postgresql': Icons.table_chart,
      'clickhouse': Icons.analytics,
      'service': Icons.settings,
      'integrations': Icons.extension,
    };

    final testableCategories = {'mongodb', 'postgresql', 'clickhouse'};

    final widgets = <Widget>[];
    for (final category in categoryNames.keys) {
      final items = _configsByCategory[category] ?? [];
      if (items.isEmpty) continue;

      // Integrations ใช้ UI ใหม่แยก AI Provider cards
      if (category == 'integrations') {
        widgets.add(_buildIntegrationsSection(items));
        continue;
      }

      final testResult = _testResults[category];
      final canTest = testableCategories.contains(category);

      widgets.add(
        Card(
          margin: const EdgeInsets.only(bottom: 12),
          child: Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                Row(
                  children: [
                    Icon(categoryIcons[category] ?? Icons.settings, color: global.theme.primaryColor),
                    const SizedBox(width: 12),
                    Text(
                      categoryNames[category] ?? category,
                      style: const TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
                    ),
                    if (testResult != null) ...[
                      const SizedBox(width: 8),
                      _buildTestBadge(testResult),
                    ],
                  ],
                ),
                const SizedBox(height: 16),
                ...items.map((item) => _buildConfigField(item)),
                if (canTest) ...[
                  const SizedBox(height: 4),
                  OutlinedButton.icon(
                    onPressed: () => _testConnection(category),
                    icon: const Icon(Icons.wifi_tethering, size: 18),
                    label: Text('${_t('test_connection')} ${categoryNames[category]}'),
                    style: OutlinedButton.styleFrom(
                      foregroundColor: global.theme.primaryColor,
                    ),
                  ),
                  if (testResult != null && testResult.message.isNotEmpty) ...[
                    const SizedBox(height: 8),
                    Text(
                      '${testResult.message}${testResult.latencyMs != null ? ' (${testResult.latencyMs}ms)' : ''}',
                      style: TextStyle(
                        fontSize: 12,
                        color: testResult.status == 'success' ? Colors.green : Colors.red,
                      ),
                    ),
                    // ปุ่มสร้าง database เมื่อ ClickHouse error code 81
                    if (category == 'clickhouse' &&
                        testResult.errorCode == 'database_not_exist' &&
                        testResult.database != null) ...[
                      const SizedBox(height: 8),
                      ElevatedButton.icon(
                        onPressed: () => _createClickHouseDatabase(testResult.database!),
                        icon: const Icon(Icons.add_circle_outline, size: 18),
                        label: Text('สร้าง database "${testResult.database}"'),
                        style: ElevatedButton.styleFrom(
                          backgroundColor: Colors.orange,
                          foregroundColor: global.theme.onPrimaryColor,
                        ),
                      ),
                    ],
                  ],
                ],
              ],
            ),
          ),
        ),
      );
    }

    return widgets;
  }

  // ============================================
  // AI Providers — Integrations Section (UX ใหม่)
  // ============================================

  /// ข้อมูล AI Provider สำหรับ UI
  static final List<_AIProviderInfo> _aiProviders = [
    // ลำดับตาม backend GetProvider() priority
    // 1. OpenRouter — backend priority #1
    _AIProviderInfo(
      id: 'openrouter',
      name: 'OpenRouter',
      icon: Icons.router,
      color: Color(0xFF6366F1),
      badge: 'ฟรี + 200+ model',
      badgeColor: Colors.green,
      registerUrl: 'https://openrouter.ai/settings/keys',
      description: 'รวม 200+ โมเดลจากทุกค่าย — มีฟรีเยอะ',
      freeModels: [
        'nvidia/nemotron-3-nano-30b-a3b:free',
        'stepfun/step-3.5-flash:free',
        'upstage/solar-pro-3:free',
        'arcee-ai/trinity-large-preview:free',
        'arcee-ai/trinity-mini:free',
      ],
      paidModels: [
        'anthropic/claude-sonnet-4.6',
        'openai/gpt-4o',
        'google/gemini-2.5-pro',
      ],
    ),
    // 2. Groq — backend priority #2
    _AIProviderInfo(
      id: 'groq',
      name: 'Groq',
      icon: Icons.bolt,
      color: Color(0xFFF55036),
      badge: 'ฟรี + เร็วมาก',
      badgeColor: Colors.green,
      registerUrl: 'https://console.groq.com/keys',
      description: 'ฟรีทุก model — Inference เร็วที่สุด (LPU chip)',
      freeModels: [
        'meta-llama/llama-4-maverick-17b-128e-instruct',
        'qwen/qwen-3-32b',
        'deepseek-r1-distill-llama-70b',
        'llama-3.3-70b-versatile',
      ],
      paidModels: [
        'meta-llama/llama-4-maverick-17b-128e-instruct',
      ],
    ),
    // 3. DeepSeek — backend priority #3
    _AIProviderInfo(
      id: 'deepseek',
      name: 'DeepSeek',
      icon: Icons.psychology,
      color: Color(0xFF0EA5E9),
      badge: 'ราคาถูก',
      badgeColor: Colors.teal,
      registerUrl: 'https://platform.deepseek.com/api_keys',
      description: 'คุณภาพสูง ราคาถูกมาก — \$0.28/1M input, ฟรี 5M tokens แรก',
      freeModels: [],
      paidModels: [
        'deepseek-chat',
        'deepseek-reasoner',
      ],
    ),
    // 4. Gemini — default fallback
    _AIProviderInfo(
      id: 'gemini',
      name: 'Google Gemini',
      icon: Icons.auto_awesome,
      color: Color(0xFF4285F4),
      badge: 'Default (Fallback)',
      badgeColor: Colors.blue,
      registerUrl: 'https://aistudio.google.com/apikey',
      description: 'Default fallback — ถ้า provider อื่นไม่มี API key จะใช้ตัวนี้',
      freeModels: [
        'gemini-2.5-flash',
        'gemini-2.5-flash-lite',
        'gemini-2.5-pro',
      ],
      paidModels: [
        'gemini-2.5-pro',
        'gemini-2.5-flash',
      ],
    ),
  ];

  Widget _buildIntegrationsSection(List<_ConfigItem> items) {
    // จัดกลุ่ม items ตาม provider prefix
    final Map<String, List<_ConfigItem>> providerItems = {};
    final List<_ConfigItem> otherItems = [];
    _ConfigItem? aiProviderItem;

    for (final item in items) {
      if (item.key == 'ai_provider') {
        aiProviderItem = item;
        continue;
      }
      bool matched = false;
      for (final provider in _aiProviders) {
        if (item.key.startsWith('${provider.id}_')) {
          providerItems.putIfAbsent(provider.id, () => []).add(item);
          matched = true;
          break;
        }
      }
      if (!matched) otherItems.add(item);
    }

    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            // Section header
            Row(
              children: [
                Icon(Icons.smart_toy, color: global.theme.primaryColor),
                const SizedBox(width: 12),
                const Text(
                  'AI Providers',
                  style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
                ),
                const Spacer(),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                  decoration: BoxDecoration(
                    color: Colors.green.withValues(alpha: 0.1),
                    borderRadius: BorderRadius.circular(12),
                    border: Border.all(color: Colors.green.withValues(alpha: 0.3)),
                  ),
                  child: const Text(
                    'เน้นของฟรี',
                    style: TextStyle(fontSize: 11, color: Colors.green, fontWeight: FontWeight.bold),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 6),
            Text(
              'ใส่ API Key ที่มี — ระบบจะ auto-fallback ลองทีละตัวอัตโนมัติ',
              style: TextStyle(fontSize: 12, color: global.theme.iconSecondaryColor),
            ),
            const SizedBox(height: 16),

            // AI Provider selector (ตัวเลือกหลัก)
            if (aiProviderItem != null) _buildAIProviderSelector(aiProviderItem),

            // AI Provider cards
            ..._aiProviders.map((provider) {
              final pItems = providerItems[provider.id] ?? [];
              return _buildAIProviderCard(provider, pItems);
            }),

            // Other integration items (S3, etc.)
            if (otherItems.isNotEmpty) ...[
              Divider(height: 32),
              Row(
                children: [
                  Icon(Icons.cloud, color: global.theme.iconSecondaryColor, size: 20),
                  SizedBox(width: 8),
                  Text(global.language('others'), style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold, color: global.theme.textColor)),
                ],
              ),
              const SizedBox(height: 12),
              ...otherItems.map((item) => _buildConfigField(item)),
            ],
          ],
        ),
      ),
    );
  }

  /// AI Provider selector — ตัวเลือกหลักที่ backend จะลองก่อน (auto-fallback ตัวอื่นถ้า fail)
  Widget _buildAIProviderSelector(_ConfigItem item) {
    final providerOptions = [
      {'value': '', 'label': 'Auto', 'desc': 'ลองทุกตัวที่มี key อัตโนมัติ', 'icon': Icons.auto_mode, 'color': Colors.green},
      {'value': 'openrouter', 'label': 'OpenRouter', 'desc': global.language('try_this_first'), 'icon': Icons.router, 'color': Color(0xFF6366F1)},
      {'value': 'groq', 'label': 'Groq', 'desc': global.language('try_this_first'), 'icon': Icons.bolt, 'color': Color(0xFFF55036)},
      {'value': 'deepseek', 'label': 'DeepSeek', 'desc': global.language('try_this_first'), 'icon': Icons.psychology, 'color': Color(0xFF0EA5E9)},
      {'value': 'gemini', 'label': 'Gemini', 'desc': global.language('try_this_first'), 'icon': Icons.auto_awesome, 'color': Color(0xFF4285F4)},
    ];

    final currentValue = item.controller.text.trim().toLowerCase();

    return Container(
      margin: const EdgeInsets.only(bottom: 16),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.green.withValues(alpha: 0.05),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.green.withValues(alpha: 0.3)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              const Icon(Icons.auto_mode, size: 20, color: Colors.green),
              const SizedBox(width: 8),
              const Text(
                'AI Provider — Auto Fallback',
                style: TextStyle(fontWeight: FontWeight.bold, fontSize: 14),
              ),
              const SizedBox(width: 8),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                decoration: BoxDecoration(
                  color: Colors.green.withValues(alpha: 0.1),
                  borderRadius: BorderRadius.circular(6),
                  border: Border.all(color: Colors.green.withValues(alpha: 0.4)),
                ),
                child: const Text(
                  'auto',
                  style: TextStyle(fontSize: 9, color: Colors.green, fontWeight: FontWeight.bold),
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          Container(
            padding: const EdgeInsets.all(10),
            decoration: BoxDecoration(
              color: Colors.blue.withValues(alpha: 0.05),
              borderRadius: BorderRadius.circular(8),
              border: Border.all(color: Colors.blue.withValues(alpha: 0.2)),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Icon(Icons.info_outline, size: 14, color: Colors.blue[700]),
                    const SizedBox(width: 6),
                    Text('วิธีทำงาน', style: TextStyle(fontSize: 11, fontWeight: FontWeight.bold, color: Colors.blue[700])),
                  ],
                ),
                const SizedBox(height: 4),
                Text(
                  '• เอาเฉพาะ provider ที่มี API Key ใส่ไว้\n'
                  '• ลองทีละตัวตามลำดับ — ถ้าตัวไหน fail ข้ามไปตัวถัดไป\n'
                  '• ตัวที่ fail จะ cooldown 1 ชั่วโมง ก่อนลองใหม่',
                  style: TextStyle(fontSize: 11, color: global.theme.textColor, height: 1.5),
                ),
              ],
            ),
          ),
          const SizedBox(height: 12),
          Text(
            'เลือก provider ที่จะลองก่อน (ตัวอื่นจะ fallback อัตโนมัติ)',
            style: TextStyle(fontSize: 11, color: global.theme.iconSecondaryColor),
          ),
          const SizedBox(height: 8),
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: providerOptions.map((opt) {
              final value = opt['value'] as String;
              final label = opt['label'] as String;
              final desc = opt['desc'] as String;
              final icon = opt['icon'] as IconData;
              final color = opt['color'] as Color;
              final isSelected = currentValue == value || (currentValue.isEmpty && value.isEmpty);

              return InkWell(
                onTap: () {
                  setState(() {
                    item.controller.text = value;
                  });
                },
                borderRadius: BorderRadius.circular(8),
                child: Container(
                  padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                  decoration: BoxDecoration(
                    color: isSelected ? color.withValues(alpha: 0.15) : global.theme.surfaceColor,
                    borderRadius: BorderRadius.circular(8),
                    border: Border.all(
                      color: isSelected ? color : global.theme.dividerBorderColor,
                      width: isSelected ? 2 : 1,
                    ),
                  ),
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Icon(icon, size: 16, color: isSelected ? color : global.theme.iconSecondaryColor),
                          const SizedBox(width: 6),
                          Text(
                            label,
                            style: TextStyle(
                              fontSize: 12,
                              fontWeight: isSelected ? FontWeight.bold : FontWeight.normal,
                              color: isSelected ? color : global.theme.textColor,
                            ),
                          ),
                          if (isSelected) ...[
                            const SizedBox(width: 4),
                            Icon(Icons.check_circle, size: 14, color: color),
                          ],
                        ],
                      ),
                      Text(desc, style: TextStyle(fontSize: 9, color: global.theme.iconSecondaryColor)),
                    ],
                  ),
                ),
              );
            }).toList(),
          ),
        ],
      ),
    );
  }

  Widget _buildAIProviderCard(_AIProviderInfo provider, List<_ConfigItem> items) {
    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      decoration: BoxDecoration(
        border: Border.all(color: provider.color.withValues(alpha: 0.3)),
        borderRadius: BorderRadius.circular(12),
      ),
      child: Theme(
        data: Theme.of(context).copyWith(dividerColor: Colors.transparent),
        child: ExpansionTile(
          tilePadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
          childrenPadding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
          leading: Container(
            width: 40,
            height: 40,
            decoration: BoxDecoration(
              color: provider.color.withValues(alpha: 0.1),
              borderRadius: BorderRadius.circular(10),
            ),
            child: Icon(provider.icon, color: provider.color, size: 22),
          ),
          title: Row(
            children: [
              Text(provider.name, style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 15)),
              const SizedBox(width: 8),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                decoration: BoxDecoration(
                  color: provider.badgeColor.withValues(alpha: 0.1),
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(color: provider.badgeColor.withValues(alpha: 0.4)),
                ),
                child: Text(
                  provider.badge,
                  style: TextStyle(fontSize: 10, color: provider.badgeColor, fontWeight: FontWeight.bold),
                ),
              ),
            ],
          ),
          subtitle: Text(provider.description, style: const TextStyle(fontSize: 12)),
          initiallyExpanded: provider.id == 'openrouter', // OpenRouter เปิดค้างไว้ (backend priority #1)
          children: [
            // API Key field
            ...items.where((i) => i.key.endsWith('_api_key')).map((item) => _buildConfigField(item)),

            // Model selection — checkbox ฟรี auto หรือกรอก model เสียเงิน
            _buildModelSelector(provider, items),

            // ปุ่มลงทะเบียน API Key
            const SizedBox(height: 12),
            SizedBox(
              width: double.infinity,
              child: OutlinedButton.icon(
                onPressed: () => _launchUrl(provider.registerUrl),
                icon: const Icon(Icons.open_in_new, size: 16),
                label: Text('ลงทะเบียนรับ API Key — ${provider.name}'),
                style: OutlinedButton.styleFrom(
                  foregroundColor: provider.color,
                  side: BorderSide(color: provider.color.withValues(alpha: 0.4)),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  /// Model selector — checkbox ฟรี (auto) หรือกรอก model เสียเงิน
  Widget _buildModelSelector(_AIProviderInfo provider, List<_ConfigItem> items) {
    final modelMatches = items.where((i) => i.key.endsWith('_model'));
    final modelItem = modelMatches.isNotEmpty ? modelMatches.first : null;
    if (modelItem == null) return const SizedBox.shrink();

    final currentModel = modelItem.controller.text.trim();
    // ถ้า model ว่าง หรือลงท้าย :free → ถือว่าใช้โหมดฟรี
    final isFreeMode = currentModel.isEmpty ||
        currentModel.endsWith(':free') ||
        provider.freeModels.contains(currentModel);
    final scanned = _scannedFreeModels[provider.id];

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // Checkbox ใช้โมเดลฟรี
        Container(
          decoration: BoxDecoration(
            color: isFreeMode ? Colors.green.withValues(alpha: 0.05) : Colors.transparent,
            borderRadius: BorderRadius.circular(8),
            border: Border.all(
              color: isFreeMode ? Colors.green.withValues(alpha: 0.3) : global.theme.dividerBorderColor,
            ),
          ),
          child: CheckboxListTile(
            value: isFreeMode,
            dense: true,
            contentPadding: const EdgeInsets.symmetric(horizontal: 8),
            title: Row(
              children: [
                const Icon(Icons.card_giftcard, size: 16, color: Colors.green),
                const SizedBox(width: 6),
                const Text('ใช้โมเดลฟรีอัตโนมัติ', style: TextStyle(fontSize: 13, fontWeight: FontWeight.w600)),
                const SizedBox(width: 8),
                if (isFreeMode)
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 1),
                    decoration: BoxDecoration(
                      color: Colors.green.withValues(alpha: 0.15),
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: const Text('AUTO', style: TextStyle(fontSize: 9, color: Colors.green, fontWeight: FontWeight.bold)),
                  ),
              ],
            ),
            subtitle: Text(
              isFreeMode ? 'Backend เลือกโมเดลฟรีให้อัตโนมัติ' : 'กดเปิดเพื่อใช้โมเดลฟรี (ไม่ต้องเลือก)',
              style: TextStyle(fontSize: 11, color: global.theme.iconSecondaryColor),
            ),
            onChanged: (val) {
              setState(() {
                if (val == true) {
                  // เปิดโหมดฟรี → เคลียร์ model field ให้ backend เลือกเอง
                  modelItem.controller.text = '';
                } else {
                  // ปิดโหมดฟรี → ใส่ paid model แรกเป็น default
                  modelItem.controller.text = provider.paidModels.isNotEmpty ? provider.paidModels.first : '';
                }
              });
            },
            activeColor: Colors.green,
          ),
        ),

        // เมื่อเปิดโหมดฟรี — แสดงปุ่ม Scan + รายการ model ฟรี (reference)
        if (isFreeMode && (provider.freeModels.isNotEmpty || scanned != null)) ...[
          const SizedBox(height: 8),
          Row(
            children: [
              Text(
                scanned != null ? 'โมเดลฟรี (${scanned.length} ตัว)' : 'โมเดลฟรีตัวอย่าง',
                style: TextStyle(fontSize: 11, color: global.theme.iconSecondaryColor),
              ),
              const Spacer(),
              SizedBox(
                height: 24,
                child: TextButton.icon(
                  onPressed: _isScanning ? null : () => _scanFreeModels(provider.id),
                  icon: _isScanning
                      ? const SizedBox(width: 12, height: 12, child: CircularProgressIndicator(strokeWidth: 2))
                      : const Icon(Icons.radar, size: 14),
                  label: Text(_isScanning ? _t('scanning_in_progress') : _t('scan_latest')),
                  style: TextButton.styleFrom(
                    foregroundColor: Colors.green,
                    padding: const EdgeInsets.symmetric(horizontal: 6),
                    textStyle: const TextStyle(fontSize: 10),
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 4),
          _buildFreeModelPreview(scanned ?? provider.freeModels),
        ],

        // เมื่อปิดโหมดฟรี — แสดง model text field + paid chips
        if (!isFreeMode) ...[
          const SizedBox(height: 8),
          _buildConfigField(modelItem),
          if (provider.paidModels.isNotEmpty) ...[
            _buildModelChips('โมเดลเสียเงิน', provider.paidModels, Colors.orange, items),
            const SizedBox(height: 4),
          ],
        ],
      ],
    );
  }

  /// แสดง preview รายการ model ฟรี (แบบ compact wrap chips)
  Widget _buildFreeModelPreview(List<String> models) {
    if (models.isEmpty) return const SizedBox.shrink();
    // แสดงแค่ 8 ตัวแรก ถ้ามีมากกว่า
    final display = models.length > 8 ? models.sublist(0, 8) : models;
    final remaining = models.length - display.length;

    return Wrap(
      spacing: 4,
      runSpacing: 4,
      children: [
        ...display.map((m) => Container(
          padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
          decoration: BoxDecoration(
            color: Colors.green.withValues(alpha: 0.06),
            borderRadius: BorderRadius.circular(10),
            border: Border.all(color: Colors.green.withValues(alpha: 0.15)),
          ),
          child: Text(m, style: TextStyle(fontSize: 10, color: Colors.green[700])),
        )),
        if (remaining > 0)
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
            decoration: BoxDecoration(
              color: global.theme.surfaceColor,
              borderRadius: BorderRadius.circular(10),
            ),
            child: Text('+$remaining อีก', style: TextStyle(fontSize: 10, color: global.theme.iconSecondaryColor)),
          ),
      ],
    );
  }

  Widget _buildModelChips(String title, List<String> models, Color color, List<_ConfigItem> items) {
    // หา model field controller เพื่อ fill ค่าเมื่อกด chip
    final modelMatches = items.where((i) => i.key.endsWith('_model'));
    final modelItem = modelMatches.isNotEmpty ? modelMatches.first : null;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        if (title.isNotEmpty) ...[
          Row(
            children: [
              Icon(
                color == Colors.green ? Icons.card_giftcard : Icons.paid,
                size: 14,
                color: color,
              ),
              const SizedBox(width: 4),
              Text(title, style: TextStyle(fontSize: 12, fontWeight: FontWeight.bold, color: color)),
            ],
          ),
          const SizedBox(height: 6),
        ],
        Wrap(
          spacing: 6,
          runSpacing: 6,
          children: models.map((model) {
            return InkWell(
              onTap: modelItem != null
                  ? () {
                      setState(() {
                        modelItem.controller.text = model;
                      });
                    }
                  : null,
              borderRadius: BorderRadius.circular(16),
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
                decoration: BoxDecoration(
                  color: color.withValues(alpha: 0.08),
                  borderRadius: BorderRadius.circular(16),
                  border: Border.all(color: color.withValues(alpha: 0.25)),
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Text(
                      model,
                      style: TextStyle(fontSize: 11, color: color.withValues(alpha: 0.8)),
                    ),
                    if (modelItem != null) ...[
                      const SizedBox(width: 4),
                      Icon(Icons.arrow_forward, size: 10, color: color.withValues(alpha: 0.5)),
                    ],
                  ],
                ),
              ),
            );
          }).toList(),
        ),
      ],
    );
  }

  /// Scan free models จาก provider APIs (public endpoints ไม่ต้องใช้ key)
  Future<void> _scanFreeModels(String providerId) async {
    setState(() => _isScanning = true);

    try {
      List<String> models = [];

      if (providerId == 'openrouter') {
        // OpenRouter public API — ไม่ต้องใช้ API key
        final dio = Dio()..options.connectTimeout = const Duration(seconds: 10);
        final resp = await dio.get('https://openrouter.ai/api/v1/models');
        if (resp.statusCode == 200 && resp.data['data'] is List) {
          final allModels = resp.data['data'] as List;
          models = allModels
              .where((m) => (m['id'] as String? ?? '').endsWith(':free'))
              .map<String>((m) => m['id'] as String)
              .toList()
            ..sort();
        }
      } else if (providerId == 'groq') {
        // Groq — ต้องใช้ API key ดึงจาก config field
        final keyItem = _configsByCategory['integrations']
            ?.firstWhere((i) => i.key == 'groq_api_key', orElse: () => _ConfigItem(category: '', key: '', value: '', controller: TextEditingController()));
        final apiKey = keyItem?.controller.text ?? '';
        if (apiKey.isNotEmpty) {
          final dio = Dio()..options.connectTimeout = const Duration(seconds: 10);
          final resp = await dio.get(
            'https://api.groq.com/openai/v1/models',
            options: Options(headers: {'Authorization': 'Bearer $apiKey'}),
          );
          if (resp.statusCode == 200 && resp.data['data'] is List) {
            models = (resp.data['data'] as List)
                .map<String>((m) => m['id'] as String)
                .where((id) => !id.contains('whisper') && !id.contains('guard') && !id.contains('tool-use'))
                .toList()
              ..sort();
          }
        }
      }

      if (models.isNotEmpty) {
        setState(() {
          _scannedFreeModels[providerId] = models;
        });
      }
    } catch (e) {
      AppLogger.error('[SetupScreen] scan $providerId models error: $e');
    } finally {
      if (mounted) setState(() => _isScanning = false);
    }
  }

  Future<void> _launchUrl(String url) async {
    final uri = Uri.parse(url);
    if (await canLaunchUrl(uri)) {
      await launchUrl(uri, mode: LaunchMode.externalApplication);
    }
  }

  Widget _buildConfigField(_ConfigItem item) {
    // แปล label จาก field key เช่น 'host' → _t('field_host')
    final fieldKey = 'field_${item.key}';
    final label = _setupTexts.containsKey(fieldKey) ? _t(fieldKey) : item.key;
    // แปล description — ถ้ามี descKey ใช้ _t(), ไม่งั้นใช้ description เดิม
    final desc = item.descKey.isNotEmpty ? _t(item.descKey) : item.description;

    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: TextField(
        controller: item.controller,
        decoration: InputDecoration(
          labelText: label,
          helperText: desc.isNotEmpty ? desc : null,
          border: const OutlineInputBorder(),
          isDense: true,
        ),
      ),
    );
  }

  Widget _buildTestBadge(_TestResult result) {
    Color color;
    IconData icon;
    switch (result.status) {
      case 'success':
        color = Colors.green;
        icon = Icons.check_circle;
        break;
      case 'failed':
        color = Colors.red;
        icon = Icons.cancel;
        break;
      case 'testing':
        color = Colors.orange;
        icon = Icons.hourglass_empty;
        break;
      default:
        color = global.theme.iconSecondaryColor;
        icon = Icons.help;
    }

    return Icon(icon, color: color, size: 18);
  }
}

// ==========================================
// Models
// ==========================================

class _ConfigItem {
  final String category;
  final String key;
  String value;
  final bool isSecret;
  final String description;
  final String descKey; // key สำหรับ _t() — ถ้ามีจะใช้แทน description
  final TextEditingController controller;

  _ConfigItem({
    required this.category,
    required this.key,
    this.value = '',
    this.isSecret = false,
    this.description = '',
    this.descKey = '',
    required this.controller,
  });
}

class _TestResult {
  final String status; // 'success', 'failed', 'testing'
  final String message;
  final int? latencyMs;
  final String? errorCode; // 'database_not_exist' สำหรับ ClickHouse error code 81
  final String? database;

  _TestResult({
    required this.status,
    required this.message,
    this.latencyMs,
    this.errorCode,
    this.database,
  });
}

class _AIProviderInfo {
  final String id;
  final String name;
  final IconData icon;
  final Color color;
  final String badge;
  final Color badgeColor;
  final String registerUrl;
  final String description;
  final List<String> freeModels;
  final List<String> paidModels;

  const _AIProviderInfo({
    required this.id,
    required this.name,
    required this.icon,
    required this.color,
    required this.badge,
    required this.badgeColor,
    required this.registerUrl,
    required this.description,
    this.freeModels = const [],
    this.paidModels = const [],
  });
}
