import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:mobile_scanner/mobile_scanner.dart';

import 'api_client.dart';
import 'batch_scan_models.dart';

// BatchScannerPage keeps the camera open and accumulates ISBNs as they are
// scanned. The same ISBN is only added once per session. Each new ISBN
// kicks off a background lookup so the list is ready by the time the user
// hits "Done".
//
// Pops back with the list of ScannedItem when the user is done.
class BatchScannerPage extends StatefulWidget {
  const BatchScannerPage({super.key, required this.client});
  final ApiClient client;

  @override
  State<BatchScannerPage> createState() => _BatchScannerPageState();
}

class _BatchScannerPageState extends State<BatchScannerPage> {
  final _controller = MobileScannerController(
    formats: const [BarcodeFormat.ean13],
  );
  final _items = <ScannedItem>[];
  final _seen = <String>{};
  ListType _listType = ListType.owned;

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  void _onDetect(BarcodeCapture capture) {
    for (final b in capture.barcodes) {
      final raw = b.rawValue;
      if (raw == null) continue;
      // Only ISBN-13 (book/manga) for now. 978/979 prefix.
      if (raw.length != 13) continue;
      if (!raw.startsWith('978') && !raw.startsWith('979')) continue;
      if (!_seen.add(raw)) continue;

      final item = ScannedItem(
        isbn: raw,
        mediaType: 'book',
        listType: _listType,
      );
      setState(() => _items.add(item));
      HapticFeedback.lightImpact();
      _lookupAndCheck(item);
    }
  }

  // Fires the lookup, then the duplicate check. Both errors are non-fatal —
  // they just leave the item in a state the review page surfaces to the user.
  Future<void> _lookupAndCheck(ScannedItem item) async {
    try {
      final result = await widget.client.lookupBookByIsbn(item.isbn);
      item.applyLookup(result);
    } catch (e) {
      item.markLookupFailed(_humanError(e));
      return;
    }
    try {
      final dupes = await widget.client.checkDuplicate(
        mediaType: item.mediaType,
        isbn: item.isbn,
      );
      item.applyDuplicates(dupes);
    } catch (_) {
      // Non-fatal: review page can still save; backend's on_duplicate=error
      // will catch the conflict if the check failed.
    }
  }

  String _humanError(Object e) {
    if (e is ApiException) {
      if (e.status == 404) return 'No book found for this ISBN.';
      return 'Lookup failed (${e.status}): ${e.message}';
    }
    return 'Lookup failed: $e';
  }

  void _removeItem(ScannedItem item) {
    setState(() {
      _items.remove(item);
      _seen.remove(item.isbn);
    });
  }

  void _finish() {
    Navigator.of(context).pop(List<ScannedItem>.from(_items));
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text('Batch scan (${_items.length})')),
      body: Column(
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(12, 8, 12, 4),
            child: Row(
              children: [
                const Text('Adding to:'),
                const SizedBox(width: 8),
                Expanded(
                  child: SegmentedButton<ListType>(
                    segments: const [
                      ButtonSegment(
                        value: ListType.owned,
                        label: Text('Collection'),
                        icon: Icon(Icons.inventory_2_outlined),
                      ),
                      ButtonSegment(
                        value: ListType.wishlist,
                        label: Text('Wishlist'),
                        icon: Icon(Icons.bookmark_border),
                      ),
                    ],
                    selected: {_listType},
                    onSelectionChanged: (s) =>
                        setState(() => _listType = s.first),
                  ),
                ),
              ],
            ),
          ),
          Expanded(
            child: Stack(
              children: [
                MobileScanner(controller: _controller, onDetect: _onDetect),
                Positioned(
                  left: 16,
                  right: 16,
                  bottom: 16,
                  child: _ScanHint(count: _items.length),
                ),
              ],
            ),
          ),
          if (_items.isNotEmpty)
            SizedBox(
              height: 120,
              child: ListView.separated(
                padding: const EdgeInsets.all(8),
                scrollDirection: Axis.horizontal,
                itemCount: _items.length,
                separatorBuilder: (_, _) => const SizedBox(width: 8),
                itemBuilder: (_, i) => _ScannedThumb(
                  item: _items[i],
                  onRemove: () => _removeItem(_items[i]),
                ),
              ),
            ),
        ],
      ),
      bottomNavigationBar: SafeArea(
        child: Padding(
          padding: const EdgeInsets.fromLTRB(12, 4, 12, 8),
          child: FilledButton.icon(
            icon: const Icon(Icons.check),
            label: Text(
              _items.isEmpty
                  ? 'Done (no items)'
                  : 'Done — review ${_items.length}',
            ),
            style: FilledButton.styleFrom(
              minimumSize: const Size.fromHeight(52),
            ),
            onPressed: _items.isEmpty ? null : _finish,
          ),
        ),
      ),
    );
  }
}

class _ScanHint extends StatelessWidget {
  const _ScanHint({required this.count});
  final int count;

  @override
  Widget build(BuildContext context) {
    final text = count == 0
        ? 'Point the camera at a book barcode'
        : 'Keep scanning, or tap Done below to review';
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
      decoration: BoxDecoration(
        color: Colors.black.withValues(alpha: 0.6),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Text(
        text,
        textAlign: TextAlign.center,
        style: const TextStyle(color: Colors.white),
      ),
    );
  }
}

class _ScannedThumb extends StatelessWidget {
  const _ScannedThumb({required this.item, required this.onRemove});
  final ScannedItem item;
  final VoidCallback onRemove;

  @override
  Widget build(BuildContext context) {
    return ListenableBuilder(
      listenable: item,
      builder: (_, _) {
        Widget body;
        switch (item.state) {
          case ScanItemState.lookingUp:
            body = const Center(
              child: SizedBox(
                width: 24,
                height: 24,
                child: CircularProgressIndicator(strokeWidth: 2),
              ),
            );
            break;
          case ScanItemState.lookupFailed:
            body = const Padding(
              padding: EdgeInsets.all(8),
              child: Icon(Icons.error_outline, color: Colors.redAccent),
            );
            break;
          default:
            body = Padding(
              padding: const EdgeInsets.all(8),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    item.title.isEmpty ? item.isbn : item.title,
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                    style: const TextStyle(fontWeight: FontWeight.w600),
                  ),
                  if (item.hasDuplicates)
                    const Padding(
                      padding: EdgeInsets.only(top: 4),
                      child: Text(
                        'Already in collection',
                        style: TextStyle(
                          fontSize: 11,
                          color: Colors.orange,
                        ),
                      ),
                    ),
                ],
              ),
            );
        }
        return SizedBox(
          width: 160,
          child: Card(
            child: Stack(
              children: [
                Positioned.fill(child: body),
                Positioned(
                  top: 0,
                  right: 0,
                  child: IconButton(
                    iconSize: 18,
                    icon: const Icon(Icons.close),
                    onPressed: onRemove,
                    tooltip: 'Remove',
                  ),
                ),
              ],
            ),
          ),
        );
      },
    );
  }
}
