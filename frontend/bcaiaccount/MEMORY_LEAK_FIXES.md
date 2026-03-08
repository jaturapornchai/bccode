# แก้ไขปัญหา Frontend Hang หลังเปิดทิ้งไว้นาน

## สาเหตุหลักที่พบ

### 1. **Debouncer Memory Leak** (Fixed ✅)
- **ปัญหา**: Timer ใน Debouncer ไม่ถูก cancel เมื่อ widget dispose
- **ผลกระทบ**: Timer สะสมเรื่อยๆ ทำให้ memory leak
- **การแก้ไข**: เพิ่ม `cancel()` และ `dispose()` methods

### 2. **Timer Accumulation in PO List Screen** (Fixed ✅)
- **ปัญหา**: Multiple timers running concurrently
- **ผลกระทบ**: CPU usage สูง, memory leak
- **การแก้ไข**: Add checks to prevent duplicate timers

### 3. **PO Approval Polling Service** (Needs Fix ⚠️)
- **ปัญหา**: Timer อาจซ้อนกันหลายตัวถ้าเรียก startPolling ซ้ำ
- **ผลกระทบ**: API calls ซ้ำซ้อน, memory leak
- **การแก้ไข**: Add guard to prevent duplicate timers

### 4. **Branch Config Watcher Service** (Needs Fix ⚠️)
- **ปัญหา**: Timer อาจซ้อนกัน
- **ผลกระทบ**: Memory leak, unnecessary API calls
- **การแก้ไข**: Add guard to prevent duplicate timers

### 5. **CartWebSocketService Connection Leak** (Needs Fix ⚠️)
- **ปัญหา**: WebSocket connection ไม่ถูก cleanup ถูกต้อง
- **ผลกระทบ**: Connection สะสม, memory leak
- **การแก้ไข**: Improve disconnect logic, add connection state check

### 6. **Debouncer Instances Not Disposed** (Needs Fix ⚠️)
- **ปัญหา**: Debouncer instances ทั่วทั้ง app ไม่ถูก dispose
- **ผลกระทบ**: Timer สะสม
- **การแก้ไข**: Add dispose calls in all screens

---

## วิธีแก้ไขที่ทำแล้ว

### 1. Debouncer Class (global.dart)
```dart
class Debouncer {
  final int? milliseconds;
  VoidCallback? action;
  Timer? _timer;

  Debouncer(this.milliseconds);

  void run(VoidCallback action) {
    _timer?.cancel();
    _timer = Timer(Duration(milliseconds: milliseconds!), action);
  }

  /// Cancel any pending timer to prevent memory leaks
  void cancel() {
    _timer?.cancel();
    _timer = null;
  }

  /// Dispose the debouncer
  void dispose() {
    cancel();
  }
}
```

---

## การทดสอบ

### วิธีตรวจสอบว่าแก้ไขสำเร็จ:
1. เปิดแอปทิ้งไว้ 30 นาที
2. สลับหน้าจอไปมา (PO List → PO Edit → PO List)
3. ตรวจสอบ memory usage ใน DevTools
4. ตรวจสอบว่าไม่มี Timer/Stream ค้างใน console

### สัญญาณเตือนที่ควรเฝ้าระวัง:
- แอปเริ่มช้าลงหลังใช้งานนาน
- UI ไม่ตอบสนอง
- ต้อง force close แอป
