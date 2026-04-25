import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

abstract class TokenStorage {
  Future<String?> read(String key);
  Future<void> write(String key, String value);
  Future<void> delete(String key);
}

class SecureTokenStorage implements TokenStorage {
  const SecureTokenStorage([this._impl = const FlutterSecureStorage()]);
  final FlutterSecureStorage _impl;

  @override
  Future<String?> read(String key) => _impl.read(key: key);

  @override
  Future<void> write(String key, String value) =>
      _impl.write(key: key, value: value);

  @override
  Future<void> delete(String key) => _impl.delete(key: key);
}

class InMemoryTokenStorage implements TokenStorage {
  final Map<String, String> _store = {};

  @override
  Future<String?> read(String key) async => _store[key];

  @override
  Future<void> write(String key, String value) async => _store[key] = value;

  @override
  Future<void> delete(String key) async => _store.remove(key);
}

class AuthException implements Exception {
  AuthException({required this.status, required this.message});
  final int status;
  final String message;

  @override
  String toString() => 'AuthException($status): $message';
}

class AuthController extends ChangeNotifier {
  AuthController({TokenStorage? storage})
      : _storage = storage ?? const SecureTokenStorage();

  static const _kBaseUrl = 'baseUrl';
  static const _kAccessToken = 'accessToken';
  static const _kRefreshToken = 'refreshToken';
  // Tim's dev machine on the home network. Android emulator users would use
  // http://10.0.2.2:8080; iOS simulator users http://localhost:8080.
  static const _defaultBaseUrl = 'http://10.0.1.18:8080';

  final TokenStorage _storage;

  String _baseUrl = _defaultBaseUrl;
  String? _accessToken;
  String? _refreshToken;
  bool _ready = false;

  String get baseUrl => _baseUrl;
  String? get accessToken => _accessToken;
  bool get isAuthenticated =>
      _accessToken != null && _accessToken!.isNotEmpty;
  bool get isReady => _ready;

  Future<void> bootstrap() async {
    final stored = await _storage.read(_kBaseUrl);
    if (stored != null && stored.isNotEmpty) _baseUrl = stored;
    _accessToken = await _storage.read(_kAccessToken);
    _refreshToken = await _storage.read(_kRefreshToken);
    _ready = true;
    notifyListeners();
  }

  Future<void> setBaseUrl(String url) async {
    _baseUrl = url;
    await _storage.write(_kBaseUrl, url);
    notifyListeners();
  }

  Future<void> login({
    required String login,
    required String password,
  }) async {
    final dio = _bareDio();
    try {
      final response = await dio.post(
        '/api/v1/auth/login',
        data: {'login': login, 'password': password},
      );
      final data = response.data as Map<String, dynamic>;
      await _setTokens(
        data['access_token'] as String,
        data['refresh_token'] as String,
      );
    } on DioException catch (e) {
      throw _toAuthException(e);
    }
  }

  Future<bool> refresh() async {
    final token = _refreshToken;
    if (token == null || token.isEmpty) return false;
    try {
      final response = await _bareDio().post(
        '/api/v1/auth/refresh',
        data: {'refresh_token': token},
      );
      final data = response.data as Map<String, dynamic>;
      await _setTokens(
        data['access_token'] as String,
        data['refresh_token'] as String,
      );
      return true;
    } catch (_) {
      return false;
    }
  }

  Future<void> logout() async {
    _accessToken = null;
    _refreshToken = null;
    await _storage.delete(_kAccessToken);
    await _storage.delete(_kRefreshToken);
    notifyListeners();
  }

  Future<void> _setTokens(String access, String refresh) async {
    _accessToken = access;
    _refreshToken = refresh;
    await _storage.write(_kAccessToken, access);
    await _storage.write(_kRefreshToken, refresh);
    notifyListeners();
  }

  Dio _bareDio() => Dio(BaseOptions(
        baseUrl: _baseUrl,
        connectTimeout: const Duration(seconds: 10),
        receiveTimeout: const Duration(seconds: 15),
        contentType: 'application/json',
      ));

  AuthException _toAuthException(DioException e) {
    final status = e.response?.statusCode ?? 0;
    final data = e.response?.data;
    final msg = (data is Map && data['error'] is String)
        ? data['error'] as String
        : (e.message ?? 'request failed');
    return AuthException(status: status, message: msg);
  }
}
