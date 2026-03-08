import 'package:flutter/material.dart';
import 'package:smlaicloud/screens/purchase/components/body.dart';
import 'package:smlaicloud/screens/purchase/purchase_add.dart';
import 'package:smlaicloud/components/appbar.dart';
import 'package:smlaicloud/global.dart' as global;

class PurchaseScreen extends StatelessWidget {
  const PurchaseScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: BaseAppBar(
        title: Text(global.language('purchase_product')),
        appBar: AppBar(),
        widgets: const <Widget>[],
      ),
      body: const Body(),
      floatingActionButton: FloatingActionButton(
        onPressed: () {
          Navigator.push(
            context,
            MaterialPageRoute(
              builder: (context) {
                return const PurchaseAdd();
              },
            ),
          );
        },
        backgroundColor: Colors.blue,
        child: const Icon(Icons.add),
      ),
    );
  }
}
