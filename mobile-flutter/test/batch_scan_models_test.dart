import 'package:alaya_archive/api_client.dart';
import 'package:alaya_archive/batch_scan_models.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  group('ScannedItem', () {
    test('lookup populates editable fields and flips state to ready', () {
      final item = ScannedItem(isbn: '9780441172719', mediaType: 'book', listType: ListType.owned);
      expect(item.state, ScanItemState.lookingUp);

      item.applyLookup(LookupResult(
        provider: 'google_books',
        title: 'Dune',
        authors: const ['Frank Herbert'],
        year: 1965,
        coverImage: 'https://example.com/dune.jpg',
        description: 'Sci-fi classic',
      ));

      expect(item.state, ScanItemState.ready);
      expect(item.title, 'Dune');
      expect(item.creator, 'Frank Herbert');
      expect(item.yearReleased, '1965');
    });

    test('detected duplicates flip default policy from error to skip', () {
      final item = ScannedItem(isbn: '9780441172719', mediaType: 'book', listType: ListType.owned)
        ..applyLookup(LookupResult(
          provider: 'google_books',
          title: 'Dune',
          authors: const [],
        ));
      expect(item.policy, DuplicatePolicy.error);

      item.applyDuplicates([
        MediaItem(
          id: 'abc',
          mediaType: 'book',
          title: 'Dune',
          status: 'planned',
          isbn: '9780441172719',
        ),
      ]);

      expect(item.hasDuplicates, true);
      expect(item.policy, DuplicatePolicy.skip);
    });

    test('user policy choice survives subsequent applyDuplicates', () {
      final item = ScannedItem(isbn: '9780441172719', mediaType: 'book', listType: ListType.owned)
        ..applyLookup(LookupResult(
          provider: 'google_books',
          title: 'Dune',
          authors: const [],
        ));
      item.applyDuplicates([
        MediaItem(
          id: 'abc',
          mediaType: 'book',
          title: 'Dune',
          status: 'planned',
          isbn: '9780441172719',
        ),
      ]);
      item.update(policy: DuplicatePolicy.overwrite);

      // Second call (e.g. retry) shouldn't clobber the user's choice.
      item.applyDuplicates([
        MediaItem(
          id: 'abc',
          mediaType: 'book',
          title: 'Dune',
          status: 'planned',
          isbn: '9780441172719',
        ),
      ]);
      expect(item.policy, DuplicatePolicy.overwrite);
    });

    test('toCreatePayload includes year when parseable, omits when blank', () {
      final item = ScannedItem(isbn: '9780441172719', mediaType: 'book', listType: ListType.owned)
        ..applyLookup(LookupResult(
          provider: 'google_books',
          title: 'Dune',
          authors: const ['Frank Herbert'],
          year: 1965,
        ));

      var payload = item.toCreatePayload();
      expect(payload['title'], 'Dune');
      expect(payload['creator'], 'Frank Herbert');
      expect(payload['isbn'], '9780441172719');
      expect(payload['year_released'], 1965);

      item.update(yearReleased: '');
      payload = item.toCreatePayload();
      expect(payload.containsKey('year_released'), false);
    });

    test('blank title falls back to (untitled) so server validation passes',
        () {
      final item = ScannedItem(isbn: '9780441172719', mediaType: 'book', listType: ListType.owned)
        ..applyLookup(LookupResult(
          provider: 'google_books',
          title: '',
          authors: const [],
        ));
      expect(item.toCreatePayload()['title'], '(untitled)');
    });

    test('payload includes list_type for both owned and wishlist', () {
      final owned = ScannedItem(
        isbn: '9780441172719',
        mediaType: 'book',
        listType: ListType.owned,
      )..applyLookup(LookupResult(
          provider: 'google_books',
          title: 'Dune',
          authors: const [],
        ));
      expect(owned.toCreatePayload()['list_type'], 'owned');

      final wishlist = ScannedItem(
        isbn: '9780062315007',
        mediaType: 'book',
        listType: ListType.wishlist,
      )..applyLookup(LookupResult(
          provider: 'google_books',
          title: 'The Alchemist',
          authors: const [],
        ));
      expect(wishlist.toCreatePayload()['list_type'], 'wishlist');
    });

    test('lookup failure transitions state and stores message', () {
      final item = ScannedItem(
        isbn: '9780000000000',
        mediaType: 'book',
        listType: ListType.owned,
      );
      item.markLookupFailed('No book found for this ISBN.');
      expect(item.state, ScanItemState.lookupFailed);
      expect(item.errorMessage, 'No book found for this ISBN.');
    });
  });
}
