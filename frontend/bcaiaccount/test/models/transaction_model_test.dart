import 'dart:convert';
import 'package:flutter_test/flutter_test.dart';
import 'package:smlaicloud/model/transaction_model.dart';

void main() {
  group('TransactionModel', () {
    test('fromJson → toJson roundtrip preserves key fields', () {
      final json = {
        'guidref': 'abc-123',
        'docno': 'PO20260107001',
        'docdatetime': '2026-01-07T06:00:00.000Z',
        'docdatelocal': '20260107',
        'doctimelocal': '13:00:00',
        'timezone': 'UTC+07:00',
        'docrefno': '',
        'docrefdate': '',
        'docreftype': 0,
        'doctype': 0,
        'vattype': 1,
        'inquirytype': 0,
        'transflag': 12,
        'custcode': 'CUST001',
        'salecode': 'SALE01',
        'salename': 'Sales Person',
        'discountword': '',
        'totalcost': 0.0,
        'totalvalue': 1000.0,
        'totaldiscount': 50.0,
        'totalvatvalue': 66.5,
        'totalbeforevat': 950.0,
        'totalaftervat': 1016.5,
        'totalexceptvat': 0.0,
        'totalamount': 1016.5,
        'cashiercode': 'CASH01',
        'posid': 'POS01',
        'membercode': '',
        'vatrate': 7.0,
        'status': 0,
        'iscancel': false,
        'ismanualamount': false,
        'taxdocno': '',
        'taxdocdate': '',
        'creator_code': 'admin',
        'creator_name': 'Admin User',
      };

      final model = TransactionModel.fromJson(json);
      expect(model.guidref, 'abc-123');
      expect(model.docno, 'PO20260107001');
      expect(model.docdatelocal, '20260107');
      expect(model.doctimelocal, '13:00:00');
      expect(model.timezone, 'UTC+07:00');
      expect(model.transflag, 12);
      expect(model.totalamount, 1016.5);
      expect(model.creatorcode, 'admin');
      expect(model.creatorname, 'Admin User');

      final output = model.toJson();
      expect(output['guidref'], 'abc-123');
      expect(output['docdatelocal'], '20260107');
      expect(output['totalamount'], 1016.5);
      expect(output['creator_code'], 'admin');
    });

    test('fromJson handles null optional fields', () {
      final json = {
        'guidref': '',
        'docno': '',
        'docdatetime': '',
        'docrefno': '',
        'docrefdate': '',
        'docreftype': 0,
        'doctype': 0,
        'vattype': 0,
        'inquirytype': 0,
        'transflag': 0,
        'custcode': '',
        'salecode': '',
        'salename': '',
        'discountword': '',
        'totalcost': 0.0,
        'totalvalue': 0.0,
        'totaldiscount': 0.0,
        'totalvatvalue': 0.0,
        'totalbeforevat': 0.0,
        'totalaftervat': 0.0,
        'totalexceptvat': 0.0,
        'totalamount': 0.0,
        'cashiercode': '',
        'posid': '',
        'membercode': '',
        'vatrate': 0.0,
        'status': 0,
        'iscancel': false,
        'ismanualamount': false,
        'taxdocno': '',
        'taxdocdate': '',
        // optional fields omitted
      };

      final model = TransactionModel.fromJson(json);
      expect(model.docdatelocal, isNull);
      expect(model.doctimelocal, isNull);
      expect(model.timezone, isNull);
      expect(model.details, isNotNull); // defaults to empty list
    });

    test('toJson output can be re-parsed (JSON encode/decode roundtrip)', () {
      final json = {
        'guidref': 'test-guid',
        'docno': 'INV001',
        'docdatetime': '2026-02-01T10:00:00.000Z',
        'docrefno': 'REF001',
        'docrefdate': '2026-01-30',
        'docreftype': 1,
        'doctype': 2,
        'vattype': 1,
        'inquirytype': 0,
        'transflag': 12,
        'custcode': 'C001',
        'salecode': 'S001',
        'salename': 'Test',
        'discountword': '10%',
        'totalcost': 900.0,
        'totalvalue': 1000.0,
        'totaldiscount': 100.0,
        'totalvatvalue': 63.0,
        'totalbeforevat': 900.0,
        'totalaftervat': 963.0,
        'totalexceptvat': 0.0,
        'totalamount': 963.0,
        'cashiercode': 'CASH',
        'posid': 'POS',
        'membercode': 'M001',
        'vatrate': 7.0,
        'status': 1,
        'iscancel': false,
        'ismanualamount': false,
        'taxdocno': 'TAX001',
        'taxdocdate': '2026-02-01',
      };

      final model = TransactionModel.fromJson(json);
      final jsonString = jsonEncode(model.toJson());
      final reparsed = TransactionModel.fromJson(jsonDecode(jsonString));

      expect(reparsed.guidref, model.guidref);
      expect(reparsed.docno, model.docno);
      expect(reparsed.totalamount, model.totalamount);
      expect(reparsed.discountword, model.discountword);
      expect(reparsed.vatrate, model.vatrate);
    });
  });
}
