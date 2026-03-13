import 'package:smlaicloud/screens/Option/option_add.dart';
import 'package:flutter/material.dart';
import 'package:smlaicloud/screens/Option/components/body.dart';
import 'package:smlaicloud/components/appbar.dart';
import 'package:smlaicloud/global.dart' as global;

class OptionScreen extends StatelessWidget {
  const OptionScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: BaseAppBar(
        title: Text(global.language('supplement_options')),
        appBar: AppBar(),
        leading: IconButton(
          color: global.theme.textColor,
          icon: Icon(Icons.arrow_back),
          onPressed: () {},
        ),
        widgets: const <Widget>[],
      ),
      body: const Body(),
      floatingActionButton: FloatingActionButton(
        onPressed: () {
          Navigator.of(
            context,
          ).pushReplacement(MaterialPageRoute(builder: (_) => OptionAdd()));
        },
        backgroundColor: global.theme.infoHighlightTextColor,
        child: Icon(Icons.add),
      ),
    );
  }
}
