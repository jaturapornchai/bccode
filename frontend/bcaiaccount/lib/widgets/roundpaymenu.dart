import 'package:flutter/material.dart';

class RoundPayMenu extends StatelessWidget {
  final String label;

  final Function()? onPressed;

  final int actived;
  final String img;
  const RoundPayMenu({
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
        style: TextButton.styleFrom(backgroundColor: Colors.white),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.start,
          children: <Widget>[
            Text(
              label,
              style: const TextStyle(color: Colors.white, fontSize: 15),
            ),
          ],
        ),
      ),
    );
  }
}
