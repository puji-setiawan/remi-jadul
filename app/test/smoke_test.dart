import 'package:flutter_test/flutter_test.dart';
import 'package:remi_jadul/main.dart';

void main() {
  testWidgets('renders the app shell', (tester) async {
    await tester.pumpWidget(const RemiJadulApp());

    expect(find.text('Remi Jadul'), findsOneWidget);
  });
}
