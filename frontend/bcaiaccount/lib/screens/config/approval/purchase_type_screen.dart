import 'package:flutter/material.dart';
import 'package:smlaicloud/model/approval_model.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/services/approval_api_service.dart';
import 'package:smlaicloud/global.dart' as global;

/// หน้าจอประเภทการจัดซื้อ (Purchase Type Screen)
/// จัดการประเภทการจัดซื้อ เช่น ซื้อเพื่อใช้, ซื้อเพื่อขาย
class PurchaseTypeScreen extends StatefulWidget {
  const PurchaseTypeScreen({super.key});

  @override
  State<PurchaseTypeScreen> createState() => _PurchaseTypeScreenState();
}

class _PurchaseTypeScreenState extends State<PurchaseTypeScreen>
    with global.ThemeRefreshMixin {
  bool _isLoading = true;
  List<PurchaseTypeModel> _items = [];
  PurchaseTypeModel? _selectedItem;
  bool _isEditing = false;

  // Controllers
  final _formKey = GlobalKey<FormState>();
  final _codeController = TextEditingController();
  final _nameController = TextEditingController();
  final _descriptionController = TextEditingController();

  @override
  void initState() {
    super.initState();
    _loadData();
  }

  @override
  void dispose() {
    _codeController.dispose();
    _nameController.dispose();
    _descriptionController.dispose();
    super.dispose();
  }

  Future<void> _loadData() async {
    setState(() => _isLoading = true);

    final result = await ApprovalApiService.getPurchaseTypes();

    if (result.isSuccess) {
      setState(() {
        _items = result.items;
        _isLoading = false;
      });
    } else {
      setState(() => _isLoading = false);
      if (mounted) {
        _showError('${global.language("load_data_failed")}: ${result.errorMessage}');
      }
    }
  }

  void _showError(String message) {
    global.showErrorSnackBar(context, message);
  }

  void _showSuccess(String message) {
    global.showSuccessSnackBar(context, message);
  }

  void _clearForm() {
    _codeController.clear();
    _nameController.clear();
    _descriptionController.clear();
    _selectedItem = null;
    _isEditing = false;
  }

  void _loadItemToForm(PurchaseTypeModel item) {
    _codeController.text = item.code;
    _nameController.text = item.getName(global.systemLanguage);
    _descriptionController.text = item.getDescription(global.systemLanguage);
    _selectedItem = item;
    _isEditing = true;
  }

  Future<void> _saveItem() async {
    if (!_formKey.currentState!.validate()) return;

    setState(() => _isLoading = true);

    final model = PurchaseTypeModel(
      guid: _selectedItem?.guid,
      code: _codeController.text.trim().toUpperCase(),
      names: [LanguageDataModel(code: global.systemLanguage, name: _nameController.text.trim())],
      descriptions: _descriptionController.text.trim().isNotEmpty
          ? [LanguageDataModel(code: global.systemLanguage, name: _descriptionController.text.trim())]
          : [],
    );

    final result = await ApprovalApiService.savePurchaseType(model);

    if (result.isSuccess) {
      _showSuccess(global.language('save_success'));
      _clearForm();
      await _loadData();
    } else {
      setState(() => _isLoading = false);
      _showError('${global.language("save_failed")}: ${result.errorMessage}');
    }
  }

  Future<void> _deleteItem(PurchaseTypeModel item) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
        title: Row(
          children: [
            Icon(Icons.warning_amber_rounded, color: global.theme.warningHighlightTextColor),
            SizedBox(width: 8),
            Text(global.language('alert_confirm_delete')),
          ],
        ),
        content: Text('${global.language("confirm_delete")} "${item.getName(global.systemLanguage)}"?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: Text(global.language('cancel')),
          ),
          ElevatedButton(
            style: ElevatedButton.styleFrom(
              backgroundColor: global.theme.negativeHighlightTextColor,
              foregroundColor: global.theme.onPrimaryColor,
            ),
            onPressed: () => Navigator.pop(context, true),
            child: Text(global.language('delete')),
          ),
        ],
      ),
    );

    if (confirmed != true) return;

    setState(() => _isLoading = true);

    final success = await ApprovalApiService.deletePurchaseType(item.guid ?? '');

    if (success) {
      _showSuccess(global.language('delete_success'));
      if (_selectedItem?.code == item.code) {
        _clearForm();
      }
      await _loadData();
    } else {
      setState(() => _isLoading = false);
      _showError(global.language('delete_failed'));
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: global.theme.surfaceColor,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        elevation: 0,
        title: Row(
          children: [
            Container(
              padding: EdgeInsets.all(8),
              decoration: BoxDecoration(
                color: global.theme.cardColor.withValues(alpha: 0.2),
                borderRadius: BorderRadius.circular(8),
              ),
              child: Icon(Icons.category_outlined, size: 20),
            ),
            SizedBox(width: 12),
            Text(global.language('purchase_type')),
          ],
        ),
        leading: IconButton(
          icon: Icon(Icons.arrow_back),
          onPressed: () => global.gotoMainMenu(context),
        ),
        actions: [
          IconButton(
            icon: Icon(Icons.refresh),
            onPressed: _loadData,
            tooltip: global.language('database_master_info.refresh'),
          ),
        ],
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : _buildBody(),
    );
  }

  Widget _buildBody() {
    return LayoutBuilder(
      builder: (context, constraints) {
        if (constraints.maxWidth > 800) {
          // Wide screen - split view
          return Row(
            children: [
              Expanded(flex: 1, child: _buildListPanel()),
              Expanded(flex: 1, child: _buildFormPanel()),
            ],
          );
        } else {
          // Narrow screen - tabs or single view
          return _isEditing ? _buildFormPanel() : _buildListPanel();
        }
      },
    );
  }

  Widget _buildListPanel() {
    return Container(
      color: global.theme.cardColor,
      child: Column(
        children: [
          // Header
          Container(
            padding: EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: global.theme.cardColor,
              boxShadow: [
                BoxShadow(
                  color: global.theme.dividerBorderColor,
                  blurRadius: 4,
                  offset: const Offset(0, 2),
                ),
              ],
            ),
            child: Row(
              children: [
                Icon(Icons.list_alt, color: global.theme.appBarColor, size: 24),
                const SizedBox(width: 12),
                Text(
                  global.language('purchase_type_list'),
                  style: TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.bold,
                    color: global.theme.textColor,
                  ),
                ),
                const Spacer(),
                Container(
                  padding:
                      EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                  decoration: BoxDecoration(
                    color: global.theme.appBarColor.withValues(alpha: 0.1),
                    borderRadius: BorderRadius.circular(20),
                  ),
                  child: Text(
                    '${_items.length} ${global.language("items")}',
                    style: TextStyle(
                      fontSize: 12,
                      fontWeight: FontWeight.w600,
                      color: global.theme.appBarColor,
                    ),
                  ),
                ),
              ],
            ),
          ),
          // List
          Expanded(
            child: _items.isEmpty
                ? _buildEmptyState()
                : ListView.builder(
                    padding: const EdgeInsets.all(16),
                    itemCount: _items.length,
                    itemBuilder: (context, index) {
                      final item = _items[index];
                      final isSelected = _selectedItem?.code == item.code;
                      return _buildItemCard(item, isSelected);
                    },
                  ),
          ),
          // Add Button
          Padding(
            padding: EdgeInsets.all(16),
            child: SizedBox(
              width: double.infinity,
              child: ElevatedButton.icon(
                onPressed: () {
                  setState(() {
                    _clearForm();
                    _isEditing = true;
                  });
                },
                icon: Icon(Icons.add),
                label: Text(global.language('add_new_purchase_type')),
                style: ElevatedButton.styleFrom(
                  backgroundColor: global.theme.appBarColor,
                  foregroundColor: global.theme.onPrimaryColor,
                  padding: const EdgeInsets.symmetric(vertical: 14),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(12),
                  ),
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildEmptyState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Container(
            padding: EdgeInsets.all(24),
            decoration: BoxDecoration(
              color: global.theme.surfaceColor,
              shape: BoxShape.circle,
            ),
            child: Icon(
              Icons.category_outlined,
              size: 64,
              color: global.theme.iconSecondaryColor,
            ),
          ),
          const SizedBox(height: 24),
          Text(
            global.language('no_purchase_types'),
            style: TextStyle(
              fontSize: 18,
              fontWeight: FontWeight.w600,
              color: global.theme.textSecondaryColor,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            global.language('press_below_to_add_purchase_type'),
            style: TextStyle(
              fontSize: 14,
              color: global.theme.textSecondaryColor,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildItemCard(PurchaseTypeModel item, bool isSelected) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: () {
            setState(() {
              _loadItemToForm(item);
            });
          },
          borderRadius: BorderRadius.circular(12),
          child: Container(
            padding: EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: isSelected
                  ? global.theme.appBarColor.withValues(alpha: 0.1)
                  : global.theme.onPrimaryColor,
              borderRadius: BorderRadius.circular(12),
              border: Border.all(
                color: isSelected
                    ? global.theme.appBarColor
                    : global.theme.dividerBorderColor,
                width: isSelected ? 2 : 1,
              ),
              boxShadow: [
                BoxShadow(
                  color: global.theme.surfaceColor,
                  blurRadius: 4,
                  offset: const Offset(0, 2),
                ),
              ],
            ),
            child: Row(
              children: [
                // Code Badge
                Container(
                  padding:
                      EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                  decoration: BoxDecoration(
                    color: isSelected
                        ? global.theme.appBarColor
                        : global.theme.surfaceColor,
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Text(
                    item.code,
                    style: TextStyle(
                      fontWeight: FontWeight.bold,
                      fontSize: 13,
                      color: isSelected ? global.theme.onPrimaryColor : global.theme.textColor,
                    ),
                  ),
                ),
                const SizedBox(width: 16),
                // Name and Description
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        item.getName(global.systemLanguage),
                        style: TextStyle(
                          fontWeight: FontWeight.w600,
                          fontSize: 15,
                          color: global.theme.textColor,
                        ),
                      ),
                      if (item.getDescription(global.systemLanguage).isNotEmpty) ...[
                        const SizedBox(height: 4),
                        Text(
                          item.getDescription(global.systemLanguage),
                          style: TextStyle(
                            fontSize: 12,
                            color: global.theme.textSecondaryColor,
                          ),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),
                      ],
                    ],
                  ),
                ),
                // Actions
                Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    IconButton(
                      icon: Icon(Icons.edit_outlined,
                          color: global.theme.appBarColor),
                      onPressed: () {
                        setState(() {
                          _loadItemToForm(item);
                        });
                      },
                      tooltip: global.language('edit'),
                      splashRadius: 24,
                    ),
                    IconButton(
                      icon: Icon(Icons.delete_outline, color: global.theme.negativeHighlightTextColor),
                      onPressed: () => _deleteItem(item),
                      tooltip: global.language('delete'),
                      splashRadius: 24,
                    ),
                  ],
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildFormPanel() {
    return Container(
      color: global.theme.surfaceColor,
      child: Align(
        alignment: Alignment.topCenter,
        child: SingleChildScrollView(
          padding: EdgeInsets.all(24),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
            // Form Card
            Container(
              padding: EdgeInsets.all(24),
              color: global.theme.searchBarColor,
              child: Form(
                key: _formKey,
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    // Form Header
                    Row(
                      children: [
                        Container(
                          padding: EdgeInsets.all(12),
                          decoration: BoxDecoration(
                            color: _selectedItem == null
                                ? global.theme.positiveHighlightColor
                                : global.theme.warningHighlightColor,
                            borderRadius: BorderRadius.circular(12),
                          ),
                          child: Icon(
                            _selectedItem == null
                                ? Icons.add_circle_outline
                                : Icons.edit_outlined,
                            color: _selectedItem == null
                                ? global.theme.positiveHighlightTextColor
                                : global.theme.warningHighlightTextColor,
                            size: 28,
                          ),
                        ),
                        SizedBox(width: 16),
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                _selectedItem == null
                                    ? global.language('add_new_purchase_type')
                                    : global.language('edit_purchase_type'),
                                style: TextStyle(
                                  fontSize: 18,
                                  fontWeight: FontWeight.bold,
                                ),
                              ),
                              SizedBox(height: 4),
                              Text(
                                _selectedItem == null
                                    ? global.language('fill_data_to_add_new_type')
                                    : '${global.language("edit_type_data")}: ${_selectedItem!.code}',
                                style: TextStyle(
                                  fontSize: 13,
                                  color: global.theme.textSecondaryColor,
                                ),
                              ),
                            ],
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 32),

                    // Code Field
                    _buildTextField(
                      controller: _codeController,
                      label: global.language('code'),
                      hint: 'เช่น USE, SELL, STOCK',
                      icon: Icons.tag,
                      enabled: _selectedItem == null,
                      textCapitalization: TextCapitalization.characters,
                      validator: (value) {
                        if (value == null || value.trim().isEmpty) {
                          return global.language('please_enter_code');
                        }
                        return null;
                      },
                    ),
                    const SizedBox(height: 20),

                    // Name Field
                    _buildTextField(
                      controller: _nameController,
                      label: global.language('name'),
                      hint: 'เช่น ซื้อเพื่อใช้, ซื้อเพื่อขาย',
                      icon: Icons.label_outline,
                      enabled: _isEditing,
                      validator: (value) {
                        if (value == null || value.trim().isEmpty) {
                          return global.language('please_enter_name');
                        }
                        return null;
                      },
                    ),
                    const SizedBox(height: 20),

                    // Description Field
                    _buildTextField(
                      controller: _descriptionController,
                      label: global.language('bill_design_detail'),
                      hint: global.language('additional_description_optional'),
                      icon: Icons.notes,
                      enabled: _isEditing,
                      maxLines: 3,
                    ),
                    const SizedBox(height: 32),

                    // Action Buttons
                    if (_isEditing) ...[
                      Row(
                        children: [
                          Expanded(
                            child: OutlinedButton.icon(
                              onPressed: () {
                                setState(() {
                                  _clearForm();
                                });
                              },
                              icon: Icon(Icons.close),
                              label: Text(global.language('cancel')),
                              style: OutlinedButton.styleFrom(
                                padding:
                                    const EdgeInsets.symmetric(vertical: 14),
                                shape: RoundedRectangleBorder(
                                  borderRadius: BorderRadius.circular(12),
                                ),
                              ),
                            ),
                          ),
                          SizedBox(width: 16),
                          Expanded(
                            flex: 2,
                            child: ElevatedButton.icon(
                              onPressed: _saveItem,
                              icon: Icon(Icons.save),
                              label: Text(global.language('form_design_save')),
                              style: ElevatedButton.styleFrom(
                                backgroundColor: global.theme.appBarColor,
                                foregroundColor: global.theme.onPrimaryColor,
                                padding:
                                    const EdgeInsets.symmetric(vertical: 14),
                                shape: RoundedRectangleBorder(
                                  borderRadius: BorderRadius.circular(12),
                                ),
                              ),
                            ),
                          ),
                        ],
                      ),
                    ],
                  ],
                ),
              ),
            ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildTextField({
    required TextEditingController controller,
    required String label,
    required String hint,
    required IconData icon,
    bool enabled = true,
    int maxLines = 1,
    TextCapitalization textCapitalization = TextCapitalization.none,
    String? Function(String?)? validator,
  }) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Icon(icon, size: 18, color: global.theme.appBarColor),
            const SizedBox(width: 8),
            Text(
              label,
              style: TextStyle(
                fontSize: 14,
                fontWeight: FontWeight.w600,
                color: global.theme.textColor,
              ),
            ),
            if (validator != null) ...[
              const SizedBox(width: 4),
              Text(
                '*',
                style: TextStyle(
                  color: global.theme.negativeHighlightTextColor,
                  fontWeight: FontWeight.bold,
                ),
              ),
            ],
          ],
        ),
        const SizedBox(height: 8),
        TextFormField(
          controller: controller,
          enabled: enabled,
          maxLines: maxLines,
          textCapitalization: textCapitalization,
          decoration: InputDecoration(
            hintText: hint,
            hintStyle: TextStyle(color: global.theme.formHintColor),
            filled: true,
            fillColor: enabled ? global.theme.surfaceColor : global.theme.surfaceColor,
            border: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: BorderSide(color: global.theme.dividerBorderColor),
            ),
            enabledBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: BorderSide(color: global.theme.dividerBorderColor),
            ),
            focusedBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide:
                  BorderSide(color: global.theme.appBarColor, width: 2),
            ),
            disabledBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: BorderSide(color: global.theme.dividerBorderColor),
            ),
            contentPadding:
                const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
          ),
          validator: validator,
        ),
      ],
    );
  }
}
