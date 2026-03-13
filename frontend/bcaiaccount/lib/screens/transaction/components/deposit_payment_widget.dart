import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/screen_search/transaction_search_screen.dart';
import 'dart:async';

class DepositPaymentWidget extends StatefulWidget {
  final List<DepositDocModel> depositDocs;
  final Function(List<DepositDocModel>) onDepositDocsChanged;
  final Function(double) onSumDepositChanged;
  final String custcode; // เพิ่มสำหรับ TransSearchScreen
  final global.TransactionTypeEnum type; // เพิ่มสำหรับกำหนดประเภทธุรกรรม

  const DepositPaymentWidget({
    super.key,
    required this.depositDocs,
    required this.onDepositDocsChanged,
    required this.onSumDepositChanged,
    required this.custcode,
    required this.type,
  });

  @override
  State<DepositPaymentWidget> createState() => _DepositPaymentWidgetState();
}

class _DepositPaymentWidgetState extends State<DepositPaymentWidget>
    with global.ThemeRefreshMixin {
  late List<DepositDocModel> _depositDocs;
  double _sumDeposit = 0.0;
  Timer? _debounceTimer;
  final Map<int, bool> _errorStates = {}; // เก็บสถานะ error ของแต่ละ field

  @override
  void initState() {
    super.initState();
    _depositDocs = List.from(widget.depositDocs);
    _calculateSumDeposit();
  }

  @override
  void dispose() {
    _debounceTimer?.cancel();
    super.dispose();
  }

  @override
  void didUpdateWidget(DepositPaymentWidget oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (widget.depositDocs != oldWidget.depositDocs) {
      _depositDocs = List.from(widget.depositDocs);
      _calculateSumDeposit();
    }
  }

  void _calculateSumDeposit() {
    final newSum = _depositDocs.fold(0.0, (sum, doc) => sum + doc.useamount);
    if (_sumDeposit != newSum) {
      _sumDeposit = newSum;
      WidgetsBinding.instance.addPostFrameCallback((_) {
        widget.onSumDepositChanged(_sumDeposit);
      });
    }
  }

  // Helper method to get the transaction type for search screen
  global.TransactionTypeEnum _getSearchTransactionType() {
    if (widget.type == global.TransactionTypeEnum.purchase ||
        widget.type == global.TransactionTypeEnum.accrualreceive) {
      return global.TransactionTypeEnum.deposit;
    } else if (widget.type == global.TransactionTypeEnum.sale) {
      return global.TransactionTypeEnum.receiveDeposit;
    }
    return global.TransactionTypeEnum.deposit; // default
  }

  void _addDepositDoc() {
    if (widget.custcode.isEmpty) {
      String message =
          widget.type == global.TransactionTypeEnum.purchase ||
              widget.type == global.TransactionTypeEnum.accrualreceive
          ? global.language("please_input_custcode_creditor")
          : global.language("please_input_custcode_debtor");

      showDialog(
        context: context,
        builder: (BuildContext context) {
          return AlertDialog(
            title: Text(global.language("alert")),
            content: Text(message),
            actions: [
              TextButton(
                child: Text(global.language("close")),
                onPressed: () {
                  Navigator.pop(context);
                },
              ),
            ],
          );
        },
      );
      return;
    }

    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => TransSearchScreen(
          custcode: widget.custcode,
          type: _getSearchTransactionType(),
        ),
      ),
    ).then((value) {
      if (value != null) {
        TransactionModel result = value;
        if (result.docno.isNotEmpty) {
          // ตรวจสอบว่าเอกสารนี้ถูกเพิ่มแล้วหรือไม่
          bool alreadyExists = _depositDocs.any(
            (doc) => doc.docno == result.docno,
          );

          if (!alreadyExists) {
            // สร้าง DepositDocModel จากข้อมูลที่ได้รับ
            DepositDocModel newDepositDoc = DepositDocModel(
              balance: result.totalamount, // ยอดคงเหลือเริ่มต้นเท่ากับยอดรวม
              docdatetime: result.docdatetime,
              docno: result.docno,
              guidfixed: result.guidfixed ?? '',
              remark: '',
              totalamount: result.totalamount,
              useamount: 0, // ยอดตัดเริ่มต้น = 0
            );

            setState(() {
              _depositDocs.add(newDepositDoc);
              _calculateSumDeposit();
            });

            WidgetsBinding.instance.addPostFrameCallback((_) {
              widget.onDepositDocsChanged(_depositDocs);
            });
          } else {
            // แสดงข้อความแจ้งเตือนว่าเอกสารถูกเพิ่มแล้ว
            global.showWarningSnackBar(context, '${global.language("document_already_added")} ${result.docno}');
          }
        }
      }
    });
  }

  void _selectNewDocument(int index) {
    if (widget.custcode.isEmpty) {
      String message = global.language("please_input_custcode_creditor");

      showDialog(
        context: context,
        builder: (BuildContext context) {
          return AlertDialog(
            title: Text(global.language("alert")),
            content: Text(message),
            actions: [
              TextButton(
                child: Text(global.language("close")),
                onPressed: () {
                  Navigator.pop(context);
                },
              ),
            ],
          );
        },
      );
      return;
    }

    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => TransSearchScreen(
          custcode: widget.custcode,
          type: _getSearchTransactionType(),
        ),
      ),
    ).then((value) {
      if (value != null) {
        TransactionModel result = value;
        if (result.docno.isNotEmpty) {
          // ตรวจสอบว่าเอกสารนี้ถูกเพิ่มแล้วหรือไม่ (ยกเว้นตำแหน่งปัจจุบัน)
          bool alreadyExists = _depositDocs.asMap().entries.any(
            (entry) => entry.key != index && entry.value.docno == result.docno,
          );

          if (!alreadyExists) {
            // อัพเดทข้อมูลในตำแหน่งเดิม
            DepositDocModel updatedDepositDoc = DepositDocModel(
              balance: result.totalamount, // รีเซ็ตยอดคงเหลือ
              docdatetime: result.docdatetime,
              docno: result.docno,
              guidfixed: result.guidfixed ?? '',
              remark: _depositDocs[index].remark, // เก็บหมายเหตุเดิม
              totalamount: result.totalamount,
              useamount: 0, // รีเซ็ตยอดตัดเป็น 0
            );

            setState(() {
              _depositDocs[index] = updatedDepositDoc;
              _calculateSumDeposit();
            });

            WidgetsBinding.instance.addPostFrameCallback((_) {
              widget.onDepositDocsChanged(_depositDocs);
            });
          } else {
            // แสดงข้อความแจ้งเตือนว่าเอกสารถูกเพิ่มแล้ว
            global.showWarningSnackBar(context, '${global.language("document_already_used")} ${result.docno}');
          }
        }
      }
    });
  }

  void _removeDepositDoc(int index) {
    setState(() {
      _depositDocs.removeAt(index);
      _calculateSumDeposit();
    });
    WidgetsBinding.instance.addPostFrameCallback((_) {
      widget.onDepositDocsChanged(_depositDocs);
    });
  }

  void _validateAndUpdateUseAmount(
    String value,
    int index,
    DepositDocModel depositDoc,
  ) {
    final useAmount = _parseAmount(value);

    // ตรวจสอบไม่ให้ยอดตัดเกินยอดเงินมัดจำ
    if (useAmount > depositDoc.totalamount) {
      // เซ็ต error state
      setState(() {
        _errorStates[index] = true;
      });

      showDialog(
        context: context,
        builder: (BuildContext context) {
          return AlertDialog(
            title: Text(global.language("alert")),
            content: Text(
              '${global.language("deposit_deduction_exceed")} ${_formatAmount(depositDoc.totalamount)} ${global.language("baht")}',
            ),
            actions: [
              TextButton(
                child: Text(global.language("close")),
                onPressed: () {
                  Navigator.pop(context);
                  // ไม่รีเซ็ต error state - ให้ error state ยังคงอยู่จนกว่าผู้ใช้จะแก้ไขค่าให้ถูกต้อง
                },
              ),
            ],
          );
        },
      );
      return;
    }

    // ล้าง error state หากไม่มีข้อผิดพลาด
    setState(() {
      _errorStates[index] = false;
    });

    final balance = depositDoc.totalamount - useAmount;
    _updateDepositDoc(
      index,
      depositDoc.copyWith(useamount: useAmount, balance: balance),
    );
  }

  void _updateDepositDoc(int index, DepositDocModel updatedDoc) {
    setState(() {
      _depositDocs[index] = updatedDoc;
      _calculateSumDeposit();
    });
    WidgetsBinding.instance.addPostFrameCallback((_) {
      widget.onDepositDocsChanged(_depositDocs);
    });
  }

  String _formatAmount(double amount) {
    if (amount == 0) return '0.00';
    return global.formatNumber(amount);
  }

  double _parseAmount(String text) {
    if (text.isEmpty) return 0;
    text = text.replaceAll(',', '');
    return double.tryParse(text) ?? 0;
  }

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: EdgeInsets.all(0.0),
      child: Padding(
        padding: EdgeInsets.all(16.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Header
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  global.language('deduct_deposit'),
                  style: Theme.of(context).textTheme.titleMedium?.copyWith(
                    fontWeight: FontWeight.bold,
                    color: Colors.blue[700],
                  ),
                ),
                Row(
                  children: [
                    Text(
                      '${global.language("total")}: ${_formatAmount(_sumDeposit)} ${global.language("baht")}',
                      style: Theme.of(context).textTheme.titleSmall?.copyWith(
                        fontWeight: FontWeight.bold,
                        color: Colors.green[700],
                      ),
                    ),
                    SizedBox(width: 10),
                    ElevatedButton.icon(
                      onPressed: _addDepositDoc,
                      icon: Icon(Icons.add, size: 18),
                      label: Text(global.language('add')),
                      style: ElevatedButton.styleFrom(
                        backgroundColor: global.theme.infoHighlightTextColor,
                        foregroundColor: global.theme.onPrimaryColor,
                        padding: const EdgeInsets.symmetric(
                          horizontal: 12,
                          vertical: 8,
                        ),
                      ),
                    ),
                  ],
                ),
              ],
            ),

            // ถ้าไม่มีข้อมูลใน depositDocs ให้ซ่อนตาราง
            if (_depositDocs.isNotEmpty)
              Column(
                children: [const SizedBox(height: 8), _buildResponsiveLayout()],
              ),
          ],
        ),
      ),
    );
  }

  Widget _buildResponsiveLayout() {
    return LayoutBuilder(
      builder: (context, constraints) {
        // ถ้าหน้าจอกว้างมากกว่า 800px ใช้ตาราง, ถ้าไม่ใช้ card layout
        if (constraints.maxWidth > 800) {
          return _buildDataTable();
        } else {
          return _buildMobileCardList();
        }
      },
    );
  }

  Widget _buildMobileCardList() {
    return Column(
      children: [
        // Simple Summary bar
        Container(
          width: double.infinity,
          margin: const EdgeInsets.only(bottom: 12),
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
          decoration: BoxDecoration(
            color: global.theme.infoHighlightColor,
            borderRadius: BorderRadius.circular(8),
            border: Border.all(color: Colors.blue[100]!),
          ),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(
                global.language("total_deduction"),
                style: TextStyle(fontSize: 14, color: global.theme.textSecondaryColor),
              ),
              Text(
                '${_formatAmount(_sumDeposit)} ${global.language("baht")}',
                style: TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.bold,
                  color: Colors.blue[700],
                ),
              ),
            ],
          ),
        ),

        // Simple deposit cards
        ...List.generate(_depositDocs.length, (index) {
          return _buildMobileDepositCard(index, _depositDocs[index]);
        }),
      ],
    );
  }

  Widget _buildMobileDepositCard(int index, DepositDocModel depositDoc) {
    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: global.theme.cardColor,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: global.theme.dividerBorderColor),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Header
          Row(
            children: [
              Text(
                '${global.language('deposit')} #${index + 1}',
                style: TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.w600,
                ),
              ),
              const Spacer(),
              GestureDetector(
                onTap: () => _removeDepositDoc(index),
                child: Icon(Icons.close, size: 20, color: global.theme.iconSecondaryColor),
              ),
            ],
          ),

          if (depositDoc.docno.isNotEmpty) ...[
            const SizedBox(height: 8),
            GestureDetector(
              onTap: () => _selectNewDocument(index),
              child: Container(
                padding: const EdgeInsets.symmetric(vertical: 4, horizontal: 8),
                decoration: BoxDecoration(
                  color: global.theme.infoHighlightColor,
                  borderRadius: BorderRadius.circular(4),
                  border: Border.all(color: Colors.blue[200]!),
                ),
                child: Row(
                  children: [
                    Expanded(
                      child: Text(
                        '${global.language("docno")}: ${depositDoc.docno}',
                        style: TextStyle(
                          fontSize: 14,
                          color: Colors.blue[700],
                          fontWeight: FontWeight.w500,
                        ),
                      ),
                    ),
                    Icon(Icons.edit, size: 16, color: global.theme.infoHighlightTextColor),
                  ],
                ),
              ),
            ),
          ] else ...[
            SizedBox(height: 8),
            GestureDetector(
              onTap: () => _selectNewDocument(index),
              child: Container(
                padding: const EdgeInsets.symmetric(vertical: 8, horizontal: 8),
                decoration: BoxDecoration(
                  color: global.theme.surfaceColor,
                  borderRadius: BorderRadius.circular(4),
                  border: Border.all(color: global.theme.dividerBorderColor),
                ),
                child: Row(
                  children: [
                    Expanded(
                      child: Text(
                        global.language('tap_to_select_document'),
                        style: TextStyle(fontSize: 14, color: global.theme.textSecondaryColor),
                      ),
                    ),
                    Icon(Icons.search, size: 16, color: global.theme.iconSecondaryColor),
                  ],
                ),
              ),
            ),
          ],

          if (depositDoc.docdatetime.isNotEmpty) ...[
            const SizedBox(height: 4),
            Text(
              '${global.language("date")}:  ${global.dateTimeBuddhist(DateTime.parse(depositDoc.docdatetime), format: global.DateTimeFormatEnum.dateDay)}',
              style: TextStyle(fontSize: 14, color: global.theme.textSecondaryColor),
            ),
          ],

          const SizedBox(height: 16),

          // Amount info
          Row(
            children: [
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      global.language("deposit_amount"),
                      style: TextStyle(fontSize: 12, color: global.theme.textSecondaryColor),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      _formatAmount(depositDoc.totalamount),
                      style: TextStyle(
                        fontSize: 14,
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                  ],
                ),
              ),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.end,
                  children: [
                    Text(
                      global.language('stock_balance'),
                      style: TextStyle(fontSize: 12, color: global.theme.textSecondaryColor),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      _formatAmount(depositDoc.balance),
                      style: TextStyle(
                        fontSize: 14,
                        fontWeight: FontWeight.w600,
                        color: depositDoc.balance >= 0
                            ? Colors.green[700]
                            : Colors.red[700],
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),

          const SizedBox(height: 16),

          // Editable use amount
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                global.language("deposit_deduction_label"),
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.w500,
                  color: global.theme.textColor,
                ),
              ),
              const SizedBox(height: 8),
              TextFormField(
                initialValue: _formatAmount(depositDoc.useamount),
                keyboardType: const TextInputType.numberWithOptions(
                  decimal: true,
                ),
                inputFormatters: [global.NumberInputFormatter()],
                style: TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.w500,
                ),
                decoration: InputDecoration(
                  hintText: '0.00',
                  suffixText: global.language("baht"),
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(8),
                    borderSide: BorderSide(
                      color: _errorStates[index] == true
                          ? global.theme.negativeHighlightTextColor
                          : global.theme.dividerBorderColor,
                      width: _errorStates[index] == true ? 2 : 1,
                    ),
                  ),
                  enabledBorder: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(8),
                    borderSide: BorderSide(
                      color: _errorStates[index] == true
                          ? global.theme.negativeHighlightTextColor
                          : global.theme.dividerBorderColor,
                      width: _errorStates[index] == true ? 2 : 1,
                    ),
                  ),
                  focusedBorder: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(8),
                    borderSide: BorderSide(
                      color: _errorStates[index] == true
                          ? global.theme.negativeHighlightTextColor
                          : global.theme.infoHighlightTextColor,
                      width: 2,
                    ),
                  ),
                  contentPadding: const EdgeInsets.symmetric(
                    horizontal: 12,
                    vertical: 12,
                  ),
                  filled: true,
                  fillColor: _errorStates[index] == true
                      ? Colors.red[50]
                      : global.theme.surfaceColor,
                ),
                onChanged: (value) {
                  // ยกเลิก timer เก่า (ถ้ามี)
                  _debounceTimer?.cancel();

                  // ตั้ง timer ใหม่ให้รอ 1 วินาทีหลังจากผู้ใช้หยุดพิมพ์
                  _debounceTimer = Timer(
                    const Duration(milliseconds: 1000),
                    () {
                      _validateAndUpdateUseAmount(value, index, depositDoc);
                    },
                  );
                },
              ),
            ],
          ),

          const SizedBox(height: 16),

          // Remark
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                global.language('remark'),
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.w500,
                  color: global.theme.textColor,
                ),
              ),
              SizedBox(height: 8),
              TextFormField(
                initialValue: depositDoc.remark,
                maxLines: 2,
                style: TextStyle(fontSize: 14),
                decoration: InputDecoration(
                  hintText: global.language('add_remark'),
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(8),
                    borderSide: BorderSide(color: global.theme.dividerBorderColor),
                  ),
                  focusedBorder: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(8),
                    borderSide: BorderSide(color: global.theme.infoHighlightTextColor, width: 2),
                  ),
                  contentPadding: const EdgeInsets.symmetric(
                    horizontal: 12,
                    vertical: 12,
                  ),
                  filled: true,
                  fillColor: global.theme.formFillColor,
                ),
                onChanged: (value) {
                  _updateDepositDoc(index, depositDoc.copyWith(remark: value));
                },
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildDataTable() {
    const double rowHeight = 68.0; // เพิ่มความสูงแถวจาก 60 เป็น 68

    return Container(
      decoration: BoxDecoration(
        border: Border.all(color: global.theme.dividerBorderColor),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Column(
        children: [
          // Header Row
          Container(
            height: 30,
            padding: const EdgeInsets.symmetric(horizontal: 8),
            decoration: BoxDecoration(
              color: global.theme.infoHighlightColor,
              borderRadius: const BorderRadius.only(
                topLeft: Radius.circular(8),
                topRight: Radius.circular(8),
              ),
            ),
            child: Row(
              children: [
                _buildHeaderCell(global.language('deposit_docno'), flex: 2),
                _buildHeaderCell(global.language('date'), flex: 2),
                _buildHeaderCell(global.language('deposit_amount'), flex: 2),
                _buildHeaderCell(global.language('stock_balance'), flex: 2),
                _buildHeaderCell(global.language('deposit_deduction_label'), flex: 2),
                _buildHeaderCell(global.language('remark'), flex: 3),
                _buildHeaderCell(global.language('manage'), flex: 0, width: 60),
              ],
            ),
          ),

          // Data Rows
          ...List.generate(_depositDocs.length, (index) {
            return _buildDataRow(index, _depositDocs[index], rowHeight);
          }),
        ],
      ),
    );
  }

  Widget _buildHeaderCell(String text, {int flex = 1, double? width}) {
    Widget cell = Container(
      height: double.infinity,
      alignment: Alignment.center,
      child: Text(
        text,
        style: TextStyle(fontWeight: FontWeight.bold, fontSize: 12),
        textAlign: TextAlign.center,
      ),
    );

    if (width != null) {
      return SizedBox(width: width, child: cell);
    }
    return Expanded(flex: flex, child: cell);
  }

  Widget _buildDataRow(int index, DepositDocModel depositDoc, double height) {
    final isEven = index % 2 == 0;

    return Container(
      height: height,
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 8),
      decoration: BoxDecoration(
        color: isEven ? global.theme.cardColor : global.theme.surfaceColor,
        border: Border(bottom: BorderSide(color: global.theme.dividerBorderColor, width: 1)),
      ),
      child: Row(
        children: [
          // เลขที่${_getDisplayText()} (คลิกได้)
          _buildDataCell(
            child: MouseRegion(
              cursor: SystemMouseCursors.click,
              child: _buildTextField(
                value: depositDoc.docno.isEmpty
                    ? global.language('tap_to_select_document')
                    : depositDoc.docno,
                isReadOnly: true,
                onChanged: null,
                onTap: () => _selectNewDocument(index),
              ),
            ),
            flex: 2,
          ),

          // วันที่
          _buildDataCell(
            child: _buildTextField(
              value: depositDoc.docdatetime.isNotEmpty
                  ? global.dateTimeBuddhist(
                      DateTime.parse(depositDoc.docdatetime),
                      format: global.DateTimeFormatEnum.dateDay,
                    )
                  : '',
              isReadOnly: true,
              onChanged: null,
            ),
            flex: 2,
          ),

          // ยอด${_getDisplayText()}
          _buildDataCell(
            child: _buildTextField(
              value: _formatAmount(depositDoc.totalamount),
              isReadOnly: true,
              onChanged: null,
              textAlign: TextAlign.right,
            ),
            flex: 2,
          ),

          // ยอดคงเหลือ
          _buildDataCell(
            child: _buildTextField(
              value: _formatAmount(depositDoc.balance),
              isReadOnly: true,
              textAlign: TextAlign.right,
              fontWeight: FontWeight.w600,
              onChanged: null,
            ),
            flex: 2,
          ),

          // ยอดตัด${_getDisplayText()} (แก้ไขได้)
          _buildDataCell(
            child: _buildTextField(
              value: _formatAmount(depositDoc.useamount),
              isReadOnly: false,
              textAlign: TextAlign.right,
              keyboardType: const TextInputType.numberWithOptions(
                decimal: true,
              ),
              inputFormatters: [global.NumberInputFormatter()],
              hasError: _errorStates[index] == true,
              onChanged: (value) {
                // ยกเลิก timer เก่า (ถ้ามี)
                _debounceTimer?.cancel();

                // ตั้ง timer ใหม่ให้รอ 1 วินาทีหลังจากผู้ใช้หยุดพิมพ์
                _debounceTimer = Timer(const Duration(milliseconds: 1000), () {
                  _validateAndUpdateUseAmount(value, index, depositDoc);
                });
              },
            ),
            flex: 2,
          ),

          // หมายเหตุ (แก้ไขได้)
          _buildDataCell(
            child: _buildTextField(
              value: depositDoc.remark,
              isReadOnly: false,
              onChanged: (value) {
                _updateDepositDoc(index, depositDoc.copyWith(remark: value));
              },
            ),
            flex: 3,
          ),

          // ปุ่มลบ
          SizedBox(
            width: 60,
            child: Center(
              child: IconButton(
                padding: EdgeInsets.zero,
                constraints: const BoxConstraints(),
                icon: Icon(Icons.delete, size: 18, color: global.theme.negativeHighlightTextColor),
                onPressed: () => _removeDepositDoc(index),
                tooltip: global.language('delete'),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildDataCell({required Widget child, int flex = 1}) {
    return Expanded(
      flex: flex,
      child: Container(
        height: double.infinity,
        alignment: Alignment.center,
        padding: const EdgeInsets.symmetric(horizontal: 4),
        child: child,
      ),
    );
  }

  Widget _buildTextField({
    required String value,
    required bool isReadOnly,
    TextInputType? keyboardType,
    List<TextInputFormatter>? inputFormatters,
    required Function(String)? onChanged,
    TextAlign? textAlign,
    FontWeight? fontWeight,
    Widget? suffixIcon,
    VoidCallback? onTap,
    bool hasError = false,
  }) {
    return Container(
      height: 48, // เพิ่มความสูงจาก 40 เป็น 48
      alignment: Alignment.center,
      child: TextFormField(
        key: ValueKey(
          value,
        ), // เพิ่ม key เพื่อบังคับให้ rebuild เมื่อค่าเปลี่ยน
        initialValue: value,
        readOnly: isReadOnly,
        keyboardType: keyboardType,
        inputFormatters: inputFormatters,
        onChanged: onChanged,
        onTap: onTap,
        textAlign: textAlign ?? TextAlign.start,
        textAlignVertical: TextAlignVertical.center,
        style: TextStyle(fontSize: 12, fontWeight: fontWeight),
        decoration: InputDecoration(
          border: OutlineInputBorder(
            borderSide: BorderSide(
              color: hasError ? global.theme.negativeHighlightTextColor : global.theme.dividerBorderColor,
              width: 1,
            ),
          ),
          enabledBorder: OutlineInputBorder(
            borderSide: BorderSide(
              color: hasError ? global.theme.negativeHighlightTextColor : global.theme.dividerBorderColor,
              width: 1,
            ),
          ),
          focusedBorder: OutlineInputBorder(
            borderSide: BorderSide(
              color: hasError ? global.theme.negativeHighlightTextColor : global.theme.infoHighlightTextColor,
              width: 1,
            ),
          ),
          contentPadding: const EdgeInsets.symmetric(
            horizontal: 10,
            vertical: 12,
          ), // เพิ่ม padding
          isDense: true,
          filled: true,
          fillColor: hasError
              ? Colors.red[50]
              : (isReadOnly ? global.theme.surfaceColor : global.theme.cardColor),
          suffixIcon: suffixIcon,
        ),
      ),
    );
  }
}
