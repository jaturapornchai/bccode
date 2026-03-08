import 'dart:convert';

import 'package:smlaicloud/api/client.dart';
import 'package:smlaicloud/model/order_type_model.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  String jsonStr =
      "{ \"success\": true, \"data\": [{ \"code\" : \"test\", \"names\" : [{\"code\" : \"th\", \"name\":\"name in thai\"}], \"remark1\": \"\" }]}";

  test('Test deserializing json', () {
    final json = jsonDecode(jsonStr);
    ApiResponse response = ApiResponse.fromMap(json);

    List<OrderTypeModel> datas = (response.data as List)
        .map((ordertype) => OrderTypeModel.fromJson(ordertype))
        .toList();
    expect(datas.isNotEmpty, true);
    expect(datas[0].remarks, isNot(null));
    expect(datas[0].names.length, 1);
    expect(datas[0].remarks!.length, 0);
    //expect(model.remarks, '123');
  });
}
