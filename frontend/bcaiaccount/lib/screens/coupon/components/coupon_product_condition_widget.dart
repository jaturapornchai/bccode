import 'package:flutter/material.dart';
import '../../../global.dart' as global;
import '../../../model/product_condition_model.dart';

class CouponProductConditionWidget extends StatelessWidget {
  final ProductConditionModel? productCondition;
  final bool isEditMode;
  final VoidCallback onAddProductCodes;
  final VoidCallback onAddGroupCodes;
  final VoidCallback onAddGroupSuboneCodes;
  final VoidCallback onAddGroupSubtwoCodes;
  final VoidCallback onAddBrandCodes;
  final VoidCallback onAddDesignCodes;
  final VoidCallback onAddModelCodes;
  final VoidCallback onAddPatternCodes;
  final VoidCallback onAddGradeCodes;
  final VoidCallback onAddCategoryCodes;
  final VoidCallback onAddClassCodes;
  final VoidCallback onEditMinimumAmount;
  final Function(String code, String type) onRemoveCode;

  const CouponProductConditionWidget({
    super.key,
    required this.productCondition,
    required this.isEditMode,
    required this.onAddProductCodes,
    required this.onAddGroupCodes,
    required this.onAddGroupSuboneCodes,
    required this.onAddGroupSubtwoCodes,
    required this.onAddBrandCodes,
    required this.onAddDesignCodes,
    required this.onAddModelCodes,
    required this.onAddPatternCodes,
    required this.onAddGradeCodes,
    required this.onAddCategoryCodes,
    required this.onAddClassCodes,
    required this.onEditMinimumAmount,
    required this.onRemoveCode,
  });

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // Section Header
        _buildSectionHeader(),
        const SizedBox(height: 16),

        // Description
        _buildDescription(),
        const SizedBox(height: 20),

        // Two Column Layout
        Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Left Column
            Expanded(
              child: Column(
                children: [
                  // Product Codes
                  _buildConditionRow(
                    label: global.language("product_codes"),
                    codes: productCondition?.productCodes,
                    type: 'product',
                    icon: Icons.inventory_2,
                    color: Colors.blue,
                    onAdd: onAddProductCodes,
                  ),

                  // Group Codes
                  _buildConditionRow(
                    label: global.language("group_codes"),
                    codes: productCondition?.groupCodes,
                    type: 'group',
                    icon: Icons.category,
                    color: Colors.purple,
                    onAdd: onAddGroupCodes,
                  ),

                  // Group Sub One Codes
                  _buildConditionRow(
                    label: global.language("group_subone_codes"),
                    codes: productCondition?.groupSuboneCodes,
                    type: 'group_subone',
                    icon: Icons.subdirectory_arrow_right,
                    color: Colors.deepPurple,
                    onAdd: onAddGroupSuboneCodes,
                  ),

                  // Group Sub Two Codes
                  _buildConditionRow(
                    label: global.language("group_subtwo_codes"),
                    codes: productCondition?.groupSubtwoCodes,
                    type: 'group_subtwo',
                    icon: Icons.double_arrow,
                    color: Colors.indigo,
                    onAdd: onAddGroupSubtwoCodes,
                  ),

                  // Brand Codes
                  _buildConditionRow(
                    label: global.language("brand_codes"),
                    codes: productCondition?.brandCodes,
                    type: 'brand',
                    icon: Icons.label,
                    color: Colors.orange,
                    onAdd: onAddBrandCodes,
                  ),

                  // Design Codes
                  _buildConditionRow(
                    label: global.language("design_codes"),
                    codes: productCondition?.designCodes,
                    type: 'design',
                    icon: Icons.design_services,
                    color: Colors.pink,
                    onAdd: onAddDesignCodes,
                  ),
                ],
              ),
            ),

            const SizedBox(width: 20),

            // Right Column
            Expanded(
              child: Column(
                children: [
                  // Model Codes
                  _buildConditionRow(
                    label: global.language("model_codes"),
                    codes: productCondition?.modelCodes,
                    type: 'model',
                    icon: Icons.style,
                    color: Colors.teal,
                    onAdd: onAddModelCodes,
                  ),

                  // Pattern Codes
                  _buildConditionRow(
                    label: global.language("pattern_codes"),
                    codes: productCondition?.patternCodes,
                    type: 'pattern',
                    icon: Icons.pattern,
                    color: Colors.green,
                    onAdd: onAddPatternCodes,
                  ),

                  // Grade Codes
                  _buildConditionRow(
                    label: global.language("grade_codes"),
                    codes: productCondition?.gradeCodes,
                    type: 'grade',
                    icon: Icons.grade,
                    color: Colors.amber,
                    onAdd: onAddGradeCodes,
                  ),

                  // Category Codes
                  _buildConditionRow(
                    label: global.language("category_codes"),
                    codes: productCondition?.categoryCodes,
                    type: 'category',
                    icon: Icons.dashboard,
                    color: Colors.cyan,
                    onAdd: onAddCategoryCodes,
                  ),

                  // Class Codes
                  _buildConditionRow(
                    label: global.language("class_codes"),
                    codes: productCondition?.classCodes,
                    type: 'class',
                    icon: Icons.class_,
                    color: Colors.brown,
                    onAdd: onAddClassCodes,
                  ),

                  // Minimum Amount
                  _buildMinimumAmountRow(),
                ],
              ),
            ),
          ],
        ),

        const SizedBox(height: 16),
      ],
    );
  }

  Widget _buildSectionHeader() {
    return Container(
      padding: const EdgeInsets.symmetric(vertical: 10, horizontal: 16),
      decoration: BoxDecoration(
        color: global.theme.appBarColor.withValues(alpha: 0.08),
        borderRadius: BorderRadius.circular(6),
        border: Border.all(color: global.theme.appBarColor.withValues(alpha: 0.2)),
      ),
      child: Row(
        children: [
          Icon(Icons.filter_alt, size: 20, color: global.theme.appBarColor),
          SizedBox(width: 10),
          Text(
            global.language("coupon_product_conditions"),
            style: TextStyle(
              fontSize: 16,
              fontWeight: FontWeight.w600,
              color: global.theme.appBarColor,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildDescription() {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: global.theme.appBarColor.withValues(alpha: 0.05),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: global.theme.appBarColor.withValues(alpha: 0.2)),
      ),
      child: Row(
        children: [
          Icon(
            Icons.info_outline,
            size: 20,
            color: global.theme.appBarColor.withValues(alpha: 0.7),
          ),
          SizedBox(width: 10),
          Expanded(
            child: Text(
              global.language("coupon_product_condition_description"),
              style: TextStyle(
                fontSize: 13,
                color: Colors.grey[700],
                height: 1.4,
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildConditionRow({
    required String label,
    required List<String>? codes,
    required String type,
    required IconData icon,
    required Color color,
    required VoidCallback onAdd,
  }) {
    final hasData = codes != null && codes.isNotEmpty;

    return Padding(
      padding: const EdgeInsets.only(bottom: 16),
      child: Container(
        padding: const EdgeInsets.all(12),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(8),
          border: Border.all(color: Colors.grey[300]!, width: 1),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(
                  icon,
                  size: 18,
                  color: global.theme.appBarColor.withValues(alpha: 0.7),
                ),
                const SizedBox(width: 8),
                Text(
                  label,
                  style: TextStyle(
                    fontSize: 14,
                    fontWeight: FontWeight.w500,
                    color: Colors.grey[700],
                  ),
                ),
                const Spacer(),
                if (isEditMode)
                  ElevatedButton.icon(
                    style: ElevatedButton.styleFrom(
                      backgroundColor: global.theme.appBarColor.withValues(alpha:
                        0.08,
                      ),
                      foregroundColor: global.theme.appBarColor,
                      elevation: 0,
                      padding: const EdgeInsets.symmetric(
                        horizontal: 12,
                        vertical: 8,
                      ),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(6),
                        side: BorderSide(
                          color: global.theme.appBarColor.withValues(alpha: 0.2),
                        ),
                      ),
                    ),
                    onPressed: onAdd,
                    icon: Icon(Icons.add, size: 16),
                    label: Text(
                      global.language("add"),
                      style: const TextStyle(fontSize: 13),
                    ),
                  ),
              ],
            ),
            const SizedBox(height: 8),
            if (hasData)
              Wrap(
                spacing: 8,
                runSpacing: 8,
                children: codes
                    .map((code) => _buildChip(code, type, color))
                    .toList(),
              )
            else
              Container(
                width: double.infinity,
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: Colors.grey[100],
                  borderRadius: BorderRadius.circular(6),
                  border: Border.all(color: Colors.grey[300]!),
                ),
                child: Text(
                  global.language("coupon_no_condition_set"),
                  style: TextStyle(
                    fontSize: 13,
                    color: Colors.grey[600],
                    fontStyle: FontStyle.italic,
                  ),
                ),
              ),
          ],
        ),
      ),
    );
  }

  Widget _buildChip(String code, String type, Color color) {
    return Chip(
      label: Text(
        code,
        style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w500),
      ),
      deleteIcon: isEditMode ? const Icon(Icons.close, size: 16) : null,
      onDeleted: isEditMode ? () => onRemoveCode(code, type) : null,
      backgroundColor: global.theme.appBarColor.withValues(alpha: 0.1),
      labelStyle: TextStyle(color: global.theme.appBarColor),
      deleteIconColor: Colors.grey[600],
      side: BorderSide(color: global.theme.appBarColor.withValues(alpha: 0.3)),
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
    );
  }

  Widget _buildMinimumAmountRow() {
    final hasMinimum =
        productCondition?.minimumAmount != null &&
        productCondition!.minimumAmount! > 0;

    return Padding(
      padding: EdgeInsets.only(bottom: 8),
      child: Container(
        padding: const EdgeInsets.all(12),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(8),
          border: Border.all(color: Colors.grey[300]!, width: 1),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(
                  Icons.payments,
                  size: 18,
                  color: global.theme.appBarColor.withValues(alpha: 0.7),
                ),
                SizedBox(width: 8),
                Text(
                  global.language("minimum_purchase_amount"),
                  style: TextStyle(
                    fontSize: 14,
                    fontWeight: FontWeight.w500,
                    color: Colors.grey[700],
                  ),
                ),
                const Spacer(),
                if (isEditMode)
                  ElevatedButton.icon(
                    style: ElevatedButton.styleFrom(
                      backgroundColor: global.theme.appBarColor.withValues(alpha:
                        0.08,
                      ),
                      foregroundColor: global.theme.appBarColor,
                      elevation: 0,
                      padding: const EdgeInsets.symmetric(
                        horizontal: 12,
                        vertical: 8,
                      ),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(6),
                        side: BorderSide(
                          color: global.theme.appBarColor.withValues(alpha: 0.2),
                        ),
                      ),
                    ),
                    onPressed: onEditMinimumAmount,
                    icon: Icon(hasMinimum ? Icons.edit : Icons.add, size: 16),
                    label: Text(
                      hasMinimum
                          ? global.language("edit")
                          : global.language("set_amount"),
                      style: const TextStyle(fontSize: 13),
                    ),
                  ),
              ],
            ),
            const SizedBox(height: 8),
            Container(
              width: double.infinity,
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: hasMinimum
                    ? global.theme.appBarColor.withValues(alpha: 0.05)
                    : Colors.grey[100],
                borderRadius: BorderRadius.circular(6),
                border: Border.all(
                  color: hasMinimum
                      ? global.theme.appBarColor.withValues(alpha: 0.2)
                      : Colors.grey[300]!,
                ),
              ),
              child: Row(
                children: [
                  Icon(
                    hasMinimum ? Icons.check_circle : Icons.info_outline,
                    size: 18,
                    color: hasMinimum
                        ? global.theme.appBarColor
                        : Colors.grey[600],
                  ),
                  SizedBox(width: 10),
                  Expanded(
                    child: Text(
                      hasMinimum
                          ? '${global.language("minimum_amount")}: ${productCondition!.minimumAmount!.toStringAsFixed(2)} ${global.language("baht")}'
                          : global.language("coupon_no_minimum_amount"),
                      style: TextStyle(
                        fontSize: 14,
                        fontWeight: hasMinimum
                            ? FontWeight.w600
                            : FontWeight.normal,
                        color: hasMinimum
                            ? global.theme.appBarColor
                            : Colors.grey[600],
                        fontStyle: hasMinimum
                            ? FontStyle.normal
                            : FontStyle.italic,
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}
