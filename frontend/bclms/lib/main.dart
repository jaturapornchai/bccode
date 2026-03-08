import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:bclms/app/core/theme/app_theme.dart';
import 'package:bclms/app/core/utils/app_logger.dart';
import 'package:bclms/app/providers/service_providers.dart';
import 'package:bclms/app/routes/app_routes.dart';
import 'package:bclms/app/services/api_service.dart';
import 'package:bclms/app/services/storage_service.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();

  AppLogger.clearLogFile();
  AppLogger.info('[App] ล้างข้อมูล — log file ถูกลบแล้ว');
  AppLogger.info('[App] กำลังเริ่มต้น services...');

  final storage = await StorageService().init();
  final api = await ApiService(storage).init();

  AppLogger.info('[App] เริ่มต้น services ทั้งหมดสำเร็จ');

  runApp(ProviderScope(
    overrides: [
      storageServiceProvider.overrideWithValue(storage),
      apiServiceProvider.overrideWithValue(api),
    ],
    child: const BclmsApp(),
  ));
}

class BclmsApp extends ConsumerWidget {
  const BclmsApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final router = ref.watch(goRouterProvider);

    return MaterialApp.router(
      title: 'BC LMS',
      debugShowCheckedModeBanner: false,
      theme: AppTheme.light,
      locale: const Locale('th', 'TH'),
      localizationsDelegates: const [
        GlobalMaterialLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
      ],
      supportedLocales: const [
        Locale('th', 'TH'),
        Locale('en', 'US'),
      ],
      routerConfig: router,
    );
  }
}
