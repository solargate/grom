/// One index-aligned sample from Strava activity streams.
///
/// GPS is optional: indoor / strength activities may have HR (and other sensors)
/// without latlng. Invalid or (0,0) coordinates are stored as null lat/lon.
class StravaStreamSample {
  const StravaStreamSample({
    this.timeSeconds,
    this.lat,
    this.lon,
    this.elevation,
    this.distanceMeters,
    this.speedMps,
    this.heartRateBpm,
    this.cadenceRpm,
    this.watts,
    this.temperatureC,
    this.gradePercent,
  });

  final int? timeSeconds;
  final double? lat;
  final double? lon;
  final double? elevation;
  final double? distanceMeters;
  final double? speedMps;
  final int? heartRateBpm;
  final int? cadenceRpm;
  final int? watts;
  final int? temperatureC;
  final double? gradePercent;

  bool get hasGps =>
      lat != null && lon != null && !(lat == 0 && lon == 0);

  bool get hasSensorData =>
      heartRateBpm != null ||
      speedMps != null ||
      distanceMeters != null ||
      cadenceRpm != null ||
      watts != null ||
      temperatureC != null ||
      gradePercent != null ||
      elevation != null;

  bool get isUseful => hasGps || hasSensorData;
}

/// Strava stream keys requested on activity streams fetch.
const kStravaStreamKeys =
    'time,latlng,distance,altitude,velocity_smooth,heartrate,cadence,watts,temp,moving,grade_smooth';

/// Builds index-aligned samples from a Strava `key_by_type` streams map.
List<StravaStreamSample> parseStravaStreamsByType(Map<String, dynamic> byType) {
  List? dataOf(String key) {
    final entry = byType[key];
    if (entry is Map<String, dynamic>) {
      final data = entry['data'];
      return data is List ? data : null;
    }
    return null;
  }

  final timeData = dataOf('time');
  final latlngData = dataOf('latlng');
  final distanceData = dataOf('distance');
  final altData = dataOf('altitude');
  final speedData = dataOf('velocity_smooth');
  final hrData = dataOf('heartrate');
  final cadenceData = dataOf('cadence');
  final wattsData = dataOf('watts');
  final tempData = dataOf('temp');
  final gradeData = dataOf('grade_smooth');

  var length = 0;
  for (final list in [
    timeData,
    latlngData,
    distanceData,
    altData,
    speedData,
    hrData,
    cadenceData,
    wattsData,
    tempData,
    gradeData,
  ]) {
    if (list != null && list.length > length) {
      length = list.length;
    }
  }
  if (timeData != null && timeData.isNotEmpty) {
    length = timeData.length;
  }
  if (length < 2) {
    return const [];
  }

  final samples = <StravaStreamSample>[];
  for (var i = 0; i < length; i++) {
    int? timeSeconds;
    if (timeData != null && i < timeData.length && timeData[i] is num) {
      timeSeconds = (timeData[i] as num).toInt();
    }

    double? lat;
    double? lon;
    if (latlngData != null && i < latlngData.length) {
      final pair = latlngData[i];
      if (pair is List && pair.length >= 2) {
        final rawLat = (pair[0] as num?)?.toDouble();
        final rawLon = (pair[1] as num?)?.toDouble();
        if (rawLat != null &&
            rawLon != null &&
            !(rawLat == 0 && rawLon == 0) &&
            rawLat.abs() <= 90 &&
            rawLon.abs() <= 180) {
          lat = rawLat;
          lon = rawLon;
        }
      }
    }

    double? elevation;
    if (altData != null && i < altData.length && altData[i] is num) {
      elevation = (altData[i] as num).toDouble();
    }

    double? distanceMeters;
    if (distanceData != null &&
        i < distanceData.length &&
        distanceData[i] is num) {
      distanceMeters = (distanceData[i] as num).toDouble();
    }

    double? speedMps;
    if (speedData != null && i < speedData.length && speedData[i] is num) {
      final v = (speedData[i] as num).toDouble();
      if (v >= 0 && !v.isNaN) {
        speedMps = v;
      }
    }

    int? heartRateBpm;
    if (hrData != null && i < hrData.length && hrData[i] is num) {
      final bpm = (hrData[i] as num).toInt();
      if (bpm >= 0) {
        heartRateBpm = bpm;
      }
    }

    int? cadenceRpm;
    if (cadenceData != null && i < cadenceData.length && cadenceData[i] is num) {
      final cad = (cadenceData[i] as num).toInt();
      if (cad >= 0) {
        cadenceRpm = cad;
      }
    }

    int? watts;
    if (wattsData != null && i < wattsData.length && wattsData[i] is num) {
      final w = (wattsData[i] as num).toInt();
      if (w >= 0) {
        watts = w;
      }
    }

    int? temperatureC;
    if (tempData != null && i < tempData.length && tempData[i] is num) {
      temperatureC = (tempData[i] as num).toInt();
    }

    double? gradePercent;
    if (gradeData != null && i < gradeData.length && gradeData[i] is num) {
      gradePercent = (gradeData[i] as num).toDouble();
    }

    final sample = StravaStreamSample(
      timeSeconds: timeSeconds,
      lat: lat,
      lon: lon,
      elevation: elevation,
      distanceMeters: distanceMeters,
      speedMps: speedMps,
      heartRateBpm: heartRateBpm,
      cadenceRpm: cadenceRpm,
      watts: watts,
      temperatureC: temperatureC,
      gradePercent: gradePercent,
    );
    if (sample.isUseful || sample.timeSeconds != null) {
      samples.add(sample);
    }
  }
  return samples;
}
