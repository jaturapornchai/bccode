import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/company_branch_model.dart';

class PaymentRoundingWidget extends StatefulWidget {
  final PaymentRoundingModel? paymentRounding;
  final Map<String, List<List<TextEditingController>>> roundingControllers;
  final Function(String method, int index) onRemoveRule;
  final Function(String method) onAddRule;
  final ValueChanged<bool> onDataChanged;

  const PaymentRoundingWidget({
    super.key,
    required this.paymentRounding,
    required this.roundingControllers,
    required this.onRemoveRule,
    required this.onAddRule,
    required this.onDataChanged,
  });

  @override
  State<PaymentRoundingWidget> createState() => _PaymentRoundingWidgetState();
}

class _PaymentRoundingWidgetState extends State<PaymentRoundingWidget> {
  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: EdgeInsets.all(10.0),
      child: Container(
        margin: const EdgeInsets.only(top: 10, bottom: 10),
        padding: const EdgeInsets.all(10),
        decoration: BoxDecoration(
          border: Border.all(color: Colors.blue),
          borderRadius: BorderRadius.circular(5),
          color: Colors.white,
        ),
        child: Column(
          children: [
            Text(
              global.language("payment_rounding"),
              style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold),
            ),
            SizedBox(height: 20),
            DefaultTabController(
              length: 7,
              child: Column(
                children: [
                  TabBar(
                    isScrollable: true,
                    labelColor: Colors.blue,
                    unselectedLabelColor: Colors.grey,
                    tabs: [
                      Tab(text: global.language("cash")),
                      Tab(text: global.language("credit_card")),
                      Tab(text: global.language("bank_transfer")),
                      Tab(text: global.language("cheque")),
                      Tab(text: global.language("coupon")),
                      Tab(text: global.language("delivery")),
                      Tab(text: global.language("qrcode")),
                    ],
                  ),
                  const SizedBox(height: 10),
                  SizedBox(
                    height: 500,
                    child: TabBarView(
                      children: [
                        _buildPaymentMethodRounding('cash'),
                        _buildPaymentMethodRounding('creditcard'),
                        _buildPaymentMethodRounding('banktransfer'),
                        _buildPaymentMethodRounding('cheque'),
                        _buildPaymentMethodRounding('coupon'),
                        _buildPaymentMethodRounding('delivery'),
                        _buildPaymentMethodRounding('qrcode'),
                      ],
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

  Widget _buildPaymentMethodRounding(String method) {
    bool isEnabled = _getEnabledState(method);

    return Column(
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.start,
          children: [
            Switch(
              value: isEnabled,
              onChanged: (value) {
                setState(() {
                  _setEnabledState(method, value);
                  widget.onDataChanged(true);
                });
              },
            ),
            Text(global.language("enable_rounding")),
          ],
        ),
        const SizedBox(height: 10),
        if (isEnabled) _buildRulesTable(method),
      ],
    );
  }

  Widget _buildRulesTable(String method) {
    return Expanded(
      child: Column(
        children: [
          Container(
            padding: const EdgeInsets.all(8),
            color: Colors.grey[200],
            child: Row(
              children: [
                SizedBox(
                  width: 30,
                  child: Text(
                    global.language("#"),
                    style: const TextStyle(fontWeight: FontWeight.bold),
                  ),
                ),
                Expanded(
                  flex: 3,
                  child: Text(
                    global.language("lower_bound"),
                    style: const TextStyle(fontWeight: FontWeight.bold),
                  ),
                ),
                Expanded(
                  flex: 3,
                  child: Text(
                    global.language("upper_bound"),
                    style: const TextStyle(fontWeight: FontWeight.bold),
                  ),
                ),
                Expanded(
                  flex: 3,
                  child: Text(
                    global.language("round_to"),
                    style: const TextStyle(fontWeight: FontWeight.bold),
                  ),
                ),
                Expanded(flex: 1, child: Container()),
              ],
            ),
          ),
          Expanded(
            child: ListView.builder(
              itemCount: widget.roundingControllers[method]!.length,
              itemBuilder: (context, index) {
                return _buildRuleRow(method, index);
              },
            ),
          ),
          Padding(
            padding: EdgeInsets.only(top: 8.0),
            child: ElevatedButton.icon(
              icon: Icon(Icons.add),
              label: Text(global.language("add_rule")),
              onPressed: () => widget.onAddRule(method),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildRuleRow(String method, int index) {
    return Padding(
      padding: EdgeInsets.symmetric(vertical: 4.0),
      child: Row(
        children: [
          SizedBox(
            width: 30,
            child: Padding(
              padding: const EdgeInsets.all(8.0),
              child: Text(
                "${index + 1}",
                style: const TextStyle(fontWeight: FontWeight.bold),
              ),
            ),
          ),
          SizedBox(width: 8),
          Expanded(
            flex: 3,
            child: TextField(
              controller: widget.roundingControllers[method]![index][0],
              decoration: InputDecoration(
                border: const OutlineInputBorder(),
                contentPadding: EdgeInsets.symmetric(horizontal: 8),
                hintText: global.language("lower_bound"),
              ),
              keyboardType: const TextInputType.numberWithOptions(
                decimal: true,
              ),
              onChanged: (value) {
                widget.onDataChanged(true);
                _updateRuleValue(method, index, 'lowerbound', value);
              },
            ),
          ),
          SizedBox(width: 8),
          Expanded(
            flex: 3,
            child: TextField(
              controller: widget.roundingControllers[method]![index][1],
              decoration: InputDecoration(
                border: const OutlineInputBorder(),
                contentPadding: EdgeInsets.symmetric(horizontal: 8),
                hintText: global.language("upper_bound"),
              ),
              keyboardType: const TextInputType.numberWithOptions(
                decimal: true,
              ),
              onChanged: (value) {
                widget.onDataChanged(true);
                _updateRuleValue(method, index, 'upperbound', value);
              },
            ),
          ),
          SizedBox(width: 8),
          Expanded(
            flex: 3,
            child: TextField(
              controller: widget.roundingControllers[method]![index][2],
              decoration: InputDecoration(
                border: const OutlineInputBorder(),
                contentPadding: EdgeInsets.symmetric(horizontal: 8),
                hintText: global.language("round_to"),
              ),
              keyboardType: const TextInputType.numberWithOptions(
                decimal: true,
              ),
              onChanged: (value) {
                widget.onDataChanged(true);
                _updateRuleValue(method, index, 'roundto', value);
              },
            ),
          ),
          const SizedBox(width: 8),
          Expanded(
            flex: 1,
            child: IconButton(
              icon: const Icon(Icons.delete, color: Colors.red),
              onPressed: () => widget.onRemoveRule(method, index),
            ),
          ),
        ],
      ),
    );
  }

  bool _getEnabledState(String method) {
    switch (method) {
      case 'cash':
        return widget.paymentRounding?.cash.enabled ?? false;
      case 'creditcard':
        return widget.paymentRounding?.creditcard.enabled ?? false;
      case 'banktransfer':
        return widget.paymentRounding?.banktransfer.enabled ?? false;
      case 'cheque':
        return widget.paymentRounding?.cheque.enabled ?? false;
      case 'coupon':
        return widget.paymentRounding?.coupon.enabled ?? false;
      case 'delivery':
        return widget.paymentRounding?.delivery.enabled ?? false;
      case 'qrcode':
        return widget.paymentRounding?.qrcode.enabled ?? false;
      default:
        return false;
    }
  }

  void _setEnabledState(String method, bool value) {
    switch (method) {
      case 'cash':
        widget.paymentRounding?.cash.enabled = value;
        break;
      case 'creditcard':
        widget.paymentRounding?.creditcard.enabled = value;
        break;
      case 'banktransfer':
        widget.paymentRounding?.banktransfer.enabled = value;
        break;
      case 'cheque':
        widget.paymentRounding?.cheque.enabled = value;
        break;
      case 'coupon':
        widget.paymentRounding?.coupon.enabled = value;
        break;
      case 'delivery':
        widget.paymentRounding?.delivery.enabled = value;
        break;
      case 'qrcode':
        widget.paymentRounding?.qrcode.enabled = value;
        break;
    }
  }

  void _updateRuleValue(String method, int index, String field, String value) {
    double val = double.tryParse(value) ?? 0;
    List<RoundingRuleModel> rules = _getRules(method);

    if (index < rules.length) {
      switch (field) {
        case 'lowerbound':
          rules[index].lowerbound = val;
          break;
        case 'upperbound':
          rules[index].upperbound = val;
          break;
        case 'roundto':
          rules[index].roundto = val;
          break;
      }
    }
  }

  List<RoundingRuleModel> _getRules(String method) {
    switch (method) {
      case 'cash':
        return widget.paymentRounding!.cash.rules;
      case 'creditcard':
        return widget.paymentRounding!.creditcard.rules;
      case 'banktransfer':
        return widget.paymentRounding!.banktransfer.rules;
      case 'cheque':
        return widget.paymentRounding!.cheque.rules;
      case 'coupon':
        return widget.paymentRounding!.coupon.rules;
      case 'delivery':
        return widget.paymentRounding!.delivery.rules;
      case 'qrcode':
        return widget.paymentRounding!.qrcode.rules;
      default:
        return [];
    }
  }
}
