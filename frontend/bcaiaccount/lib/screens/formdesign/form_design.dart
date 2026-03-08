import 'dart:convert';
import 'package:smlaicloud/widgets/manual_button.dart';
import 'dart:io';
import 'package:flutter/material.dart';
import 'package:file_picker/file_picker.dart';
import 'package:image_picker/image_picker.dart';
import 'package:smlaicloud/global.dart' as global;
import 'models/form_template.dart';
import 'models/form_element.dart';
import 'models/paper_size.dart';
import 'widgets/form_preview.dart';
import 'widgets/full_form_editor.dart';

class FormDesignScreen extends StatefulWidget {
  const FormDesignScreen({super.key});

  @override
  State<FormDesignScreen> createState() => _FormDesignScreenState();
}

class _FormDesignScreenState extends State<FormDesignScreen> {
  late FormTemplate currentTemplate;
  bool _isPreviewMode = false;
  final ImagePicker _picker = ImagePicker();
  double _zoomLevel = 1.0; // 100%
  bool _showRulers = true;
  bool _showJsonPanel = true; // Show JSON Panel
  late TextEditingController _jsonController;
  String _jsonError = '';
  FormElement? selectedElement;

  @override
  void initState() {
    super.initState();
    currentTemplate = FormTemplate(
      name: global.language('form_design_new_form'),
      description: global.language('form_design_created_with_designer'),
    );
    _jsonController = TextEditingController(
      text: _formatJson(currentTemplate.toJson()),
    );
  }

  @override
  void dispose() {
    _jsonController.dispose();
    super.dispose();
  }

  String _formatJson(Map<String, dynamic> json) {
    const encoder = JsonEncoder.withIndent('  ');
    return encoder.convert(json);
  }

  void _updateJsonController() {
    _jsonController.text = _formatJson(currentTemplate.toJson());
    _jsonError = '';
  }

  void _updateTemplate(FormTemplate template) {
    setState(() {
      currentTemplate = template;
      _updateJsonController();
    });
  }

  Future<void> _saveTemplate() async {
    try {
      final String? outputPath = await FilePicker.platform.saveFile(
        dialogTitle: global.language('form_design_save_form'),
        fileName: '${currentTemplate.name}.json',
        type: FileType.custom,
        allowedExtensions: ['json'],
      );

      if (outputPath != null) {
        final jsonString = jsonEncode(currentTemplate.toJson());
        final file = File(outputPath);
        await file.writeAsString(jsonString);

        if (mounted) {
          global.showSuccessSnackBar(context, global.language('form_design_save_success'));
        }
      }
    } catch (e) {
      if (mounted) {
        global.showErrorSnackBar(context, '${global.language('form_design_error')}: $e');
      }
    }
  }

  Future<void> _loadTemplate() async {
    try {
      final result = await FilePicker.platform.pickFiles(
        type: FileType.custom,
        allowedExtensions: ['json'],
      );

      if (result != null && result.files.single.path != null) {
        final file = File(result.files.single.path!);
        final jsonString = await file.readAsString();
        final jsonData = jsonDecode(jsonString);
        final template = FormTemplate.fromJson(jsonData);

        setState(() {
          currentTemplate = template;
        });

        if (mounted) {
          global.showSuccessSnackBar(context, global.language('form_design_load_success'));
        }
      }
    } catch (e) {
      if (mounted) {
        global.showErrorSnackBar(context, '${global.language('form_design_error')}: $e');
      }
    }
  }

  Future<void> _setGlobalBackground() async {
    final XFile? image = await _picker.pickImage(source: ImageSource.gallery);
    if (image != null) {
      _updateTemplate(
        currentTemplate.copyWith(globalBackgroundImagePath: image.path),
      );
    }
  }

  void _showNewTemplateDialog() {
    final nameController = TextEditingController(text: currentTemplate.name);
    final descController = TextEditingController(
      text: currentTemplate.description,
    );
    PaperSize selectedSize = currentTemplate.paperSize;

    showDialog(
      context: context,
      builder: (context) => StatefulBuilder(
        builder: (context, setDialogState) => AlertDialog(
          title: Text(global.language('form_design_form_info')),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              TextField(
                controller: nameController,
                decoration: InputDecoration(
                  labelText: global.language('form_design_form_name'),
                ),
              ),
              SizedBox(height: 16),
              TextField(
                controller: descController,
                decoration: InputDecoration(
                  labelText: global.language('form_design_description'),
                ),
                maxLines: 2,
              ),
              SizedBox(height: 16),
              DropdownButtonFormField<PaperSize>(
                initialValue: selectedSize,
                decoration: InputDecoration(
                  labelText: global.language('form_design_paper_size'),
                ),
                items: PaperSize.values.map((size) {
                  return DropdownMenuItem(value: size, child: Text(size.name));
                }).toList(),
                onChanged: (value) {
                  if (value != null) {
                    setDialogState(() {
                      selectedSize = value;
                    });
                  }
                },
              ),
            ],
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context),
              child: Text(global.language('form_design_cancel')),
            ),
            ElevatedButton(
              onPressed: () {
                _updateTemplate(
                  currentTemplate.copyWith(
                    name: nameController.text,
                    description: descController.text,
                    paperSize: selectedSize,
                  ),
                );
                Navigator.pop(context);
              },
              child: Text(global.language('form_design_save')),
            ),
          ],
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(
          _isPreviewMode
              ? global.language('form_design_preview_form')
              : global.language('form_design_design_form'),
        ),
        actions: [
          const ManualButton(path: 'settings-form-design'),
          if (!_isPreviewMode) ...[
            // Zoom controls
            PopupMenuButton<double>(
              icon: Icon(Icons.zoom_in),
              tooltip: global.language('form_design_zoom'),
              onSelected: (value) {
                setState(() {
                  _zoomLevel = value;
                });
              },
              itemBuilder: (context) => [
                PopupMenuItem(
                  value: 2.0,
                  child: Text('200% ${_zoomLevel == 2.0 ? "✓" : ""}'),
                ),
                PopupMenuItem(
                  value: 1.0,
                  child: Text('100% ${_zoomLevel == 1.0 ? "✓" : ""}'),
                ),
                PopupMenuItem(
                  value: 0.5,
                  child: Text('50% ${_zoomLevel == 0.5 ? "✓" : ""}'),
                ),
                PopupMenuItem(
                  value: 0.25,
                  child: Text('25% ${_zoomLevel == 0.25 ? "✓" : ""}'),
                ),
                const PopupMenuDivider(),
                PopupMenuItem(
                  value: -1,
                  child: const Text('Full Screen'),
                  onTap: () {
                    // Will be handled after menu closes
                    Future.delayed(Duration.zero, () {
                      setState(() {
                        _zoomLevel = -1; // Special value for full screen
                      });
                    });
                  },
                ),
              ],
            ),
            // Ruler toggle
            IconButton(
              icon: Icon(
                _showRulers ? Icons.straighten : Icons.straighten_outlined,
              ),
              onPressed: () {
                setState(() {
                  _showRulers = !_showRulers;
                });
              },
              tooltip: _showRulers
                  ? global.language('form_design_hide_ruler')
                  : global.language('form_design_show_ruler'),
            ),
            // JSON Panel toggle
            IconButton(
              icon: Icon(_showJsonPanel ? Icons.code : Icons.code_off),
              onPressed: () {
                setState(() {
                  _showJsonPanel = !_showJsonPanel;
                });
              },
              tooltip: _showJsonPanel
                  ? global.language('form_design_hide_json')
                  : global.language('form_design_show_json'),
            ),
          ],
          IconButton(
            icon: const Icon(Icons.info_outline),
            onPressed: _showNewTemplateDialog,
            tooltip: global.language('form_design_form_info'),
          ),
          IconButton(
            icon: const Icon(Icons.wallpaper),
            onPressed: _setGlobalBackground,
            tooltip: global.language('form_design_set_global_bg'),
          ),
          IconButton(
            icon: Icon(_isPreviewMode ? Icons.edit : Icons.preview),
            onPressed: () {
              setState(() {
                _isPreviewMode = !_isPreviewMode;
              });
            },
            tooltip: _isPreviewMode
                ? global.language('form_design_edit')
                : global.language('form_design_preview'),
          ),
          IconButton(
            icon: const Icon(Icons.folder_open),
            onPressed: _loadTemplate,
            tooltip: global.language('form_design_load_form'),
          ),
          IconButton(
            icon: const Icon(Icons.save),
            onPressed: _saveTemplate,
            tooltip: global.language('form_design_save_form'),
          ),
        ],
      ),
      body: _isPreviewMode ? _buildPreviewMode() : _buildEditMode(),
    );
  }

  Widget _buildEditMode() {
    return Row(
      children: [
        // Main Editor Area
        Expanded(
          flex: _showJsonPanel ? 2 : 1,
          child: Column(
            children: [
              // Info Bar
              Container(
                padding: const EdgeInsets.all(8),
                color: Colors.blue[50],
                child: Row(
                  children: [
                    Text(
                      '${global.language('form_design_form_name')}: ${currentTemplate.name}',
                      style: const TextStyle(fontWeight: FontWeight.bold),
                    ),
                    SizedBox(width: 16),
                    Text(
                      '${global.language('form_design_paper_size')}: ${currentTemplate.paperSize.name}',
                    ),
                    SizedBox(width: 16),
                    Text(
                      '${global.language('form_design_description')}: ${currentTemplate.description}',
                    ),
                    SizedBox(width: 16),
                    Text(
                      '${global.language('form_design_zoom')}: ${(_zoomLevel == -1 ? "Full" : "${(_zoomLevel * 100).toInt()}%")}',
                    ),
                  ],
                ),
              ),
              // Full Form Editor
              Expanded(
                child: FullFormEditor(
                  template: currentTemplate,
                  onTemplateChanged: (updatedTemplate) {
                    setState(() {
                      currentTemplate = updatedTemplate;
                      _updateJsonController();
                    });
                  },
                  zoomLevel: _zoomLevel,
                  showRulers: _showRulers,
                  onElementSelected: (element) {
                    setState(() {
                      selectedElement = element;
                    });
                  },
                ),
              ),
            ],
          ),
        ),
        // JSON Panel
        if (_showJsonPanel)
          Container(
            width: 400,
            decoration: BoxDecoration(
              color: Colors.grey[100],
              border: Border(
                left: BorderSide(color: Colors.grey[300]!, width: 1),
              ),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // JSON Panel Header
                Container(
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: Colors.grey[200],
                    border: Border(
                      bottom: BorderSide(color: Colors.grey[300]!, width: 1),
                    ),
                  ),
                  child: Row(
                    children: [
                      const Icon(Icons.code, size: 20),
                      const SizedBox(width: 8),
                      const Text(
                        'JSON Editor',
                        style: TextStyle(
                          fontWeight: FontWeight.bold,
                          fontSize: 16,
                        ),
                      ),
                      Spacer(),
                      IconButton(
                        icon: const Icon(Icons.refresh, size: 20),
                        onPressed: _applyJsonChanges,
                        tooltip: global.language('form_design_apply_changes'),
                      ),
                    ],
                  ),
                ),
                // Error message
                if (_jsonError.isNotEmpty)
                  Container(
                    padding: const EdgeInsets.all(8),
                    color: Colors.red[100],
                    child: Row(
                      children: [
                        const Icon(Icons.error, color: Colors.red, size: 16),
                        const SizedBox(width: 8),
                        Expanded(
                          child: Text(
                            _jsonError,
                            style: const TextStyle(
                              color: Colors.red,
                              fontSize: 12,
                            ),
                          ),
                        ),
                      ],
                    ),
                  ),
                // JSON Editor
                Expanded(
                  child: Padding(
                    padding: const EdgeInsets.all(8),
                    child: TextField(
                      controller: _jsonController,
                      maxLines: null,
                      expands: true,
                      style: const TextStyle(
                        fontFamily: 'Courier',
                        fontSize: 12,
                      ),
                      decoration: InputDecoration(
                        border: const OutlineInputBorder(),
                        hintText: global.language('json_code_hint'),
                        contentPadding: EdgeInsets.all(12),
                      ),
                    ),
                  ),
                ),
                // Action Buttons
                Container(
                  padding: const EdgeInsets.all(8),
                  decoration: BoxDecoration(
                    color: Colors.grey[200],
                    border: Border(
                      top: BorderSide(color: Colors.grey[300]!, width: 1),
                    ),
                  ),
                  child: Row(
                    children: [
                      Expanded(
                        child: ElevatedButton.icon(
                          onPressed: _applyJsonChanges,
                          icon: Icon(Icons.check, size: 18),
                          label: Text(
                            global.language('form_design_apply_changes'),
                          ),
                          style: ElevatedButton.styleFrom(
                            backgroundColor: Colors.green,
                            foregroundColor: Colors.white,
                          ),
                        ),
                      ),
                      SizedBox(width: 8),
                      Expanded(
                        child: ElevatedButton.icon(
                          onPressed: () {
                            setState(() {
                              _jsonController.text = _formatJson(
                                currentTemplate.toJson(),
                              );
                              _jsonError = '';
                            });
                          },
                          icon: Icon(Icons.refresh, size: 18),
                          label: Text(global.language('form_design_reset')),
                          style: ElevatedButton.styleFrom(
                            backgroundColor: Colors.orange,
                            foregroundColor: Colors.white,
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
      ],
    );
  }

  void _applyJsonChanges() {
    try {
      final jsonData = jsonDecode(_jsonController.text);
      final template = FormTemplate.fromJson(jsonData);
      setState(() {
        currentTemplate = template;
        _jsonError = '';
      });
      global.showSuccessSnackBar(context, global.language('form_design_apply_changes_success'));
    } catch (e) {
      setState(() {
        _jsonError = 'JSON Error: ${e.toString()}';
      });
      global.showErrorSnackBar(context, '${global.language('form_design_json_invalid')}: $e');
    }
  }

  Widget _buildPreviewMode() {
    return Center(
      child: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: FormPreview(template: currentTemplate, scale: 0.8),
      ),
    );
  }
}
