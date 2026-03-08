# BC Ai Cloud

*by jead*
## วิธีการ Build และ Deploy ตาม Flavor

แอปพลิเคชันนี้มี 3 flavors

1. BC Ai Cloud DEV
2. BC Ai Cloud UAT
3. BC Ai Cloud PROD

## การ Build สำหรับ Web

# Deploy ไป Development
.\scripts\deploy_dev.bat

# Deploy ไป UAT
.\scripts\deploy_uat.bat

# Deploy ไป Production (ต้องพิมพ์ YES เพื่อยืนยัน)
.\scripts\deploy_prod.bat

### BC Ai Cloud DEV

เพื่อทดสอบล่าสุด
host: https://smlai-cloud-dev-2.web.app

```bash
flutter build web -t lib/main_bcaidev.dart --release --no-tree-shake-icons --pwa-strategy none
firebase deploy --only hosting:bcai-dev
```

### BC Ai Cloud UAT

host: https://smlai-cloud-uat.web.app

```bash
set DART_VM_OPTIONS=--max-heap-size=4096m
flutter build web -t lib/main_bcaiuat.dart --release --no-tree-shake-icons --pwa-strategy none
firebase deploy --only hosting:bcai-uat
```

### BC Ai Cloud PROD

host: https://bcai-cloud.web.app

```bash
flutter build web -t lib/main_bcaiprod.dart --release --no-tree-shake-icons --pwa-strategy none
firebase deploy --only hosting:bcaicloud
```

## Code Generation

```bash
flutter pub run build_runner build
flutter pub run build_runner build --delete-conflicting-outputs
```
