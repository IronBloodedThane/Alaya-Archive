import 'package:flutter/foundation.dart';

import 'api_client.dart';

// Lifecycle of a single scanned item as it moves from "just scanned" to
// "ready to save". Pages render different UI per state.
enum ScanItemState {
  lookingUp,
  ready,
  lookupFailed,
  saving,
  saved,
  saveFailed,
  skipped,
}

// Which list this item goes onto. owned = "I have it"; wishlist = "I want it".
// Set per scanning session, not per item — the user picks once at the top
// of the scanner. Maps to the backend's `list_type` field.
enum ListType { owned, wishlist }

String listTypeName(ListType l) {
  switch (l) {
    case ListType.owned:
      return 'owned';
    case ListType.wishlist:
      return 'wishlist';
  }
}

String listTypeLabel(ListType l) {
  switch (l) {
    case ListType.owned:
      return 'Collection';
    case ListType.wishlist:
      return 'Wishlist';
  }
}

// Per-item duplicate policy chosen by the user during the review step.
// Mapped 1:1 to the backend's `on_duplicate` field.
enum DuplicatePolicy {
  error, // default — should never be sent if a known duplicate exists
  overwrite,
  skip,
  allow,
}

String duplicatePolicyName(DuplicatePolicy p) {
  switch (p) {
    case DuplicatePolicy.error:
      return 'error';
    case DuplicatePolicy.overwrite:
      return 'overwrite';
    case DuplicatePolicy.skip:
      return 'skip';
    case DuplicatePolicy.allow:
      return 'allow';
  }
}

// ScannedItem is the editable in-memory record for one barcode in the
// current batch session. It owns both the lookup result and the user's
// edits — the merged values are what we POST.
//
// We use ChangeNotifier so individual cards can rebuild without rebuilding
// the whole list when fields change.
class ScannedItem extends ChangeNotifier {
  ScannedItem({
    required this.isbn,
    required this.mediaType,
    this.listType = ListType.owned,
  });

  final String isbn;
  final String mediaType;
  ListType listType;

  ScanItemState state = ScanItemState.lookingUp;
  LookupResult? lookup;
  String? errorMessage;

  // Editable fields — pre-filled from lookup, then user can override.
  String title = '';
  String creator = '';
  String? yearReleased;
  String status = 'planned';
  String notes = '';

  // Duplicate detection result. Empty list = no duplicates.
  List<MediaItem> duplicates = const [];
  DuplicatePolicy policy = DuplicatePolicy.error;

  // Result of the save attempt.
  MediaItem? savedItem;
  String? saveError;

  bool get hasDuplicates => duplicates.isNotEmpty;

  void applyLookup(LookupResult result) {
    lookup = result;
    title = result.title;
    creator = result.authors.join(', ');
    yearReleased = result.year?.toString();
    state = ScanItemState.ready;
    notifyListeners();
  }

  void markLookupFailed(String message) {
    errorMessage = message;
    state = ScanItemState.lookupFailed;
    notifyListeners();
  }

  void applyDuplicates(List<MediaItem> dupes) {
    duplicates = dupes;
    if (dupes.isNotEmpty && policy == DuplicatePolicy.error) {
      // Default to "skip" when a duplicate is detected so Save All is safe
      // by default. User can flip to overwrite or allow.
      policy = DuplicatePolicy.skip;
    }
    notifyListeners();
  }

  void update({
    String? title,
    String? creator,
    String? yearReleased,
    String? status,
    String? notes,
    DuplicatePolicy? policy,
    ListType? listType,
  }) {
    if (title != null) this.title = title;
    if (creator != null) this.creator = creator;
    if (yearReleased != null) this.yearReleased = yearReleased;
    if (status != null) this.status = status;
    if (notes != null) this.notes = notes;
    if (policy != null) this.policy = policy;
    if (listType != null) this.listType = listType;
    notifyListeners();
  }

  void markSaving() {
    state = ScanItemState.saving;
    saveError = null;
    notifyListeners();
  }

  void markSaved(MediaItem item) {
    savedItem = item;
    state = ScanItemState.saved;
    notifyListeners();
  }

  void markSaveFailed(String message) {
    saveError = message;
    state = ScanItemState.saveFailed;
    notifyListeners();
  }

  void markSkipped() {
    state = ScanItemState.skipped;
    notifyListeners();
  }

  // Build the JSON payload sent to POST /api/v1/media. on_duplicate is
  // appended by the api client.
  Map<String, dynamic> toCreatePayload() {
    final payload = <String, dynamic>{
      'media_type': mediaType,
      'title': title.trim().isEmpty ? '(untitled)' : title.trim(),
      'creator': creator.trim(),
      'status': status,
      'isbn': isbn,
      'list_type': listTypeName(listType),
      'notes': notes.trim(),
    };
    final year = int.tryParse(yearReleased ?? '');
    if (year != null) payload['year_released'] = year;
    final cover = lookup?.coverImage;
    if (cover != null && cover.isNotEmpty) payload['cover_image'] = cover;
    final desc = lookup?.description;
    if (desc != null && desc.isNotEmpty) payload['description'] = desc;
    return payload;
  }
}
