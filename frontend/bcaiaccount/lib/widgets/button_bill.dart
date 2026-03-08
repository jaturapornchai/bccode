import 'package:flutter/material.dart';

class ButtonBill extends StatelessWidget {
  final String label;
  final Function()? onPressed;
  final Color color;

  const ButtonBill({
    super.key,
    required this.label,
    required this.onPressed,
    required this.color,
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.all(8.0),
      child: ElevatedButton(
        style: ButtonStyle(
          backgroundColor: WidgetStateProperty.all(color),
          foregroundColor: WidgetStateProperty.all<Color>(Colors.white),
        ),
        onPressed: onPressed,
        child: Text(label),
      ),
    );
  }
}
