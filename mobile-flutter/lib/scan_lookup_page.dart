import 'package:flutter/material.dart';
import 'package:mobile_scanner/mobile_scanner.dart';

import 'api_client.dart';
import 'auth_controller.dart';
import 'batch_review_page.dart';
import 'batch_scan_models.dart';
import 'batch_scan_service.dart';
import 'batch_scanner_page.dart';

// ScanLookupPage is the home screen for authenticated users. Two entry
// points to the same downstream UX:
//
//   1. "Scan a book barcode"  → one-shot scanner returns one ISBN; we
//      build a single ScannedItem and route to BatchReviewPage with it.
//   2. "Batch scan"           → continuous scanner accumulates many.
//
// Both flows hand off to BatchReviewPage so saving, duplicate handling,
// and list-type selection happen in exactly one place.
class ScanLookupPage extends StatefulWidget {
  const ScanLookupPage({super.key, required this.auth});
  final AuthController auth;

  @override
  State<ScanLookupPage> createState() => _ScanLookupPageState();
}

class _ScanLookupPageState extends State<ScanLookupPage> {
  late final ApiClient _client = ApiClient(widget.auth);

  Future<void> _scanSingle() async {
    final isbn = await Navigator.of(context).push<String>(
      MaterialPageRoute(builder: (_) => const _SingleScannerPage()),
    );
    if (isbn == null || !mounted) return;

    final item = ScannedItem(isbn: isbn, mediaType: 'book');
    // Fire and forget — the review screen listens to the item and
    // updates as lookup/check resolve.
    lookupAndCheckIsbn(_client, item);

    await Navigator.of(context).push(
      MaterialPageRoute(
        builder: (_) => BatchReviewPage(client: _client, items: [item]),
      ),
    );
  }

  Future<void> _scanBatch() async {
    final items = await Navigator.of(context).push<List<ScannedItem>>(
      MaterialPageRoute(builder: (_) => BatchScannerPage(client: _client)),
    );
    if (items == null || items.isEmpty || !mounted) return;
    await Navigator.of(context).push(
      MaterialPageRoute(
        builder: (_) => BatchReviewPage(client: _client, items: items),
      ),
    );
  }

  Future<void> _editBaseUrl() async {
    final controller = TextEditingController(text: widget.auth.baseUrl);
    final newUrl = await showDialog<String>(
      context: context,
      builder: (_) => AlertDialog(
        title: const Text('API base URL'),
        content: TextField(
          controller: controller,
          decoration: const InputDecoration(
            helperText: 'e.g. http://10.0.1.18:8080',
          ),
          keyboardType: TextInputType.url,
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () => Navigator.of(context).pop(controller.text.trim()),
            child: const Text('Save'),
          ),
        ],
      ),
    );
    controller.dispose();
    if (newUrl != null && newUrl.isNotEmpty) {
      await widget.auth.setBaseUrl(newUrl);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Alaya Archive'),
        actions: [
          IconButton(
            icon: const Icon(Icons.dns_outlined),
            tooltip: 'Change base URL',
            onPressed: _editBaseUrl,
          ),
          IconButton(
            icon: const Icon(Icons.logout),
            tooltip: 'Sign out',
            onPressed: () => widget.auth.logout(),
          ),
        ],
      ),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              FilledButton.icon(
                icon: const Icon(Icons.qr_code_scanner),
                label: const Text('Scan a book barcode'),
                style: FilledButton.styleFrom(
                  minimumSize: const Size.fromHeight(56),
                ),
                onPressed: _scanSingle,
              ),
              const SizedBox(height: 12),
              OutlinedButton.icon(
                icon: const Icon(Icons.library_add_outlined),
                label: const Text('Batch scan'),
                style: OutlinedButton.styleFrom(
                  minimumSize: const Size.fromHeight(56),
                ),
                onPressed: _scanBatch,
              ),
              const SizedBox(height: 24),
              Text(
                'Scan a book to look it up, edit details, and add it to '
                'your collection or wishlist.',
                style: Theme.of(context).textTheme.bodySmall,
              ),
            ],
          ),
        ),
      ),
    );
  }
}

// _SingleScannerPage is the one-shot scanner: pops the first valid ISBN-13
// it sees and exits. Same camera setup as BatchScannerPage but stops on
// the first hit instead of accumulating.
class _SingleScannerPage extends StatefulWidget {
  const _SingleScannerPage();

  @override
  State<_SingleScannerPage> createState() => _SingleScannerPageState();
}

class _SingleScannerPageState extends State<_SingleScannerPage> {
  final _controller = MobileScannerController(
    formats: const [BarcodeFormat.ean13],
  );
  bool _handled = false;

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  void _onDetect(BarcodeCapture capture) {
    if (_handled) return;
    for (final b in capture.barcodes) {
      final raw = b.rawValue;
      if (raw == null) continue;
      // Book ISBN-13s start with 978 or 979.
      if (raw.length == 13 &&
          (raw.startsWith('978') || raw.startsWith('979'))) {
        _handled = true;
        Navigator.of(context).pop(raw);
        return;
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Scan ISBN')),
      body: MobileScanner(controller: _controller, onDetect: _onDetect),
    );
  }
}
