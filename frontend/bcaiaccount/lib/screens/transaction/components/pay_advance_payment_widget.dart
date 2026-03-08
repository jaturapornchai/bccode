import 'package:flutter/material.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/screen_search/transaction_search_screen.dart';
import 'dart:async';

class PayAdvancePaymentWidget extends StatefulWidget {
  final List<AdvancePaymentDocModel> advancePaymentDocs;
  final Function(List<AdvancePaymentDocModel>) onAdvancePaymentDocsChanged;
  final Function(double) onSumAdvancePaymentChanged;
  final String custcode; // เพิ่มสำหรับ TransSearchScreen
  final int inquirytype; // เพิ่มสำหรับการเช็ค enabled
  final global.TransactionTypeEnum type; // เพิ่มสำหรับตรวจสอบ transaction type

  const PayAdvancePaymentWidget({
    super.key,
    required this.advancePaymentDocs,
    required this.onAdvancePaymentDocsChanged,
    required this.onSumAdvancePaymentChanged,
    required this.custcode,
    required this.inquirytype,
    required this.type,
  });

  @override
  State<PayAdvancePaymentWidget> createState() =>
      _PayAdvancePaymentWidgetState();
}

class _PayAdvancePaymentWidgetState extends State<PayAdvancePaymentWidget> {
  late List<AdvancePaymentDocModel> _advancePaymentDocs;
  double _sumAdvancePayment = 0.0;
  Timer? _debounceTimer;
  final Map<int, bool> _errorStates = {}; // เก็บสถานะ error ของแต่ละ field

  @override
  void initState() {
    super.initState();
    _advancePaymentDocs = List.from(widget.advancePaymentDocs);
    _calculateSumAdvancePayment();
  }

  @override
  void dispose() {
    _debounceTimer?.cancel();
    super.dispose();
  }

  @override
  void didUpdateWidget(PayAdvancePaymentWidget oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (widget.advancePaymentDocs != oldWidget.advancePaymentDocs) {
      _advancePaymentDocs = List.from(widget.advancePaymentDocs);
      _calculateSumAdvancePayment();
    }
  }

  void _calculateSumAdvancePayment() {
    final newSum = _advancePaymentDocs.fold(
      0.0,
      (sum, doc) => sum + doc.useamount,
    );
    if (_sumAdvancePayment != newSum) {
      _sumAdvancePayment = newSum;
      WidgetsBinding.instance.addPostFrameCallback((_) {
        widget.onSumAdvancePaymentChanged(_sumAdvancePayment);
      });
    }
  }

  // Helper method to determine the correct transaction type for TransSearchScreen
  global.TransactionTypeEnum _getSearchTransactionType() {
    if (widget.type == global.TransactionTypeEnum.purchase) {
      return global.TransactionTypeEnum.advancePayment; // จ่ายเงินล่วงหน้า
    } else if (widget.type == global.TransactionTypeEnum.sale) {
      return global.TransactionTypeEnum.paidAdvance; // รับเงินล่วงหน้า
    }
    // Default fallback
    return global.TransactionTypeEnum.advancePayment;
  }

  void _addAdvancePaymentDoc() {
    if (widget.custcode.isEmpty) {
      String message = widget.type == global.TransactionTypeEnum.purchase
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
          bool alreadyExists = _advancePaymentDocs.any(
            (doc) => doc.docno == result.docno,
          );

          if (!alreadyExists) {
            // สร้าง AdvancePaymentDocModel จากข้อมูลที่ได้รับ
            AdvancePaymentDocModel newAdvancePaymentDoc =
                AdvancePaymentDocModel(
                  balance:
                      result.totalamount, // ยอดคงเหลือเริ่มต้นเท่ากับยอดรวม
                  docdatetime: result.docdatetime,
                  docno: result.docno,
                  guidfixed: result.guidfixed ?? '',
                  remark: '',
                  totalamount: result.totalamount,
                  useamount: 0, // ยอดตัดเริ่มต้น = 0
                );

            setState(() {
              _advancePaymentDocs.add(newAdvancePaymentDoc);
              _calculateSumAdvancePayment();
            });

            WidgetsBinding.instance.addPostFrameCallback((_) {
              widget.onAdvancePaymentDocsChanged(_advancePaymentDocs);
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
      String message = widget.type == global.TransactionTypeEnum.purchase
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
          // ตรวจสอบว่าเอกสารนี้ถูกเพิ่มแล้วหรือไม่ (ยกเว้นตำแหน่งปัจจุบัน)
          bool alreadyExists = _advancePaymentDocs.asMap().entries.any(
            (entry) => entry.key != index && entry.value.docno == result.docno,
          );

          if (!alreadyExists) {
            // อัพเดทข้อมูลในตำแหน่งเดิม
            AdvancePaymentDocModel updatedAdvancePaymentDoc =
                AdvancePaymentDocModel(
                  balance: result.totalamount, // รีเซ็ตยอดคงเหลือ
                  docdatetime: result.docdatetime,
                  docno: result.docno,
                  guidfixed: result.guidfixed ?? '',
                  remark: _advancePaymentDocs[index].remark, // เก็บหมายเหตุเดิม
                  totalamount: result.totalamount,
                  useamount: 0, // รีเซ็ตยอดตัดเป็น 0
                );

            setState(() {
              _advancePaymentDocs[index] = updatedAdvancePaymentDoc;
              _calculateSumAdvancePayment();
            });

            WidgetsBinding.instance.addPostFrameCallback((_) {
              widget.onAdvancePaymentDocsChanged(_advancePaymentDocs);
            });
          } else {
            // แสดงข้อความแจ้งเตือนว่าเอกสารถูกเพิ่มแล้ว
            global.showWarningSnackBar(context, '${global.language("document_already_used")} ${result.docno}');
          }
        }
      }
    });
  }

  void _removeAdvancePaymentDoc(int index) {
    setState(() {
      _advancePaymentDocs.removeAt(index);
      _errorStates.remove(index); // ลบ error state ของ index นี้
      _calculateSumAdvancePayment();
    });
    WidgetsBinding.instance.addPostFrameCallback((_) {
      widget.onAdvancePaymentDocsChanged(_advancePaymentDocs);
    });
  }

  void _validateAndUpdateUseAmount(
    String value,
    int index,
    AdvancePaymentDocModel advancePaymentDoc,
  ) {
    final useAmount = _parseAmount(value);

    // ตรวจสอบไม่ให้ยอดตัดเกินยอดเงินล่วงหน้า
    if (useAmount > advancePaymentDoc.totalamount) {
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
              '${global.language("advance_payment_deduction")} ${_formatAmount(advancePaymentDoc.totalamount)} ${global.language("baht")}',
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

    // คำนวณยอดคงเหลือใหม่
    final balance = advancePaymentDoc.totalamount - useAmount;
    _updateAdvancePaymentDoc(
      index,
      advancePaymentDoc.copyWith(
        useamount: useAmount,
        balance: balance, // อัพเดต balance
      ),
    );
  }

  void _updateAdvancePaymentDoc(int index, AdvancePaymentDocModel updatedDoc) {
    setState(() {
      _advancePaymentDocs[index] = updatedDoc;
      _calculateSumAdvancePayment();
    });
    WidgetsBinding.instance.addPostFrameCallback((_) {
      widget.onAdvancePaymentDocsChanged(_advancePaymentDocs);
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
                  global.language('deduct_advance_payment'),
                  style: Theme.of(context).textTheme.titleMedium?.copyWith(
                    fontWeight: FontWeight.bold,
                    color: Colors.green[700],
                  ),
                ),
                Row(
                  children: [
                    Text(
                      '${global.language("total")}: ${_formatAmount(_sumAdvancePayment)} ${global.language("baht")}',
                      style: Theme.of(context).textTheme.titleSmall?.copyWith(
                        fontWeight: FontWeight.bold,
                        color: Colors.orange[700],
                      ),
                    ),
                    SizedBox(width: 10),
                    ElevatedButton.icon(
                      onPressed: (widget.inquirytype != 0)
                          ? _addAdvancePaymentDoc
                          : null,
                      icon: Icon(Icons.add, size: 18),
                      label: Text(global.language('add')),
                      style: ElevatedButton.styleFrom(
                        backgroundColor: (widget.inquirytype != 0)
                            ? Colors.green[600]
                            : Colors.grey[400],
                        foregroundColor: Colors.white,
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

            // ถ้าไม่มีข้อมูลใน advancePaymentDocs ให้ซ่อนรายการ
            if (_advancePaymentDocs.isNotEmpty)
              Column(
                children: [
                  const SizedBox(height: 16),
                  _buildAdvancePaymentCardList(),
                ],
              ),
          ],
        ),
      ),
    );
  }

  Widget _buildAdvancePaymentCardList() {
    return Column(
      children: [
        // Advance payment cards
        ...List.generate(_advancePaymentDocs.length, (index) {
          return _buildAdvancePaymentCard(index, _advancePaymentDocs[index]);
        }),
      ],
    );
  }

  Widget _buildAdvancePaymentCard(
    int index,
    AdvancePaymentDocModel advancePaymentDoc,
  ) {
    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: Colors.grey[300]!),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Header
          Row(
            children: [
              Text(
                '${global.language('advance_payment')} #${index + 1}',
                style: const TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.w600,
                ),
              ),
              const Spacer(),
              MouseRegion(
                cursor: SystemMouseCursors.click,
                child: GestureDetector(
                  onTap: () => _removeAdvancePaymentDoc(index),
                  child: Icon(Icons.close, size: 20, color: Colors.grey[600]),
                ),
              ),
            ],
          ),

          if (advancePaymentDoc.docno.isNotEmpty) ...[
            const SizedBox(height: 8),
            MouseRegion(
              cursor: SystemMouseCursors.click,
              child: GestureDetector(
                onTap: () => _selectNewDocument(index),
                child: Container(
                  padding: const EdgeInsets.symmetric(
                    vertical: 4,
                    horizontal: 8,
                  ),
                  decoration: BoxDecoration(
                    color: Colors.green[50],
                    borderRadius: BorderRadius.circular(4),
                    border: Border.all(color: Colors.green[200]!),
                  ),
                  child: Row(
                    children: [
                      Expanded(
                        child: Text(
                          '${global.language("docno")}: ${advancePaymentDoc.docno} / ${global.dateTimeBuddhist(DateTime.parse(advancePaymentDoc.docdatetime))}',
                          style: TextStyle(
                            fontSize: 14,
                            color: Colors.green[700],
                            fontWeight: FontWeight.w500,
                          ),
                        ),
                      ),
                      Icon(Icons.edit, size: 16, color: Colors.green[600]),
                    ],
                  ),
                ),
              ),
            ),
          ] else ...[
            SizedBox(height: 8),
            MouseRegion(
              cursor: SystemMouseCursors.click,
              child: GestureDetector(
                onTap: () => _selectNewDocument(index),
                child: Container(
                  padding: const EdgeInsets.symmetric(
                    vertical: 8,
                    horizontal: 8,
                  ),
                  decoration: BoxDecoration(
                    color: Colors.grey[100],
                    borderRadius: BorderRadius.circular(4),
                    border: Border.all(color: Colors.grey[300]!),
                  ),
                  child: Row(
                    children: [
                      Expanded(
                        child: Text(
                          global.language('tap_to_select_document'),
                          style: TextStyle(
                            fontSize: 14,
                            color: Colors.grey[600],
                          ),
                        ),
                      ),
                      Icon(Icons.search, size: 16, color: Colors.grey[600]),
                    ],
                  ),
                ),
              ),
            ),
          ],

          const SizedBox(height: 16),

          // Amount info And Editable use amount - จัด layout ให้เท่ากัน
          Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // ส่วนแสดงยอดเงินล่วงหน้าและยอดคงเหลือ
              Expanded(
                child: Container(
                  padding: const EdgeInsets.all(16),
                  decoration: BoxDecoration(
                    color: Colors.blue[50],
                    borderRadius: BorderRadius.circular(8),
                    border: Border.all(color: Colors.blue[200]!),
                  ),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        global.language("payment_info"),
                        style: TextStyle(
                          fontSize: 14,
                          fontWeight: FontWeight.w500,
                          color: Colors.blue[800],
                        ),
                      ),
                      const SizedBox(height: 12),
                      // ยอดเงินล่วงหน้า
                      Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          Text(
                            '${global.language("advance_payment_amount")}:',
                            style: TextStyle(
                              fontSize: 13,
                              color: Colors.grey[700],
                            ),
                          ),
                          Text(
                            _formatAmount(advancePaymentDoc.totalamount),
                            style: const TextStyle(
                              fontSize: 14,
                              fontWeight: FontWeight.w600,
                              color: Colors.black, // เปลี่ยนเป็นสีดำ
                            ),
                          ),
                        ],
                      ),
                      const SizedBox(height: 8),
                      // ยอดคงเหลือ
                      Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          Text(
                            '${global.language("remaining_balance")}:',
                            style: TextStyle(
                              fontSize: 13,
                              color: Colors.grey[700],
                            ),
                          ),
                          Text(
                            key: ValueKey(
                              'balance_${index}_${advancePaymentDoc.balance}',
                            ),
                            _formatAmount(advancePaymentDoc.balance),
                            style: TextStyle(
                              fontSize: 14,
                              fontWeight: FontWeight.w600,
                              color: Colors.green[700], // เปลี่ยนเป็นสีเขียว
                            ),
                          ),
                        ],
                      ),
                    ],
                  ),
                ),
              ),

              const SizedBox(width: 16),

              // ส่วนกรอกยอดตัดเงินล่วงหน้า
              Expanded(
                child: Container(
                  padding: const EdgeInsets.all(16),
                  decoration: BoxDecoration(
                    color: _errorStates[index] == true
                        ? Colors.red[50]
                        : Colors.grey[50],
                    borderRadius: BorderRadius.circular(8),
                    border: Border.all(
                      color: _errorStates[index] == true
                          ? Colors.red[200]!
                          : Colors.grey[300]!,
                    ),
                  ),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        global.language("advance_payment_deduction_label"),
                        style: TextStyle(
                          fontSize: 14,
                          fontWeight: FontWeight.w500,
                          color: _errorStates[index] == true
                              ? Colors.red[800]
                              : Colors.grey[700],
                        ),
                      ),
                      const SizedBox(height: 12),
                      TextFormField(
                        key: ValueKey(
                          'useamount_${index}_${advancePaymentDoc.useamount}',
                        ),
                        initialValue: _formatAmount(
                          advancePaymentDoc.useamount,
                        ),
                        keyboardType: const TextInputType.numberWithOptions(
                          decimal: true,
                        ),
                        inputFormatters: [global.NumberInputFormatter()],
                        style: TextStyle(
                          fontSize: 16,
                          fontWeight: FontWeight.w600,
                          color: _errorStates[index] == true
                              ? Colors.red[700]
                              : Colors.purple[700],
                        ),
                        decoration: InputDecoration(
                          hintText: '0.00',
                          hintStyle: TextStyle(color: Colors.grey[400]),
                          suffixText: global.language("baht"),
                          suffixStyle: TextStyle(
                            color: Colors.grey[600],
                            fontWeight: FontWeight.w500,
                          ),
                          border: OutlineInputBorder(
                            borderRadius: BorderRadius.circular(8),
                            borderSide: BorderSide(
                              color: _errorStates[index] == true
                                  ? Colors.red[400]!
                                  : Colors.grey[300]!,
                              width: _errorStates[index] == true ? 2 : 1,
                            ),
                          ),
                          enabledBorder: OutlineInputBorder(
                            borderRadius: BorderRadius.circular(8),
                            borderSide: BorderSide(
                              color: _errorStates[index] == true
                                  ? Colors.red[400]!
                                  : Colors.grey[300]!,
                              width: _errorStates[index] == true ? 2 : 1,
                            ),
                          ),
                          focusedBorder: OutlineInputBorder(
                            borderRadius: BorderRadius.circular(8),
                            borderSide: BorderSide(
                              color: _errorStates[index] == true
                                  ? Colors.red[400]!
                                  : Colors.purple[400]!,
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
                              : Colors.white,
                        ),
                        onChanged: (value) {
                          // ยกเลิก timer เก่า (ถ้ามี)
                          _debounceTimer?.cancel();

                          // ตั้ง timer ใหม่ให้รอ 1 วินาทีหลังจากผู้ใช้หยุดพิมพ์
                          _debounceTimer = Timer(
                            const Duration(milliseconds: 1000),
                            () {
                              _validateAndUpdateUseAmount(
                                value,
                                index,
                                advancePaymentDoc,
                              );
                            },
                          );
                        },
                      ),
                    ],
                  ),
                ),
              ),
            ],
          ),

          const SizedBox(height: 8),

          // Remark
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                global.language('remark'),
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.w500,
                  color: Colors.grey[700],
                ),
              ),
              SizedBox(height: 8),
              TextFormField(
                key: ValueKey('remark_${index}_${advancePaymentDoc.remark}'),
                initialValue: advancePaymentDoc.remark,
                maxLines: 2,
                style: TextStyle(fontSize: 14),
                decoration: InputDecoration(
                  hintText: global.language('add_remark'),
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(8),
                    borderSide: BorderSide(color: Colors.grey[300]!),
                  ),
                  focusedBorder: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(8),
                    borderSide: BorderSide(color: Colors.green[400]!, width: 2),
                  ),
                  contentPadding: const EdgeInsets.symmetric(
                    horizontal: 12,
                    vertical: 12,
                  ),
                  filled: true,
                  fillColor: Colors.grey[50],
                ),
                onChanged: (value) {
                  _updateAdvancePaymentDoc(
                    index,
                    advancePaymentDoc.copyWith(remark: value),
                  );
                },
              ),
            ],
          ),
        ],
      ),
    );
  }
}
