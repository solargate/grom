import 'package:fit_sdk/fit_sdk.dart';

import 'strava_api_streams.dart';

const _fitEpochOffsetSeconds = 631065600;
const _semicirclesPerDegree = 11930464.7111;

/// Builds a minimal activity FIT from Strava stream samples.
///
/// GPS fields are omitted on samples without valid coordinates so indoor /
/// strength workouts can still carry HR series. Sensor channels are written
/// only when the stream has meaningful data for the activity (e.g. speed only
/// when at least one sample has speed > 0).
List<int> buildFitFromStravaStreams({
  required DateTime startDate,
  required List<StravaStreamSample> samples,
  int? elapsedSeconds,
  int? movingSeconds,
  double? distanceMeters,
  double? averageHeartrate,
  double? maxHeartrate,
  double? averageCadence,
  double? maxCadence,
  double? averageWatts,
  double? maxWatts,
  double? calories,
}) {
  final usable =
      samples.where((s) => s.isUseful || s.timeSeconds != null).toList();
  if (usable.length < 2) {
    throw ArgumentError('Need at least 2 stream samples for a FIT track');
  }

  final includeSpeed = usable.any((s) => s.speedMps != null && s.speedMps! > 0);
  final includeDistance =
      usable.any((s) => s.distanceMeters != null && s.distanceMeters! > 0);
  final includeHeartRate = usable.any((s) => s.heartRateBpm != null);
  final includeCadence = usable.any((s) => s.cadenceRpm != null);
  final includeWatts = usable.any((s) => s.watts != null);
  final includeElevation = usable.any((s) => s.elevation != null);
  final includeGrade = usable.any((s) => s.gradePercent != null);
  final includeTemperature = usable.any((s) => s.temperatureC != null);

  final startUtc = startDate.toUtc();
  final encoder = Encode();
  encoder.open();

  final createdFitTs = _toFitTimestamp(startUtc);
  final fileId = Mesg.fromMesgNum(MesgNum.fileId);
  fileId.setFieldValue(0, 4); // type = activity
  fileId.setFieldValue(1, 255); // manufacturer = development
  fileId.setFieldValue(2, 0); // product
  fileId.setFieldValue(4, createdFitTs);
  encoder.writeMesgDefinition(MesgDefinition.fromMesg(fileId));
  encoder.writeMesg(fileId);

  var lastTimestamp = createdFitTs;
  var hrSum = 0;
  var hrCount = 0;
  var hrMax = 0;
  var cadSum = 0;
  var cadCount = 0;
  var cadMax = 0;
  var wattsSum = 0;
  var wattsCount = 0;
  var wattsMax = 0;
  double? lastDistance;

  for (final sample in usable) {
    final seconds = sample.timeSeconds ?? 0;
    final ts = _toFitTimestamp(startUtc.add(Duration(seconds: seconds)));
    lastTimestamp = ts;

    final record = Mesg.fromMesgNum(MesgNum.record);
    record.setFieldValue(253, ts); // timestamp

    if (sample.hasGps) {
      record.setFieldValue(0, (sample.lat! * _semicirclesPerDegree).round());
      record.setFieldValue(1, (sample.lon! * _semicirclesPerDegree).round());
    }
    if (includeElevation && sample.elevation != null) {
      record.setFieldValue(2, sample.elevation); // altitude m (scaled by SDK)
    }
    if (includeHeartRate && sample.heartRateBpm != null) {
      final bpm = sample.heartRateBpm!.clamp(0, 254);
      record.setFieldValue(3, bpm);
      hrSum += bpm;
      hrCount++;
      if (bpm > hrMax) {
        hrMax = bpm;
      }
    }
    if (includeCadence && sample.cadenceRpm != null) {
      final cad = sample.cadenceRpm!.clamp(0, 254);
      record.setFieldValue(4, cad);
      cadSum += cad;
      cadCount++;
      if (cad > cadMax) {
        cadMax = cad;
      }
    }
    if (includeDistance && sample.distanceMeters != null) {
      record.setFieldValue(5, sample.distanceMeters);
      lastDistance = sample.distanceMeters;
    }
    if (includeSpeed && sample.speedMps != null) {
      // Keep in-series zeros when the channel is included.
      record.setFieldValue(6, sample.speedMps);
    }
    if (includeWatts && sample.watts != null) {
      final w = sample.watts!.clamp(0, 65534);
      record.setFieldValue(7, w);
      wattsSum += w;
      wattsCount++;
      if (w > wattsMax) {
        wattsMax = w;
      }
    }
    if (includeGrade && sample.gradePercent != null) {
      record.setFieldValue(9, sample.gradePercent);
    }
    if (includeTemperature && sample.temperatureC != null) {
      record.setFieldValue(13, sample.temperatureC);
    }

    encoder.writeMesgDefinition(MesgDefinition.fromMesg(record));
    encoder.writeMesg(record);
  }

  final sessionStart = createdFitTs;
  final sessionEnd = lastTimestamp;
  final elapsed = elapsedSeconds ??
      (sessionEnd > sessionStart ? sessionEnd - sessionStart : usable.length - 1)
          .clamp(0, 86400 * 7);
  final moving = (movingSeconds ?? elapsed).clamp(0, elapsed);

  final avgHr = averageHeartrate?.round() ??
      (hrCount > 0 ? (hrSum / hrCount).round() : null);
  final maxHr = maxHeartrate?.round() ?? (hrCount > 0 ? hrMax : null);
  final avgCad = averageCadence?.round() ??
      (cadCount > 0 ? (cadSum / cadCount).round() : null);
  final maxCad = maxCadence?.round() ?? (cadCount > 0 ? cadMax : null);
  final avgPower = averageWatts?.round() ??
      (wattsCount > 0 ? (wattsSum / wattsCount).round() : null);
  final maxPower = maxWatts?.round() ?? (wattsCount > 0 ? wattsMax : null);
  final dist = distanceMeters ?? lastDistance;
  final kcal = calories?.round();

  final session = Mesg.fromMesgNum(MesgNum.session);
  session.setFieldValue(253, sessionEnd); // timestamp
  session.setFieldValue(2, sessionStart); // start_time
  session.setFieldValue(7, elapsed.toDouble()); // total_elapsed_time (s)
  session.setFieldValue(8, moving.toDouble()); // total_timer_time (s)
  if (dist != null && dist > 0) {
    session.setFieldValue(9, dist);
  }
  if (kcal != null && kcal > 0) {
    session.setFieldValue(11, kcal);
  }
  if (avgHr != null && avgHr > 0) {
    session.setFieldValue(16, avgHr.clamp(0, 254));
  }
  if (maxHr != null && maxHr > 0) {
    session.setFieldValue(17, maxHr.clamp(0, 254));
  }
  if (avgCad != null && avgCad > 0) {
    session.setFieldValue(18, avgCad.clamp(0, 254));
  }
  if (maxCad != null && maxCad > 0) {
    session.setFieldValue(19, maxCad.clamp(0, 254));
  }
  if (avgPower != null && avgPower > 0) {
    session.setFieldValue(20, avgPower.clamp(0, 65534));
  }
  if (maxPower != null && maxPower > 0) {
    session.setFieldValue(21, maxPower.clamp(0, 65534));
  }
  encoder.writeMesgDefinition(MesgDefinition.fromMesg(session));
  encoder.writeMesg(session);

  final activity = Mesg.fromMesgNum(MesgNum.activity);
  activity.setFieldValue(253, sessionEnd);
  activity.setFieldValue(0, elapsed.toDouble()); // total_timer_time
  activity.setFieldValue(1, 1); // num_sessions
  activity.setFieldValue(2, 0); // type = manual
  activity.setFieldValue(3, 0); // event = timer
  activity.setFieldValue(4, 1); // event_type = stop
  encoder.writeMesgDefinition(MesgDefinition.fromMesg(activity));
  encoder.writeMesg(activity);

  return encoder.close();
}

/// Whether [samples] are worth attaching as a FIT track.
bool stravaSamplesWorthTrack(List<StravaStreamSample> samples) {
  final useful = samples.where((s) => s.isUseful).length;
  return useful >= 2;
}

int _toFitTimestamp(DateTime utc) {
  return (utc.toUtc().millisecondsSinceEpoch ~/ 1000) - _fitEpochOffsetSeconds;
}
