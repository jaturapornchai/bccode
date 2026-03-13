import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/global.dart' as global;

// Custom Label TextField Component
class LabelTextField extends StatelessWidget {
  final String label;
  final String? hint;
  final String initialValue;
  final Widget? prefix;
  final Widget? suffix;
  final bool isRequired;
  final bool readOnly;
  final int maxLines;
  final TextInputAction? textInputAction;
  final TextInputType? keyboardType;
  final Function(String) onChanged;

  const LabelTextField({
    super.key,
    required this.label,
    this.hint,
    required this.initialValue,
    this.prefix,
    this.suffix,
    this.isRequired = false,
    this.readOnly = false,
    this.maxLines = 1,
    this.textInputAction,
    this.keyboardType,
    required this.onChanged,
  });

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // Label outside the TextField
        Row(
          children: [
            if (prefix != null) ...[prefix!, const SizedBox(width: 8)],
            Text(
              label,
              style: TextStyle(fontWeight: FontWeight.w500, fontSize: 14),
            ),
            if (isRequired) ...[
              const SizedBox(width: 4),
              Text(
                '*',
                style: TextStyle(
                  color: Theme.of(context).colorScheme.error,
                  fontWeight: FontWeight.bold,
                ),
              ),
            ],
          ],
        ),
        const SizedBox(height: 6),
        // TextField
        TextFormField(
          initialValue: initialValue,
          readOnly: readOnly,
          maxLines: maxLines,
          textInputAction: textInputAction,
          keyboardType: keyboardType,
          decoration: InputDecoration(
            hintText: hint,
            suffixIcon: suffix,
            contentPadding: const EdgeInsets.symmetric(
              horizontal: 12,
              vertical: 10,
            ),
            border: OutlineInputBorder(
              borderRadius: BorderRadius.circular(4),
              borderSide: BorderSide(color: Theme.of(context).dividerColor),
            ),
            isDense: true,
          ),
          onChanged: onChanged,
          validator: isRequired
              ? (value) {
                  if (value == null || value.isEmpty) {
                    return global.language('required_field');
                  }
                  return null;
                }
              : null,
        ),
      ],
    );
  }
}

// Language Item Component
class LanguageItem extends StatelessWidget {
  final int index;
  final LanguageModel language;
  final bool isDefault;
  final VoidCallback onSelect;
  final VoidCallback? onDelete;

  const LanguageItem({
    super.key,
    required this.index,
    required this.language,
    required this.isDefault,
    required this.onSelect,
    this.onDelete,
  });

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        SizedBox(
          width: 28,
          child: Text(
            (index + 1).toString(),
            style: TextStyle(fontWeight: FontWeight.w500),
          ),
        ),
        Expanded(
          child: InkWell(
            onTap: onSelect,
            borderRadius: BorderRadius.circular(4),
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
              decoration: BoxDecoration(
                border: Border.all(color: Theme.of(context).dividerColor),
                borderRadius: BorderRadius.circular(4),
                color: isDefault
                    ? Theme.of(
                        context,
                      ).colorScheme.primaryContainer.withValues(alpha: 0.3)
                    : null,
              ),
              child: Row(
                children: [
                  if (language.code != null && language.code!.isNotEmpty) ...[
                    Image.asset(
                      'assets/flags/${language.code}.png',
                      width: 20,
                      height: 14,
                    ),
                    const SizedBox(width: 12),
                  ],
                  Expanded(
                    child: Text(
                      (language.code != null && language.code!.isNotEmpty)
                          ? language.name!
                          : global.language('select_language'),
                    ),
                  ),
                  Icon(
                    Icons.arrow_drop_down,
                    color: Theme.of(context).colorScheme.primary,
                  ),
                ],
              ),
            ),
          ),
        ),
        if (onDelete != null)
          IconButton(
            onPressed: onDelete,
            icon: Icon(
              Icons.delete_outline,
              color: Theme.of(context).colorScheme.error,
              size: 20,
            ),
            constraints: const BoxConstraints(minWidth: 32, minHeight: 32),
            padding: EdgeInsets.zero,
            visualDensity: VisualDensity.compact,
          ),
      ],
    );
  }
}

// Section Container Component
class SectionContainer extends StatelessWidget {
  final String title;
  final IconData icon;
  final List<Widget> children;
  final EdgeInsetsGeometry padding;

  const SectionContainer({
    super.key,
    required this.title,
    required this.icon,
    required this.children,
    this.padding = const EdgeInsets.all(16),
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      margin: const EdgeInsets.only(bottom: 16),
      decoration: BoxDecoration(
        border: Border.all(color: Theme.of(context).dividerColor),
        borderRadius: BorderRadius.circular(4),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Title bar
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
            decoration: BoxDecoration(
              color: Theme.of(context).colorScheme.surfaceContainerHighest,
              borderRadius: const BorderRadius.vertical(
                top: Radius.circular(3),
              ),
            ),
            child: Row(
              children: [
                Icon(
                  icon,
                  size: 18,
                  color: Theme.of(context).colorScheme.primary,
                ),
                const SizedBox(width: 8),
                Text(
                  title,
                  style: TextStyle(
                    fontWeight: FontWeight.bold,
                    fontSize: 15,
                  ),
                ),
              ],
            ),
          ),
          // Content
          Padding(
            padding: padding,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: children,
            ),
          ),
        ],
      ),
    );
  }
}

// API Key Field Component
class ApiKeyField extends StatelessWidget {
  final TextEditingController controller;
  final bool isLoading;
  final VoidCallback onGetApiKey;

  const ApiKeyField({
    super.key,
    required this.controller,
    required this.isLoading,
    required this.onGetApiKey,
  });

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          'BC API KEY',
          style: TextStyle(fontWeight: FontWeight.w500, fontSize: 14),
        ),
        SizedBox(height: 6),
        Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Expanded(
              child: TextFormField(
                readOnly: true,
                controller: controller,
                decoration: InputDecoration(
                  hintText: global.language("no_api_key"),
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(4),
                  ),
                  contentPadding: const EdgeInsets.symmetric(
                    horizontal: 12,
                    vertical: 10,
                  ),
                  isDense: true,
                  suffixIcon: controller.text.isNotEmpty
                      ? IconButton(
                          icon: Icon(Icons.copy, size: 20),
                          onPressed: () {
                            Clipboard.setData(
                              ClipboardData(text: controller.text),
                            );
                            global.showSuccessSnackBar(context, global.language("copy_api_key_success"));
                          },
                        )
                      : null,
                ),
              ),
            ),
            SizedBox(width: 8),
            SizedBox(
              height: 40,
              child: ElevatedButton(
                onPressed: isLoading ? null : onGetApiKey,
                style: ElevatedButton.styleFrom(
                  padding: const EdgeInsets.symmetric(horizontal: 8),
                ),
                child: isLoading
                    ? const SizedBox(
                        width: 16,
                        height: 16,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      )
                    : Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Icon(Icons.vpn_key, size: 16),
                          SizedBox(width: 4),
                          Text(
                            controller.text.isEmpty
                                ? global.language("get")
                                : global.language("refresh"),
                            style: TextStyle(fontSize: 13),
                          ),
                        ],
                      ),
              ),
            ),
          ],
        ),
      ],
    );
  }
}

// Language Selection Bottom Sheet
Future<String?> showLanguageSelectionSheet(
  BuildContext context,
  List<LanguageModel> availableLanguages,
) async {
  return showModalBottomSheet<String>(
    context: context,
    builder: (BuildContext context) {
      return Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Padding(
            padding: EdgeInsets.symmetric(horizontal: 16, vertical: 12),
            child: Row(
              children: [
                Text(
                  global.language('select_language'),
                  style: TextStyle(
                    fontWeight: FontWeight.bold,
                    fontSize: 16,
                  ),
                ),
                const Spacer(),
                IconButton(
                  icon: Icon(Icons.close),
                  onPressed: () => Navigator.pop(context),
                  padding: EdgeInsets.zero,
                  constraints: const BoxConstraints(
                    minWidth: 32,
                    minHeight: 32,
                  ),
                ),
              ],
            ),
          ),
          const Divider(height: 1),
          Expanded(
            child: ListView.builder(
              itemCount: availableLanguages.length,
              itemBuilder: (context, index) {
                return ListTile(
                  leading: Image.asset(
                    'assets/flags/${availableLanguages[index].code}.png',
                    width: 24,
                    height: 16,
                  ),
                  title: Text(availableLanguages[index].name!),
                  onTap: () {
                    Navigator.pop(context, availableLanguages[index].code);
                  },
                  dense: true,
                );
              },
            ),
          ),
        ],
      );
    },
  );
}

// Confirmation Dialog
Future<bool> showConfirmationDialog(
  BuildContext context,
  String title,
  String content,
) async {
  final result = await showDialog<bool>(
    context: context,
    builder: (BuildContext context) {
      return AlertDialog(
        title: Text(title),
        content: Text(content),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(false),
            child: Text(global.language('cancel')),
          ),
          TextButton(
            onPressed: () => Navigator.of(context).pop(true),
            style: TextButton.styleFrom(
              foregroundColor: Theme.of(context).colorScheme.error,
            ),
            child: Text(global.language('delete')),
          ),
        ],
      );
    },
  );
  return result ?? false;
}

// Custom Action Button
class ActionButton extends StatelessWidget {
  final String label;
  final IconData icon;
  final bool isLoading;
  final VoidCallback onPressed;
  final bool isPrimary;

  const ActionButton({
    super.key,
    required this.label,
    required this.icon,
    this.isLoading = false,
    required this.onPressed,
    this.isPrimary = true,
  });

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      width: double.infinity,
      height: 44,
      child: ElevatedButton.icon(
        onPressed: isLoading ? null : onPressed,
        style: ElevatedButton.styleFrom(
          backgroundColor: isPrimary
              ? Theme.of(context).colorScheme.primary
              : Theme.of(context).colorScheme.surfaceContainerHighest,
          foregroundColor: isPrimary
              ? Theme.of(context).colorScheme.onPrimary
              : Theme.of(context).colorScheme.onSurfaceVariant,
        ),
        icon: isLoading
            ? SizedBox(
                width: 16,
                height: 16,
                child: CircularProgressIndicator(
                  strokeWidth: 2,
                  color: isPrimary
                      ? Theme.of(context).colorScheme.onPrimary
                      : Theme.of(context).colorScheme.primary,
                ),
              )
            : Icon(icon, size: 18),
        label: Text(label),
      ),
    );
  }
}
