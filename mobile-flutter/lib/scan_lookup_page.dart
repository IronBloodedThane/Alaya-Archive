import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:mobile_scanner/mobile_scanner.dart';

import 'api_client.dart';
import 'auth_controller.dart';
import 'batch_review_page.dart';
import 'batch_scan_models.dart';
import 'batch_scanner_page.dart';

class ScanLookupPage extends StatefulWidget {
  const ScanLookupPage({super.key, required this.auth});
  final AuthController auth;

  @override
  State<ScanLookupPage> createState() => _ScanLookupPageState();
}

class _ScanLookupPageState extends State<ScanLookupPage> {
  late final ApiClient _client = ApiClient(widget.auth);

  String? _scannedIsbn;
  LookupResult? _result;
  String? _error;
  bool _busy = false;

  Future<void> _scanAndLookup() async {
    final isbn = await Navigator.of(context).push<String>(
      MaterialPageRoute(builder: (_) => const _ScannerPage()),
    );
    if (isbn == null || !mounted) return;

    setState(() {
      _scannedIsbn = isbn;
      _result = null;
      _error = null;
      _busy = true;
    });

    try {
      final result = await _client.lookupBookByIsbn(isbn);
      if (!mounted) return;
      setState(() {
        _result = result;
        _busy = false;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _error = _humanError(e);
        _busy = false;
      });
    }
  }

  String _humanError(Object e) {
    if (e is ApiException) {
      if (e.status == 401) return 'Session expired. Please sign in again.';
      if (e.status == 404) return 'No book found for that ISBN.';
      if (e.status == 502) return 'Lookup provider failed: ${e.message}';
      return 'API error ${e.status}: ${e.message}';
    }
    if (e is DioException) {
      return 'Network error: ${e.message ?? e.type.name}';
    }
    return 'Unexpected error: $e';
  }

  Future<void> _startBatchScan() async {
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
            onPressed: () =>
                Navigator.of(context).pop(controller.text.trim()),
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
        title: const Text('Alaya Archive — Lookup'),
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
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              FilledButton.icon(
                icon: const Icon(Icons.qr_code_scanner),
                label: const Text('Scan a book barcode'),
                onPressed: _busy ? null : _scanAndLookup,
              ),
              const SizedBox(height: 8),
              OutlinedButton.icon(
                icon: const Icon(Icons.library_add_outlined),
                label: const Text('Batch scan'),
                onPressed: _busy ? null : _startBatchScan,
              ),
              const SizedBox(height: 16),
              if (_scannedIsbn != null)
                Text(
                  'Scanned ISBN: $_scannedIsbn',
                  style: Theme.of(context).textTheme.bodyMedium,
                ),
              const SizedBox(height: 16),
              if (_busy) const Center(child: CircularProgressIndicator()),
              if (_error != null) _ErrorCard(message: _error!),
              if (_result != null) _ResultCard(result: _result!),
              if (_result == null && _error == null && !_busy)
                Text(
                  'Tap "Scan a book barcode" to begin.',
                  style: Theme.of(context).textTheme.bodySmall,
                ),
            ],
          ),
        ),
      ),
    );
  }
}

class _ScannerPage extends StatefulWidget {
  const _ScannerPage();

  @override
  State<_ScannerPage> createState() => _ScannerPageState();
}

class _ScannerPageState extends State<_ScannerPage> {
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

class _ResultCard extends StatelessWidget {
  const _ResultCard({required this.result});
  final LookupResult result;

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            if (result.coverImage != null)
              Padding(
                padding: const EdgeInsets.only(right: 16),
                child: Image.network(
                  result.coverImage!,
                  width: 80,
                  errorBuilder: (_, _, _) => const SizedBox(width: 80),
                ),
              ),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    result.title,
                    style: Theme.of(context).textTheme.titleMedium,
                  ),
                  if (result.subtitle != null && result.subtitle!.isNotEmpty)
                    Text(
                      result.subtitle!,
                      style: Theme.of(context).textTheme.bodyMedium,
                    ),
                  if (result.authors.isNotEmpty)
                    Padding(
                      padding: const EdgeInsets.only(top: 4),
                      child: Text('by ${result.authors.join(", ")}'),
                    ),
                  if (result.year != null || result.publisher != null)
                    Padding(
                      padding: const EdgeInsets.only(top: 4),
                      child: Text(
                        [
                          if (result.publisher != null) result.publisher,
                          if (result.year != null) result.year.toString(),
                        ].whereType<String>().join(' · '),
                      ),
                    ),
                  if (result.isbn13 != null)
                    Padding(
                      padding: const EdgeInsets.only(top: 4),
                      child: Text(
                        'ISBN-13: ${result.isbn13}',
                        style: Theme.of(context).textTheme.bodySmall,
                      ),
                    ),
                  Padding(
                    padding: const EdgeInsets.only(top: 8),
                    child: Text(
                      'Provider: ${result.provider}',
                      style: Theme.of(context).textTheme.bodySmall,
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _ErrorCard extends StatelessWidget {
  const _ErrorCard({required this.message});
  final String message;

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    return Card(
      color: scheme.errorContainer,
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Row(
          children: [
            Icon(Icons.error_outline, color: scheme.onErrorContainer),
            const SizedBox(width: 12),
            Expanded(
              child: Text(
                message,
                style: TextStyle(color: scheme.onErrorContainer),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
