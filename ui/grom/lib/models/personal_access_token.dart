class PersonalAccessToken {
  PersonalAccessToken({
    required this.id,
    required this.name,
    required this.tokenPrefix,
    required this.scopes,
    required this.createdAt,
    this.expiresAt,
    this.lastUsedAt,
  });

  final String id;
  final String name;
  final String tokenPrefix;
  final List<String> scopes;
  final DateTime createdAt;
  final DateTime? expiresAt;
  final DateTime? lastUsedAt;

  factory PersonalAccessToken.fromJson(Map<String, dynamic> json) {
    return PersonalAccessToken(
      id: json['id'] as String,
      name: json['name'] as String,
      tokenPrefix: json['token_prefix'] as String? ?? '',
      scopes: (json['scopes'] as List<dynamic>? ?? const [])
          .map((item) => item.toString())
          .toList(),
      createdAt: DateTime.parse(json['created_at'] as String),
      expiresAt: _parseOptionalDate(json['expires_at']),
      lastUsedAt: _parseOptionalDate(json['last_used_at']),
    );
  }

  static DateTime? _parseOptionalDate(Object? value) {
    if (value == null) {
      return null;
    }
    final text = value.toString().trim();
    if (text.isEmpty) {
      return null;
    }
    return DateTime.parse(text);
  }
}

class CreatePersonalAccessTokenResult {
  CreatePersonalAccessTokenResult({
    required this.token,
    required this.pat,
  });

  final String token;
  final PersonalAccessToken pat;

  factory CreatePersonalAccessTokenResult.fromJson(Map<String, dynamic> json) {
    return CreatePersonalAccessTokenResult(
      token: json['token'] as String,
      pat: PersonalAccessToken.fromJson(
        json['pat'] as Map<String, dynamic>,
      ),
    );
  }
}
