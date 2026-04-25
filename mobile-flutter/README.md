# Alaya Archive — Mobile (Flutter)

Native Android + iOS app for the Alaya Archive media catalog. Complements the
React PWA in `../frontend/` with mobile-first features (barcode scanning,
camera-driven metadata lookup).

Bundle id: `com.dewees.alaya_archive`

## Run

```sh
flutter pub get
flutter run            # pick a device when prompted
flutter run -d <id>    # or target one directly (see `flutter devices`)
```

## Test

```sh
flutter test
```

## Layout

- `lib/` — Dart source
- `android/`, `ios/` — platform projects
- `test/` — widget + unit tests

## Backend

Talks to the Go API in `../backend-go/` (default `http://localhost:8080`).
