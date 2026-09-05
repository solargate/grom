import 'dart:convert';

import 'package:http/http.dart' as http;

import 'strava_api_constants.dart';
import 'strava_api_gpx.dart';

class StravaApiException implements Exception {
  StravaApiException(this.message, {this.statusCode});

  final String message;
  final int? statusCode;

  @override
  String toString() => message;
}

class StravaSummaryActivity {
  const StravaSummaryActivity({
    required this.id,
    required this.name,
    required this.sportType,
    required this.type,
    required this.startDate,
    required this.movingTime,
    required this.elapsedTime,
    required this.distanceMeters,
    this.maxSpeedMs,
    this.averageSpeedMs,
    this.totalElevationGain,
    this.description,
    this.totalPhotoCount = 0,
    this.hasMapPolyline = false,
  });

  final int id;
  final String name;
  final String sportType;
  final String type;
  final DateTime startDate;
  final int movingTime;
  final int elapsedTime;
  final double distanceMeters;
  final double? maxSpeedMs;
  final double? averageSpeedMs;
  final double? totalElevationGain;
  final String? description;
  final int totalPhotoCount;
  final bool hasMapPolyline;

  factory StravaSummaryActivity.fromJson(Map<String, dynamic> json) {
    final startRaw = json['start_date'] as String? ?? '';
    final startDate = DateTime.tryParse(startRaw)?.toUtc() ??
        DateTime.fromMillisecondsSinceEpoch(0, isUtc: true);
    final map = json['map'];
    var hasPolyline = false;
    if (map is Map<String, dynamic>) {
      final summary = map['summary_polyline'];
      hasPolyline = summary is String && summary.isNotEmpty;
    }
    return StravaSummaryActivity(
      id: (json['id'] as num).toInt(),
      name: (json['name'] as String?)?.trim() ?? '',
      sportType: (json['sport_type'] as String?)?.trim() ?? '',
      type: (json['type'] as String?)?.trim() ?? '',
      startDate: startDate,
      movingTime: (json['moving_time'] as num?)?.toInt() ?? 0,
      elapsedTime: (json['elapsed_time'] as num?)?.toInt() ?? 0,
      distanceMeters: (json['distance'] as num?)?.toDouble() ?? 0,
      maxSpeedMs: (json['max_speed'] as num?)?.toDouble(),
      averageSpeedMs: (json['average_speed'] as num?)?.toDouble(),
      totalElevationGain: (json['total_elevation_gain'] as num?)?.toDouble(),
      description: (json['description'] as String?)?.trim(),
      totalPhotoCount: (json['total_photo_count'] as num?)?.toInt() ?? 0,
      hasMapPolyline: hasPolyline,
    );
  }
}

class StravaActivityPhoto {
  const StravaActivityPhoto({
    required this.uniqueId,
    required this.url,
  });

  final String uniqueId;
  final String url;
}

class StravaApiClient {
  StravaApiClient({http.Client? httpClient})
      : _http = httpClient ?? http.Client();

  final http.Client _http;

  Future<List<StravaSummaryActivity>> listAthleteActivities({
    required String accessToken,
    int perPage = kStravaApiSyncLimit,
    int page = 1,
  }) async {
    final uri = Uri.parse('$kStravaApiBase/athlete/activities').replace(
      queryParameters: {
        'per_page': '$perPage',
        'page': '$page',
      },
    );
    final response = await _http.get(
      uri,
      headers: {'Authorization': 'Bearer $accessToken'},
    );
    if (response.statusCode != 200) {
      throw StravaApiException(
        _errorMessage(response),
        statusCode: response.statusCode,
      );
    }
    final decoded = jsonDecode(response.body);
    if (decoded is! List) {
      return const [];
    }
    return decoded
        .whereType<Map<String, dynamic>>()
        .map(StravaSummaryActivity.fromJson)
        .toList();
  }

  Future<StravaSummaryActivity> getActivity({
    required String accessToken,
    required int activityId,
  }) async {
    final uri = Uri.parse('$kStravaApiBase/activities/$activityId');
    final response = await _http.get(
      uri,
      headers: {'Authorization': 'Bearer $accessToken'},
    );
    if (response.statusCode != 200) {
      throw StravaApiException(
        _errorMessage(response),
        statusCode: response.statusCode,
      );
    }
    final json = jsonDecode(response.body) as Map<String, dynamic>;
    return StravaSummaryActivity.fromJson(json);
  }

  /// Returns GPS points from streams, or empty if no usable latlng stream.
  Future<List<StravaStreamPoint>> getActivityTrackPoints({
    required String accessToken,
    required int activityId,
  }) async {
    final uri = Uri.parse('$kStravaApiBase/activities/$activityId/streams')
        .replace(
      queryParameters: {
        'keys': 'latlng,time,altitude',
        'key_by_type': 'true',
      },
    );
    final response = await _http.get(
      uri,
      headers: {'Authorization': 'Bearer $accessToken'},
    );
    if (response.statusCode == 404) {
      return const [];
    }
    if (response.statusCode != 200) {
      throw StravaApiException(
        _errorMessage(response),
        statusCode: response.statusCode,
      );
    }

    final decoded = jsonDecode(response.body);
    Map<String, dynamic>? byType;
    if (decoded is Map<String, dynamic>) {
      byType = decoded;
    } else if (decoded is List) {
      byType = {
        for (final item in decoded.whereType<Map<String, dynamic>>())
          if (item['type'] is String) item['type'] as String: item,
      };
    }
    if (byType == null) {
      return const [];
    }

    final latlng = byType['latlng'];
    if (latlng is! Map<String, dynamic>) {
      return const [];
    }
    final latlngData = latlng['data'];
    if (latlngData is! List || latlngData.length < 2) {
      return const [];
    }

    final timeData = (byType['time'] is Map<String, dynamic>)
        ? (byType['time'] as Map<String, dynamic>)['data']
        : null;
    final altData = (byType['altitude'] is Map<String, dynamic>)
        ? (byType['altitude'] as Map<String, dynamic>)['data']
        : null;

    final points = <StravaStreamPoint>[];
    for (var i = 0; i < latlngData.length; i++) {
      final pair = latlngData[i];
      if (pair is! List || pair.length < 2) {
        continue;
      }
      final lat = (pair[0] as num?)?.toDouble();
      final lon = (pair[1] as num?)?.toDouble();
      if (lat == null || lon == null) {
        continue;
      }
      if (lat == 0 && lon == 0) {
        continue;
      }
      int? timeSeconds;
      if (timeData is List && i < timeData.length && timeData[i] is num) {
        timeSeconds = (timeData[i] as num).toInt();
      }
      double? elevation;
      if (altData is List && i < altData.length && altData[i] is num) {
        elevation = (altData[i] as num).toDouble();
      }
      points.add(
        StravaStreamPoint(
          lat: lat,
          lon: lon,
          timeSeconds: timeSeconds,
          elevation: elevation,
        ),
      );
    }
    return points;
  }

  /// Best-effort photo list; returns empty on failure or missing endpoint.
  Future<List<StravaActivityPhoto>> listActivityPhotos({
    required String accessToken,
    required int activityId,
    int size = 2048,
  }) async {
    final uri =
        Uri.parse('$kStravaApiBase/activities/$activityId/photos').replace(
      queryParameters: {
        'photo_sources': 'true',
        'size': '$size',
      },
    );
    try {
      final response = await _http.get(
        uri,
        headers: {'Authorization': 'Bearer $accessToken'},
      );
      if (response.statusCode != 200) {
        return const [];
      }
      final decoded = jsonDecode(response.body);
      if (decoded is! List) {
        return const [];
      }
      final photos = <StravaActivityPhoto>[];
      for (final item in decoded.whereType<Map<String, dynamic>>()) {
        final urls = item['urls'];
        if (urls is! Map) {
          continue;
        }
        String? bestUrl;
        var bestSize = -1;
        for (final entry in urls.entries) {
          final keySize = int.tryParse('${entry.key}') ?? 0;
          final value = entry.value;
          if (value is String && value.isNotEmpty && keySize >= bestSize) {
            bestSize = keySize;
            bestUrl = value;
          }
        }
        if (bestUrl == null || bestUrl.isEmpty) {
          continue;
        }
        final uniqueId = (item['unique_id'] as String?)?.trim();
        photos.add(
          StravaActivityPhoto(
            uniqueId: (uniqueId != null && uniqueId.isNotEmpty)
                ? uniqueId
                : 'photo_${photos.length}',
            url: bestUrl,
          ),
        );
      }
      return photos;
    } catch (_) {
      return const [];
    }
  }

  Future<List<int>?> downloadBytes(String url) async {
    try {
      final response = await _http.get(Uri.parse(url));
      if (response.statusCode != 200 || response.bodyBytes.isEmpty) {
        return null;
      }
      return response.bodyBytes;
    } catch (_) {
      return null;
    }
  }

  String _errorMessage(http.Response response) {
    try {
      final json = jsonDecode(response.body);
      if (json is Map<String, dynamic>) {
        final message = json['message'];
        if (message is String && message.isNotEmpty) {
          return message;
        }
      }
    } catch (_) {
      // Fall through.
    }
    return 'Strava API error (${response.statusCode})';
  }
}
