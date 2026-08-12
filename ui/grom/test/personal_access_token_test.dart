import 'package:flutter_test/flutter_test.dart';
import 'package:grom/models/personal_access_token.dart';

void main() {
  test('PersonalAccessToken.fromJson parses metadata', () {
    final token = PersonalAccessToken.fromJson({
      'id': 'id-1',
      'name': 'Script',
      'token_prefix': 'grom_pat_ab',
      'scopes': ['workouts:read', 'equipment:read'],
      'created_at': '2026-08-12T10:00:00Z',
      'expires_at': '2026-11-10T10:00:00Z',
      'last_used_at': '2026-08-12T15:00:00Z',
    });

    expect(token.id, 'id-1');
    expect(token.name, 'Script');
    expect(token.tokenPrefix, 'grom_pat_ab');
    expect(token.scopes, ['workouts:read', 'equipment:read']);
    expect(token.expiresAt, isNotNull);
    expect(token.lastUsedAt, isNotNull);
  });

  test('CreatePersonalAccessTokenResult.fromJson parses token once payload', () {
    final result = CreatePersonalAccessTokenResult.fromJson({
      'token': 'grom_pat_secret',
      'pat': {
        'id': 'id-1',
        'name': 'Script',
        'token_prefix': 'grom_pat_se',
        'scopes': ['workouts:read'],
        'created_at': '2026-08-12T10:00:00Z',
      },
    });

    expect(result.token, 'grom_pat_secret');
    expect(result.pat.name, 'Script');
  });
}
