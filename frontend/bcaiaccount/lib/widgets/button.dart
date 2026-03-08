import 'package:flutter/material.dart';

class NumPadButton extends StatelessWidget {
  final String? text;
  final IconData? icon;
  final Function callBack;
  final Color? color;
  final Color? textAndIconColor;
  final double margin;

  const NumPadButton({
    super.key,
    this.text,
    this.icon,
    required this.callBack,
    this.color,
    this.margin = 0,
    this.textAndIconColor,
  });

  @override
  Widget build(BuildContext context) {
    Widget label = icon != null
        ? FittedBox(
            fit: BoxFit.fill,
            child: Icon(
              icon,
              shadows: const <Shadow>[
                Shadow(
                  blurRadius: 1.0,
                  color: Colors.black,
                  offset: Offset(2.0, 2.0),
                ),
              ],
              color: (textAndIconColor == null)
                  ? Colors.white
                  : textAndIconColor,
            ),
          )
        : Text(
            text ?? "",
            style: TextStyle(
              fontSize: 80,
              fontWeight: FontWeight.bold,
              shadows: const [
                Shadow(
                  blurRadius: 1.0,
                  color: Colors.black,
                  offset: Offset(2.0, 2.0),
                ),
              ],
              color: ((textAndIconColor == null)
                  ? Colors.white
                  : textAndIconColor),
            ),
          );
    ElevatedButton button = ElevatedButton(
      style: ElevatedButton.styleFrom(
        backgroundColor: (color == null) ? Colors.blue : color,
        minimumSize: Size.zero,
      ),
      onPressed: () {
        callBack.call();
      },
      child: FittedBox(fit: BoxFit.scaleDown, child: label),
    );
    return (margin == 0)
        ? button
        : Padding(padding: EdgeInsets.all(margin), child: button);
  }
}

class CommandButton extends StatelessWidget {
  final String label;
  final IconData? icon;
  final bool haveBorder;
  final VoidCallback onPressed;
  final Color? labelColor;
  final Color? primaryColor;
  final double? height;
  final double width;
  final Color? iconColor;
  final String imgAssetPath;
  final String imgNetworkPath;

  const CommandButton({
    super.key,
    this.label = "",
    this.icon,
    this.haveBorder = true,
    required this.onPressed,
    this.labelColor = Colors.black,
    this.primaryColor,
    this.height = 50,
    this.width = 100,
    this.imgAssetPath = "",
    this.imgNetworkPath = "",
    this.iconColor = Colors.black,
  });

  @override
  Widget build(BuildContext context) {
    TextStyle buttonStyle = TextStyle(
      fontSize: 20.0,
      fontWeight: FontWeight.bold,
      color: (labelColor == null)
          ? Theme.of(context).scaffoldBackgroundColor
          : labelColor,
    );
    Widget labelAndStyle = Text(
      label,
      style: buttonStyle,
      textAlign: TextAlign.center,
      overflow: TextOverflow.clip,
    );
    bool useImage = (imgAssetPath.isNotEmpty || imgNetworkPath.isNotEmpty);

    return Container(
      width: width,
      height: height,
      padding: const EdgeInsets.all(4),
      decoration: BoxDecoration(
        color: primaryColor,
        border: Border.all(color: Colors.grey, width: 1),
        borderRadius: BorderRadius.circular(4),
      ),
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: () {
            onPressed();
          },
          child: (useImage == false)
              ? Center(child: labelAndStyle)
              : (label.isEmpty)
              ? FittedBox(
                  fit: BoxFit.fill,
                  child: (imgAssetPath.isNotEmpty)
                      ? Image.asset(imgAssetPath)
                      : Image.network(imgNetworkPath),
                )
              : Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: <Widget>[
                    Expanded(
                      child: (imgAssetPath.isNotEmpty)
                          ? Image.asset(imgAssetPath)
                          : Image.network(imgNetworkPath),
                    ),
                    FittedBox(fit: BoxFit.fitWidth, child: labelAndStyle),
                  ],
                ),
        ),
      ),
    );
  }
}
