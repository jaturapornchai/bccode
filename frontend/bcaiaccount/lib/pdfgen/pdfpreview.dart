
import 'package:flutter/foundation.dart';
import 'package:smlaicloud/bloc/company_branch/company_branch_bloc.dart';
import 'package:smlaicloud/bloc/creditor/creditor_bloc.dart';
import 'package:smlaicloud/bloc/debtor/debtor_bloc.dart';
import 'package:smlaicloud/global.dart';
import 'package:smlaicloud/model/company_branch_model.dart';
import 'package:smlaicloud/model/customer_address_model.dart';
import 'package:smlaicloud/model/debtor_creditor_model.dart';
import 'package:smlaicloud/pdfgen/pdf/accrual_receive.dart';
import 'package:smlaicloud/pdfgen/pdf/advance_payment.dart';
import 'package:smlaicloud/pdfgen/pdf/paid.dart';
import 'package:smlaicloud/pdfgen/pdf/pay.dart';
import 'package:smlaicloud/pdfgen/pdf/purchase.dart';
import 'package:smlaicloud/pdfgen/pdf/purchase_order.dart';
import 'package:smlaicloud/pdfgen/pdf/purchase_partial.dart';
import 'package:smlaicloud/pdfgen/pdf/purchase_return.dart';
import 'package:smlaicloud/pdfgen/pdf/sale_invoice_order.dart';
import 'package:smlaicloud/pdfgen/pdf/sale_invoice_return.dart';
import 'package:smlaicloud/pdfgen/pdf/stock_adjustment.dart';
import 'package:smlaicloud/pdfgen/pdf/stock_pickup_product.dart';
import 'package:smlaicloud/pdfgen/pdf/stock_receive_product.dart';
import 'package:smlaicloud/pdfgen/pdf/stock_transfer.dart';
import 'package:smlaicloud/pdfgen/pdf/transaction_returnproduct.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:pdf/pdf.dart';
import 'package:printing/printing.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'pdf/sale_invoice.dart';
import 'package:smlaicloud/global.dart' as global;

class PdfPreviewPage extends StatefulWidget {
  final TransactionModel? screenData;
  final TransactionPaidPayModel? screenDataPayPaid;
  final TransactionTypeEnum type;

  const PdfPreviewPage({
    super.key,
    this.screenData,
    this.screenDataPayPaid,
    required this.type,
  });

  @override
  PdfPreviewPageState createState() => PdfPreviewPageState();
}

class PdfPreviewPageState extends State<PdfPreviewPage> {
  late CompanyBranchModel companyBranchData;
  late DebtorCreditorModel debtorCreditorData;
  late List<TransactionModel>? originalTransactionData;

  @override
  void initState() {
    super.initState();

    companyBranchData = CompanyBranchModel(guidfixed: '', code: '');
    debtorCreditorData = DebtorCreditorModel(
      addressforbilling: CustomerAddressModel(
        guid: '',
        address: [],
        countrycode: '',
        provincecode: '',
        districtcode: '',
        subdistrictcode: '',
        zipcode: '',
        contactnames: [],
        phoneprimary: '',
        phonesecondary: '',
        latitude: 0,
        longitude: 0,
      ),
    );

    originalTransactionData = [];

    loadDataForm();
  }

  Future<void> loadDataForm() async {
    /// Load ข้อมูลบริษัท

    if (widget.type == TransactionTypeEnum.pay ||
        widget.type == TransactionTypeEnum.paid) {
      context.read<CompanyBranchBloc>().add(
        CompanyBranchGet(guid: widget.screenDataPayPaid!.branch!.guidfixed!),
      );
      if (widget.screenDataPayPaid!.custcode.isNotEmpty) {
        if (widget.type == TransactionTypeEnum.paid) {
          /// Load ข้อมูลลูกหนี้
          context.read<DebtorBloc>().add(
            DebtorGetBycode(custcode: widget.screenDataPayPaid!.custcode),
          );
        } else if (widget.type == TransactionTypeEnum.pay) {
          /// Load ข้อมูลเจ้าหนี้
          context.read<CreditorBloc>().add(
            CreditorGetBycode(custcode: widget.screenDataPayPaid!.custcode),
          );
        }
      }
    } else {
      context.read<CompanyBranchBloc>().add(
        CompanyBranchGet(guid: widget.screenData!.branch!.guidfixed!),
      );
      if (widget.screenData!.custcode.isNotEmpty) {
        if (widget.type == TransactionTypeEnum.sale ||
            widget.type == TransactionTypeEnum.salereturn ||
            widget.type == TransactionTypeEnum.paid ||
            widget.type == TransactionTypeEnum.saleorder) {
          /// Load ข้อมูลลูกหนี้
          context.read<DebtorBloc>().add(
            DebtorGetBycode(custcode: widget.screenData!.custcode),
          );
        } else if (widget.type == TransactionTypeEnum.purchase ||
            widget.type == TransactionTypeEnum.purchasereturn ||
            widget.type == TransactionTypeEnum.pay ||
            widget.type == TransactionTypeEnum.stocktransfer ||
            widget.type == TransactionTypeEnum.stockreceiveproduct ||
            widget.type == TransactionTypeEnum.stockpickupproduct ||
            widget.type == TransactionTypeEnum.stockreturnproduct ||
            widget.type == TransactionTypeEnum.adjust ||
            widget.type == TransactionTypeEnum.purchaseorder ||
            widget.type == TransactionTypeEnum.purchasepartial ||
            widget.type == TransactionTypeEnum.advancePayment) {
          /// Load ข้อมูลเจ้าหนี้
          context.read<CreditorBloc>().add(
            CreditorGetBycode(custcode: widget.screenData!.custcode),
          );
        }
      }
    }
  }

  Future<Uint8List> getGenPdf() async {
    PdfPageFormat format = PdfPageFormat.a4.copyWith(
      marginTop: 30, // Top margin
      marginBottom: 30, // Bottom margin
      marginLeft: 30, // Left margin
      marginRight: 30, // Right margin
    );

    await Future.delayed(const Duration(seconds: 1));

    if (widget.type == TransactionTypeEnum.sale) {
      return generateInvoice(
        format,
        widget.screenData!,
        companyBranchData,
        debtorCreditorData,
      );
    } else if (widget.type == TransactionTypeEnum.salereturn) {
      return generateInvoiceReturn(
        format,
        widget.screenData!,
        companyBranchData,
        debtorCreditorData,
      );
    } else if (widget.type == TransactionTypeEnum.purchase) {
      return generatePurchase(
        format,
        widget.screenData!,
        companyBranchData,
        debtorCreditorData,
      );
    } else if (widget.type == TransactionTypeEnum.purchasereturn) {
      return generatePurchaseReturn(
        format,
        widget.screenData!,
        companyBranchData,
        debtorCreditorData,
      );
    } else if (widget.type == TransactionTypeEnum.paid) {
      return generatePaid(
        format,
        widget.screenDataPayPaid!,
        companyBranchData,
        debtorCreditorData,
      );
    } else if (widget.type == TransactionTypeEnum.pay) {
      return generatePay(
        format,
        widget.screenDataPayPaid!,
        companyBranchData,
        debtorCreditorData,
      );
    } else if (widget.type == TransactionTypeEnum.stocktransfer) {
      return generateStockTransfer(
        format,
        widget.screenData!,
        companyBranchData,
      );
    } else if (widget.type == TransactionTypeEnum.stockreceiveproduct) {
      return generateStockReceiveProduct(
        format,
        widget.screenData!,
        companyBranchData,
      );
    } else if (widget.type == TransactionTypeEnum.stockpickupproduct) {
      return generateStockPickupProduct(
        format,
        widget.screenData!,
        companyBranchData,
      );
    } else if (widget.type == TransactionTypeEnum.stockreturnproduct) {
      return generateTransactionReturnProduct(
        format,
        widget.screenData!,
        companyBranchData,
      );
    } else if (widget.type == TransactionTypeEnum.adjust) {
      return generateStockAdjustment(
        format,
        widget.screenData!,
        companyBranchData,
      );
    } else if (widget.type == TransactionTypeEnum.saleorder) {
      return generateSaleOrder(
        format,
        widget.screenData!,
        companyBranchData,
        debtorCreditorData,
      );
    } else if (widget.type == TransactionTypeEnum.purchaseorder) {
      return generatePurchaseOrder(
        format,
        widget.screenData!,
        companyBranchData,
        debtorCreditorData,
      );
    } else if (widget.type == TransactionTypeEnum.purchasepartial) {
      return generatePurchasePartial(
        format,
        widget.screenData!,
        companyBranchData,
        debtorCreditorData,
      );
    } else if (widget.type == TransactionTypeEnum.accrualreceive) {
      return generateAccrualReceive(
        format,
        widget.screenData!,
        companyBranchData,
        debtorCreditorData,
      );
    } else if (widget.type == TransactionTypeEnum.advancePayment) {
      return generateAdvancePayment(
        format,
        widget.screenData!,
        companyBranchData,
        debtorCreditorData,
      );
    } else {
      return Future.value(Uint8List(0));
    }
  }

  @override
  Widget build(BuildContext context) {
    return MultiBlocListener(
      listeners: [
        BlocListener<CompanyBranchBloc, CompanyBranchState>(
          listener: (context, state) {
            if (state is CompanyBranchGetSuccess) {
              companyBranchData = state.companyBranch;
            } else if (state is CompanyBranchGetFailed) {
              setState(() {
                global.showSnackBar(
                  context,
                  const Icon(Icons.error, color: Colors.white),
                  state.message,
                  Colors.red,
                );
              });
            }
          },
        ),
        BlocListener<DebtorBloc, DebtorState>(
          listener: (context, state) {
            if (state is DebtorGetBycodeSuccess) {
              debtorCreditorData = DebtorCreditorModel(
                addressforbilling: state.debtors.addressforbilling,
                code: state.debtors.code,
                names: state.debtors.names,
                taxid: state.debtors.taxid,
              );
            } else if (state is DebtorGetBycodeFailed) {
              setState(() {
                global.showSnackBar(
                  context,
                  const Icon(Icons.error, color: Colors.white),
                  state.message,
                  Colors.red,
                );
              });
            }
          },
        ),
        BlocListener<CreditorBloc, CreditorState>(
          listener: (context, state) {
            if (state is CreditorGetBycodeSuccess) {
              debtorCreditorData = DebtorCreditorModel(
                addressforbilling: state.creditors.addressforbilling,
                code: state.creditors.code,
                names: state.creditors.names,
                taxid: state.creditors.taxid,
              );
            } else if (state is CreditorGetBycodeFailed) {
              setState(() {
                global.showSnackBar(
                  context,
                  const Icon(Icons.error, color: Colors.white),
                  state.message,
                  Colors.red,
                );
              });
            }
          },
        ),
      ],
      child: Scaffold(
        appBar: AppBar(
          backgroundColor: global.theme.appBarColor,
          title: Text(global.language("print")),
          actions: [
            /// print
            Padding(
              padding: const EdgeInsets.only(right: 10),
              child: IconButton(
                onPressed: () {
                  Printing.layoutPdf(
                    onLayout: (PdfPageFormat format) => getGenPdf(),
                  );
                },
                icon: const Icon(Icons.print),
              ),
            ),
          ],
        ),
        body: kIsWeb
            ? // สำหรับ Web - ใช้ PdfPreview โดยไม่มี InteractiveViewer
              SizedBox(
                width: double.infinity,
                height: double.infinity,
                child: PdfPreview(
                  allowSharing: false,
                  canChangePageFormat: false,
                  canChangeOrientation: false,
                  allowPrinting: false,
                  maxPageWidth: 700,
                  canDebug: false, // ปิด debug mode สำหรับ production
                  onError: (context, error) {
                    return Center(
                      child: Text('${global.language("error_loading_pdf")}: ${error.toString()}'),
                    );
                  },
                  build: (format) async => await getGenPdf(),
                ),
              )
            : // สำหรับ Mobile - ใช้ InteractiveViewer
              Expanded(
                child: InteractiveViewer(
                  constrained: true,
                  boundaryMargin: const EdgeInsets.all(20.0),
                  minScale: 0.5,
                  maxScale: 3.0,
                  child: PdfPreview(
                    allowSharing: false,
                    canChangePageFormat: false,
                    canChangeOrientation: false,
                    allowPrinting: false,
                    maxPageWidth: 700,
                    canDebug: false,
                    build: (format) async => await getGenPdf(),
                  ),
                ),
              ),
        floatingActionButton: FloatingActionButton(
          onPressed: () {
            Printing.layoutPdf(onLayout: (PdfPageFormat format) => getGenPdf());
          },
          backgroundColor: global.theme.appBarColor,
          child: const Icon(Icons.print),
        ),
      ),
    );
  }
}
