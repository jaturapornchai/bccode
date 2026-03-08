import 'package:json_annotation/json_annotation.dart';
import '../global.dart' as global;

part 'product_condition_model.g.dart';

@JsonSerializable(explicitToJson: true)
class ProductConditionModel {
  @JsonKey(name: 'product_codes')
  List<String>? productCodes;

  @JsonKey(name: 'group_codes')
  List<String>? groupCodes;

  @JsonKey(name: 'group_subone_codes')
  List<String>? groupSuboneCodes;

  @JsonKey(name: 'group_subtwo_codes')
  List<String>? groupSubtwoCodes;

  @JsonKey(name: 'brand_codes')
  List<String>? brandCodes;

  @JsonKey(name: 'design_codes')
  List<String>? designCodes;

  @JsonKey(name: 'model_codes')
  List<String>? modelCodes;

  @JsonKey(name: 'pattern_codes')
  List<String>? patternCodes;

  @JsonKey(name: 'grade_codes')
  List<String>? gradeCodes;

  @JsonKey(name: 'category_codes')
  List<String>? categoryCodes;

  @JsonKey(name: 'class_codes')
  List<String>? classCodes;

  @JsonKey(name: 'minimum_amount')
  double? minimumAmount;

  ProductConditionModel({
    this.productCodes,
    this.groupCodes,
    this.groupSuboneCodes,
    this.groupSubtwoCodes,
    this.brandCodes,
    this.designCodes,
    this.modelCodes,
    this.patternCodes,
    this.gradeCodes,
    this.categoryCodes,
    this.classCodes,
    this.minimumAmount,
  });

  factory ProductConditionModel.fromJson(Map<String, dynamic> json) =>
      _$ProductConditionModelFromJson(json);

  Map<String, dynamic> toJson() => _$ProductConditionModelToJson(this);

  ProductConditionModel copyWith({
    List<String>? productCodes,
    List<String>? groupCodes,
    List<String>? groupSuboneCodes,
    List<String>? groupSubtwoCodes,
    List<String>? brandCodes,
    List<String>? designCodes,
    List<String>? modelCodes,
    List<String>? patternCodes,
    List<String>? gradeCodes,
    List<String>? categoryCodes,
    List<String>? classCodes,
    double? minimumAmount,
  }) {
    return ProductConditionModel(
      productCodes: productCodes ?? this.productCodes,
      groupCodes: groupCodes ?? this.groupCodes,
      groupSuboneCodes: groupSuboneCodes ?? this.groupSuboneCodes,
      groupSubtwoCodes: groupSubtwoCodes ?? this.groupSubtwoCodes,
      brandCodes: brandCodes ?? this.brandCodes,
      designCodes: designCodes ?? this.designCodes,
      modelCodes: modelCodes ?? this.modelCodes,
      patternCodes: patternCodes ?? this.patternCodes,
      gradeCodes: gradeCodes ?? this.gradeCodes,
      categoryCodes: categoryCodes ?? this.categoryCodes,
      classCodes: classCodes ?? this.classCodes,
      minimumAmount: minimumAmount ?? this.minimumAmount,
    );
  }

  // Helper method to check if any product condition is set
  bool hasConditions() {
    return (productCodes?.isNotEmpty ?? false) ||
        (groupCodes?.isNotEmpty ?? false) ||
        (groupSuboneCodes?.isNotEmpty ?? false) ||
        (groupSubtwoCodes?.isNotEmpty ?? false) ||
        (brandCodes?.isNotEmpty ?? false) ||
        (designCodes?.isNotEmpty ?? false) ||
        (modelCodes?.isNotEmpty ?? false) ||
        (patternCodes?.isNotEmpty ?? false) ||
        (gradeCodes?.isNotEmpty ?? false) ||
        (categoryCodes?.isNotEmpty ?? false) ||
        (classCodes?.isNotEmpty ?? false) ||
        (minimumAmount != null && minimumAmount! > 0);
  }

  // Helper method to get minimum amount text
  String getMinimumAmountText() {
    if (minimumAmount == null || minimumAmount! <= 0) return global.language('not_set');
    return '${minimumAmount!.toStringAsFixed(0)} บาท';
  }
}
