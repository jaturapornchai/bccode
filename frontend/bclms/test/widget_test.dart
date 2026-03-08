import 'package:flutter_test/flutter_test.dart';
import 'package:bclms/main.dart';

void main() {
  testWidgets('แอปเริ่มต้นได้', (WidgetTester tester) async {
    await tester.pumpWidget(const BclmsApp());
    await tester.pumpAndSettle();
    expect(find.text('BCTMS'), findsOneWidget);
  });
}
