// Stub implementation for non-web platforms
/// Trigger download for web platform (stub for non-web platforms)
void triggerWebDownload(List<int> bytes, String fileName) {
  // This function should never be called on non-web platforms
  throw UnsupportedError('Web download is not supported on this platform');
}
