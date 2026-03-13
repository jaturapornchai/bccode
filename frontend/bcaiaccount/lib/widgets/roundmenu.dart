import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;

class RoundMenu extends StatelessWidget {
  final String label;
  final Color fontColor;
  final int index;
  final int actived;
  final String img;
  final Function callBack;

  const RoundMenu({
    super.key,
    required this.label,
    required this.index,
    required this.actived,
    this.img = "setting.svg",
    required this.callBack,
    required this.fontColor,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 100,
      height: 100,
      color: global.theme.infoHighlightTextColor,
      child: Material(
        type: MaterialType.transparency,
        child: InkWell(
          onTap: () {
            callBack.call();
          },
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: <Widget>[
              Flexible(
                child: Text(label, style: TextStyle(color: fontColor)),
              ),
            ],
          ),
          /*        style: TextButton.styleFrom(
          backgroundColor: index == actived
              ? global.posTheme.secondary
              : global.posTheme.background,
        ),*/
        ),
      ),
    );
  }
}
