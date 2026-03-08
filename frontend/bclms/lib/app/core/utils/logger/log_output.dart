// Conditional export: เลือก implementation ตาม platform
// - Web → log_output_stub.dart (console only)
// - Native (Windows/Android/iOS) → log_output_io.dart (console + file)
export 'log_output_stub.dart'
    if (dart.library.io) 'log_output_io.dart';
