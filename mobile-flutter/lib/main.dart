import 'package:flutter/material.dart';

import 'auth_controller.dart';
import 'login_page.dart';
import 'scan_lookup_page.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  final auth = AuthController();
  await auth.bootstrap();
  runApp(AlayaArchiveApp(auth: auth));
}

class AlayaArchiveApp extends StatelessWidget {
  const AlayaArchiveApp({super.key, required this.auth});
  final AuthController auth;

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Alaya Archive',
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: Colors.deepPurple),
        useMaterial3: true,
      ),
      home: ListenableBuilder(
        listenable: auth,
        builder: (_, _) => auth.isAuthenticated
            ? ScanLookupPage(auth: auth)
            : LoginPage(auth: auth),
      ),
    );
  }
}
