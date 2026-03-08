import 'package:flutter/material.dart';

/// แสดง Dialog เลือกภาษา
/// Returns: LanguageCode ที่เลือก (en, th, lo, cn, ja, ko, my, km, vi) หรือ null ถ้ายกเลิก
Future<String?> showLanguageSelectionDialog(BuildContext context) async {
  // ชื่อภาษาแบบ hard-code เป็นภาษานั้นๆ โดยตรง
  final List<String> languageNames = [
    "English",           // อังกฤษ
    "ไทย",               // ไทย
    "ລາວ",               // ลาว
    "中文",              // จีน
    "日本語",            // ญี่ปุ่น
    "한국어",            // เกาหลี
    "မြန်မာ",           // พม่า
    "ខ្មែរ",             // เขมร
    "Tiếng Việt",       // เวียดนาม
  ];

  final List<String> countryCodes = [
    "en",
    "th",
    "lo",
    "cn",
    "ja",
    "ko",
    "my",
    "km",
    "vi",
  ];

  return showDialog<String>(
    context: context,
    builder: (BuildContext context) {
      return Dialog(
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(10),
        ),
        child: Container(
          constraints: const BoxConstraints(
            maxWidth: 700,
          ),
          padding: const EdgeInsets.all(20),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              // หัวข้อ
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  const Text(
                    'Select Language / เลือกภาษา',
                    style: TextStyle(
                      fontSize: 20,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  IconButton(
                    icon: const Icon(Icons.close),
                    onPressed: () => Navigator.of(context).pop(),
                  ),
                ],
              ),
              const Divider(),
              const SizedBox(height: 10),
              // รายการภาษาแบบ Grid
              GridView.builder(
                shrinkWrap: true,
                physics: const NeverScrollableScrollPhysics(),
                gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                  crossAxisCount: 3, // 3 คอลัมน์
                  childAspectRatio: 2.5, // อัตราส่วน กว้าง:สูง
                  crossAxisSpacing: 10,
                  mainAxisSpacing: 10,
                ),
                itemCount: countryCodes.length,
                itemBuilder: (context, index) {
                  return InkWell(
                    onTap: () {
                      Navigator.of(context).pop(countryCodes[index]);
                    },
                    borderRadius: BorderRadius.circular(8),
                    child: Container(
                      decoration: BoxDecoration(
                        border: Border.all(color: Colors.grey.shade300, width: 1),
                        borderRadius: BorderRadius.circular(8),
                      ),
                      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                      child: Row(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Image.asset(
                            'assets/flags/${countryCodes[index]}.png',
                            width: 40,
                            height: 40,
                          ),
                          const SizedBox(width: 10),
                          Flexible(
                            child: Text(
                              languageNames[index],
                              style: const TextStyle(
                                fontSize: 16,
                                fontWeight: FontWeight.w600,
                              ),
                              overflow: TextOverflow.ellipsis,
                            ),
                          ),
                        ],
                      ),
                    ),
                  );
                },
              ),
            ],
          ),
        ),
      );
    },
  );
}

/// Legacy Screen - deprecated, ใช้ showLanguageSelectionDialog() แทน
@Deprecated('Use showLanguageSelectionDialog() instead')
class SelectLanguageScreen extends StatefulWidget {
  const SelectLanguageScreen({super.key});

  @override
  SelectLanguageScreenState createState() => SelectLanguageScreenState();
}

class SelectLanguageScreenState extends State<SelectLanguageScreen> {
  final List<String> languageNames = [
    "English",
    "ไทย",
    "ລາວ",
    "中文",
    "日本語",
    "한국어",
    "မြန်မာ",
    "ខ្មែរ",
    "Tiếng Việt",
  ];

  final List<String> countryCodes = [
    "en",
    "th",
    "lo",
    "cn",
    "ja",
    "ko",
    "my",
    "km",
    "vi",
  ];

  @override
  Widget build(BuildContext context) {
    return Container(
      color: Colors.white,
      child: Scaffold(
        resizeToAvoidBottomInset: false,
        body: Container(
          width: double.infinity,
          padding: const EdgeInsets.all(10),
          child: ListView.builder(
            itemCount: countryCodes.length,
            itemBuilder: (context, index) {
              return ListTile(
                dense: true,
                shape: RoundedRectangleBorder(
                  side: const BorderSide(color: Colors.grey, width: 1),
                  borderRadius: BorderRadius.circular(5),
                ),
                title: Row(
                  children: [
                    Image.asset(
                      'assets/flags/${countryCodes[index]}.png',
                      width: 100,
                      height: 100,
                    ),
                    const SizedBox(width: 10),
                    Text(
                      languageNames[index],
                      style: const TextStyle(
                        fontSize: 20,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ],
                ),
                onTap: () {
                  Navigator.of(context).pop(countryCodes[index]);
                },
              );
            },
          ),
        ),
      ),
    );
  }
}
