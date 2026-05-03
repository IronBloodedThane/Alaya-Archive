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
  ListType _sessionListType = ListType.owned;

  // Each item is a ChangeNotifier; the cards already listen to their own
  // item, but the bottom Save button reads aggregate state (count of
  // ready/saveable items) and lives on the parent. Subscribe at the page
  // level so the parent rebuilds when any item transitions states (e.g.
  // lookup completes), otherwise Save stays disabled until something else
  // forces a setState.
  @override
  void initState() {
    super.initState();
    for (final item in widget.items) {
      item.addListener(_onItemChanged);
    }
  }

  @override
  void dispose() {
    for (final item in widget.items) {
      item.removeListener(_onItemChanged);
    }
    super.dispose();
  }

  void _onItemChanged() {
    if (mounted) setState(() {});
  }

  void _applyListTypeToAll(ListType lt) {
    setState(() => _sessionListType = lt);
    for (final item in widget.items) {
      item.update(listType: lt);
    }
  }

  bool get _allDone => widget.items.every(_isTerminal);
  bool _isTerminal(ScannedItem i) =>
      i.state == ScanItemState.saved ||
      i.state == ScanItemState.skipped ||
      i.state == ScanItemState.lookupFailed;

  // Items the Save loop will visit: ready to send, or previously failed
  // and worth retrying. _saveAll iterates this and may locally mark
  // skip-duplicates as skipped without a network call.
  Iterable<ScannedItem> get _pending => widget.items.where(
        (i) =>
            i.state == ScanItemState.ready ||
            i.state == ScanItemState.saveFailed,
      );

  // Items the user actually wants pushed to the server. Excludes
  // skip-on-duplicate, which would be a local no-op — Save All shouldn't
  // light up just to mark items skipped on the client. The button reads
  // its enabled state and count from this so "all items set to Skip"
  // leaves only the Close button as a meaningful action.
  Iterable<ScannedItem> get _serverWork =>
      _pending.where((i) =>
          !(i.hasDuplicates && i.policy == DuplicatePolicy.skip));

  Future<void> _saveAll() async {
    if (_saving) return;
    setState(() => _saving = true);

    for (final item in _pending.toList()) {
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
    final remaining = _serverWork.length;
    return Scaffold(
      appBar: AppBar(
        title: Text('Review (${widget.items.length})'),
      ),
      body: Column(
        children: [
          _SessionListTypeToggle(
            value: _sessionListType,
            onChanged: _applyListTypeToAll,
          ),
          Expanded(
            child: ListView.separated(
              padding: const EdgeInsets.fromLTRB(12, 4, 12, 12),
              itemCount: widget.items.length,
              separatorBuilder: (_, _) => const SizedBox(height: 8),
              itemBuilder: (_, i) => _ReviewCard(item: widget.items[i]),
            ),
          ),
        ],
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
                            : remaining == 1
                                ? 'Save'
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
        // Visually flag duplicate cards with a colored border so the user
        // can scan the list and immediately spot anything needing a decision.
        final borderSide = item.hasDuplicates
            ? BorderSide(color: scheme.tertiary, width: 2)
            : BorderSide.none;
        return Card(
          shape: RoundedRectangleBorder(
            side: borderSide,
            borderRadius: BorderRadius.circular(12),
          ),
          child: Padding(
            padding: const EdgeInsets.all(12),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                if (item.hasDuplicates) ...[
                  _DuplicateBanner(item: item),
                  const SizedBox(height: 12),
                ],
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
                      child: Text('· Skipped (kept existing item)'),
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
          child: Row(
            children: [
              _ListTypeChip(item: item),
              const SizedBox(width: 8),
              Expanded(
                child: Text(
                  'ISBN ${item.isbn}',
                  style: Theme.of(context).textTheme.bodySmall,
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }
}

// _ListTypeChip is tappable: the session-level toggle at the top of the
// review screen sets a default for every item, but the user can tap any
// individual chip to override that single item's destination.
class _ListTypeChip extends StatelessWidget {
  const _ListTypeChip({required this.item});
  final ScannedItem item;

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    final isWishlist = item.listType == ListType.wishlist;
    final bg = isWishlist ? scheme.tertiaryContainer : scheme.primaryContainer;
    final fg =
        isWishlist ? scheme.onTertiaryContainer : scheme.onPrimaryContainer;
    return InkWell(
      onTap: () => item.update(
        listType: isWishlist ? ListType.owned : ListType.wishlist,
      ),
      borderRadius: BorderRadius.circular(12),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
        decoration: BoxDecoration(
          color: bg,
          borderRadius: BorderRadius.circular(12),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(
              isWishlist ? Icons.bookmark_border : Icons.inventory_2_outlined,
              size: 14,
              color: fg,
            ),
            const SizedBox(width: 4),
            Text(
              listTypeLabel(item.listType),
              style: TextStyle(
                color: fg,
                fontSize: 11,
                fontWeight: FontWeight.w600,
              ),
            ),
          ],
        ),
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

// _DuplicateBanner sits at the very top of a card whose ISBN already exists
// in the user's collection. It surfaces the existing item, asks the user
// what to do, and offers three explicit choices with a one-line description
// each — so the user doesn't have to remember what "skip"/"overwrite"/
// "add as copy" mean.
class _DuplicateBanner extends StatelessWidget {
  const _DuplicateBanner({required this.item});
  final ScannedItem item;

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    final existing = item.duplicates.first;
    final n = item.duplicates.length;
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: scheme.tertiaryContainer,
        borderRadius: BorderRadius.circular(8),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Icon(Icons.warning_amber_rounded,
                  color: scheme.onTertiaryContainer),
              const SizedBox(width: 8),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      n == 1
                          ? 'Already in your collection'
                          : 'Already in your collection ($n copies)',
                      style: TextStyle(
                        fontWeight: FontWeight.w700,
                        color: scheme.onTertiaryContainer,
                      ),
                    ),
                    const SizedBox(height: 2),
                    Text(
                      _existingLabel(existing),
                      style: TextStyle(
                        fontSize: 12,
                        color: scheme.onTertiaryContainer,
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          Text(
            'What do you want to do?',
            style: TextStyle(
              fontWeight: FontWeight.w600,
              color: scheme.onTertiaryContainer,
            ),
          ),
          const SizedBox(height: 6),
          _PolicyOption(
            item: item,
            policy: DuplicatePolicy.skip,
            icon: Icons.skip_next,
            label: 'Skip',
            description: 'Keep what\'s already there. Don\'t change anything.',
          ),
          _PolicyOption(
            item: item,
            policy: DuplicatePolicy.overwrite,
            icon: Icons.edit_note,
            label: 'Overwrite',
            description: 'Replace the existing entry with this scan\'s info.',
          ),
          _PolicyOption(
            item: item,
            policy: DuplicatePolicy.allow,
            icon: Icons.add_box_outlined,
            label: 'Add as a second copy',
            description: 'Create a new entry alongside the existing one.',
          ),
        ],
      ),
    );
  }

  String _existingLabel(MediaItem m) {
    final parts = <String>[m.title];
    if (m.creator != null && m.creator!.isNotEmpty) parts.add(m.creator!);
    final label = parts.join(' — ');
    return m.mediaType == 'book' ? '$label · ${_listLabel(m)}' : label;
  }

  String _listLabel(MediaItem m) =>
      m.status.isEmpty ? 'in collection' : m.status;
}

class _PolicyOption extends StatelessWidget {
  const _PolicyOption({
    required this.item,
    required this.policy,
    required this.icon,
    required this.label,
    required this.description,
  });
  final ScannedItem item;
  final DuplicatePolicy policy;
  final IconData icon;
  final String label;
  final String description;

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    final selected = item.policy == policy;
    return InkWell(
      onTap: () => item.update(policy: policy),
      borderRadius: BorderRadius.circular(6),
      child: Padding(
        padding: const EdgeInsets.symmetric(vertical: 6, horizontal: 4),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Icon(
              selected ? Icons.radio_button_checked : Icons.radio_button_off,
              size: 20,
              color: selected
                  ? scheme.primary
                  : scheme.onTertiaryContainer.withValues(alpha: 0.6),
            ),
            const SizedBox(width: 8),
            Icon(icon, size: 18, color: scheme.onTertiaryContainer),
            const SizedBox(width: 8),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    label,
                    style: TextStyle(
                      fontWeight: FontWeight.w600,
                      color: scheme.onTertiaryContainer,
                    ),
                  ),
                  Text(
                    description,
                    style: TextStyle(
                      fontSize: 12,
                      color: scheme.onTertiaryContainer.withValues(alpha: 0.85),
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

// _SessionListTypeToggle sits at the top of the review screen. Picking a
// value applies it to every item in the batch — the common case is that
// the user is either at home (Collection) or in a store (Wishlist) and
// wants the same destination for everything they just scanned. Per-item
// overrides happen by tapping the chip on individual cards.
class _SessionListTypeToggle extends StatelessWidget {
  const _SessionListTypeToggle({required this.value, required this.onChanged});
  final ListType value;
  final ValueChanged<ListType> onChanged;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(12, 12, 12, 4),
      child: Row(
        children: [
          const Text('Add all to:'),
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
              selected: {value},
              onSelectionChanged: (s) => onChanged(s.first),
            ),
          ),
        ],
      ),
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
