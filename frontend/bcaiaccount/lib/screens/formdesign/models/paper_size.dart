enum PaperSize {
  a4,
  a5,
  b5,
  letter,
  legal,
  receipt80mm,
  receipt58mm,
}

extension PaperSizeExtension on PaperSize {
  String get name {
    switch (this) {
      case PaperSize.a4:
        return 'A4';
      case PaperSize.a5:
        return 'A5';
      case PaperSize.b5:
        return 'B5';
      case PaperSize.letter:
        return 'Letter';
      case PaperSize.legal:
        return 'Legal';
      case PaperSize.receipt80mm:
        return 'Receipt 80mm';
      case PaperSize.receipt58mm:
        return 'Receipt 58mm';
    }
  }

  String get description {
    switch (this) {
      case PaperSize.a4:
        return '210 x 297 mm';
      case PaperSize.a5:
        return '148 x 210 mm';
      case PaperSize.b5:
        return '182 x 257 mm';
      case PaperSize.letter:
        return '216 x 279 mm';
      case PaperSize.legal:
        return '216 x 356 mm';
      case PaperSize.receipt80mm:
        return '80 mm (thermal)';
      case PaperSize.receipt58mm:
        return '58 mm (thermal)';
    }
  }

  // ขนาดในหน่วย points (72 points = 1 inch)
  double get width {
    switch (this) {
      case PaperSize.a4:
        return 595.0; // 210mm
      case PaperSize.a5:
        return 420.0; // 148mm
      case PaperSize.b5:
        return 516.0; // 182mm
      case PaperSize.letter:
        return 612.0; // 8.5in
      case PaperSize.legal:
        return 612.0; // 8.5in
      case PaperSize.receipt80mm:
        return 227.0; // 80mm
      case PaperSize.receipt58mm:
        return 164.0; // 58mm
    }
  }

  double get height {
    switch (this) {
      case PaperSize.a4:
        return 842.0; // 297mm
      case PaperSize.a5:
        return 595.0; // 210mm
      case PaperSize.b5:
        return 729.0; // 257mm
      case PaperSize.letter:
        return 792.0; // 11in
      case PaperSize.legal:
        return 1008.0; // 14in
      case PaperSize.receipt80mm:
        return 0; // continuous
      case PaperSize.receipt58mm:
        return 0; // continuous
    }
  }

  bool get isContinuous =>
      this == PaperSize.receipt80mm || this == PaperSize.receipt58mm;

  String toJson() => name;

  static PaperSize fromJson(String json) {
    final upper = json.toUpperCase();
    for (final size in PaperSize.values) {
      if (size.name.toUpperCase() == upper) return size;
    }
    return PaperSize.a4;
  }
}
