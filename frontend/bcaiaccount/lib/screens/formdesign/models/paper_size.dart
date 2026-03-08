enum PaperSize {
  a4,
  b5,
}

extension PaperSizeExtension on PaperSize {
  String get name {
    switch (this) {
      case PaperSize.a4:
        return 'A4';
      case PaperSize.b5:
        return 'B5';
    }
  }

  // ขนาดในหน่วย points (72 points = 1 inch)
  double get width {
    switch (this) {
      case PaperSize.a4:
        return 595.0; // 210mm
      case PaperSize.b5:
        return 516.0; // 182mm
    }
  }

  double get height {
    switch (this) {
      case PaperSize.a4:
        return 842.0; // 297mm
      case PaperSize.b5:
        return 729.0; // 257mm
    }
  }

  String toJson() => name;

  static PaperSize fromJson(String json) {
    switch (json.toUpperCase()) {
      case 'A4':
        return PaperSize.a4;
      case 'B5':
        return PaperSize.b5;
      default:
        return PaperSize.a4;
    }
  }
}
