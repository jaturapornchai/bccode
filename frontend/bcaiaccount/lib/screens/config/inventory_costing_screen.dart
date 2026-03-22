import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/bloc/inventory_costing/inventory_costing_bloc.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/inventory_costing_model.dart';
import 'package:smlaicloud/repositories/inventory_costing_repository.dart';

class InventoryCostingScreen extends StatefulWidget {
  const InventoryCostingScreen({super.key});

  @override
  State<InventoryCostingScreen> createState() => _InventoryCostingScreenState();
}

class _InventoryCostingScreenState extends State<InventoryCostingScreen>
    with global.ThemeRefreshMixin {
  TextEditingController itemCodeController = TextEditingController();
  TextEditingController standardCostController = TextEditingController();
  TextEditingController expiryAlertDaysController = TextEditingController();
  FocusNode searchFocusNode = FocusNode(skipTraversal: true);

  String selectedMethod = CostingMethod.movingAverage;
  bool lotTrackingEnabled = false;
  bool expiryTrackingEnabled = false;
  bool autoBlockExpired = false;
  bool allowNegativeStock = false;
  bool isLoaded = false;
  String currentItemCode = '';

  @override
  void initState() {
    super.initState();
    expiryAlertDaysController.text = '30';
    standardCostController.text = '0';
  }

  @override
  void dispose() {
    itemCodeController.dispose();
    standardCostController.dispose();
    expiryAlertDaysController.dispose();
    searchFocusNode.dispose();
    super.dispose();
  }

  void _loadConfig(BuildContext context) {
    final code = itemCodeController.text.trim();
    if (code.isEmpty) return;
    currentItemCode = code;
    context.read<InventoryCostingBloc>().add(
          InventoryCostingLoadConfig(itemCode: code),
        );
  }

  void _saveConfig(BuildContext context) {
    if (currentItemCode.isEmpty) return;
    final config = ProductCostingConfigModel(
      itemCode: currentItemCode,
      costingMethod: selectedMethod,
      lotTrackingEnabled: lotTrackingEnabled,
      expiryTrackingEnabled: expiryTrackingEnabled,
      expiryAlertDays: int.tryParse(expiryAlertDaysController.text) ?? 30,
      autoBlockExpired: autoBlockExpired,
      allowNegativeStock: allowNegativeStock,
      standardCost: double.tryParse(standardCostController.text) ?? 0,
    );
    context.read<InventoryCostingBloc>().add(
          InventoryCostingSaveConfig(config: config),
        );
  }

  void _populateForm(ProductCostingConfigModel config) {
    setState(() {
      selectedMethod = config.costingMethod;
      lotTrackingEnabled = config.lotTrackingEnabled;
      expiryTrackingEnabled = config.expiryTrackingEnabled;
      expiryAlertDaysController.text = config.expiryAlertDays.toString();
      autoBlockExpired = config.autoBlockExpired;
      allowNegativeStock = config.allowNegativeStock;
      standardCostController.text = config.standardCost.toStringAsFixed(4);
      isLoaded = true;
    });
  }

  @override
  Widget build(BuildContext context) {
    return BlocProvider(
      create: (context) => InventoryCostingBloc(
        repository: InventoryCostingRepository(),
      ),
      child: BlocListener<InventoryCostingBloc, InventoryCostingState>(
        listener: (context, state) {
          if (state is InventoryCostingConfigLoaded) {
            _populateForm(state.config);
          } else if (state is InventoryCostingConfigSaveSuccess) {
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(
                content: Text(global.language('save_success')),
                backgroundColor: global.theme.buttonYesColor,
              ),
            );
          } else if (state is InventoryCostingFailed) {
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(
                content: Text(state.message),
                backgroundColor: global.theme.buttonNoColor,
              ),
            );
          }
        },
        child: Scaffold(
          backgroundColor: global.theme.backgroundColor,
          appBar: AppBar(
            title: Text(
              global.language('inventory_costing_config'),
              style: const TextStyle(color: Colors.white),
            ),
            backgroundColor: global.theme.primaryColor,
            iconTheme: const IconThemeData(color: Colors.white),
          ),
          body: _buildBody(),
        ),
      ),
    );
  }

  Widget _buildBody() {
    return Builder(
      builder: (context) {
        return SingleChildScrollView(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              _buildSearchSection(context),
              if (isLoaded) ...[
                const SizedBox(height: 16),
                _buildCostingMethodSection(),
                const SizedBox(height: 16),
                _buildTrackingSection(),
                if (selectedMethod == CostingMethod.standard) ...[
                  const SizedBox(height: 16),
                  _buildStandardCostSection(),
                ],
                const SizedBox(height: 16),
                _buildOptionsSection(),
                const SizedBox(height: 24),
                _buildSaveButton(context),
              ],
            ],
          ),
        );
      },
    );
  }

  Widget _buildSearchSection(BuildContext context) {
    return Card(
      color: global.theme.cardColor,
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Row(
          children: [
            Expanded(
              child: TextField(
                controller: itemCodeController,
                focusNode: searchFocusNode,
                decoration: InputDecoration(
                  labelText: global.language('item_code'),
                  labelStyle: TextStyle(color: global.theme.textSecondaryColor),
                  border: const OutlineInputBorder(),
                  enabledBorder: OutlineInputBorder(
                    borderSide: BorderSide(color: global.theme.dividerBorderColor),
                  ),
                ),
                style: TextStyle(color: global.theme.textColor),
                onSubmitted: (_) => _loadConfig(context),
              ),
            ),
            const SizedBox(width: 12),
            ElevatedButton.icon(
              onPressed: () => _loadConfig(context),
              icon: const Icon(Icons.search),
              label: Text(global.language('search')),
              style: ElevatedButton.styleFrom(
                backgroundColor: global.theme.primaryColor,
                foregroundColor: Colors.white,
                padding:
                    const EdgeInsets.symmetric(horizontal: 20, vertical: 16),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildCostingMethodSection() {
    return Card(
      color: global.theme.cardColor,
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              global.language('costing_method'),
              style: TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.bold,
                color: global.theme.textColor,
              ),
            ),
            const SizedBox(height: 12),
            ...CostingMethod.all.map((method) => RadioListTile<String>(
                  title: Text(
                    CostingMethod.displayName(method),
                    style: TextStyle(color: global.theme.textColor),
                  ),
                  subtitle: Text(
                    _methodDescription(method),
                    style: TextStyle(
                      color: global.theme.textSecondaryColor,
                      fontSize: 12,
                    ),
                  ),
                  value: method,
                  groupValue: selectedMethod,
                  activeColor: global.theme.primaryColor,
                  onChanged: (value) {
                    if (value != null) {
                      setState(() {
                        selectedMethod = value;
                        if (value == CostingMethod.fefo) {
                          expiryTrackingEnabled = true;
                        }
                        if (value == CostingMethod.lot) {
                          lotTrackingEnabled = true;
                        }
                      });
                    }
                  },
                )),
          ],
        ),
      ),
    );
  }

  Widget _buildTrackingSection() {
    return Card(
      color: global.theme.cardColor,
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              global.language('tracking_settings'),
              style: TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.bold,
                color: global.theme.textColor,
              ),
            ),
            SwitchListTile(
              title: Text(
                global.language('lot_tracking'),
                style: TextStyle(color: global.theme.textColor),
              ),
              value: lotTrackingEnabled,
              activeThumbColor: global.theme.primaryColor,
              onChanged: selectedMethod == CostingMethod.lot
                  ? null
                  : (value) => setState(() => lotTrackingEnabled = value),
            ),
            SwitchListTile(
              title: Text(
                global.language('expiry_tracking'),
                style: TextStyle(color: global.theme.textColor),
              ),
              value: expiryTrackingEnabled,
              activeThumbColor: global.theme.primaryColor,
              onChanged: selectedMethod == CostingMethod.fefo
                  ? null
                  : (value) => setState(() => expiryTrackingEnabled = value),
            ),
            if (expiryTrackingEnabled) ...[
              Padding(
                padding:
                    const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                child: TextField(
                  controller: expiryAlertDaysController,
                  keyboardType: TextInputType.number,
                  decoration: InputDecoration(
                    labelText: global.language('expiry_alert_days'),
                    labelStyle:
                        TextStyle(color: global.theme.textSecondaryColor),
                    border: const OutlineInputBorder(),
                    enabledBorder: OutlineInputBorder(
                      borderSide:
                          BorderSide(color: global.theme.dividerBorderColor),
                    ),
                    suffixText: global.language('days'),
                  ),
                  style: TextStyle(color: global.theme.textColor),
                ),
              ),
              SwitchListTile(
                title: Text(
                  global.language('auto_block_expired'),
                  style: TextStyle(color: global.theme.textColor),
                ),
                value: autoBlockExpired,
                activeThumbColor: global.theme.primaryColor,
                onChanged: (value) =>
                    setState(() => autoBlockExpired = value),
              ),
            ],
          ],
        ),
      ),
    );
  }

  Widget _buildStandardCostSection() {
    return Card(
      color: global.theme.cardColor,
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              global.language('standard_cost'),
              style: TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.bold,
                color: global.theme.textColor,
              ),
            ),
            const SizedBox(height: 12),
            TextField(
              controller: standardCostController,
              keyboardType:
                  const TextInputType.numberWithOptions(decimal: true),
              decoration: InputDecoration(
                labelText: global.language('standard_cost_per_unit'),
                labelStyle: TextStyle(color: global.theme.textSecondaryColor),
                border: const OutlineInputBorder(),
                enabledBorder: OutlineInputBorder(
                  borderSide: BorderSide(color: global.theme.dividerBorderColor),
                ),
                prefixIcon: Icon(Icons.monetization_on,
                    color: global.theme.textSecondaryColor),
              ),
              style: TextStyle(color: global.theme.textColor),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildOptionsSection() {
    return Card(
      color: global.theme.cardColor,
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              global.language('options'),
              style: TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.bold,
                color: global.theme.textColor,
              ),
            ),
            SwitchListTile(
              title: Text(
                global.language('allow_negative_stock'),
                style: TextStyle(color: global.theme.textColor),
              ),
              subtitle: Text(
                global.language('allow_negative_stock_desc'),
                style: TextStyle(
                  color: global.theme.textSecondaryColor,
                  fontSize: 12,
                ),
              ),
              value: allowNegativeStock,
              activeThumbColor: global.theme.primaryColor,
              onChanged: (value) =>
                  setState(() => allowNegativeStock = value),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildSaveButton(BuildContext context) {
    return BlocBuilder<InventoryCostingBloc, InventoryCostingState>(
      builder: (context, state) {
        final isLoading = state is InventoryCostingConfigSaveInProgress;
        return SizedBox(
          width: double.infinity,
          child: ElevatedButton.icon(
            onPressed: isLoading ? null : () => _saveConfig(context),
            icon: isLoading
                ? const SizedBox(
                    width: 20,
                    height: 20,
                    child: CircularProgressIndicator(
                      strokeWidth: 2,
                      color: Colors.white,
                    ),
                  )
                : const Icon(Icons.save),
            label: Text(
              global.language('save'),
              style: const TextStyle(fontSize: 16),
            ),
            style: ElevatedButton.styleFrom(
              backgroundColor: global.theme.primaryColor,
              foregroundColor: Colors.white,
              padding: const EdgeInsets.symmetric(vertical: 16),
            ),
          ),
        );
      },
    );
  }

  String _methodDescription(String method) {
    switch (method) {
      case CostingMethod.movingAverage:
        return 'Weighted average recalculated on each receipt';
      case CostingMethod.periodicAverage:
        return 'Average calculated at end of accounting period';
      case CostingMethod.fifo:
        return 'First In, First Out — oldest stock issued first';
      case CostingMethod.lifo:
        return 'Last In, First Out — newest stock issued first';
      case CostingMethod.fefo:
        return 'First Expiry, First Out — earliest expiry issued first';
      case CostingMethod.lot:
        return 'Track cost per lot/batch number';
      case CostingMethod.standard:
        return 'Predetermined cost with variance tracking';
      default:
        return '';
    }
  }
}
