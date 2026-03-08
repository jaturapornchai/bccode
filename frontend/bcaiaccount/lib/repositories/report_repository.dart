import 'package:flutter/foundation.dart';
import 'dart:convert';

import 'package:smlaicloud/environment.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/select_colums_csv_model.dart';
import 'package:intl/intl.dart';
import 'client.dart';
import 'package:dio/dio.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

class ReportRepository {
  Future<ApiResponse> getReportMovement(
    String barcode,
    String fromdate,
    String todate,
    String whcode,
    String lccode,
  ) async {
    Dio dio = Dio();
    final token = global.appConfig.getString("token");
    try {
      final response = await dio.get(
        '${Environment().config.reportApi}/movement?token=$token$barcode$fromdate$todate$whcode$lccode',
      );
      try {
        return ApiResponse.fromMap(response.data);
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      throw Exception(errorMessage);
    }
  }

  Future<ApiResponse> getReportProductBalance(String barcode) async {
    Dio dio = Dio();
    final token = global.appConfig.getString("token");
    try {
      final response = await dio.get(
        '${Environment().config.reportApi}/productbalance?token=$token$barcode',
      );
      try {
        return ApiResponse.fromMap(response.data);
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      throw Exception(errorMessage);
    }
  }

  Future<ApiResponse> activePos(
    String pincode,
    String devicenumber,
    int isdev,
    String apikey,
  ) async {
    Dio dio = Dio();
    final shopid = global.appConfig.getString("shopid");
    final token = global.appConfig.getString("refreshtoken") ?? "";
    final actoken = global.appConfig.getString("token");

    // Parse the JSON string to extract the email
    final userString = global.appConfig.getString("user") ?? '{}';
    final usernameData = json.decode(userString);
    final email = usernameData['email'];

    try {
      final response = await dio.get(
        '${Environment().config.reportApi}/poscenter/active?shopid=$shopid&pin=$pincode&token=$token&deviceid=$devicenumber&actoken=$actoken&isdev=$isdev&apikey=$apikey&username=$email',
      );
      try {
        return ApiResponse.fromMap(response.data);
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      throw Exception(errorMessage);
    }
  }

  /// delete pos
  Future<ApiResponse> deletePos(String pincode) async {
    Dio dio = Dio();
    final shopid = global.getShopId();

    try {
      final response = await dio.get(
        '${Environment().config.reportApi}/poscenter/delete?shopid=$shopid&pin=$pincode',
      );
      try {
        return ApiResponse.fromMap(response.data);
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      throw Exception(errorMessage);
    }
  }

  /// get apikey
  Future<ApiResponse> getApiKey(String pincode) async {
    Dio dio = Dio();
    final shopid = global.getShopId();

    try {
      final response = await dio.get(
        '${Environment().config.reportApi}/poscenter/getapikey?shopid=$shopid&pin=$pincode',
      );
      try {
        return ApiResponse.fromMap(response.data);
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      throw Exception(errorMessage);
    }
  }

  /// getReport by type
  Future<ApiResponse> getReport(
    global.ReportEnum type,
    String fromdate,
    String todate,
    int page,
    int perPage,
  ) async {
    Dio dio = Dio();
    final token = global.appConfig.getString("token");
    String urlapi = "";
    try {
      if (type == global.ReportEnum.salebydate) {
        urlapi =
            '/salebydate/sale?token=$token$fromdate$todate&page=$page&pagesize=$perPage';
      } else if (type == global.ReportEnum.receivemoney) {
        urlapi =
            '/salebydate/receivemoney?token=$token$fromdate$todate&page=$page&pagesize=$perPage';
      } else if (type == global.ReportEnum.saleinvoice) {
        urlapi =
            '/saleinvoice?token=$token$fromdate$todate&page=$page&pagesize=$perPage';
      }
      final response = await dio.get(
        '${Environment().config.reportApi}$urlapi',
      );
      try {
        return ApiResponse.fromMap(response.data);
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      throw Exception(errorMessage);
    }
  }

  /// view by type
  Future<String> getUriViewReport(
    global.ReportEnum type,
    String fromdate,
    String todate,
  ) async {
    final token = global.appConfig.getString("token");
    String urlapi = "";

    if (type == global.ReportEnum.salebydate) {
      urlapi =
          '${Environment().config.reportApi}/salebydate/sale/pdfview?token=$token$fromdate$todate';
    }

    return urlapi;
  }

  Future<ApiResponse> getPDFDownload(
    global.ReportEnum type,
    String fromdate,
    String todate,
    int showDetail,
    int showSumByDate,
    String search,
    String yearnum,
    String monthnum,
    String fromcustcode,
    String tocustcode,
    String branch,
    int iscancel,
    String fromsalecode,
    String tosalecode,
    String inquirytype,
    String ispos,
    String frombarcode,
    String tobarcode,
    String fromgroup,
    String togroup,
    int showcost,
    String barcode,
    String typefile,
    List<ListColumsCsvModel>? listcolumscsv,
  ) async {
    Dio dio = Dio();
    final token = global.appConfig.getString("token");
    final user = global.appConfig.getString("user");
    try {
      String urlapi = "";
      String queryFromdate = "";
      String queryTodate = "";
      String selectColumnsCsv = "";

      if (fromdate != '') {
        DateTime parsedDate = DateFormat("dd/MM/yyyy").parse(fromdate);
        String formattedDate = DateFormat("yyyy-MM-dd").format(parsedDate);
        queryFromdate = "&fromdate=$formattedDate";
      }

      if (todate != '') {
        DateTime parsedToDate = DateFormat("dd/MM/yyyy").parse(todate);
        String formattedToDate = DateFormat("yyyy-MM-dd").format(parsedToDate);

        queryTodate = "&todate=$formattedToDate";
      }

      if (search.isNotEmpty) {
        search = "&search=$search";
      }

      if (yearnum.isNotEmpty) {
        yearnum = "&year=$yearnum";
      }

      if (monthnum.isNotEmpty) {
        monthnum = "&month=$monthnum";
      }

      if (fromcustcode.isNotEmpty) {
        fromcustcode = "&fromcustcode=$fromcustcode";
      }

      if (tocustcode.isNotEmpty) {
        tocustcode = "&tocustcode=$tocustcode";
      }

      if (branch.isNotEmpty) {
        branch = "&branchcode=$branch";
      }

      if (fromsalecode.isNotEmpty) {
        fromsalecode = "&fromsalecode=$fromsalecode";
      }

      if (tosalecode.isNotEmpty) {
        tosalecode = "&tosalecode=$tosalecode";
      }

      if (inquirytype == "inquiry_all") {
        inquirytype = "";
      } else if (inquirytype == "inquiry_credit") {
        inquirytype = "&inquirytype=0";
      } else if (inquirytype == "inquiry_cash") {
        inquirytype = "&inquirytype=1";
      }

      if (ispos == "sale_all") {
        ispos = "";
      } else if (ispos == "sale_pos") {
        ispos = "&ispos=1";
      } else if (ispos == "sale_merchant") {
        ispos = "&ispos=0";
      }

      if (frombarcode.isNotEmpty) {
        frombarcode = "&frombarcode=$frombarcode";
      }

      if (tobarcode.isNotEmpty) {
        tobarcode = "&tobarcode=$tobarcode";
      }

      if (fromgroup.isNotEmpty) {
        fromgroup = "&fromgroup=$fromgroup";
      }

      if (togroup.isNotEmpty) {
        togroup = "&togroup=$togroup";
      }

      if (typefile == "pdf") {
        typefile = "&typefile=pdf";
      } else if (typefile == "excel") {
        typefile = "&typefile=excel";
      }

      /// columns csv
      if (listcolumscsv!.isNotEmpty) {
        selectColumnsCsv = "&columns=";
        for (var element in listcolumscsv) {
          /// last element not add comma
          if (listcolumscsv.indexOf(element) == listcolumscsv.length - 1) {
            // ignore: unnecessary_string_interpolations
            selectColumnsCsv += "${element.code}";
          } else {
            selectColumnsCsv += "${element.code},";
          }
        }
      }

      if (type == global.ReportEnum.salebydate) {
        urlapi =
            '/salebydate/genPDFSaleByDate?token=$token$queryFromdate$queryTodate$branch$inquirytype$ispos&printby=$user';
      } else if (type == global.ReportEnum.receivemoney) {
        urlapi =
            '/receivebydate/genPDFReceiveByDate?token=$token$queryFromdate$queryTodate$branch&printby=$user';
      } else if (type == global.ReportEnum.saleinvoice ||
          type == global.ReportEnum.saleinvoicedetail) {
        urlapi =
            '/saleinvoice/genPDFSale?token=$token$queryFromdate$queryTodate&showdetail=$showDetail&showsumbydate=$showSumByDate&iscancel=$iscancel&printby=$user$fromcustcode$tocustcode$branch$fromsalecode$tosalecode$inquirytype$ispos$typefile';
      } else if (type == global.ReportEnum.product) {
        urlapi = '/product/genPDFProduct?token=$token$search';
      } else if (type == global.ReportEnum.debtor) {
        urlapi = '/debtor/genPDFDebtor?token=$token$search';
      } else if (type == global.ReportEnum.creditor) {
        urlapi = '/creditor/genPDFCreditor?token=$token$search';
      } else if (type == global.ReportEnum.bookbank) {
        urlapi = '/bookbank/genPDFBookBank?token=$token$search';
      } else if (type == global.ReportEnum.purchase) {
        urlapi =
            '/purchase/genPDFPurchase?token=$token$queryFromdate$queryTodate&showdetail=$showDetail&showsumbydate=$showSumByDate&iscancel=$iscancel&printby=$user$fromcustcode$tocustcode$branch$fromsalecode$tosalecode$inquirytype';
      } else if (type == global.ReportEnum.purchasereturn) {
        urlapi =
            '/purchasereturn/genPDFPurchaseReturn?token=$token$queryFromdate$queryTodate$search';
      } else if (type == global.ReportEnum.saleinvoicereturn) {
        urlapi =
            '/saleinvoicereturn/genPDFSaleInvReturn?token=$token$queryFromdate$queryTodate$search';
      } else if (type == global.ReportEnum.transfer) {
        urlapi =
            '/transfer/genPDFTransfer?token=$token$queryFromdate$queryTodate$search';
      } else if (type == global.ReportEnum.receive) {
        urlapi =
            '/receive/genPDFReceive?token=$token$queryFromdate$queryTodate$search';
      } else if (type == global.ReportEnum.pickup) {
        urlapi =
            '/pickup/genPDFPickup?token=$token$queryFromdate$queryTodate$search';
      } else if (type == global.ReportEnum.returnproduct) {
        urlapi =
            '/returnproduct/genPDFReturnProduct?token=$token$queryFromdate$queryTodate$search';
      } else if (type == global.ReportEnum.stockadjustment) {
        urlapi =
            '/stockadjustment/genPDFStockAdjustment?token=$token$queryFromdate$queryTodate$search';
      } else if (type == global.ReportEnum.paid) {
        urlapi =
            '/paid/genPDFPaid?token=$token$queryFromdate$queryTodate&printby=$user&showsumbydate=$showSumByDate$fromcustcode$tocustcode$branch';
      } else if (type == global.ReportEnum.pay) {
        urlapi = '/pay/genPDFPay?token=$token$queryFromdate$queryTodate$search';
      } else if (type == global.ReportEnum.getpaid) {
        urlapi =
            '/getpaid/genPDFGetPaid?token=$token$queryFromdate$queryTodate$search';
      } else if (type == global.ReportEnum.getpay) {
        urlapi =
            '/getpay/genPDFGetPay?token=$token$queryFromdate$queryTodate$search';
      } else if (type == global.ReportEnum.salebydebtor) {
        urlapi =
            '/salebydebtor/genPDFSaleByDebtor?token=$token$queryFromdate$queryTodate$search';
      } else if (type == global.ReportEnum.vatsale) {
        urlapi = '/vatsale/genPDFVatSale?token=$token$yearnum$monthnum';
      } else if (type == global.ReportEnum.vatpurchase) {
        urlapi = '/vatpurchase/genPDFVatPurchase?token=$token$yearnum$monthnum';
      } else if (type == global.ReportEnum.salebyproduct) {
        urlapi =
            '/salebyproduct/genPDFsalebyproduct?token=$token$queryFromdate$queryTodate$branch$frombarcode$tobarcode$fromgroup$togroup&printby=$user';
      } else if (type == global.ReportEnum.productmovement) {
        urlapi =
            '/movement/genPDFMovement?token=$token$queryFromdate$queryTodate&barcode=$barcode&printby=$user';
      } else if (type == global.ReportEnum.stockbalance) {
        urlapi =
            '/stockbalance/genPDFStockBalance?token=$token$queryTodate&showcost=$showcost$frombarcode$tobarcode&printby=$user';
      } else if (type == global.ReportEnum.stockcard) {
        urlapi =
            '/stockcard/genPDFStockCard?token=$token$queryFromdate$queryTodate&barcode=$barcode&printby=$user';
      } else if (type == global.ReportEnum.csvsaledetail) {
        urlapi =
            '/exportdatacsv/sale_detail/genData?token=$token$queryFromdate$queryTodate$selectColumnsCsv';
      }
      final response = await dio.get(
        '${Environment().config.reportApi}$urlapi',
      );
      try {
        return ApiResponse.fromMap(response.data);
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      throw Exception(errorMessage);
    }
  }

  Future<ApiResponse> checkFileStatus(
    global.ReportEnum type,
    String url,
    String jobId,
  ) async {
    Dio dio = Dio();
    try {
      String urlapi = "";

      if (type == global.ReportEnum.salebydate) {
        urlapi = '/salebydate/check-salebydate/$jobId/${url.split('/').last}';
      } else if (type == global.ReportEnum.receivemoney) {
        urlapi =
            '/receivebydate/check-receivebydate/$jobId/${url.split('/').last}';
      } else if (type == global.ReportEnum.saleinvoice ||
          type == global.ReportEnum.saleinvoicedetail) {
        urlapi = '/saleinvoice/check-saleinv/$jobId/${url.split('/').last}';
      } else if (type == global.ReportEnum.product) {
        urlapi = '/product/check-product/$jobId/${url.split('/').last}';
      } else if (type == global.ReportEnum.debtor) {
        urlapi = '/debtor/check-debtor/$jobId/${url.split('/').last}';
      } else if (type == global.ReportEnum.creditor) {
        urlapi = '/creditor/check-creditor/$jobId/${url.split('/').last}';
      } else if (type == global.ReportEnum.bookbank) {
        urlapi = '/bookbank/check-bookbank/$jobId/${url.split('/').last}';
      } else if (type == global.ReportEnum.purchase) {
        urlapi = '/purchase/check-purchase/$jobId/${url.split('/').last}';
      } else if (type == global.ReportEnum.purchasereturn) {
        urlapi =
            '/purchasereturn/check-purchase-return/$jobId/${url.split('/').last}';
      } else if (type == global.ReportEnum.saleinvoicereturn) {
        urlapi =
            '/saleinvoicereturn/check-saleinv-return/$jobId/${url.split('/').last}';
      } else if (type == global.ReportEnum.transfer) {
        urlapi = '/transfer/check-transfer/$jobId/${url.split('/').last}';
      } else if (type == global.ReportEnum.receive) {
        urlapi = '/receive/check-receive/$jobId/${url.split('/').last}';
      } else if (type == global.ReportEnum.pickup) {
        urlapi = '/pickup/check-pickup/$jobId/${url.split('/').last}';
      } else if (type == global.ReportEnum.returnproduct) {
        urlapi =
            '/returnproduct/check-return-product/$jobId/${url.split('/').last}';
      } else if (type == global.ReportEnum.stockadjustment) {
        urlapi =
            '/stockadjustment/check-stock-adjustment/$jobId/${url.split('/').last}';
      } else if (type == global.ReportEnum.paid) {
        urlapi = '/paid/check-paid/$jobId/${url.split('/').last}';
      } else if (type == global.ReportEnum.pay) {
        urlapi = '/pay/check-pay/$jobId/${url.split('/').last}';
      } else if (type == global.ReportEnum.getpaid) {
        urlapi = '/getpaid/check-get-paid/$jobId/${url.split('/').last}';
      } else if (type == global.ReportEnum.getpay) {
        urlapi = '/getpay/check-get-pay/$jobId/${url.split('/').last}';
      } else if (type == global.ReportEnum.salebydebtor) {
        urlapi =
            '/salebydebtor/check-sale-by-debtor/$jobId/${url.split('/').last}';
      } else if (type == global.ReportEnum.vatsale) {
        urlapi = '/vatsale/check-vat-sale/$jobId/${url.split('/').last}';
      } else if (type == global.ReportEnum.vatpurchase) {
        urlapi =
            '/vatpurchase/check-vat-purchase/$jobId/${url.split('/').last}';
      } else if (type == global.ReportEnum.salebyproduct) {
        urlapi =
            '/salebyproduct/check-salebyproduct/$jobId/${url.split('/').last}';
      } else if (type == global.ReportEnum.productmovement) {
        urlapi = '/movement/check-movement/$jobId/${url.split('/').last}';
      } else if (type == global.ReportEnum.stockbalance) {
        urlapi =
            '/stockbalance/check-stockbalance/$jobId/${url.split('/').last}';
      } else if (type == global.ReportEnum.stockcard) {
        urlapi = '/stockcard/check-stockcard/$jobId/${url.split('/').last}';
      } else if (type == global.ReportEnum.csvsaledetail) {
        urlapi =
            '/exportdatacsv/sale_detail/check-sale-detail/$jobId/${url.split('/').last}';
        if (kDebugMode) {
          AppLogger.debug(urlapi);
        }
      }
      final response = await dio.get(
        '${Environment().config.reportApi}$urlapi',
      );
      try {
        return ApiResponse.fromMap(response.data);
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      throw Exception(errorMessage);
    }
  }

  /// ดึงข้อมูล JSON จาก report service
  /// ใช้สำหรับรายงานที่มีการ pagination
  ///
  /// Parameters:
  /// - [guid]: GUID ของรายงาน
  /// - [offset]: เริ่มต้นที่แถวที่เท่าไร
  /// - [limit]: จำนวนแถวที่ต้องการดึง
  ///
  /// Returns: Map ที่มี status, data (List of JSON objects), และ count
  /// ถ้าไม่มีข้อมูลแล้ว (404) จะคืน Map ที่มี data เป็น empty list
  static Future<Map<String, dynamic>> fetchReportData({
    required String guid,
    required int offset,
    required int limit,
  }) async {
    try {
      var payload = {
        "shop_id": global.getShopId(),
        "command_id": "data_json",
        "guid": guid,
        "offset": offset.toString(),
        "limit": limit.toString(),
      };

      var jsonResult = await global.reportServiceGet(payload);

      // Check for API-level errors
      if (jsonResult['status'] == 'error') {
        // ถ้าเป็น 404 Not Found (ไม่มีข้อมูลแล้ว) คืน empty data
        if (jsonResult['code'] == 404 ||
            jsonResult['message']?.toString().contains('Not Found') == true ||
            jsonResult['message']?.toString().contains('No data found') ==
                true) {
          if (kDebugMode) {
            AppLogger.debug("ℹ️ No more data at offset $offset");
          }
          return {
            'status': 'success',
            'code': 200,
            'data': [],
            'count': 0,
            'message': 'No more data',
          };
        }
        throw Exception('API Error: ${jsonResult['message']}');
      }

      return jsonResult;
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error("❌ Error fetching report data: $e");
      }
      rethrow;
    }
  }

  /// Generic helper function สำหรับ load ข้อมูลรายงานแบบ pagination
  /// ใช้ได้กับทุกประเภทรายงานที่มี fromJson method
  ///
  /// Type Parameters:
  /// - [T]: ประเภทของ Model ที่ต้องการแปลง (ต้องมี fromJson method)
  ///
  /// Parameters:
  /// - [guid]: GUID ของรายงาน
  /// - [offset]: เริ่มต้นที่แถวที่เท่าไร
  /// - [limit]: จำนวนแถวที่ต้องการดึง
  /// - [dataList]: List ที่จะเพิ่มข้อมูลเข้าไป
  /// - [fromJson]: Function ที่แปลง JSON เป็น Model
  ///
  /// Returns: จำนวนข้อมูลที่ได้รับ (0 ถ้าไม่มีข้อมูลแล้ว)
  static Future<int> loadReportData<T>({
    required String guid,
    required int offset,
    required int limit,
    required List<T> dataList,
    required T Function(Map<String, dynamic>) fromJson,
  }) async {
    try {
      var jsonResult = await fetchReportData(
        guid: guid,
        offset: offset,
        limit: limit,
      );

      int count = 0;

      // Process the data - data is already JSON objects array
      if (jsonResult['data'] != null && jsonResult['data'] is List) {
        for (var item in jsonResult['data']) {
          try {
            // item is already a JSON object (Map<String, dynamic>)
            dataList.add(fromJson(item));
            count++;
          } catch (e) {
            if (kDebugMode) {
              AppLogger.error("Error parsing item: $e");
              AppLogger.debug("Item data: $item");
            }
          }
        }
      }

      return count;
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error("❌ Error loading report data: $e");
      }
      rethrow;
    }
  }
}
