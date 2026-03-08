import 'package:flutter/material.dart';
import 'package:smlaicloud/screens/Product/components/body.dart';
import 'package:smlaicloud/screens/Product/product_add.dart';
import 'package:smlaicloud/components/appbar.dart';
import 'package:smlaicloud/global.dart' as global;

class ProductScreen extends StatefulWidget {
  const ProductScreen({super.key});

  @override
  State<ProductScreen> createState() => _ProductScreenState();
}

class _ProductScreenState extends State<ProductScreen> {
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: BaseAppBar(
        title: Text(global.language('product_registry')),
        appBar: AppBar(),
        leading: IconButton(
          color: Colors.black,
          icon: const Icon(Icons.arrow_back),
          onPressed: () {},
        ),
        widgets: const <Widget>[],
      ),
      body: const Body(),
      floatingActionButton: FloatingActionButton(
        onPressed: () {
          Navigator.pushReplacement(
            context,
            MaterialPageRoute(
              builder: (context) {
                return const ProductAdd();
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
