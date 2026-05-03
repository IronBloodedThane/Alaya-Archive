import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:alaya_archive/auth_controller.dart';
import 'package:alaya_archive/main.dart';

void main() {
  testWidgets('unauthenticated boot lands on the sign-in screen',
      (tester) async {
    final auth = AuthController(storage: InMemoryTokenStorage());
    await auth.bootstrap();
    await tester.pumpWidget(AlayaArchiveApp(auth: auth));
    await tester.pump();

    expect(find.text('Sign in'), findsWidgets);
    expect(find.text('Email or username'), findsOneWidget);
    expect(find.text('Password'), findsOneWidget);
    expect(find.text('API base URL'), findsOneWidget);
  });

  testWidgets('authenticated boot lands on the scan screen', (tester) async {
    final storage = InMemoryTokenStorage();
    await storage.write('accessToken', 'fake-access');
    await storage.write('refreshToken', 'fake-refresh');
    final auth = AuthController(storage: storage);
    await auth.bootstrap();

    await tester.pumpWidget(AlayaArchiveApp(auth: auth));
    await tester.pump();

    expect(find.text('Scan a book barcode'), findsOneWidget);
    expect(find.text('Batch scan'), findsOneWidget);
    expect(find.byIcon(Icons.logout), findsOneWidget);
  });
}
