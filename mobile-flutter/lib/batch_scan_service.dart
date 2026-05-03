import 'package:dio/dio.dart';

import 'api_client.dart';
import 'batch_scan_models.dart';

// lookupAndCheckIsbn drives a freshly-created ScannedItem through the two
// network calls we always do after a scan: fetch metadata from the lookup
// provider, then check whether this user already has a copy. Both single
// and batch scan flows share this so the item lifecycle stays identical.
//
// Errors during lookup transition the item to lookupFailed with a human
// message. Errors during the duplicate check are swallowed: the item is
// still saveable, and the backend's on_duplicate=error guard catches any
// missed duplicate at write time.
Future<void> lookupAndCheckIsbn(ApiClient client, ScannedItem item) async {
  try {
    final result = await client.lookupBookByIsbn(item.isbn);
    item.applyLookup(result);
  } catch (e) {
    item.markLookupFailed(humanLookupError(e));
    return;
  }
  try {
    final dupes = await client.checkDuplicate(
      mediaType: item.mediaType,
      isbn: item.isbn,
    );
    item.applyDuplicates(dupes);
  } catch (_) {
    // Non-fatal — see comment above.
  }
}

// humanLookupError turns API/network errors into the short string we put
// on a card or in a banner. Kept here so error wording is consistent
// wherever a lookup happens.
String humanLookupError(Object e) {
  if (e is ApiException) {
    if (e.status == 404) return 'No book found for this ISBN.';
    return 'Lookup failed (${e.status}): ${e.message}';
  }
  if (e is DioException) {
    return 'Network error: ${e.message ?? e.type.name}';
  }
  return 'Lookup failed: $e';
}
