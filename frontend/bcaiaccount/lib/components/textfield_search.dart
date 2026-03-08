import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;

class TextFieldSearch extends StatelessWidget {
  final void Function(String)? onChange;
  const TextFieldSearch({super.key, required this.onChange});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: EdgeInsets.symmetric(horizontal: 20, vertical: 5),
      child: TextField(
        onChanged: onChange,
        decoration: InputDecoration(
          icon: Icon(Icons.search, color: Color(0xff046AF3)),
          hintText: global.language("search"),
          border: InputBorder.none,
        ),
      ),
    );
  }
}
