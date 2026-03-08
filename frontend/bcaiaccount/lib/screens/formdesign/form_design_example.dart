import 'package:flutter/material.dart';
import '../formdesign/form_design.dart';
import 'package:smlaicloud/global.dart' as global;

/// ตัวอย่างการใช้งาน Form Design System
///
/// ใช้สำหรับทดสอบระบบออกแบบฟอร์ม
class FormDesignExample extends StatelessWidget {
  const FormDesignExample({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Form Design Example'),
      ),
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Text(
              global.language('form_design_title'),
              style: const TextStyle(fontSize: 24, fontWeight: FontWeight.bold),
            ),
            SizedBox(height: 20),
            Padding(
              padding: EdgeInsets.all(16.0),
              child: Text(
                global.language('form_design_create_forms'),
                textAlign: TextAlign.center,
                style: const TextStyle(fontSize: 16),
              ),
            ),
            SizedBox(height: 30),
            ElevatedButton.icon(
              onPressed: () {
                Navigator.push(
                  context,
                  MaterialPageRoute(
                    builder: (context) => const FormDesignScreen(),
                  ),
                );
              },
              icon: Icon(Icons.design_services),
              label: Text(global.language('form_design_open')),
              style: ElevatedButton.styleFrom(
                padding: const EdgeInsets.symmetric(
                  horizontal: 32,
                  vertical: 16,
                ),
                textStyle: const TextStyle(fontSize: 18),
              ),
            ),
            SizedBox(height: 20),
            Padding(
              padding: EdgeInsets.all(16.0),
              child: Text(
                global.language('form_design_features'),
                textAlign: TextAlign.center,
                style: const TextStyle(fontSize: 14, color: Colors.grey),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
