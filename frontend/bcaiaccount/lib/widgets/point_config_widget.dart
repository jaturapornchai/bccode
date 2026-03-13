import 'package:flutter/material.dart';
import 'package:smlaicloud/model/company_branch_model.dart';
import 'package:smlaicloud/widgets/date_picker_widget.dart';
import 'package:smlaicloud/global.dart' as global;

class PointConfigWidget extends StatefulWidget {
  final PointConfigModel pointConfig;
  final Function(PointConfigModel) onChanged;
  final bool isEditMode;

  const PointConfigWidget({
    super.key,
    required this.pointConfig,
    required this.onChanged,
    required this.isEditMode,
  });

  @override
  State<PointConfigWidget> createState() => _PointConfigWidgetState();
}

class _PointConfigWidgetState extends State<PointConfigWidget>
    with SingleTickerProviderStateMixin, global.ThemeRefreshMixin {
  late PointConfigModel _pointConfig;
  late TabController _tabController;
  @override
  void initState() {
    super.initState();
    _pointConfig = widget.pointConfig;
    _tabController = TabController(length: 2, vsync: this);

    // Ensure pointusagetype has default value of 1 if it's 0
    if (_pointConfig.pointusagetype == 0) {
      _pointConfig.pointusagetype = 1;
    }
  }

  @override
  void didUpdateWidget(PointConfigWidget oldWidget) {
    super.didUpdateWidget(oldWidget);
    // Always update when widget changes to ensure we get the latest data
    setState(() {
      _pointConfig = widget.pointConfig;
      // Ensure pointusagetype has default value of 1 if it's 0
      if (_pointConfig.pointusagetype == 0) {
        _pointConfig.pointusagetype = 1;
      }
    });
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(12.0),
      child: Container(
        decoration: BoxDecoration(
          color: global.theme.cardColor,
          borderRadius: BorderRadius.circular(12),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withValues(alpha: 0.05),
              spreadRadius: 1,
              blurRadius: 10,
              offset: const Offset(0, 2),
            ),
          ],
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Header with gradient
            Container(
              decoration: BoxDecoration(
                gradient: LinearGradient(
                  colors: [global.theme.infoHighlightTextColor, global.theme.infoHighlightTextColor],
                  begin: Alignment.topLeft,
                  end: Alignment.bottomRight,
                ),
                borderRadius:
                    const BorderRadius.vertical(top: Radius.circular(12)),
              ),
              padding: EdgeInsets.symmetric(vertical: 10, horizontal: 20),
              child: Row(
                children: [
                  Icon(Icons.card_giftcard,
                      color: global.theme.onPrimaryColor, size: 24),
                  SizedBox(width: 12),
                  Expanded(
                    child: Text(
                      global.language("point_system"),
                      style: TextStyle(
                        fontSize: 20,
                        fontWeight: FontWeight.bold,
                        color: global.theme.onPrimaryColor,
                      ),
                    ),
                  ),
                  if (widget.isEditMode)
                    Container(
                      padding: const EdgeInsets.symmetric(
                          horizontal: 12, vertical: 6),
                      decoration: BoxDecoration(
                        color: global.theme.cardColor.withValues(alpha: 0.2),
                        borderRadius: BorderRadius.circular(16),
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Icon(Icons.edit, size: 14, color: global.theme.onPrimaryColor),
                          SizedBox(width: 4),
                          Text(
                            global.language("edit_mode"),
                            style: TextStyle(
                              fontSize: 12,
                              fontWeight: FontWeight.w500,
                              color: global.theme.onPrimaryColor,
                            ),
                          ),
                        ],
                      ),
                    ),
                ],
              ),
            ),

            // Point Usage Type Section
            Padding(
              padding: const EdgeInsets.all(16),
              child: _buildPointUsageTypeSection(),
            ),

            // Tabs
            Container(
              decoration: BoxDecoration(
                color: global.theme.cardColor,
                border: Border(
                  top: BorderSide(color: global.theme.dividerBorderColor),
                  bottom: BorderSide(color: global.theme.dividerBorderColor),
                ),
              ),
              child: TabBar(
                controller: _tabController,
                labelColor: global.theme.infoHighlightTextColor,
                unselectedLabelColor: global.theme.textSecondaryColor,
                indicatorColor: global.theme.infoHighlightTextColor,
                indicatorWeight: 3,
                tabs: [
                  Tab(
                    icon: Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Icon(Icons.rule),
                        SizedBox(width: 8),
                        Text(global.language("general_rules")),
                      ],
                    ),
                  ),
                  Tab(
                    icon: Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Icon(Icons.star),
                        SizedBox(width: 8),
                        Text(global.language("special_rules")),
                      ],
                    ),
                  ),
                ],
              ),
            ),

            // Tab Content - THIS IS THE FIX: Use Expanded inside a SizedBox with fixed height instead of directly using Expanded
            SizedBox(
              height: 600, // Fixed height to prevent layout issues
              child: TabBarView(
                controller: _tabController,
                children: [
                  _buildGeneralRulesTab(),
                  _buildSpecialRulesTab(),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildPointUsageTypeSection() {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: global.theme.columnHeaderColor,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: global.theme.columnHeaderColor),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(Icons.settings, color: global.theme.infoHighlightTextColor, size: 18),
              SizedBox(width: 8),
              Text(
                global.language("point_usage_type"),
                style: TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.bold,
                  color: global.theme.columnHeaderTextColor,
                ),
              ),
            ],
          ),
          SizedBox(height: 12),
          Row(
            children: [
              Expanded(
                child: _buildUsageTypeOption(
                  value: 1,
                  title: global.language("discount"),
                  icon: Icons.discount,
                  description: global.language("use_points_for_discount"),
                ),
              ),
              SizedBox(width: 12),
              Expanded(
                child: _buildUsageTypeOption(
                  value: 2,
                  title: global.language("cash"),
                  icon: Icons.attach_money,
                  description: global.language("use_points_for_cash"),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildUsageTypeOption({
    required int value,
    required String title,
    required IconData icon,
    required String description,
  }) {
    final isSelected = _pointConfig.pointusagetype == value;

    return InkWell(
      onTap: widget.isEditMode
          ? () {
              setState(() {
                _pointConfig.pointusagetype = value;
                widget.onChanged(_pointConfig);
              });
            }
          : null,
      borderRadius: BorderRadius.circular(10),
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 200),
        padding: const EdgeInsets.all(12),
        decoration: BoxDecoration(
          color: isSelected ? global.theme.infoHighlightTextColor : global.theme.cardColor,
          borderRadius: BorderRadius.circular(10),
          border: Border.all(
            color: isSelected ? global.theme.infoHighlightTextColor : global.theme.dividerBorderColor,
            width: 2,
          ),
          boxShadow: isSelected
              ? [
                  BoxShadow(
                    color: global.theme.infoHighlightColor,
                    blurRadius: 8,
                    spreadRadius: 1,
                  )
                ]
              : null,
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(
                  icon,
                  color: isSelected ? global.theme.onPrimaryColor : global.theme.textSecondaryColor,
                  size: 20,
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    title,
                    style: TextStyle(
                      fontSize: 16,
                      fontWeight: FontWeight.bold,
                      color: isSelected ? global.theme.onPrimaryColor : global.theme.textColor,
                    ),
                  ),
                ),
                Icon(
                  isSelected
                      ? Icons.radio_button_checked
                      : Icons.radio_button_unchecked,
                  color: isSelected ? global.theme.onPrimaryColor : global.theme.iconSecondaryColor,
                ),
              ],
            ),
            const SizedBox(height: 8),
            Text(
              description,
              style: TextStyle(
                fontSize: 12,
                color: isSelected
                    ? Colors.white.withValues(alpha: 0.9)
                    : global.theme.textSecondaryColor,
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildGeneralRulesTab() {
    return SingleChildScrollView(
      padding: EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Tab title
          Row(
            children: [
              Icon(Icons.rule_folder, size: 20, color: global.theme.infoHighlightTextColor),
              SizedBox(width: 8),
              Text(
                global.language("general_rules_for_points"),
                style: TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.bold,
                  color: global.theme.columnHeaderTextColor,
                ),
              ),
            ],
          ),
          SizedBox(height: 8),
          Text(
            global.language("set_point_rate_and_value"),
            style: TextStyle(
              fontSize: 14,
              color: global.theme.textSecondaryColor,
            ),
          ),
          const SizedBox(height: 16),

          // Rules List
          ListView.builder(
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            itemCount: _pointConfig.generalrules.length,
            itemBuilder: (context, index) {
              return _buildGeneralRuleCard(index);
            },
          ),

          // Add Rule Button
          if (widget.isEditMode)
            Padding(
              padding: EdgeInsets.symmetric(vertical: 16),
              child: Center(
                child: ElevatedButton.icon(
                  onPressed: () {
                    setState(() {
                      _pointConfig.generalrules.add(GeneralRuleModel());
                      widget.onChanged(_pointConfig);
                    });
                  },
                  icon: Icon(Icons.add),
                  label: Text(global.language("add_rule")),
                  style: ElevatedButton.styleFrom(
                    foregroundColor: global.theme.onPrimaryColor,
                    backgroundColor: global.theme.infoHighlightTextColor,
                    padding: const EdgeInsets.symmetric(
                        horizontal: 20, vertical: 12),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(8),
                    ),
                  ),
                ),
              ),
            ),
        ],
      ),
    );
  }

  Widget _buildGeneralRuleCard(int index) {
    final rule = _pointConfig.generalrules[index];

    return Card(
      margin: const EdgeInsets.only(bottom: 16),
      elevation: 2,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: BorderSide(color: global.theme.columnHeaderColor),
      ),
      child: Padding(
        padding: EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Card header with delete button
            Row(
              children: [
                Container(
                  padding:
                      const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
                  decoration: BoxDecoration(
                    color: global.theme.columnHeaderColor,
                    borderRadius: BorderRadius.circular(16),
                  ),
                  child: Text(
                    "${global.language("rule")} #${index + 1}",
                    style: TextStyle(
                      fontWeight: FontWeight.bold,
                      color: global.theme.infoHighlightTextColor,
                    ),
                  ),
                ),
                const Spacer(),
                if (widget.isEditMode && _pointConfig.generalrules.length > 1)
                  IconButton(
                    icon: Icon(Icons.delete, color: global.theme.negativeHighlightTextColor),
                    tooltip: global.language("delete_rule"),
                    onPressed: () {
                      setState(() {
                        _pointConfig.generalrules.removeAt(index);
                        widget.onChanged(_pointConfig);
                      });
                    },
                  ),
              ],
            ),
            const SizedBox(height: 16),
            // Date Range Selection
            Row(
              children: [
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      DatePickerWidget(
                        label: global.language("start_date"),
                        selectedDate: _parseLocalDate(rule.startdate),
                        onDateSelected: widget.isEditMode
                            ? (date) {
                                if (date != null) {
                                  setState(() {
                                    rule.startdate = _formatToUtc(date);
                                    widget.onChanged(_pointConfig);
                                  });
                                }
                              }
                            : null,
                        isEnabled: widget.isEditMode,
                        isRequired: false,
                      ),
                    ],
                  ),
                ),
                SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      DatePickerWidget(
                        label: global.language("end_date"),
                        selectedDate: _parseLocalDate(rule.enddate),
                        onDateSelected: widget.isEditMode
                            ? (date) {
                                if (date != null) {
                                  setState(() {
                                    // Set to end of day in local time, then convert to UTC
                                    final endOfDay = DateTime(date.year,
                                        date.month, date.day, 23, 59, 59);
                                    rule.enddate = _formatToUtc(endOfDay);
                                    widget.onChanged(_pointConfig);
                                  });
                                }
                              }
                            : null,
                        isEnabled: widget.isEditMode,
                        isRequired: false,
                      ),
                    ],
                  ),
                ),
              ],
            ),

            const SizedBox(height: 24),

            // Points Configuration
            Row(
              children: [
                Expanded(
                  child: _buildNumberField(
                    label: global.language("amount_per_point"),
                    value: rule.payperpoint.toString(),
                    icon: Icons.money,
                    hint: "20",
                    onChanged: widget.isEditMode
                        ? (value) {
                            setState(() {
                              rule.payperpoint = double.tryParse(value) ?? 20;
                              widget.onChanged(_pointConfig);
                            });
                          }
                        : null,
                    helperText: global.language("amount_to_pay_for_1_point"),
                  ),
                ),
                SizedBox(width: 16),
                Expanded(
                  child: _buildNumberField(
                    label: global.language("points_per_baht"),
                    value: rule.pointvalue.toString(),
                    icon: Icons.star,
                    hint: "1",
                    onChanged: widget.isEditMode
                        ? (value) {
                            setState(() {
                              rule.pointvalue = double.tryParse(value) ?? 1;
                              widget.onChanged(_pointConfig);
                            });
                          }
                        : null,
                    helperText: "ใช้ ${rule.pointvalue} แต้ม เท่ากับ 1 บาท",
                  ),
                ),
              ],
            ),

            // Example calculation section
            SizedBox(height: 16),
            Row(
              children: [
                Expanded(
                  child: Container(
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      color: global.theme.surfaceColor,
                      borderRadius: BorderRadius.circular(8),
                      border: Border.all(color: global.theme.dividerBorderColor),
                    ),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          global.language("calculation_amount_per_point"),
                          style: TextStyle(
                            fontWeight: FontWeight.bold,
                            fontSize: 14,
                          ),
                        ),
                        const SizedBox(height: 8),
                        Text(
                          "หากคุณจ่าย 100 บาท คุณจะได้รับ ${rule.payperpoint > 0 ? (100 / rule.payperpoint).toStringAsFixed(2) : '∞'} แต้ม",
                          style: TextStyle(
                            fontSize: 12,
                            color: global.theme.textSecondaryColor,
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
                SizedBox(width: 16),
                Expanded(
                  child: Container(
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      color: global.theme.surfaceColor,
                      borderRadius: BorderRadius.circular(8),
                      border: Border.all(color: global.theme.dividerBorderColor),
                    ),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          global.language("point_redemption_example"),
                          style: TextStyle(
                            fontWeight: FontWeight.bold,
                            fontSize: 14,
                          ),
                        ),
                        const SizedBox(height: 8),
                        Text(
                          "หากคุณมี 100 แต้ม สามารถแลกได้ ${(100 / rule.pointvalue).toStringAsFixed(0)} บาท",
                          style: TextStyle(
                            fontSize: 12,
                            color: global.theme.textSecondaryColor,
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildSpecialRulesTab() {
    return SingleChildScrollView(
      padding: EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Tab title
          Row(
            children: [
              Icon(Icons.star, size: 20, color: global.theme.warningHighlightTextColor),
              SizedBox(width: 8),
              Text(
                global.language("special_rules_for_promotion"),
                style: TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.bold,
                  color: global.theme.warningHighlightTextColor,
                ),
              ),
            ],
          ),
          SizedBox(height: 8),
          Text(
            global.language("create_special_promo_multiply_points"),
            style: TextStyle(
              fontSize: 14,
              color: global.theme.textSecondaryColor,
            ),
          ),
          const SizedBox(height: 16),

          // Empty state if no special rules
          if (_pointConfig.specialrules.isEmpty) _buildEmptySpecialRules(),

          // Special Rules list
          ...List.generate(_pointConfig.specialrules.length, (index) {
            return _buildSpecialRuleCard(index);
          }),

          // Add Special Rule Button
          if (widget.isEditMode)
            Padding(
              padding: EdgeInsets.symmetric(vertical: 16),
              child: Center(
                child: ElevatedButton.icon(
                  onPressed: () {
                    setState(() {
                      _pointConfig.specialrules.add(SpecialRuleModel());
                      widget.onChanged(_pointConfig);
                    });
                  },
                  icon: Icon(Icons.add),
                  label: Text(global.language("add_special_rule")),
                  style: ElevatedButton.styleFrom(
                    foregroundColor: global.theme.onPrimaryColor,
                    backgroundColor: global.theme.warningHighlightTextColor,
                    padding: const EdgeInsets.symmetric(
                        horizontal: 20, vertical: 12),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(8),
                    ),
                  ),
                ),
              ),
            ),
        ],
      ),
    );
  }

  Widget _buildEmptySpecialRules() {
    return Center(
      child: Container(
        margin: const EdgeInsets.symmetric(vertical: 32),
        padding: const EdgeInsets.all(24),
        decoration: BoxDecoration(
          color: global.theme.surfaceColor,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: global.theme.dividerBorderColor),
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(
              Icons.star_border,
              size: 64,
              color: global.theme.warningHighlightTextColor,
            ),
            SizedBox(height: 16),
            Text(
              global.language("no_special_rules"),
              style: TextStyle(
                fontSize: 18,
                fontWeight: FontWeight.bold,
                color: global.theme.textSecondaryColor,
              ),
            ),
            SizedBox(height: 8),
            Text(
              global.language("add_special_rule_for_promotion"),
              textAlign: TextAlign.center,
              style: TextStyle(
                fontSize: 14,
                color: global.theme.textSecondaryColor,
              ),
            ),
            if (widget.isEditMode) ...[
              SizedBox(height: 24),
              ElevatedButton.icon(
                onPressed: () {
                  setState(() {
                    _pointConfig.specialrules.add(SpecialRuleModel());
                    widget.onChanged(_pointConfig);
                  });
                },
                icon: Icon(Icons.add),
                label: Text(global.language("add_first_special_rule")),
                style: ElevatedButton.styleFrom(
                  foregroundColor: global.theme.onPrimaryColor,
                  backgroundColor: global.theme.warningHighlightTextColor,
                  padding:
                      const EdgeInsets.symmetric(horizontal: 20, vertical: 12),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(8),
                  ),
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }

  Widget _buildSpecialRuleCard(int index) {
    final rule = _pointConfig.specialrules[index];

    return Card(
      margin: const EdgeInsets.only(bottom: 16),
      elevation: 2,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: BorderSide(color: global.theme.warningHighlightColor),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Card header with multiplier badge
          _buildSpecialRuleHeader(rule, index),

          // Card content
          Padding(
            padding: EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Date range
                Row(
                  children: [
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          DatePickerWidget(
                            label: global.language("start_date"),
                            selectedDate: _parseLocalDate(rule.startdate),
                            onDateSelected: widget.isEditMode
                                ? (date) {
                                    if (date != null) {
                                      setState(() {
                                        rule.startdate = _formatToUtc(date);
                                        widget.onChanged(_pointConfig);
                                      });
                                    }
                                  }
                                : null,
                            isEnabled: widget.isEditMode,
                            isRequired: false,
                          ),
                        ],
                      ),
                    ),
                    SizedBox(width: 12),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          DatePickerWidget(
                            label: global.language("end_date"),
                            selectedDate: _parseLocalDate(rule.enddate),
                            onDateSelected: widget.isEditMode
                                ? (date) {
                                    if (date != null) {
                                      setState(() {
                                        // Set to end of day in local time, then convert to UTC
                                        final endOfDay = DateTime(date.year,
                                            date.month, date.day, 23, 59, 59);
                                        rule.enddate = _formatToUtc(endOfDay);
                                        widget.onChanged(_pointConfig);
                                      });
                                    }
                                  }
                                : null,
                            isEnabled: widget.isEditMode,
                            isRequired: false,
                          ),
                        ],
                      ),
                    ),
                  ],
                ),

                const SizedBox(height: 24),

                // Points Configuration
                Row(
                  children: [
                    Expanded(
                      child: _buildNumberField(
                        label: global.language("point_multiplier"),
                        value: rule.multiplier.toString(),
                        icon: Icons.star_half,
                        hint: "2",
                        onChanged: widget.isEditMode
                            ? (value) {
                                setState(() {
                                  rule.multiplier = double.tryParse(value) ?? 2;
                                  widget.onChanged(_pointConfig);
                                });
                              }
                            : null,
                        helperText: global.language("multiplier_description"),
                      ),
                    ),
                    SizedBox(width: 16),
                    Expanded(
                      child: _buildNumberField(
                        label: global.language("max_points_per_bill"),
                        value: rule.maxpointperbill.toString(),
                        icon: Icons.receipt,
                        hint: "100",
                        onChanged: widget.isEditMode
                            ? (value) {
                                setState(() {
                                  rule.maxpointperbill =
                                      double.tryParse(value) ?? 100;
                                  widget.onChanged(_pointConfig);
                                });
                              }
                            : null,
                        helperText: global.language("max_points_per_bill_description"),
                      ),
                    ),
                  ],
                ),

                const SizedBox(height: 24),

                // Days of week
                _buildDaysOfWeekSelector(rule),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildSpecialRuleHeader(SpecialRuleModel rule, int index) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      decoration: BoxDecoration(
        gradient: LinearGradient(
          colors: [global.theme.warningHighlightTextColor, global.theme.warningHighlightTextColor],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        borderRadius: const BorderRadius.vertical(top: Radius.circular(12)),
      ),
      child: Row(
        children: [
          Icon(Icons.emoji_events, color: global.theme.onPrimaryColor, size: 20),
          SizedBox(width: 8),
          Text(
            "${global.language("special_promotion")} #${index + 1}",
            style: TextStyle(
              fontSize: 16,
              fontWeight: FontWeight.bold,
              color: global.theme.onPrimaryColor,
            ),
          ),
          const Spacer(),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
            decoration: BoxDecoration(
              color: global.theme.cardColor.withValues(alpha: 0.3),
              borderRadius: BorderRadius.circular(16),
            ),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                Icon(Icons.star, size: 14, color: global.theme.onPrimaryColor),
                const SizedBox(width: 4),
                Text(
                  "x${rule.multiplier}",
                  style: TextStyle(
                    fontSize: 14,
                    fontWeight: FontWeight.bold,
                    color: global.theme.onPrimaryColor,
                  ),
                ),
              ],
            ),
          ),
          if (widget.isEditMode)
            IconButton(
              icon: Icon(Icons.delete, color: global.theme.onPrimaryColor),
              tooltip: global.language("delete_rule"),
              onPressed: () {
                setState(() {
                  _pointConfig.specialrules.removeAt(index);
                  widget.onChanged(_pointConfig);
                });
              },
            ),
        ],
      ),
    );
  }

  Widget _buildDaysOfWeekSelector(SpecialRuleModel rule) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        _buildFormLabel(global.language("days_of_week")),
        SizedBox(height: 8),
        Text(
          global.language("select_promotion_days"),
          style: TextStyle(
            fontSize: 12,
            color: global.theme.textSecondaryColor,
            fontStyle: FontStyle.italic,
          ),
        ),
        const SizedBox(height: 12),
        Container(
          padding: const EdgeInsets.all(12),
          decoration: BoxDecoration(
            color: global.theme.surfaceColor,
            borderRadius: BorderRadius.circular(12),
            border: Border.all(color: global.theme.dividerBorderColor),
          ),
          child: Column(
            children: [
              Row(
                children: [
                  Expanded(
                      child: _buildDayChip("monday", rule.monday,
                          (value) => rule.monday = value)),
                  const SizedBox(width: 8),
                  Expanded(
                      child: _buildDayChip("tuesday", rule.tuesday,
                          (value) => rule.tuesday = value)),
                ],
              ),
              const SizedBox(height: 8),
              Row(
                children: [
                  Expanded(
                      child: _buildDayChip("wednesday", rule.wednesday,
                          (value) => rule.wednesday = value)),
                  const SizedBox(width: 8),
                  Expanded(
                      child: _buildDayChip("thursday", rule.thursday,
                          (value) => rule.thursday = value)),
                ],
              ),
              const SizedBox(height: 8),
              Row(
                children: [
                  Expanded(
                      child: _buildDayChip("friday", rule.friday,
                          (value) => rule.friday = value)),
                  const SizedBox(width: 8),
                  Expanded(
                      child: _buildDayChip("saturday", rule.saturday,
                          (value) => rule.saturday = value)),
                ],
              ),
              const SizedBox(height: 8),
              Row(
                children: [
                  Expanded(
                      child: _buildDayChip("sunday", rule.sunday,
                          (value) => rule.sunday = value)),
                  const Expanded(child: SizedBox()), // Empty to balance the row
                ],
              ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildDayChip(String day, bool isSelected, Function(bool) onChanged) {
    // Map day codes to full Thai day names
    final Map<String, String> dayNames = {
      "monday": global.language("monday"),
      "tuesday": global.language("tuesday"),
      "wednesday": global.language("wednesday"),
      "thursday": global.language("thursday"),
      "friday": global.language("friday"),
      "saturday": global.language("saturday"),
      "sunday": global.language("sunday"),
    };

    final String dayName = dayNames[day] ?? day;

    return Material(
      child: InkWell(
        onTap: widget.isEditMode
            ? () {
                setState(() {
                  onChanged(!isSelected);
                  widget.onChanged(_pointConfig);
                });
              }
            : null,
        borderRadius: BorderRadius.circular(8),
        child: Ink(
          padding: const EdgeInsets.symmetric(vertical: 10),
          decoration: BoxDecoration(
            color: isSelected ? global.theme.warningHighlightColor : global.theme.cardColor,
            border: Border.all(
              color: isSelected ? global.theme.warningHighlightTextColor : global.theme.dividerBorderColor,
              width: 1.5,
            ),
            borderRadius: BorderRadius.circular(8),
          ),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(
                isSelected ? Icons.check_circle : Icons.circle_outlined,
                color:
                    isSelected ? global.theme.warningHighlightTextColor : global.theme.iconSecondaryColor,
                size: 16,
              ),
              const SizedBox(width: 6),
              Text(
                dayName,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: isSelected ? FontWeight.bold : FontWeight.normal,
                  color:
                      isSelected ? global.theme.warningHighlightTextColor : global.theme.textColor,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildNumberField({
    required String label,
    required String value,
    required IconData icon,
    required String hint,
    required String helperText,
    Function(String)? onChanged,
  }) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        _buildFormLabel(label),
        const SizedBox(height: 8),
        TextFormField(
          initialValue: value,
          enabled: onChanged != null,
          keyboardType: TextInputType.number,
          decoration: InputDecoration(
            prefixIcon: Icon(icon),
            hintText: hint,
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(8)),
            contentPadding:
                const EdgeInsets.symmetric(horizontal: 12, vertical: 12),
            isDense: true,
          ),
          onChanged: onChanged,
        ),
        if (helperText.isNotEmpty) ...[
          const SizedBox(height: 4),
          Text(
            helperText,
            style: TextStyle(
              fontSize: 12,
              color: global.theme.textSecondaryColor,
              fontStyle: FontStyle.italic,
            ),
          ),
        ],
      ],
    );
  }

  Widget _buildFormLabel(String text) {
    return Text(
      text,
      style: TextStyle(
        fontSize: 14,
        fontWeight: FontWeight.w500,
        color: global.theme.textColor,
      ),
    );
  }

  // Helper functions for date/time conversion
  DateTime _parseLocalDate(String isoString) {
    try {
      final utcDate = DateTime.parse(isoString);
      return utcDate.toLocal();
    } catch (e) {
      return DateTime.now();
    }
  }

  String _formatToUtc(DateTime localDateTime) {
    return localDateTime.toUtc().toIso8601String();
  }
}

