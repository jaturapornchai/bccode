import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/business_type_model.dart';
import 'package:smlaicloud/screen_search/business_type_search_screen.dart';
import 'package:smlaicloud/usersystem/business_type_selector.dart';

class BusinessTypeSelectorWidget extends StatelessWidget {
  final BusinessTypeModel? businessType;
  final bool businessTypeNull;
  final ValueChanged<BusinessTypeModel> onBusinessTypeSelected;
  final VoidCallback onBusinessTypeCleared;

  const BusinessTypeSelectorWidget({
    super.key,
    required this.businessType,
    required this.businessTypeNull,
    required this.onBusinessTypeSelected,
    required this.onBusinessTypeCleared,
  });

  bool get _hasValue => businessType?.guidfixed?.isNotEmpty ?? false;

  @override
  Widget build(BuildContext context) {
    final info = getBusinessTypeInfo(businessType?.code);
    final name = _hasValue ? global.packName(businessType!.names!) : '';
    final code = _hasValue ? businessType!.code ?? '' : '';

    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 10),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Label
          Padding(
            padding: const EdgeInsets.only(bottom: 8),
            child: Text(
              global.language("business_type"),
              style: TextStyle(fontSize: 13, fontWeight: FontWeight.w600, color: global.theme.formLabelColor),
            ),
          ),
          // Card
          Material(
            color: Colors.transparent,
            child: InkWell(
              borderRadius: BorderRadius.circular(14),
              onTap: () => _openSearch(context),
              child: AnimatedContainer(
                duration: const Duration(milliseconds: 200),
                padding: const EdgeInsets.all(14),
                decoration: BoxDecoration(
                  color: _hasValue ? info.color.withValues(alpha: 0.06) : global.theme.formFillColor,
                  borderRadius: BorderRadius.circular(14),
                  border: Border.all(
                    color: businessTypeNull
                        ? global.theme.negativeHighlightTextColor
                        : (_hasValue ? info.color.withValues(alpha: 0.3) : global.theme.formBorderColor),
                    width: businessTypeNull ? 1.5 : 1,
                  ),
                ),
                child: Row(
                  children: [
                    // Icon
                    Container(
                      width: 44,
                      height: 44,
                      decoration: BoxDecoration(
                        gradient: _hasValue
                            ? LinearGradient(colors: [info.color, info.color.withValues(alpha: 0.7)])
                            : null,
                        color: _hasValue ? null : global.theme.surfaceColor,
                        borderRadius: BorderRadius.circular(12),
                        boxShadow: _hasValue
                            ? [BoxShadow(color: info.color.withValues(alpha: 0.25), blurRadius: 6, offset: const Offset(0, 2))]
                            : [],
                      ),
                      child: Icon(
                        _hasValue ? info.icon : Icons.add_business_rounded,
                        size: 22,
                        color: _hasValue ? Colors.white : global.theme.iconSecondaryColor,
                      ),
                    ),
                    const SizedBox(width: 14),
                    // Text
                    Expanded(
                      child: _hasValue
                          ? Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(
                                  name,
                                  style: TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: global.theme.textColor),
                                  overflow: TextOverflow.ellipsis,
                                ),
                                const SizedBox(height: 2),
                                Text(
                                  '${global.language("code")}: $code',
                                  style: TextStyle(fontSize: 11, color: global.theme.textSecondaryColor),
                                ),
                              ],
                            )
                          : Text(
                              global.language("please_select_business_type"),
                              style: TextStyle(fontSize: 14, color: global.theme.formHintColor),
                            ),
                    ),
                    // Actions
                    if (_hasValue)
                      IconButton(
                        icon: Icon(Icons.close_rounded, size: 18, color: global.theme.iconSecondaryColor),
                        onPressed: onBusinessTypeCleared,
                        constraints: const BoxConstraints(minWidth: 32, minHeight: 32),
                        padding: EdgeInsets.zero,
                        tooltip: global.language("clear"),
                      ),
                    Icon(Icons.chevron_right_rounded, size: 20, color: global.theme.iconSecondaryColor),
                  ],
                ),
              ),
            ),
          ),
          // Validation error
          if (businessTypeNull)
            Padding(
              padding: const EdgeInsets.only(top: 6, left: 4),
              child: Text(
                global.language("please_select_business_type"),
                style: TextStyle(fontSize: 11, color: global.theme.negativeHighlightTextColor),
              ),
            ),
        ],
      ),
    );
  }

  void _openSearch(BuildContext context) {
    Navigator.push(
      context,
      MaterialPageRoute(builder: (context) => const BusinessTypeSearchScreen(word: '')),
    ).then((value) {
      if (value != null) {
        BusinessTypeModel result = value;
        if (result.guidfixed!.isNotEmpty) {
          onBusinessTypeSelected(result);
        }
      }
    });
  }
}
