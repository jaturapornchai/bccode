import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;

class RoundPayMenu extends StatelessWidget {
  final String label;

  final Function()? onPressed;

  final int actived;
  final String img;
  RoundPayMenu({
    super.key,
    required this.label,
    this.onPressed,
    required this.actived,
    this.img = "setting.svg",
  });

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      width: double.infinity,
      child: TextButton(
        onPressed: onPressed,
        style: TextButton.styleFrom(backgroundColor: global.theme.cardColor),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.start,
          children: <Widget>[
            Text(
              label,
              style: TextStyle(color: global.theme.onPrimaryColor, fontSize: 15),
            ),
          ],
        ),
      ),
    );
  }
}
