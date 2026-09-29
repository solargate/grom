import 'dart:typed_data';

import 'package:fit_sdk/fit_sdk.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:grom/services/strava_api_fit.dart';
import 'package:grom/services/strava_api_streams.dart';

void main() {
  test('buildFitFromStravaStreams writes HR without GPS', () {
    final bytes = buildFitFromStravaStreams(
      startDate: DateTime.utc(2026, 9, 5, 10),
      samples: const [
        StravaStreamSample(timeSeconds: 0, heartRateBpm: 110),
        StravaStreamSample(timeSeconds: 30, heartRateBpm: 125),
        StravaStreamSample(timeSeconds: 60, heartRateBpm: 140),
      ],
      averageHeartrate: 125,
      maxHeartrate: 140,
    );

    final hrs = <int>[];
    var sawPosition = false;
    final decoder = Decode();
    decoder.onMesg = (Mesg mesg) {
      if (mesg.num != MesgNum.record) {
        return;
      }
      final lat = mesg.getFieldValue(0);
      final lon = mesg.getFieldValue(1);
      if (lat != null || lon != null) {
        sawPosition = true;
      }
      final hr = mesg.getFieldValue(3);
      if (hr is num) {
        hrs.add(hr.toInt());
      }
    };
    decoder.read(Uint8List.fromList(bytes));

    expect(sawPosition, isFalse);
    expect(hrs, [110, 125, 140]);
  });

  test('buildFitFromStravaStreams includes GPS speed and keeps gap HR', () {
    final bytes = buildFitFromStravaStreams(
      startDate: DateTime.utc(2026, 9, 5, 10),
      samples: const [
        StravaStreamSample(
          timeSeconds: 0,
          lat: 55.75,
          lon: 37.61,
          heartRateBpm: 120,
          speedMps: 2.0,
        ),
        StravaStreamSample(
          timeSeconds: 30,
          heartRateBpm: 130,
          speedMps: 0,
        ),
        StravaStreamSample(
          timeSeconds: 60,
          lat: 55.76,
          lon: 37.62,
          heartRateBpm: 145,
          speedMps: 2.5,
        ),
      ],
    );

    final hrs = <int>[];
    var gpsCount = 0;
    final decoder = Decode();
    decoder.onMesg = (Mesg mesg) {
      if (mesg.num != MesgNum.record) {
        return;
      }
      final lat = mesg.getFieldValue(0);
      if (lat != null) {
        gpsCount++;
      }
      final hr = mesg.getFieldValue(3);
      if (hr is num) {
        hrs.add(hr.toInt());
      }
    };
    decoder.read(Uint8List.fromList(bytes));

    expect(gpsCount, 2);
    expect(hrs, [120, 130, 145]);
  });

  test('buildFitFromStravaStreams rejects short series', () {
    expect(
      () => buildFitFromStravaStreams(
        startDate: DateTime.utc(2026, 9, 5),
        samples: const [
          StravaStreamSample(timeSeconds: 0, heartRateBpm: 100),
        ],
      ),
      throwsArgumentError,
    );
  });

  test('stravaSamplesWorthTrack requires two useful samples', () {
    expect(
      stravaSamplesWorthTrack(const [
        StravaStreamSample(timeSeconds: 0, heartRateBpm: 100),
      ]),
      isFalse,
    );
    expect(
      stravaSamplesWorthTrack(const [
        StravaStreamSample(timeSeconds: 0, heartRateBpm: 100),
        StravaStreamSample(timeSeconds: 1, heartRateBpm: 110),
      ]),
      isTrue,
    );
  });
}
