# UI Components & Conventions

## Logger

ใช้ `AppLogger` เสมอ — ห้ามใช้ `print()`

```dart
import 'package:smlaicloud/utils/logger/app_logger.dart';

AppLogger.info('message');
AppLogger.debug('message: $value');
AppLogger.error('error: $e');
AppLogger.warning('warning');
```

## Loading / Error / Empty States Pattern

ทุกหน้าที่ดึงข้อมูลจาก API ต้องมีครบ 3 states:

```dart
BlocBuilder<FeatureBloc, FeatureState>(
  builder: (context, state) {
    // Loading
    if (state is FeatureInProgress) {
      return const Center(child: CircularProgressIndicator());
    }
    // Error
    if (state is FeatureLoadFailed) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.error_outline, size: 48, color: Colors.red),
            const SizedBox(height: 8),
            Text(state.message),
            TextButton(
              onPressed: () => context.read<FeatureBloc>().add(
                FeatureLoadList(shopId: shopId)),
              child: const Text('ลองใหม่'),
            ),
          ],
        ),
      );
    }
    // Empty
    if (state is FeatureLoadSuccess && state.items.isEmpty) {
      return const Center(child: Text('ไม่มีข้อมูล'));
    }
    // Data
    if (state is FeatureLoadSuccess) {
      return FeatureListWidget(items: state.items);
    }
    return const SizedBox.shrink();
  },
)
```

## Form Save/Update Pattern

ใช้ `BlocConsumer` เมื่อ widget ต้องทั้ง rebuild และ listen side effects:

```dart
BlocConsumer<FeatureBloc, FeatureState>(
  listener: (context, state) {
    if (state is FeatureSaveSuccess || state is FeatureUpdateSuccess) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('บันทึกสำเร็จ')),
      );
      Navigator.pop(context, true); // ส่ง true กลับให้ caller refresh
    }
    if (state is FeatureSaveFailed || state is FeatureUpdateFailed) {
      final msg = state is FeatureSaveFailed
          ? (state as FeatureSaveFailed).message
          : (state as FeatureUpdateFailed).message;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(msg), backgroundColor: Colors.red),
      );
    }
  },
  builder: (context, state) {
    final isLoading = state is FeatureSaveInProgress || state is FeatureUpdateInProgress;
    return ElevatedButton(
      onPressed: isLoading ? null : _onSubmit,
      child: isLoading
          ? const SizedBox(width: 20, height: 20,
              child: CircularProgressIndicator(strokeWidth: 2))
          : const Text('บันทึก'),
    );
  },
)
```

## Document Status Badge

```dart
Widget buildStatusBadge(String status) {
  final config = {
    'draft':     ('ร่าง',        Colors.grey),
    'pending':   ('รออนุมัติ',   Colors.orange),
    'approved':  ('อนุมัติแล้ว', Colors.green),
    'completed': ('เสร็จสิ้น',   Colors.blue),
    'rejected':  ('ปฏิเสธ',      Colors.red),
    'cancelled': ('ยกเลิก',      Colors.red.shade300),
  };
  final (label, color) = config[status] ?? (status, Colors.grey);
  return Container(
    padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
    decoration: BoxDecoration(
      color: color.withOpacity(0.1),
      borderRadius: BorderRadius.circular(4),
      border: Border.all(color: color),
    ),
    child: Text(label, style: TextStyle(color: color, fontSize: 12)),
  );
}
```

## Action Button Visibility (ตาม Document State)

```dart
// ตรวจสอบ state ก่อนแสดงปุ่มเสมอ
final canEdit   = doc.docstatus == 'draft' || doc.docstatus == 'rejected';
final canSubmit = doc.docstatus == 'draft';
final canApprove = doc.docstatus == 'pending' && currentUser.isApprover;
final canCancel = !['completed', 'cancelled'].contains(doc.docstatus);

if (canEdit) ElevatedButton(onPressed: _onEdit, child: Text('แก้ไข')),
if (canSubmit) ElevatedButton(onPressed: _onSubmit, child: Text('ส่งอนุมัติ')),
if (canApprove) ElevatedButton(onPressed: _onApprove, child: Text('อนุมัติ')),
```

## Navigation

```dart
// ไปหน้าใหม่
Navigator.push(context, MaterialPageRoute(builder: (_) => FeatureScreen()));

// กลับพร้อม refresh
final result = await Navigator.push(...);
if (result == true) {
  context.read<FeatureBloc>().add(FeatureLoadList(shopId: shopId));
}

// ใช้ named routes (ถ้า project ใช้ app_routes.dart)
Navigator.pushNamed(context, AppRoutes.featureDetail, arguments: {'id': guid});
```

## Build Commands

```bash
# Run DEV
flutter run -t lib/main_bcaidev.dart

# Build Web DEV
flutter build web -t lib/main_bcaidev.dart --release

# Build Web PROD
flutter build web -t lib/main_bcaiprod.dart --release

# Build Windows DEV
flutter build windows -t lib/main_bcaidev.dart --release
```

## Code Conventions

- **ภาษา:** comment และ log ใช้ภาษาไทยได้ แต่ชื่อตัวแปร/คลาส/ฟังก์ชันใช้ภาษาอังกฤษ
- **สื่อสารกับ user:** ภาษาไทยเสมอ (error messages, SnackBar, Dialog)
- **ห้าม:** hardcode URL, API key, shopid
- **ห้าม:** Flutter เรียก AI API provider โดยตรง (Groq, OpenAI, Gemini)
