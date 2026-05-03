import 'package:flutter/material.dart';

import 'api_client.dart';
import 'batch_scan_models.dart';

// BatchReviewPage shows the items captured by BatchScannerPage as editable
// cards, lets the user pick a duplicate policy per conflict, then POSTs
// each item with the chosen on_duplicate flag.
class BatchReviewPage extends StatefulWidget {
  const BatchReviewPage({
    super.key,
    required this.client,
    required this.items,
  });

  final ApiClient client;
  final List<ScannedItem> items;

  @override
  State<BatchReviewPage> createState() => _BatchReviewPageState();
}

class _BatchReviewPageState extends State<BatchReviewPage> {
  bool _saving = false;

  bool get _allDone => widget.items.every(_isTerminal);
  bool _isTerminal(ScannedItem i) =>
      i.state == ScanItemState.saved ||
      i.state == ScanItemState.skipped ||
      i.state == ScanItemState.lookupFailed;

  // Items eligible for the next Save All pass: ready to send, or previously
  // failed and worth retrying.
  Iterable<ScannedItem> get _saveable => widget.items.where(
        (i) =>
            i.state == ScanItemState.ready ||
            i.state == ScanItemState.saveFailed,
      );

  Future<void> _saveAll() async {
    if (_saving) return;
    setState(() => _saving = true);

    for (final item in _saveable.toList()) {
      // If user wants to skip a duplicate, no need to round-trip the server.
      if (item.hasDuplicates && item.policy == DuplicatePolicy.skip) {
        item.markSkipped();
        continue;
      }
      item.markSaving();
      try {
        final policy = item.hasDuplicates
            ? duplicatePolicyName(item.policy)
            : 'error';
        final saved = await widget.client.createMedia(
          payload: item.toCreatePayload(),
          onDuplicate: policy,
        );
        item.markSaved(saved);
      } on ApiException catch (e) {
        if (e.status == 409) {
          // Server detected a duplicate we didn't know about. Surface it so
          // the user can pick a policy and retry.
          item.markSaveFailed('Already in collection — choose a policy.');
          try {
            final dupes = await widget.client.checkDuplicate(
              mediaType: item.mediaType,
              isbn: item.isbn,
            );
            item.applyDuplicates(dupes);
          } catch (_) {}
        } else {
          item.markSaveFailed('Save failed (${e.status}): ${e.message}');
        }
      } catch (e) {
        item.markSaveFailed('Save failed: $e');
      }
    }

    if (!mounted) return;
    setState(() => _saving = false);
  }

  @override
  Widget build(BuildContext context) {
    final remaining = _saveable.length;
    return Scaffold(
      appBar: AppBar(
        title: Text('Review (${widget.items.length})'),
      ),
      body: ListView.separated(
        padding: const EdgeInsets.all(12),
        itemCount: widget.items.length,
        separatorBuilder: (_, _) => const SizedBox(height: 8),
        itemBuilder: (_, i) => _ReviewCard(item: widget.items[i]),
      ),
      bottomNavigationBar: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(12),
          child: Row(
            children: [
              Expanded(
                child: FilledButton.icon(
                  icon: _saving
                      ? const SizedBox(
                          width: 18,
                          height: 18,
                          child: CircularProgressIndicator(
                            strokeWidth: 2,
                            color: Colors.white,
                          ),
                        )
                      : const Icon(Icons.cloud_upload_outlined),
                  label: Text(
                    _saving
                        ? 'Saving…'
                        : remaining == 0 && _allDone
                            ? 'All done'
                            : 'Save all ($remaining)',
                  ),
                  onPressed: _saving || remaining == 0 ? null : _saveAll,
                ),
              ),
              const SizedBox(width: 8),
              OutlinedButton(
                onPressed: _saving
                    ? null
                    : () => Navigator.of(context).pop(),
                child: const Text('Close'),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _ReviewCard extends StatelessWidget {
  const _ReviewCard({required this.item});
  final ScannedItem item;

  @override
  Widget build(BuildContext context) {
    return ListenableBuilder(
      listenable: item,
      builder: (_, _) {
        final scheme = Theme.of(context).colorScheme;
        return Card(
          child: Padding(
            padding: const EdgeInsets.all(12),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                _CardHeader(item: item),
                const SizedBox(height: 8),
                if (item.state == ScanItemState.lookingUp)
                  const Center(child: CircularProgressIndicator())
                else if (item.state == ScanItemState.lookupFailed)
                  _ErrorBanner(
                    message: item.errorMessage ?? 'Lookup failed',
                    color: scheme.errorContainer,
                    textColor: scheme.onErrorContainer,
                  )
                else ...[
                  _Editable(item: item),
                  if (item.hasDuplicates) ...[
                    const SizedBox(height: 8),
                    _DuplicatePicker(item: item),
                  ],
                  if (item.state == ScanItemState.saving)
                    const Padding(
                      padding: EdgeInsets.only(top: 8),
                      child: LinearProgressIndicator(),
                    ),
                  if (item.state == ScanItemState.saved)
                    Padding(
                      padding: const EdgeInsets.only(top: 8),
                      child: Text(
                        '✓ Saved',
                        style: TextStyle(color: scheme.primary),
                      ),
                    ),
                  if (item.state == ScanItemState.skipped)
                    const Padding(
                      padding: EdgeInsets.only(top: 8),
                      child: Text('· Skipped (already in collection)'),
                    ),
                  if (item.state == ScanItemState.saveFailed)
                    Padding(
                      padding: const EdgeInsets.only(top: 8),
                      child: _ErrorBanner(
                        message: item.saveError ?? 'Save failed',
                        color: scheme.errorContainer,
                        textColor: scheme.onErrorContainer,
                      ),
                    ),
                ],
              ],
            ),
          ),
        );
      },
    );
  }
}

class _CardHeader extends StatelessWidget {
  const _CardHeader({required this.item});
  final ScannedItem item;

  @override
  Widget build(BuildContext context) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        if (item.lookup?.coverImage != null)
          Padding(
            padding: const EdgeInsets.only(right: 12),
            child: Image.network(
              item.lookup!.coverImage!,
              width: 56,
              errorBuilder: (_, _, _) => const SizedBox(width: 56),
            ),
          ),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  _ListTypeChip(listType: item.listType),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      'ISBN ${item.isbn}',
                      style: Theme.of(context).textTheme.bodySmall,
                    ),
                  ),
                ],
              ),
              if (item.hasDuplicates)
                Padding(
                  padding: const EdgeInsets.only(top: 2),
                  child: Text(
                    '⚠ Already in collection (${item.duplicates.length})',
                    style: TextStyle(
                      color: Theme.of(context).colorScheme.tertiary,
                      fontSize: 12,
                    ),
                  ),
                ),
            ],
          ),
        ),
      ],
    );
  }
}

class _ListTypeChip extends StatelessWidget {
  const _ListTypeChip({required this.listType});
  final ListType listType;

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    final isWishlist = listType == ListType.wishlist;
    final bg = isWishlist ? scheme.tertiaryContainer : scheme.primaryContainer;
    final fg =
        isWishlist ? scheme.onTertiaryContainer : scheme.onPrimaryContainer;
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
      decoration: BoxDecoration(
        color: bg,
        borderRadius: BorderRadius.circular(10),
      ),
      child: Text(
        listTypeLabel(listType),
        style: TextStyle(color: fg, fontSize: 11, fontWeight: FontWeight.w600),
      ),
    );
  }
}

class _Editable extends StatefulWidget {
  const _Editable({required this.item});
  final ScannedItem item;

  @override
  State<_Editable> createState() => _EditableState();
}

class _EditableState extends State<_Editable> {
  late final _title = TextEditingController(text: widget.item.title);
  late final _creator = TextEditingController(text: widget.item.creator);
  late final _year = TextEditingController(
    text: widget.item.yearReleased ?? '',
  );
  late final _notes = TextEditingController(text: widget.item.notes);

  @override
  void dispose() {
    _title.dispose();
    _creator.dispose();
    _year.dispose();
    _notes.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        TextField(
          controller: _title,
          decoration: const InputDecoration(labelText: 'Title'),
          onChanged: (v) => widget.item.update(title: v),
        ),
        TextField(
          controller: _creator,
          decoration: const InputDecoration(labelText: 'Author(s)'),
          onChanged: (v) => widget.item.update(creator: v),
        ),
        TextField(
          controller: _year,
          keyboardType: TextInputType.number,
          decoration: const InputDecoration(labelText: 'Year'),
          onChanged: (v) => widget.item.update(yearReleased: v),
        ),
        TextField(
          controller: _notes,
          decoration: const InputDecoration(labelText: 'Notes'),
          maxLines: 2,
          onChanged: (v) => widget.item.update(notes: v),
        ),
      ],
    );
  }
}

class _DuplicatePicker extends StatelessWidget {
  const _DuplicatePicker({required this.item});
  final ScannedItem item;

  @override
  Widget build(BuildContext context) {
    return Wrap(
      spacing: 6,
      children: [
        _PolicyChip(
          label: 'Skip',
          policy: DuplicatePolicy.skip,
          item: item,
        ),
        _PolicyChip(
          label: 'Overwrite',
          policy: DuplicatePolicy.overwrite,
          item: item,
        ),
        _PolicyChip(
          label: 'Add as copy',
          policy: DuplicatePolicy.allow,
          item: item,
        ),
      ],
    );
  }
}

class _PolicyChip extends StatelessWidget {
  const _PolicyChip({
    required this.label,
    required this.policy,
    required this.item,
  });
  final String label;
  final DuplicatePolicy policy;
  final ScannedItem item;

  @override
  Widget build(BuildContext context) {
    return ChoiceChip(
      label: Text(label),
      selected: item.policy == policy,
      onSelected: (_) => item.update(policy: policy),
    );
  }
}

class _ErrorBanner extends StatelessWidget {
  const _ErrorBanner({
    required this.message,
    required this.color,
    required this.textColor,
  });
  final String message;
  final Color color;
  final Color textColor;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(8),
      decoration: BoxDecoration(
        color: color,
        borderRadius: BorderRadius.circular(6),
      ),
      child: Row(
        children: [
          Icon(Icons.error_outline, color: textColor, size: 18),
          const SizedBox(width: 6),
          Expanded(
            child: Text(message, style: TextStyle(color: textColor)),
          ),
        ],
      ),
    );
  }
}
