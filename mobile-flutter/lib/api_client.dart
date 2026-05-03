import 'package:dio/dio.dart';

import 'auth_controller.dart';

class ApiClient {
  ApiClient(this._auth) : _dio = Dio() {
    _dio.interceptors.add(_AuthInterceptor(_auth, _dio));
  }

  final AuthController _auth;
  final Dio _dio;

  Future<LookupResult> lookupBookByIsbn(String isbn) async {
    try {
      final response = await _dio.get(
        '/api/v1/lookup',
        queryParameters: {'type': 'book', 'isbn': isbn},
      );
      return LookupResult.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw _toApiException(e);
    }
  }

  // checkDuplicate returns existing media in the user's collection that
  // match (mediaType, isbn). Empty list = no duplicates.
  Future<List<MediaItem>> checkDuplicate({
    required String mediaType,
    required String isbn,
  }) async {
    try {
      final response = await _dio.get(
        '/api/v1/media/check',
        queryParameters: {'type': mediaType, 'isbn': isbn},
      );
      final data = response.data as Map<String, dynamic>;
      final items = (data['items'] as List?) ?? const [];
      return items
          .map((e) => MediaItem.fromJson(e as Map<String, dynamic>))
          .toList();
    } on DioException catch (e) {
      throw _toApiException(e);
    }
  }

  // createMedia POSTs a single item with the given duplicate policy.
  // Returns the saved/affected MediaItem. Throws ApiException on 409 when
  // policy is 'error' — caller should re-issue with a different policy after
  // asking the user.
  Future<MediaItem> createMedia({
    required Map<String, dynamic> payload,
    String onDuplicate = 'error',
  }) async {
    try {
      final body = Map<String, dynamic>.from(payload)
        ..['on_duplicate'] = onDuplicate;
      final response = await _dio.post('/api/v1/media/', data: body);
      return MediaItem.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw _toApiException(e);
    }
  }
}

ApiException _toApiException(DioException e) {
  final status = e.response?.statusCode ?? 0;
  final data = e.response?.data;
  final msg = (data is Map && data['error'] is String)
      ? data['error'] as String
      : (e.message ?? 'request failed');
  return ApiException(status: status, message: msg);
}

class _AuthInterceptor extends Interceptor {
  _AuthInterceptor(this._auth, this._dio);
  final AuthController _auth;
  final Dio _dio;

  @override
  void onRequest(RequestOptions options, RequestInterceptorHandler handler) {
    options.baseUrl = _auth.baseUrl;
    final token = _auth.accessToken;
    if (token != null && token.isNotEmpty) {
      options.headers['Authorization'] = 'Bearer $token';
    }
    handler.next(options);
  }

  @override
  Future<void> onError(
    DioException err,
    ErrorInterceptorHandler handler,
  ) async {
    final retried = err.requestOptions.extra['retried'] == true;
    if (err.response?.statusCode == 401 && !retried) {
      final refreshed = await _auth.refresh();
      if (refreshed) {
        final opts = err.requestOptions
          ..extra['retried'] = true
          ..headers['Authorization'] = 'Bearer ${_auth.accessToken}';
        try {
          final response = await _dio.fetch(opts);
          return handler.resolve(response);
        } on DioException catch (retryErr) {
          return handler.next(retryErr);
        }
      } else {
        await _auth.logout();
      }
    }
    handler.next(err);
  }
}

class ApiException implements Exception {
  ApiException({required this.status, required this.message});
  final int status;
  final String message;

  @override
  String toString() => 'ApiException($status): $message';
}

class LookupResult {
  LookupResult({
    required this.provider,
    required this.title,
    required this.authors,
    this.subtitle,
    this.description,
    this.coverImage,
    this.year,
    this.publisher,
    this.isbn13,
    this.isbn10,
    this.pageCount,
    this.series,
    this.seriesPosition,
  });

  final String provider;
  final String title;
  final List<String> authors;
  final String? subtitle;
  final String? description;
  final String? coverImage;
  final int? year;
  final String? publisher;
  final String? isbn13;
  final String? isbn10;
  final int? pageCount;
  final String? series;
  final int? seriesPosition;

  factory LookupResult.fromJson(Map<String, dynamic> json) {
    final provider = json['provider'] as String? ?? '';
    final result = (json['result'] as Map<String, dynamic>?) ?? const {};
    return LookupResult(
      provider: provider,
      title: (result['title'] as String?) ?? '(untitled)',
      authors: ((result['authors'] as List?) ?? const []).cast<String>(),
      subtitle: result['subtitle'] as String?,
      description: result['description'] as String?,
      coverImage: result['cover_image'] as String?,
      year: result['year'] as int?,
      publisher: result['publisher'] as String?,
      isbn13: result['isbn_13'] as String?,
      isbn10: result['isbn_10'] as String?,
      pageCount: result['page_count'] as int?,
      series: result['series'] as String?,
      seriesPosition: result['series_position'] as int?,
    );
  }
}

// MediaItem mirrors the backend's repository.Media struct — only the fields
// we need on the mobile app are decoded. Used both for duplicate-check
// results and create responses.
class MediaItem {
  MediaItem({
    required this.id,
    required this.mediaType,
    required this.title,
    required this.status,
    this.creator,
    this.isbn,
    this.coverImage,
    this.yearReleased,
    this.rating,
    this.notes,
  });

  final String id;
  final String mediaType;
  final String title;
  final String status;
  final String? creator;
  final String? isbn;
  final String? coverImage;
  final int? yearReleased;
  final int? rating;
  final String? notes;

  factory MediaItem.fromJson(Map<String, dynamic> json) {
    return MediaItem(
      id: json['id'] as String? ?? '',
      mediaType: json['media_type'] as String? ?? '',
      title: json['title'] as String? ?? '',
      status: json['status'] as String? ?? 'planned',
      creator: json['creator'] as String?,
      isbn: json['isbn'] as String?,
      coverImage: json['cover_image'] as String?,
      yearReleased: json['year_released'] as int?,
      rating: json['rating'] as int?,
      notes: json['notes'] as String?,
    );
  }
}
