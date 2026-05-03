import 'package:alaya_archive/api_client.dart';
import 'package:alaya_archive/auth_controller.dart';
import 'package:alaya_archive/batch_review_page.dart';
import 'package:alaya_archive/batch_scan_models.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

// Helpers to keep the per-test boilerplate light.
ApiClient _stubClient() {
  final auth = AuthController(storage: InMemoryTokenStorage());
  // bootstrap is async but reads from in-memory storage synchronously here;
  // we don't await because no test needs the call site to be ready.
  auth.bootstrap();
  return ApiClient(auth);
}

ScannedItem _readyItem(String isbn) {
  return ScannedItem(isbn: isbn, mediaType: 'book')
    ..applyLookup(LookupResult(
      provider: 'google_books',
      title: 'Title $isbn',
      authors: const [],
    ));
}

MediaItem _existingMedia(String isbn) => MediaItem(
      id: 'existing-$isbn',
      mediaType: 'book',
      title: 'Existing $isbn',
      status: 'planned',
      isbn: isbn,
    );

FilledButton _saveButton(WidgetTester tester) =>
    tester.widget<FilledButton>(find.byType(FilledButton).first);

void main() {
  // Regression test for the bug where the Save button stayed disabled
  // after a single-item lookup completed. The page subscribes to each
  // item's ChangeNotifier so the bottom button reflects state
  // transitions even when the user hasn't touched anything else.
  testWidgets(
      'Save enables when item transitions from lookingUp to ready',
      (tester) async {
    final item = ScannedItem(isbn: '9780441172719', mediaType: 'book');
    expect(item.state, ScanItemState.lookingUp);

    await tester.pumpWidget(MaterialApp(
      home: BatchReviewPage(client: _stubClient(), items: [item]),
    ));
    await tester.pump();

    expect(_saveButton(tester).onPressed, isNull,
        reason: 'Save must be disabled while items are still loading');

    item.applyLookup(LookupResult(
      provider: 'google_books',
      title: 'Dune',
      authors: const ['Frank Herbert'],
    ));
    await tester.pump();

    expect(find.text('Save'), findsOneWidget);
    expect(_saveButton(tester).onPressed, isNotNull,
        reason: 'Save must enable once the item is ready');
  });

  testWidgets(
      'Save disabled when single item is a duplicate set to skip',
      (tester) async {
    final item = _readyItem('9780441172719')
      ..applyDuplicates([_existingMedia('9780441172719')]);
    // applyDuplicates flips default policy to skip when a dup exists.
    expect(item.policy, DuplicatePolicy.skip);

    await tester.pumpWidget(MaterialApp(
      home: BatchReviewPage(client: _stubClient(), items: [item]),
    ));
    await tester.pump();

    expect(_saveButton(tester).onPressed, isNull,
        reason:
            'Skip on a duplicate is a no-op server-side — Save should be off');
  });

  testWidgets(
      'Save re-enables when user changes skip duplicate to overwrite',
      (tester) async {
    final item = _readyItem('9780441172719')
      ..applyDuplicates([_existingMedia('9780441172719')]);

    await tester.pumpWidget(MaterialApp(
      home: BatchReviewPage(client: _stubClient(), items: [item]),
    ));
    await tester.pump();
    expect(_saveButton(tester).onPressed, isNull);

    item.update(policy: DuplicatePolicy.overwrite);
    await tester.pump();

    expect(_saveButton(tester).onPressed, isNotNull,
        reason: 'Switching to overwrite means real server work');
  });

  testWidgets(
      'Batch count reflects only items with real server work',
      (tester) async {
    final dupSkip = _readyItem('9780441172719')
      ..applyDuplicates([_existingMedia('9780441172719')]);
    // Two ready, no-dup items — these are the actual work.
    final fresh1 = _readyItem('9780062315007');
    final fresh2 = _readyItem('9781501142970');

    await tester.pumpWidget(MaterialApp(
      home: BatchReviewPage(
        client: _stubClient(),
        items: [dupSkip, fresh1, fresh2],
      ),
    ));
    await tester.pump();

    // 2 server-bound items, not 3.
    expect(find.text('Save all (2)'), findsOneWidget);
    expect(_saveButton(tester).onPressed, isNotNull);
  });

  testWidgets(
      'Save disables when every item in batch is a duplicate set to skip',
      (tester) async {
    final a = _readyItem('9780441172719')
      ..applyDuplicates([_existingMedia('9780441172719')]);
    final b = _readyItem('9780062315007')
      ..applyDuplicates([_existingMedia('9780062315007')]);

    await tester.pumpWidget(MaterialApp(
      home: BatchReviewPage(client: _stubClient(), items: [a, b]),
    ));
    await tester.pump();

    expect(_saveButton(tester).onPressed, isNull,
        reason: 'All-skip batch should not allow Save — Close is the action');
    // Close button stays available.
    expect(find.widgetWithText(OutlinedButton, 'Close'), findsOneWidget);
  });
}
