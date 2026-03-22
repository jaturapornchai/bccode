import 'dart:async';

import 'package:smlaicloud/bloc/import_product/import_product_bloc.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/import_product_model.dart';
import 'package:smlaicloud/model/pagination.dart';
import 'package:smlaicloud/model/product_model.dart';
import 'package:smlaicloud/repositories/product_barcode_repository.dart';
import 'package:smlaicloud/screen_search/unit_search_screen.dart';
import 'package:file_picker/file_picker.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter/services.dart' show rootBundle;
import 'package:smlaicloud/screens/report/file_download.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

class ImportProductScreen extends StatefulWidget {
  const ImportProductScreen({super.key});

  @override
  State<ImportProductScreen> createState() => _ImportProductScreenState();
}

class _ImportProductScreenState extends State<ImportProductScreen>
    with global.ThemeRefreshMixin {
  final ProductBarcodeRepository _productBarcodeRepository =
      ProductBarcodeRepository();
  String fileImportName = '';
  bool isLoading = false;
  bool isShowTable = false;
  bool isLoadingTable = false;
  bool isLoadingText = false;
  bool isLoadingSaveTaskid = false;
  int page = 1;
  int limit = 20;
  String q = '';
  String taskid = '';
  final List<int> _limitOptions = <int>[10, 20, 50, 100];

  // Variables for status checking
  Timer? _statusTimer;
  ImportProductStatusModel? currentStatus;
  int currentProgress = 0;

  // Variables for compare detail
  CompareDetailModel? compareDetailModel;
  List<CompareItemModel> compareItems = [];

  List<ImportProductModel> detaillmportProductModel = [];
  Pagination pagination = Pagination(
    page: 1,
    perPage: 20,
    total: 10,
    totalPage: 1,
    next: 0,
    prev: 0,
  );

  LanguageModel languangeImport = global.config.languages[0];


  @override
  void dispose() {
    _statusTimer?.cancel();
    super.dispose();
  }

  void startStatusChecking() {
    if (taskid.isNotEmpty) {
      if (kDebugMode) {
        AppLogger.debug('Starting status checking for taskid: $taskid');
      }

      // Check status immediately
      context.read<ImportProductBloc>().add(GetStatusTaskid(taskid: taskid));

      // Start periodic checking every 2 seconds
      _statusTimer = Timer.periodic(const Duration(seconds: 2), (timer) {
        if (taskid.isNotEmpty && isLoadingSaveTaskid) {
          if (kDebugMode) {
            AppLogger.debug(
              'Checking status... taskid: $taskid, isLoadingSaveTaskid: $isLoadingSaveTaskid',
            );
          }
          context.read<ImportProductBloc>().add(
            GetStatusTaskid(taskid: taskid),
          );
        } else {
          if (kDebugMode) {
            AppLogger.debug('Stopping timer - taskid empty or not loading');
          }
          timer.cancel();
        }
      });
    }
  }

  void stopStatusChecking() {
    _statusTimer?.cancel();
    _statusTimer = null;
  }

  void fetchData(int page, int limit, String q) {
    this.page = page;
    this.limit = limit;
    this.q = q;

    context.read<ImportProductBloc>().add(
      CompareDetail(taskid: taskid, page: page, limit: limit, fast: false),
    );
  }

  String getTextError(ImportProductModel item) {
    List<String> errorMessages = [];

    if (item.isexist!) {
      errorMessages.add(global.language('product_already_exists'));
    }

    if (item.isduplicate!) {
      errorMessages.add(global.language('product_duplicate'));
    }

    if (item.isunitnotexist!) {
      errorMessages.add(global.language('unit_not_exist'));
    }

    // Join the error messages with a comma and space, and trim any trailing commas just in case
    return errorMessages.join(', ').trim();
  }

  String getTextErrorFromCompareItem(CompareItemModel item) {
    List<String> errorMessages = [];

    if (item.importData?.isexist == true) {
      errorMessages.add(global.language('product_already_exists'));
    }

    if (item.importData?.isduplicate == true) {
      errorMessages.add(global.language('product_duplicate'));
    }

    if (item.importData?.isunitnotexist == true) {
      errorMessages.add(global.language('unit_not_exist'));
    }

    // Join the error messages with a comma and space, and trim any trailing commas just in case
    return errorMessages.join(', ').trim();
  }

  Color _getStatusColor(String status) {
    switch (status.toUpperCase()) {
      case 'NEW':
        return global.theme.positiveHighlightTextColor;
      case 'EXISTING':
        return global.theme.infoHighlightTextColor;
      case 'UPDATED':
        return global.theme.warningHighlightTextColor;
      case 'CONFLICT':
        return global.theme.negativeHighlightTextColor;
      default:
        return global.theme.iconSecondaryColor;
    }
  }

  Future<void> uploadFileExcel() async {
    double filesize = 0;

    /// select file excel from device
    FilePickerResult? result = await FilePicker.platform.pickFiles(
      type: FileType.custom,
      allowedExtensions: ['xlsx'],
    );
    if (result != null) {
      filesize = result.files.first.size / 1048576;

      /// set filesize 2 decimal
      filesize = double.parse(filesize.toStringAsFixed(2));

      if (filesize <= 2.0) {
        if (mounted) {
          setState(() {
            fileImportName = result.files.first.name;
          });

          late Uint8List file = result.files.first.bytes!;

          context.read<ImportProductBloc>().add(
            UploadFileExcel(file: file, filename: fileImportName),
          );
        }
      } else {
        setState(() {
          showDialog(
            context: context,
            builder: (BuildContext context) {
              return AlertDialog(
                title: Text(global.language('file_size_limit_exceeded')),
                content: RichText(
                  text: TextSpan(
                    children: <TextSpan>[
                      TextSpan(
                        text: global.language(
                          'please_select_file_not_exceed_2mb',
                        ),
                      ),
                      TextSpan(text: ' '),
                      TextSpan(
                        text:
                            '${global.language('file_size_label')} : $filesize MB',
                        style: TextStyle(color: global.theme.negativeHighlightTextColor),
                      ),
                    ],
                  ),
                ),
                actions: [
                  TextButton(
                    onPressed: () {
                      Navigator.of(context).pop();
                    },
                    child: Text(global.language('close')),
                  ),
                ],
              );
            },
          );
        });
      }
    }
  }

  void _showBarcodeDialog(
    BuildContext context,
    ImportProductModel? item,
  ) async {
    TextEditingController barcode = TextEditingController();
    TextEditingController name = TextEditingController();
    TextEditingController unitcode = TextEditingController();
    TextEditingController price = TextEditingController();
    TextEditingController pricemember = TextEditingController();
    TextEditingController pricedelivery = TextEditingController();

    ProductBarcodeModel result = ProductBarcodeModel(guidfixed: '');

    barcode.text = "";
    name.text = "";
    unitcode.text = "";
    price.text = "";
    pricemember.text = "";
    pricedelivery.text = "";

    FocusNode barcodefocus = FocusNode();
    FocusNode namefocus = FocusNode();
    FocusNode unitcodefocus = FocusNode();
    FocusNode pricefocus = FocusNode();
    FocusNode pricememberfocus = FocusNode();

    if (item != null) {
      barcode.text = item.barcode!;
      name.text = item.name!;
      unitcode.text = item.unitcode!;
      price.text = item.price.toString();
      pricemember.text = item.pricemember.toString();
      pricedelivery.text = item.pricedelivery.toString();
    }

    return showDialog(
      context: context,
      barrierDismissible: true, // Allows tapping outside the dialog to close it
      builder: (BuildContext context) {
        return AlertDialog(
          title: Text(global.language('barcode')),
          content: SizedBox(
            width: (global.isMobileScreen(context)) ? 350 : 500,
            height: 200,
            child: Column(
              children: [
                Row(
                  children: [
                    Expanded(
                      child: TextField(
                        autofocus: true,
                        controller: barcode,
                        focusNode: barcodefocus,
                        decoration: InputDecoration(
                          labelText: '',
                          border: const OutlineInputBorder(),
                          suffixIcon: IconButton(
                            icon: Icon(Icons.clear),
                            onPressed: () {
                              barcode.text = '';
                              name.text = '';
                              unitcode.text = '';
                              price.text = "";
                              pricemember.text = "";
                              pricedelivery.text = "";
                              Future.delayed(
                                const Duration(milliseconds: 200),
                                () {
                                  FocusScope.of(
                                    context,
                                  ).requestFocus(barcodefocus);
                                },
                              );
                              setState(() {});
                            },
                          ),
                        ),
                        onSubmitted: (value) {
                          if (value.trim().isNotEmpty) {
                            String barcodeValue = "";
                            barcodeValue = value.trim();
                            _productBarcodeRepository
                                .getProductBarcodeDetail(barcodeValue)
                                .then((value) {
                                  if (value.success && value.data != null) {
                                    result = ProductBarcodeModel.fromJson(
                                      value.data,
                                    );

                                    if (result.itemtype == 3) {
                                      return;
                                    }
                                    name.text = global.activeLangName(
                                      result.names!,
                                    );
                                    unitcode.text = result.itemunitcode;

                                    Future.delayed(
                                      const Duration(milliseconds: 200),
                                      () {
                                        FocusScope.of(
                                          context,
                                        ).requestFocus(pricefocus);
                                      },
                                    );
                                  }
                                  setState(() {});
                                })
                                .onError((error, stackTrace) {
                                  barcode.text = '';
                                  Future.delayed(
                                    const Duration(milliseconds: 200),
                                    () {
                                      FocusScope.of(
                                        context,
                                      ).requestFocus(barcodefocus);
                                    },
                                  );
                                  setState(() {});
                                });
                          }
                        },
                      ),
                    ),
                  ],
                ),
                SizedBox(height: 10),
                Row(
                  children: [
                    Expanded(
                      child: TextField(
                        controller: name,
                        focusNode: namefocus,
                        decoration: InputDecoration(
                          labelText: global.language('product_name'),
                          border: const OutlineInputBorder(),
                        ),
                      ),
                    ),
                    SizedBox(width: 10),
                    Expanded(
                      child: TextField(
                        controller: unitcode,
                        focusNode: unitcodefocus,
                        decoration: InputDecoration(
                          labelText: global.language('unit_code'),
                          border: const OutlineInputBorder(),
                          prefixIcon: IconButton(
                            onPressed: () {
                              Navigator.push(
                                context,
                                MaterialPageRoute(
                                  builder: (context) =>
                                      const UnitSearchScreen(word: ''),
                                ),
                              ).then((value) {
                                global.SearchCodeNameModel result = value;
                                if (result.code.isNotEmpty) {
                                  setState(() {
                                    unitcode.text = result.code;

                                    Future.delayed(
                                      const Duration(milliseconds: 200),
                                      () {
                                        FocusScope.of(
                                          context,
                                        ).requestFocus(pricefocus);
                                      },
                                    );
                                  });
                                }
                              });
                            },
                            icon: Icon(Icons.search),
                          ),
                        ),
                      ),
                    ),
                  ],
                ),
                SizedBox(height: 10),
                Row(
                  children: [
                    Expanded(
                      child: TextField(
                        controller: price,
                        focusNode: pricefocus,
                        keyboardType: const TextInputType.numberWithOptions(
                          decimal: true,
                        ),
                        inputFormatters: [global.NumberInputFormatter()],
                        decoration: InputDecoration(
                          labelText: global.language('price'),
                          border: const OutlineInputBorder(),
                        ),
                        onChanged: (value) {
                          if (value.isNotEmpty) {
                            if (value == '0') {
                              price.selection = TextSelection.fromPosition(
                                TextPosition(offset: price.text.length),
                              );
                            }
                          } else {
                            price.text = '0';
                          }
                        },
                        onEditingComplete: () {
                          Future.delayed(const Duration(milliseconds: 200), () {
                            FocusScope.of(
                              context,
                            ).requestFocus(pricememberfocus);
                          });
                        },
                      ),
                    ),
                    SizedBox(width: 10),
                    Expanded(
                      child: TextField(
                        controller: pricemember,
                        focusNode: pricememberfocus,
                        keyboardType: const TextInputType.numberWithOptions(
                          decimal: true,
                        ),
                        inputFormatters: [global.NumberInputFormatter()],
                        decoration: InputDecoration(
                          labelText: global.language('price_member'),
                          border: const OutlineInputBorder(),
                        ),
                        onChanged: (value) {
                          if (value.isNotEmpty) {
                            if (value == '0') {
                              pricemember
                                  .selection = TextSelection.fromPosition(
                                TextPosition(offset: pricemember.text.length),
                              );
                            }
                          } else {
                            pricemember.text = '0';
                          }
                        },
                      ),
                    ),
                    // SizedBox(
                    //   width: 10,
                    // ),
                    // Expanded(
                    //   child: TextField(
                    //     controller: pricedelivery,
                    //     focusNode: pricedeliveryfocus,
                    //     keyboardType: const TextInputType.numberWithOptions(decimal: true),
                    //     inputFormatters: [global.NumberInputFormatter()],
                    //     decoration: InputDecoration(
                    //       labelText: global.language("product_price_delivery"),
                    //       border: const OutlineInputBorder(),
                    //     ),
                    //     onChanged: (value) {
                    //       if (value.isNotEmpty) {
                    //         if (value == '0') {
                    //           pricedelivery.selection = TextSelection.fromPosition(TextPosition(offset: pricedelivery.text.length));
                    //         }
                    //       } else {
                    //         pricedelivery.text = '0';
                    //       }
                    //     },
                    //   ),
                    // ),
                  ],
                ),
              ],
            ),
          ),
          actions: <Widget>[
            TextButton(
              child: Text(global.language('close')),
              onPressed: () {
                Navigator.pop(context);
              },
            ),

            /// button Save - New UI Design with double-tap prevention
            ElevatedButton(
              onPressed: (barcode.text.isEmpty || isLoadingTable)
                  ? null
                  : () async {
                      if (barcode.text.isEmpty) {
                        Future.delayed(const Duration(milliseconds: 200), () {
                          FocusScope.of(context).requestFocus(barcodefocus);
                        });
                        return;
                      }

                      // Prevent double tap by setting loading state
                      setState(() {
                        isLoadingTable = true;
                      });

                      try {
                        ImportProductModel detail = ImportProductModel(
                          guidfixed: (item == null) ? '' : item.guidfixed!,
                          barcode: barcode.text,
                          name: name.text,
                          unitcode: unitcode.text,
                          price: double.parse(price.text.replaceAll(',', '')),
                          pricemember: double.parse(
                            pricemember.text.replaceAll(',', ''),
                          ),
                          pricedelivery: double.parse(
                            pricedelivery.text.replaceAll(',', ''),
                          ),
                          taskid: taskid,
                          rownumber: (item == null) ? 0 : item.rownumber!,
                        );

                        if (item == null) {
                          context.read<ImportProductBloc>().add(
                            AddDetail(importProductModel: detail),
                          );
                        } else {
                          context.read<ImportProductBloc>().add(
                            UpdateDetail(
                              guid: item.guidfixed!,
                              importProductModel: detail,
                            ),
                          );
                        }

                        Navigator.pop(context);
                      } catch (e) {
                        // Reset loading state on error
                        setState(() {
                          isLoadingTable = false;
                        });
                        // Show error message
                        global.showErrorSnackBar(context, '${global.language('error_occurred')}: ${e.toString()}');
                      }
                    },
              style: ElevatedButton.styleFrom(
                backgroundColor: isLoadingTable
                    ? global.theme.iconSecondaryColor
                    : global.theme.appBarColor,
                foregroundColor: global.theme.onPrimaryColor,
                padding: const EdgeInsets.symmetric(
                  horizontal: 24,
                  vertical: 12,
                ),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(8),
                ),
                elevation: isLoadingTable ? 0 : 2,
              ),
              child: isLoadingTable
                  ? SizedBox(
                      width: 20,
                      height: 20,
                      child: CircularProgressIndicator(
                        strokeWidth: 2,
                        valueColor: AlwaysStoppedAnimation<Color>(global.theme.onPrimaryColor),
                      ),
                    )
                  : Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Icon(Icons.save, size: 18),
                        SizedBox(width: 8),
                        Text(
                          global.language('save'),
                          style: TextStyle(
                            fontSize: 16,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                      ],
                    ),
            ),
          ],
        );
      },
    );
  }

  Widget editProductListDetailWidget(double screenHeight) {
    return (isShowTable)
        ? Column(
            children: [
              ///
              Padding(
                padding: const EdgeInsets.only(
                  top: 20.0,
                  left: 20.0,
                  right: 20,
                ),
                child: Row(
                  /// between
                  children: [
                    Expanded(
                      flex: 1,
                      child: Row(
                        children: [
                          /// frist page
                          IconButton(
                            icon: Icon(Icons.first_page),
                            onPressed: pagination.page > 1
                                ? () => fetchData(1, limit, q)
                                : null,
                          ),
                          IconButton(
                            icon: Icon(Icons.chevron_left),
                            onPressed: pagination.page > 1
                                ? () => fetchData(pagination.page - 1, limit, q)
                                : null,
                          ),
                          Text(
                            'Page ${pagination.page} of ${pagination.totalPage}',
                          ),
                          IconButton(
                            icon: Icon(Icons.chevron_right),
                            onPressed: pagination.page < pagination.totalPage
                                ? () => fetchData(pagination.page + 1, limit, q)
                                : null,
                          ),

                          /// last page
                          IconButton(
                            icon: Icon(Icons.last_page),
                            onPressed: pagination.page < pagination.totalPage
                                ? () =>
                                      fetchData(pagination.totalPage, limit, q)
                                : null,
                          ),

                          /// size box width 10
                          SizedBox(width: 10),
                          DropdownButton<int>(
                            value: limit,
                            items: _limitOptions.map<DropdownMenuItem<int>>((
                              int value,
                            ) {
                              return DropdownMenuItem<int>(
                                value: value,
                                child: Text(value.toString()),
                              );
                            }).toList(),
                            onChanged: (int? newValue) {
                              setState(() {
                                limit = newValue!;
                              });
                              fetchData(
                                page,
                                limit,
                                q,
                              ); // Go to the first page with the new rows per page
                            },
                          ),
                        ],
                      ),
                    ),
                    Spacer(),
                    Expanded(
                      child: SizedBox(
                        height: 49,
                        child: InputDecorator(
                          decoration: InputDecoration(
                            border: OutlineInputBorder(),
                            labelText: global.language(
                              'select_import_language',
                            ),
                          ),
                          child: DropdownButtonHideUnderline(
                            child: DropdownButton<LanguageModel>(
                              value: languangeImport,
                              icon: Icon(Icons.arrow_drop_down),
                              style: TextStyle(color: Colors.deepPurple),
                              underline: Container(
                                color: Colors.deepPurpleAccent,
                              ),
                              onChanged: (LanguageModel? value) {
                                setState(() {
                                  languangeImport = value!;
                                });
                              },
                              isDense: true,
                              isExpanded: true,
                              items: global.config.languages
                                  .map<DropdownMenuItem<LanguageModel>>((
                                    LanguageModel value,
                                  ) {
                                    return DropdownMenuItem<LanguageModel>(
                                      value: value,
                                      child: Row(
                                        children: <Widget>[
                                          Image.asset(
                                            'assets/flags/${value.code}.png', // Ensure the image path is correct
                                            width: 30,
                                            height: 30,
                                            errorBuilder:
                                                (context, error, stackTrace) {
                                                  return Icon(
                                                    Icons.error,
                                                  ); // Error icon if the image fails to load
                                                },
                                          ),
                                          SizedBox(
                                            width: 10,
                                          ), // Spacing between the image and text
                                          Text(value.name!),
                                        ],
                                      ),
                                    );
                                  })
                                  .toList(),
                            ),
                          ),
                        ),
                      ),
                    ),
                    const Spacer(),
                    Expanded(
                      flex: 1,
                      child: TextField(
                        decoration: InputDecoration(
                          labelText: global.language('search'),
                          suffixIcon: isLoadingText
                              ? Padding(
                                  padding: EdgeInsets.all(8.0),
                                  child: CircularProgressIndicator(),
                                )
                              : null, // Loading indicator
                          prefixIcon: Icon(Icons.search),
                        ),
                        onChanged: (value) {
                          setState(() {
                            q = value;
                          });
                          fetchData(1, limit, value);
                        },
                      ),
                    ),
                  ],
                ),
              ),
              SizedBox(
                width: double.infinity,
                height: screenHeight * 0.7,
                child: isLoadingTable
                    ? Center(
                        child: Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            const CircularProgressIndicator(),
                            SizedBox(height: 10),
                            Text(global.language('loading_data')),
                          ],
                        ),
                      )
                    : SingleChildScrollView(
                        child: Column(
                          children: [
                            SizedBox(
                              width: double.infinity,
                              child: DataTable(
                                columns: <DataColumn>[
                                  DataColumn(
                                    label: Text(global.language('barcode')),
                                  ),
                                  DataColumn(
                                    label: Text(
                                      global.language('product_name'),
                                    ),
                                  ),
                                  DataColumn(
                                    label: Text(global.language('unit_code')),
                                  ),
                                  DataColumn(
                                    label: Text(
                                      global.language('selling_price'),
                                    ),
                                  ),
                                  DataColumn(
                                    label: Text(
                                      global.language('price_member'),
                                    ),
                                  ),
                                  // DataColumn(label: Text(global.language('product_price_delivery'))),
                                  DataColumn(
                                    label: Text(global.language('status')),
                                  ),
                                  DataColumn(
                                    label: Text(
                                      global.language('import_status'),
                                    ),
                                  ),
                                  // DataColumn(
                                  //     label: Text(global.language('edit'))),
                                  // DataColumn(
                                  //     label: Text(global.language('delete'))),
                                  // Add more columns as needed
                                ],
                                rows: (!isLoadingTable)
                                    ? compareItems
                                          .map<DataRow>(
                                            (item) => DataRow(
                                              cells: <DataCell>[
                                                DataCell(
                                                  Text(
                                                    item.importData?.barcode ??
                                                        '',
                                                    style:
                                                        (item
                                                                    .importData
                                                                    ?.isexist ==
                                                                true ||
                                                            item
                                                                    .importData
                                                                    ?.isduplicate ==
                                                                true ||
                                                            item
                                                                    .importData
                                                                    ?.isunitnotexist ==
                                                                true)
                                                        ? TextStyle(
                                                            fontWeight:
                                                                FontWeight.bold,
                                                            color: global.theme.negativeHighlightTextColor,
                                                          )
                                                        : TextStyle(
                                                            fontWeight:
                                                                FontWeight
                                                                    .normal,
                                                          ),
                                                  ),
                                                ),
                                                DataCell(
                                                  Text(
                                                    item.importData?.name ?? '',
                                                    style:
                                                        (item
                                                                    .importData
                                                                    ?.isexist ==
                                                                true ||
                                                            item
                                                                    .importData
                                                                    ?.isduplicate ==
                                                                true ||
                                                            item
                                                                    .importData
                                                                    ?.isunitnotexist ==
                                                                true)
                                                        ? TextStyle(
                                                            fontWeight:
                                                                FontWeight.bold,
                                                            color: global.theme.negativeHighlightTextColor,
                                                          )
                                                        : TextStyle(
                                                            fontWeight:
                                                                FontWeight
                                                                    .normal,
                                                          ),
                                                  ),
                                                ),
                                                DataCell(
                                                  Text(
                                                    item.importData?.unitcode ??
                                                        '',
                                                    style:
                                                        (item
                                                                    .importData
                                                                    ?.isexist ==
                                                                true ||
                                                            item
                                                                    .importData
                                                                    ?.isduplicate ==
                                                                true ||
                                                            item
                                                                    .importData
                                                                    ?.isunitnotexist ==
                                                                true)
                                                        ? TextStyle(
                                                            fontWeight:
                                                                FontWeight.bold,
                                                            color: global.theme.negativeHighlightTextColor,
                                                          )
                                                        : TextStyle(
                                                            fontWeight:
                                                                FontWeight
                                                                    .normal,
                                                          ),
                                                  ),
                                                ),
                                                DataCell(
                                                  Align(
                                                    alignment:
                                                        Alignment.centerRight,
                                                    child: Text(
                                                      (item.importData?.price ??
                                                              0)
                                                          .toString(),
                                                      style:
                                                          (item
                                                                      .importData
                                                                      ?.isexist ==
                                                                  true ||
                                                              item
                                                                      .importData
                                                                      ?.isduplicate ==
                                                                  true ||
                                                              item
                                                                      .importData
                                                                      ?.isunitnotexist ==
                                                                  true)
                                                          ? TextStyle(
                                                              fontWeight:
                                                                  FontWeight
                                                                      .bold,
                                                              color: global.theme.negativeHighlightTextColor,
                                                            )
                                                          : TextStyle(
                                                              fontWeight:
                                                                  FontWeight
                                                                      .normal,
                                                            ),
                                                    ),
                                                  ),
                                                ),
                                                DataCell(
                                                  Align(
                                                    alignment:
                                                        Alignment.centerRight,
                                                    child: Text(
                                                      (item
                                                                  .importData
                                                                  ?.pricemember ??
                                                              0)
                                                          .toString(),
                                                      style:
                                                          (item
                                                                      .importData
                                                                      ?.isexist ==
                                                                  true ||
                                                              item
                                                                      .importData
                                                                      ?.isduplicate ==
                                                                  true ||
                                                              item
                                                                      .importData
                                                                      ?.isunitnotexist ==
                                                                  true)
                                                          ? TextStyle(
                                                              fontWeight:
                                                                  FontWeight
                                                                      .bold,
                                                              color: global.theme.negativeHighlightTextColor,
                                                            )
                                                          : TextStyle(
                                                              fontWeight:
                                                                  FontWeight
                                                                      .normal,
                                                            ),
                                                    ),
                                                  ),
                                                ),
                                                // DataCell(
                                                //   Align(
                                                //     alignment: Alignment.centerRight,
                                                //     child: Text(
                                                //       item.pricedelivery!.toString(),
                                                //       style: (item.isexist! || item.isduplicate! || item.isunitnotexist!)
                                                //           ? TextStyle(
                                                //               fontWeight: FontWeight.bold,
                                                //               color: global.theme.negativeHighlightTextColor,
                                                //             )
                                                //           : TextStyle(
                                                //               fontWeight: FontWeight.normal,
                                                //             ),
                                                //     ),
                                                //   ),
                                                // ),
                                                DataCell(
                                                  Align(
                                                    alignment: Alignment.center,
                                                    child:
                                                        (item
                                                                    .importData
                                                                    ?.isduplicate ==
                                                                false &&
                                                            item
                                                                    .importData
                                                                    ?.isexist ==
                                                                false &&
                                                            item
                                                                    .importData
                                                                    ?.isunitnotexist ==
                                                                false)
                                                        ? Icon(
                                                            Icons.check,
                                                            color: global.theme.positiveHighlightTextColor,
                                                          )
                                                        : ElevatedButton(
                                                            style: ElevatedButton.styleFrom(
                                                              foregroundColor:
                                                                  global.theme.onPrimaryColor,
                                                              backgroundColor:
                                                                  Colors
                                                                      .red, // Text color
                                                            ),
                                                            onPressed: () {
                                                              showDialog(
                                                                context:
                                                                    context,
                                                                builder:
                                                                    (
                                                                      BuildContext
                                                                      context,
                                                                    ) {
                                                                      return AlertDialog(
                                                                        title: Text(
                                                                          'Error',
                                                                        ),
                                                                        content: Text(
                                                                          getTextErrorFromCompareItem(
                                                                            item,
                                                                          ),
                                                                        ),
                                                                        actions:
                                                                            <
                                                                              Widget
                                                                            >[
                                                                              TextButton(
                                                                                onPressed: () => Navigator.of(
                                                                                  context,
                                                                                ).pop(),
                                                                                child: Text(
                                                                                  'Close',
                                                                                ),
                                                                              ),
                                                                            ],
                                                                      );
                                                                    },
                                                              );
                                                            },
                                                            child: Icon(
                                                              Icons.error,
                                                              color:
                                                                  global.theme.onPrimaryColor,
                                                            ), // Icon color
                                                          ),
                                                  ),
                                                ),
                                                DataCell(
                                                  Align(
                                                    alignment: Alignment.center,
                                                    child: Container(
                                                      padding:
                                                          const EdgeInsets.symmetric(
                                                            horizontal: 12,
                                                            vertical: 6,
                                                          ),
                                                      decoration: BoxDecoration(
                                                        color: _getStatusColor(
                                                          item.status ?? '',
                                                        ),
                                                        borderRadius:
                                                            BorderRadius.circular(
                                                              12,
                                                            ),
                                                      ),
                                                      child: Text(
                                                        item.status ?? '',
                                                        style: TextStyle(
                                                          color: global.theme.onPrimaryColor,
                                                          fontWeight:
                                                              FontWeight.bold,
                                                        ),
                                                      ),
                                                    ),
                                                  ),
                                                ),
                                                // DataCell(Align(
                                                //   alignment: Alignment.center,
                                                //   child: IconButton(
                                                //     icon: Icon(Icons.edit),
                                                //     onPressed: () {
                                                //       _showBarcodeDialog(
                                                //           context, item);
                                                //     },
                                                //   ),
                                                // )),
                                                // DataCell(
                                                //   Align(
                                                //     alignment: Alignment.center,
                                                //     child: IconButton(
                                                //       icon: Icon(
                                                //           Icons.delete),
                                                //       onPressed: () {
                                                //         /// show dialog confirm delete
                                                //         showDialog(
                                                //           context: context,
                                                //           builder: (BuildContext
                                                //               context) {
                                                //             return AlertDialog(
                                                //               title: Text(
                                                //
                                                //                       'confirm_delete')),
                                                //               content: Text(
                                                //
                                                //                       'are_you_sure_you_want_to_delete_this_item')),
                                                //               actions: <Widget>[
                                                //                 TextButton(
                                                //                   onPressed: () {
                                                //                     Navigator.of(
                                                //                             context)
                                                //                         .pop();
                                                //                   },
                                                //                   child: Text(global
                                                //                       .language(
                                                //                           'close')),
                                                //                 ),
                                                //                 ElevatedButton(
                                                //                   style: ElevatedButton
                                                //                       .styleFrom(
                                                //                           backgroundColor:
                                                //                               global.theme.negativeHighlightTextColor),
                                                //                   onPressed: () {
                                                //                     Navigator.pop(
                                                //                         context);
                                                //                   },
                                                //                   child: Text(
                                                //
                                                //                         'delete'),
                                                //                   ),
                                                //                 ),
                                                //               ],
                                                //             );
                                                //           },
                                                //         );
                                                //       },
                                                //     ),
                                                //   ),
                                                // ),
                                              ],
                                            ),
                                          )
                                          .toList()
                                    : [],
                              ),
                            ),
                          ],
                        ),
                      ),
              ),
            ],
          )
        : Container();
  }

  Widget menutListButtomBar() {
    return SizedBox(
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Row(
            children: [
              (fileImportName.isNotEmpty)
                  ? Padding(
                      padding: EdgeInsets.all(8.0),
                      child: ElevatedButton(
                        onPressed: () {
                          _showBarcodeDialog(context, null);
                        },
                        child: Text(
                          global.language('add_new_line_with_barcode'),
                        ),
                      ),
                    )
                  : Container(),
              Padding(
                padding: const EdgeInsets.all(8.0),
                child: (fileImportName.isEmpty)
                    ? Container()
                    : ElevatedButton.icon(
                        /// button red
                        style: ElevatedButton.styleFrom(
                          foregroundColor: global.theme.onPrimaryColor,
                          backgroundColor: global.theme.negativeHighlightTextColor,
                        ),

                        onPressed: () {
                          discardData(
                            callBack: () {
                              // Clear data without API call
                              setState(() {
                                fileImportName = '';
                                taskid = '';
                                compareItems.clear();
                                pagination = Pagination(
                                  page: 1,
                                  perPage: 10,
                                  total: 0,
                                  totalPage: 1,
                                  next: 0,
                                  prev: 0,
                                );
                              });
                            },
                          );
                        },
                        icon: Icon(Icons.clear),
                        label: Text(fileImportName),
                      ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  void discardData({required Function callBack}) {
    if (fileImportName.isNotEmpty) {
      showDialog<String>(
        context: context,
        builder: (BuildContext context) => AlertDialog(
          title: Text(global.language('data_uploaded')),
          content: Text(
            '${global.language('confirm_delete_data')} : $fileImportName ?',
          ),
          actions: <Widget>[
            ElevatedButton(
              style: ElevatedButton.styleFrom(backgroundColor: global.theme.negativeHighlightTextColor),
              onPressed: () => Navigator.pop(context),
              child: Text(global.language('no')),
            ),
            ElevatedButton(
              style: ElevatedButton.styleFrom(backgroundColor: global.theme.infoHighlightTextColor),
              onPressed: () {
                Navigator.pop(context);
                callBack();
              },
              child: Text(global.language('yes')),
            ),
          ],
        ),
      );
    } else {
      callBack();
    }
  }

  Future<void> downloadAssetFile(String assetPath, String saveFileName) async {
    // Load the asset file
    final ByteData data = await rootBundle.load(assetPath);
    final Uint8List bytes = data.buffer.asUint8List();

    // Use the download function
    bool success = await downloadAssetFileBytes(bytes, saveFileName);

    if (success) {
      if (kDebugMode) {
        AppLogger.info("File successfully downloaded.");
      }
    } else {
      if (kDebugMode) {
        AppLogger.error("Failed to download file.");
      }
    }
  }

  Future<void> saveTaskIdOperation() async {
    // Start the apply import task (async mode)
    if (taskid.isNotEmpty) {
      if (mounted) {
        context.read<ImportProductBloc>().add(
          ApplyImport(
            taskid: taskid,
            async: true,
            batchSize: 200,
            forceUpdate: true,
            importMode: 'AUTO',
          ),
        );
        // Start checking status after initiating apply
        startStatusChecking();
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    double screenHeight = MediaQuery.of(
      context,
    ).size.height; // Get the screen height

    return Scaffold(
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language('import_product')),
        leading: IconButton(
          focusNode: FocusNode(skipTraversal: true),
          icon: Icon(Icons.arrow_back),
          onPressed: isLoadingSaveTaskid
              ? null
              : () {
                  global.gotoMainMenu(context);
                },
        ),
        actions: [
          ElevatedButton.icon(
            icon: Icon(Icons.save, size: 26.0),
            label: Text(
              global.language('save'),
            ), // Adjust text based on your localization setup
            onPressed: (fileImportName.isEmpty || isLoadingSaveTaskid)
                ? null
                : () async {
                    // Show confirmation dialog
                    final confirm = await showDialog<bool>(
                      context: context,
                      builder: (BuildContext context) {
                        return AlertDialog(
                          title: Text(
                            global.language('confirm_import_product'),
                          ),
                          content: Text(
                            '${global.language('confirm_import_file')} $fileImportName',
                          ),
                          actions: <Widget>[
                            TextButton(
                              onPressed: () => Navigator.of(context).pop(false),
                              child: Text(global.language('close')),
                            ),
                            ElevatedButton(
                              style: ElevatedButton.styleFrom(
                                backgroundColor: global.theme.infoHighlightTextColor,
                              ),
                              onPressed: () => Navigator.of(context).pop(true),
                              child: Text(global.language('save')),
                            ),
                          ],
                        );
                      },
                    );

                    // If confirmed, start save operation
                    if (confirm == true) {
                      // BlocListener will handle loading state automatically
                      await saveTaskIdOperation();
                    }
                  },
          ),
        ],
      ),
      body: MultiBlocListener(
        listeners: [
          BlocListener<ImportProductBloc, ImportProductState>(
            listener: (context, state) {
              // Handle InProgress states
              if (state is UploadFileExcelInProgress) {
                setState(() {
                  isLoading = true;
                });
              } else if (state is LoadImportProductByTaskidInProgress ||
                  state is CompareDetailInProgress) {
                setState(() {
                  isLoadingTable = true;
                });
              } else if (state is VerifyTaskidInProgress) {
                setState(() {
                  isLoadingText = true;
                });
              } else if (state is UpdateDetailInProgress ||
                  state is AddDetailInProgress ||
                  state is DeleteDetailByGuidInProgress) {
                setState(() {
                  isLoadingTable = true;
                });
              } else if (state is SaveTaskidInProgress ||
                  state is ApplyImportInProgress) {
                setState(() {
                  isLoadingSaveTaskid = true;
                });
              }

              if (state is UploadFileExcelSuccess) {
                taskid = state.response.id;
                // เรียก CompareDetail แทน VerifyTaskid
                context.read<ImportProductBloc>().add(
                  CompareDetail(
                    taskid: taskid,
                    page: 1,
                    limit: 20,
                    fast: false,
                  ),
                );
              } else if (state is UploadFileExcelFailed) {
                setState(() {
                  fileImportName = '';
                  isLoading = false;

                  /// show dialog error
                  showDialog(
                    context: context,
                    builder: (BuildContext context) {
                      return AlertDialog(
                        title: Text(
                          '${global.language('error')} : ${global.language('invalid_file_format')}',
                        ),
                        content: Text(state.message),
                        actions: [
                          TextButton(
                            onPressed: () {
                              Navigator.of(context).pop();
                            },
                            child: Text(global.language('close')),
                          ),
                        ],
                      );
                    },
                  );
                });
              }

              if (state is VerifyTaskidSuccess) {
                /// delay 1 second
                Future.delayed(const Duration(seconds: 1), () {
                  fetchData(1, 20, "");
                });
              }

              if (state is CompareDetailSuccess) {
                setState(() {
                  compareDetailModel = state.data;
                  compareItems = state.data.items ?? [];
                  pagination = state.pagination;

                  isLoading = false;
                  isLoadingText = false;
                  isLoadingTable = false;
                  isShowTable = true;
                });
              }

              if (state is CompareDetailFailed) {
                setState(() {
                  isLoading = false;
                  isLoadingText = false;
                  isLoadingTable = false;
                  isShowTable = false;

                  global.showSnackBar(
                    context,
                    Icon(Icons.error, color: global.theme.onPrimaryColor),
                    state.message,
                    global.theme.negativeHighlightTextColor,
                  );
                });
              }

              if (state is LoadImportProductByTaskidSuccess) {
                setState(() {
                  detaillmportProductModel = [];
                  pagination = state.pagination;
                  detaillmportProductModel = state.data;

                  isLoading = false;
                  isLoadingText = false;
                  isLoadingTable = false;
                  isShowTable = true;
                });
              }
              if (state is LoadImportProductByTaskidFailed) {
                setState(() {
                  isLoading = false;
                  isLoadingText = false;
                  isLoadingTable = false;
                  isShowTable = false;

                  global.showSnackBar(
                    context,
                    Icon(Icons.error, color: global.theme.onPrimaryColor),
                    state.message,
                    global.theme.negativeHighlightTextColor,
                  );
                });
              }

              if (state is DeleteDetailByGuidSuccess) {
                setState(() {
                  isLoadingTable = false;
                });
                fetchData(page, limit, q);
              }

              if (state is UpdateDetailSuccess) {
                setState(() {
                  isLoadingTable = false;
                });
                // เรียก CompareDetail แทน VerifyTaskid
                fetchData(page, limit, q);
                global.showSnackBar(
                  context,
                  Icon(Icons.save, color: global.theme.onPrimaryColor),
                  global.language('update_success'),
                  global.theme.infoHighlightTextColor,
                );
              }

              if (state is AddDetailSuccess) {
                setState(() {
                  isLoadingTable = false;
                });
                // เรียก CompareDetail แทน VerifyTaskid
                fetchData(page, limit, q);
                global.showSnackBar(
                  context,
                  Icon(Icons.save, color: global.theme.onPrimaryColor),
                  global.language('save_success'),
                  global.theme.infoHighlightTextColor,
                );
              }

              // if (state is SaveTaskidSuccess) {
              //   // Don't stop loading immediately, let status checking handle completion
              //   // Just show initial success message
              //   global.showSnackBar(
              //     context,
              //     Icon(
              //       Icons.info,
              //       color: global.theme.onPrimaryColor,
              //     ),
              //     'เริ่มดำเนินการแล้ว',
              //     global.theme.infoHighlightTextColor,
              //   );
              // }

              if (state is ApplyImportSuccess) {
                // Show initial success message
                global.showSnackBar(
                  context,
                  Icon(Icons.info, color: global.theme.onPrimaryColor),
                  global.language('operation_started'),
                  global.theme.infoHighlightTextColor,
                );
              }

              if (state is ApplyImportFailed) {
                setState(() {
                  isLoadingSaveTaskid = false;
                  currentProgress = 0;
                  currentStatus = null;
                });
                stopStatusChecking();
                global.showSnackBar(
                  context,
                  Icon(Icons.error, color: global.theme.onPrimaryColor),
                  '${global.language('save_failed')} : ${state.message}',
                  global.theme.negativeHighlightTextColor,
                );
              }

              if (state is SaveTaskidFailed) {
                setState(() {
                  isLoadingSaveTaskid = false;
                  currentProgress = 0;
                  currentStatus = null;
                });
                stopStatusChecking();
                global.showSnackBar(
                  context,
                  Icon(Icons.error, color: global.theme.onPrimaryColor),
                  '${global.language('save_failed')} : ${state.message}',
                  global.theme.negativeHighlightTextColor,
                );
              }

              // Handle other failed states
              if (state is UpdateDetailFailed) {
                setState(() {
                  isLoadingTable = false;
                });
                global.showSnackBar(
                  context,
                  Icon(Icons.error, color: global.theme.onPrimaryColor),
                  '${global.language('update_failed')} : ${state.message}',
                  global.theme.negativeHighlightTextColor,
                );
              }

              if (state is AddDetailFailed) {
                setState(() {
                  isLoadingTable = false;
                });
                global.showSnackBar(
                  context,
                  Icon(Icons.error, color: global.theme.onPrimaryColor),
                  '${global.language('save_failed')} : ${state.message}',
                  global.theme.negativeHighlightTextColor,
                );
              }

              if (state is DeleteDetailByGuidFailed) {
                setState(() {
                  isLoadingTable = false;
                });
                global.showSnackBar(
                  context,
                  Icon(Icons.error, color: global.theme.onPrimaryColor),
                  '${global.language('delete_failed')} : ${state.message}',
                  global.theme.negativeHighlightTextColor,
                );
              }

              if (state is VerifyTaskidFailed) {
                setState(() {
                  isLoadingText = false;
                });
                global.showSnackBar(
                  context,
                  Icon(Icons.error, color: global.theme.onPrimaryColor),
                  '${global.language('verify_failed')} : ${state.message}',
                  global.theme.negativeHighlightTextColor,
                );
              }

              // Handle GetStatusTaskid states
              if (state is GetStatusTaskidSuccess) {
                if (kDebugMode) {
                  AppLogger.debug(
                    'GetStatusTaskidSuccess: status=${state.data.status}, progress=${state.data.progress}',
                  );
                }

                setState(() {
                  currentStatus = state.data;
                  currentProgress = state.data.progress;
                });

                // Check if status is COMPLETED
                if (state.data.status == "COMPLETED") {
                  if (kDebugMode) {
                    AppLogger.debug(
                      'Status is COMPLETED - stopping timer and showing success',
                    );
                  }

                  stopStatusChecking();

                  // Show success dialog with option to start new import
                  showDialog(
                    context: context,
                    barrierDismissible: false,
                    builder: (BuildContext dialogContext) {
                      return AlertDialog(
                        title: Row(
                          children: [
                            Icon(
                              Icons.check_circle,
                              color: global.theme.positiveHighlightTextColor,
                              size: 32,
                            ),
                            SizedBox(width: 12),
                            Text(global.language('import_completed')),
                          ],
                        ),
                        content: Column(
                          mainAxisSize: MainAxisSize.min,
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              global.language('product_import_completed'),
                              style: TextStyle(fontSize: 16),
                            ),
                            SizedBox(height: 16),
                            Text(
                              'Progress: ${state.data.progress}%',
                              style: TextStyle(color: global.theme.textSecondaryColor),
                            ),
                          ],
                        ),
                        actions: [
                          ElevatedButton(
                            style: ElevatedButton.styleFrom(
                              backgroundColor: global.theme.appBarColor,
                              foregroundColor: global.theme.onPrimaryColor,
                            ),
                            onPressed: () {
                              Navigator.pop(dialogContext);

                              // Clear all data and reset for new import
                              setState(() {
                                fileImportName = '';
                                taskid = '';
                                isLoadingSaveTaskid = false;
                                isLoading = false;
                                isShowTable = false;
                                isLoadingTable = false;
                                isLoadingText = false;
                                currentProgress = 0;
                                currentStatus = null;
                                compareItems = [];
                                compareDetailModel = null;
                                detaillmportProductModel = [];
                                pagination = Pagination(
                                  page: 1,
                                  perPage: 20,
                                  total: 10,
                                  totalPage: 1,
                                  next: 0,
                                  prev: 0,
                                );
                              });

                              // Show ready for new import message
                              global.showSnackBar(
                                context,
                                Icon(Icons.info, color: global.theme.onPrimaryColor),
                                global.language('ready_for_new_import'),
                                global.theme.infoHighlightTextColor,
                              );
                            },
                            child: Text(global.language('ok')),
                          ),
                        ],
                      );
                    },
                  );
                }
              }

              if (state is GetStatusTaskidFailed) {
                // If task not found or completed, stop loading
                if (state.message.contains("Task not found or completed")) {
                  setState(() {
                    isLoadingSaveTaskid = false;
                    currentProgress = 100;
                  });
                  stopStatusChecking();

                  // Reset all states as task is completed
                  setState(() {
                    pagination = Pagination(
                      page: 1,
                      perPage: 20,
                      total: 10,
                      totalPage: 1,
                      next: 0,
                      prev: 0,
                    );
                    detaillmportProductModel = <ImportProductModel>[];
                    isLoading = false;
                    isLoadingText = false;
                    isLoadingTable = false;
                    isShowTable = false;
                    fileImportName = '';
                    currentStatus = null;
                    currentProgress = 0;
                  });

                  global.showSnackBar(
                    context,
                    Icon(Icons.check_circle, color: global.theme.onPrimaryColor),
                    global.language('save_success'),
                    global.theme.infoHighlightTextColor,
                  );
                }
              }
            },
          ),
        ],
        child: SafeArea(
          child: Stack(
            children: [
              (isLoading)
                  ? Center(
                      child: SizedBox(
                        width: 100, // Set the desired width
                        height: 100, // Set the desired height
                        child: CircularProgressIndicator(
                          strokeWidth:
                              4, // Optional: Set the thickness of the indicator
                          valueColor: AlwaysStoppedAnimation<Color>(
                            global.theme.infoHighlightTextColor,
                          ), // Optional: Set the color
                        ),
                      ),
                    )
                  : SizedBox(
                      width: double.infinity,
                      child: (fileImportName.isEmpty)
                          ? Column(
                              mainAxisAlignment: MainAxisAlignment.center,
                              crossAxisAlignment: CrossAxisAlignment.center,
                              children: [
                                IconButton(
                                  iconSize: 150,
                                  color: global.theme.textSecondaryColor,
                                  tooltip: global.language('import_product'),
                                  icon: Icon(Icons.upload_file),
                                  onPressed: () {
                                    uploadFileExcel();
                                  },
                                ),
                                Text(
                                  global.language('import_product_from_excel'),
                                  style: TextStyle(
                                    fontSize: 20,
                                    color: global.theme.textSecondaryColor,
                                  ),
                                ),

                                /// download file excel ex sample stock balance
                                TextButton(
                                  onPressed: () {
                                    try {
                                      downloadAssetFile(
                                        'assets/file_import/import_product_final.xlsx',
                                        'import_product.xlsx',
                                      );
                                    } catch (e) {
                                      if (kDebugMode) {
                                        AppLogger.error(
                                          'An error occurred: //',
                                        );
                                      }
                                      // Handle the error or show a message to the user

                                      /// show dialog error
                                      showDialog(
                                        context: context,
                                        builder: (BuildContext context) {
                                          return AlertDialog(
                                            title: Text(
                                              global.language('error'),
                                            ),
                                            content: Text(e.toString()),
                                            actions: [
                                              TextButton(
                                                child: Text(
                                                  global.language('confirm'),
                                                ),
                                                onPressed: () {
                                                  Navigator.of(context).pop();
                                                },
                                              ),
                                            ],
                                          );
                                        },
                                      );
                                    }
                                  },
                                  child: Text(
                                    global.language('download_sample_excel'),
                                    style: TextStyle(fontSize: 20),
                                  ),
                                ),
                              ],
                            )
                          : Column(
                              children: [
                                Expanded(
                                  child: editProductListDetailWidget(
                                    screenHeight,
                                  ),
                                ),
                                menutListButtomBar(),
                              ],
                            ),
                    ),
              // Full-screen loading overlay when saving - with Progress
              if (isLoadingSaveTaskid)
                Container(
                  width: double.infinity,
                  height: double.infinity,
                  color: global.theme.cardColor.withValues(alpha: 0.95),
                  child: Center(
                    child: Container(
                      padding: const EdgeInsets.all(40),
                      margin: const EdgeInsets.symmetric(horizontal: 40),
                      decoration: BoxDecoration(
                        color: global.theme.cardColor,
                        borderRadius: BorderRadius.circular(16),
                        boxShadow: [
                          BoxShadow(
                            color: Colors.black.withValues(alpha: 0.1),
                            blurRadius: 20,
                            offset: const Offset(0, 4),
                          ),
                        ],
                      ),
                      child: Column(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          // Progress indicator with percentage
                          Stack(
                            alignment: Alignment.center,
                            children: [
                              SizedBox(
                                width: 80,
                                height: 80,
                                child: CircularProgressIndicator(
                                  strokeWidth: 6,
                                  value: currentProgress / 100.0,
                                  valueColor: AlwaysStoppedAnimation<Color>(
                                    global.theme.appBarColor,
                                  ),
                                  backgroundColor: global.theme.dividerBorderColor,
                                ),
                              ),
                              Text(
                                '$currentProgress%',
                                style: TextStyle(
                                  fontSize: 16,
                                  fontWeight: FontWeight.bold,
                                  color: global.theme.appBarColor,
                                ),
                              ),
                            ],
                          ),
                          SizedBox(height: 24),
                          // Main message
                          Text(
                            global.language('saving_data'),
                            style: TextStyle(
                              fontSize: 18,
                              fontWeight: FontWeight.w600,
                              color: global.theme.textColor,
                            ),
                          ),
                          SizedBox(height: 8),
                          // Status message
                          Text(
                            currentStatus != null
                                ? '${global.language('status_label')}: ${currentStatus!.status}'
                                : global.language('please_wait'),
                            style: TextStyle(
                              fontSize: 14,
                              color: global.theme.textSecondaryColor,
                            ),
                          ),
                          if (currentStatus != null) ...[
                            SizedBox(height: 16),
                            // Progress bar
                            Container(
                              width: 200,
                              height: 8,
                              decoration: BoxDecoration(
                                color: global.theme.dividerBorderColor,
                                borderRadius: BorderRadius.circular(4),
                              ),
                              child: FractionallySizedBox(
                                alignment: Alignment.centerLeft,
                                widthFactor: currentProgress / 100.0,
                                child: Container(
                                  decoration: BoxDecoration(
                                    color: global.theme.appBarColor,
                                    borderRadius: BorderRadius.circular(4),
                                  ),
                                ),
                              ),
                            ),
                          ],
                        ],
                      ),
                    ),
                  ),
                ),
            ],
          ),
        ),
      ),
    );
  }
}
