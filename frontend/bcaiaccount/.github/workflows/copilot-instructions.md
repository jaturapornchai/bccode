# คำแนะนำสำหรับ Flutter Project

## ภาษาที่ใช้
- **ตอบทุกคำถามเป็นภาษาไทยเสมอ**
- เขียน comment และ documentation เป็นภาษาไทยทั้งหมด
- อธิบายโค้ดและ logic เป็นภาษาไทย
- ชื่อ class, method, variable ใช้ภาษาอังกฤษ แต่อธิบายเป็นภาษาไทย

## Flutter Coding Style
- ใช้ Dart 3+ features (records, patterns, sealed classes)
- ใช้ const constructors เมื่อเป็นไปได้
- ใช้ Riverpod หรือ Provider สำหรับ state management
- แยก business logic ออกจาก UI
- สร้าง responsive UI ที่รองรับทั้ง Desktop (Windows) และ Web

## Project Structure
- `/lib/features/` - แยก features ตาม module
- `/lib/core/` - shared utilities, constants
- `/lib/widgets/` - reusable widgets
- เขียน documentation เป็นภาษาไทยทุกไฟล์

## Platform Support
- รองรับ Windows desktop (ความกว้างหน้าจอ 1024px ขึ้นไป)
- รองรับ Web responsive (mobile, tablet, desktop)
- ใช้ Platform.isXXX หรือ kIsWeb สำหรับ platform-specific code

## Code Quality
- เขียน unit tests และ widget tests
- ใช้ meaningful names ที่อธิบายวัตถุประสงค์
- เพิ่ม error handling และ loading states
- comment ซับซ้อน logic เป็นภาษาไทย

## Documentation Update Rule (สำคัญมาก!)
- **เมื่อแก้ไข code หรือ feature ใดๆ ต้องสรุปและอัพเดท documentation ใน /.github/skills/rules ด้วยเสมอ**
- ให้ AI ตัวอื่นๆ สามารถอ่านและเข้าใจระบบได้
- อ่าน `/.github/skills/rules/README.md` สำหรับโครงสร้างและรายละเอียด
- ศึกษากฎใน folder `/.github/skills/rules` ก่อนเริ่มทำงาน