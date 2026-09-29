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
    var sawSpeed = false;
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
      if (mesg.getFieldValue(6) != null) {
        sawSpeed = true;
      }
      final hr = mesg.getFieldValue(3);
      if (hr is num) {
        hrs.add(hr.toInt());
      }
    };
    decoder.read(Uint8List.fromList(bytes));

    expect(sawPosition, isFalse);
    expect(sawSpeed, isFalse);
    expect(hrs, [110, 125, 140]);
  });

  test('buildFitFromStravaStreams omits all-zero speed stream', () {
    final bytes = buildFitFromStravaStreams(
      startDate: DateTime.utc(2026, 9, 5, 10),
      samples: const [
        StravaStreamSample(timeSeconds: 0, heartRateBpm: 110, speedMps: 0),
        StravaStreamSample(timeSeconds: 30, heartRateBpm: 120, speedMps: 0),
        StravaStreamSample(timeSeconds: 60, heartRateBpm: 130, speedMps: 0),
      ],
    );

    var sawSpeed = false;
    final decoder = Decode();
    decoder.onMesg = (Mesg mesg) {
      if (mesg.num == MesgNum.record && mesg.getFieldValue(6) != null) {
        sawSpeed = true;
      }
    };
    decoder.read(Uint8List.fromList(bytes));
    expect(sawSpeed, isFalse);
  });

  test('buildFitFromStravaStreams keeps in-series zero speed when channel used',
      () {
    final bytes = buildFitFromStravaStreams(
      startDate: DateTime.utc(2026, 9, 5, 10),
      samples: const [
        StravaStreamSample(timeSeconds: 0, speedMps: 2.0),
        StravaStreamSample(timeSeconds: 1, speedMps: 0),
        StravaStreamSample(timeSeconds: 2, speedMps: 3.0),
      ],
    );

    final speeds = <double>[];
    final decoder = Decode();
    decoder.onMesg = (Mesg mesg) {
      if (mesg.num != MesgNum.record) {
        return;
      }
      final speed = mesg.getFieldValue(6);
      if (speed is num) {
        speeds.add(speed.toDouble());
      }
    };
    decoder.read(Uint8List.fromList(bytes));
    expect(speeds, hasLength(3));
    expect(speeds[0], closeTo(2.0, 0.01));
    expect(speeds[1], closeTo(0.0, 0.01));
    expect(speeds[2], closeTo(3.0, 0.01));
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

  test('buildFitFromStravaStreams embeds Strava device as product_name', () {
    final bytes = buildFitFromStravaStreams(
      startDate: DateTime.utc(2026, 9, 5, 10),
      deviceName: 'Garmin Edge 530',
      samples: const [
        StravaStreamSample(timeSeconds: 0, heartRateBpm: 110),
        StravaStreamSample(timeSeconds: 30, heartRateBpm: 125),
      ],
    );

    String? productName;
    int? manufacturer;
    final decoder = Decode();
    decoder.onMesg = (Mesg mesg) {
      if (mesg.num != MesgNum.fileId) {
        return;
      }
      manufacturer = mesg.getFieldValue(1) as int?;
      final name = mesg.getFieldValue(8);
      if (name != null) {
        productName = name.toString();
      }
    };
    decoder.read(Uint8List.fromList(bytes));

    expect(manufacturer, isNull);
    expect(productName, 'Garmin Edge 530');
  });

  test('buildFitFromStravaStreams omits development manufacturer without device',
      () {
    final bytes = buildFitFromStravaStreams(
      startDate: DateTime.utc(2026, 9, 5, 10),
      samples: const [
        StravaStreamSample(timeSeconds: 0, heartRateBpm: 110),
        StravaStreamSample(timeSeconds: 30, heartRateBpm: 125),
      ],
    );

    int? manufacturer;
    String? productName;
    final decoder = Decode();
    decoder.onMesg = (Mesg mesg) {
      if (mesg.num != MesgNum.fileId) {
        return;
      }
      manufacturer = mesg.getFieldValue(1) as int?;
      final name = mesg.getFieldValue(8);
      if (name != null) {
        productName = name.toString();
      }
    };
    decoder.read(Uint8List.fromList(bytes));

    expect(manufacturer, isNot(255));
    expect(productName, isNull);
  });

  test('buildFitFromStravaStreams omits product_name for blank deviceName', () {
    final bytes = buildFitFromStravaStreams(
      startDate: DateTime.utc(2026, 9, 5, 10),
      deviceName: '   ',
      samples: const [
        StravaStreamSample(timeSeconds: 0, heartRateBpm: 110),
        StravaStreamSample(timeSeconds: 30, heartRateBpm: 125),
      ],
    );

    String? productName;
    final decoder = Decode();
    decoder.onMesg = (Mesg mesg) {
      if (mesg.num != MesgNum.fileId) {
        return;
      }
      final name = mesg.getFieldValue(8);
      if (name != null) {
        productName = name.toString();
      }
    };
    decoder.read(Uint8List.fromList(bytes));
    expect(productName, isNull);
  });

  test('buildFitFromStravaStreams session uses activity summary metrics', () {
    final bytes = buildFitFromStravaStreams(
      startDate: DateTime.utc(2026, 9, 5, 10),
      samples: const [
        StravaStreamSample(timeSeconds: 0, heartRateBpm: 100, cadenceRpm: 70),
        StravaStreamSample(timeSeconds: 60, heartRateBpm: 120, cadenceRpm: 90),
      ],
      elapsedSeconds: 620,
      movingSeconds: 600,
      distanceMeters: 2500,
      calories: 312,
      averageHeartrate: 140,
      maxHeartrate: 155,
      averageCadence: 85,
      maxCadence: 95,
      averageWatts: 200,
      maxWatts: 280,
    );

    num? elapsed;
    num? moving;
    num? distance;
    num? calories;
    num? avgHr;
    num? maxHr;
    num? avgCad;
    num? maxCad;
    num? avgPower;
    num? maxPower;
    final decoder = Decode();
    decoder.onMesg = (Mesg mesg) {
      if (mesg.num != MesgNum.session) {
        return;
      }
      elapsed = mesg.getFieldValue(7) as num?;
      moving = mesg.getFieldValue(8) as num?;
      distance = mesg.getFieldValue(9) as num?;
      calories = mesg.getFieldValue(11) as num?;
      avgHr = mesg.getFieldValue(16) as num?;
      maxHr = mesg.getFieldValue(17) as num?;
      avgCad = mesg.getFieldValue(18) as num?;
      maxCad = mesg.getFieldValue(19) as num?;
      avgPower = mesg.getFieldValue(20) as num?;
      maxPower = mesg.getFieldValue(21) as num?;
    };
    decoder.read(Uint8List.fromList(bytes));

    expect(elapsed, closeTo(620, 0.01));
    expect(moving, closeTo(600, 0.01));
    expect(distance, closeTo(2500, 0.1));
    expect(calories?.toInt(), 312);
    expect(avgHr?.toInt(), 140);
    expect(maxHr?.toInt(), 155);
    expect(avgCad?.toInt(), 85);
    expect(maxCad?.toInt(), 95);
    expect(avgPower?.toInt(), 200);
    expect(maxPower?.toInt(), 280);
  });

  test('buildFitFromStravaStreams prefers activity averages over samples', () {
    final bytes = buildFitFromStravaStreams(
      startDate: DateTime.utc(2026, 9, 5, 10),
      samples: const [
        StravaStreamSample(timeSeconds: 0, heartRateBpm: 100),
        StravaStreamSample(timeSeconds: 60, heartRateBpm: 200),
      ],
      averageHeartrate: 125,
      maxHeartrate: 160,
    );

    num? avgHr;
    num? maxHr;
    final decoder = Decode();
    decoder.onMesg = (Mesg mesg) {
      if (mesg.num != MesgNum.session) {
        return;
      }
      avgHr = mesg.getFieldValue(16) as num?;
      maxHr = mesg.getFieldValue(17) as num?;
    };
    decoder.read(Uint8List.fromList(bytes));

    expect(avgHr?.toInt(), 125);
    expect(maxHr?.toInt(), 160);
  });

  test('buildFitFromStravaStreams omits all-zero distance stream', () {
    final bytes = buildFitFromStravaStreams(
      startDate: DateTime.utc(2026, 9, 5, 10),
      samples: const [
        StravaStreamSample(timeSeconds: 0, heartRateBpm: 110, distanceMeters: 0),
        StravaStreamSample(
            timeSeconds: 30, heartRateBpm: 120, distanceMeters: 0),
        StravaStreamSample(
            timeSeconds: 60, heartRateBpm: 130, distanceMeters: 0),
      ],
    );

    var sawDistance = false;
    num? sessionDistance;
    final decoder = Decode();
    decoder.onMesg = (Mesg mesg) {
      if (mesg.num == MesgNum.record && mesg.getFieldValue(5) != null) {
        sawDistance = true;
      }
      if (mesg.num == MesgNum.session) {
        sessionDistance = mesg.getFieldValue(9) as num?;
      }
    };
    decoder.read(Uint8List.fromList(bytes));
    expect(sawDistance, isFalse);
    expect(sessionDistance, isNull);
  });

  test('buildFitFromStravaStreams writes distance and session fallback', () {
    final bytes = buildFitFromStravaStreams(
      startDate: DateTime.utc(2026, 9, 5, 10),
      samples: const [
        StravaStreamSample(timeSeconds: 0, distanceMeters: 0),
        StravaStreamSample(timeSeconds: 30, distanceMeters: 50),
        StravaStreamSample(timeSeconds: 60, distanceMeters: 120),
      ],
    );

    final distances = <double>[];
    num? sessionDistance;
    final decoder = Decode();
    decoder.onMesg = (Mesg mesg) {
      if (mesg.num == MesgNum.record) {
        final d = mesg.getFieldValue(5);
        if (d is num) {
          distances.add(d.toDouble());
        }
      }
      if (mesg.num == MesgNum.session) {
        sessionDistance = mesg.getFieldValue(9) as num?;
      }
    };
    decoder.read(Uint8List.fromList(bytes));

    expect(distances, hasLength(3));
    expect(distances[0], closeTo(0, 0.01));
    expect(distances[1], closeTo(50, 0.1));
    expect(distances[2], closeTo(120, 0.1));
    expect(sessionDistance, closeTo(120, 0.1));
  });

  test('buildFitFromStravaStreams writes cadence watts and elevation', () {
    final bytes = buildFitFromStravaStreams(
      startDate: DateTime.utc(2026, 9, 5, 10),
      samples: const [
        StravaStreamSample(
          timeSeconds: 0,
          elevation: 100,
          cadenceRpm: 80,
          watts: 150,
        ),
        StravaStreamSample(
          timeSeconds: 30,
          elevation: 105,
          cadenceRpm: 90,
          watts: 200,
        ),
      ],
    );

    final cadences = <int>[];
    final watts = <int>[];
    final elevations = <double>[];
    final decoder = Decode();
    decoder.onMesg = (Mesg mesg) {
      if (mesg.num != MesgNum.record) {
        return;
      }
      final cad = mesg.getFieldValue(4);
      if (cad is num) {
        cadences.add(cad.toInt());
      }
      final w = mesg.getFieldValue(7);
      if (w is num) {
        watts.add(w.toInt());
      }
      final elev = mesg.getFieldValue(2);
      if (elev is num) {
        elevations.add(elev.toDouble());
      }
    };
    decoder.read(Uint8List.fromList(bytes));

    expect(cadences, [80, 90]);
    expect(watts, [150, 200]);
    expect(elevations, hasLength(2));
    expect(elevations[0], closeTo(100, 0.5));
    expect(elevations[1], closeTo(105, 0.5));
  });

  test('buildFitFromStravaStreams omits cadence and watts when all null', () {
    final bytes = buildFitFromStravaStreams(
      startDate: DateTime.utc(2026, 9, 5, 10),
      samples: const [
        StravaStreamSample(timeSeconds: 0, heartRateBpm: 110),
        StravaStreamSample(timeSeconds: 30, heartRateBpm: 120),
      ],
    );

    var sawCadence = false;
    var sawWatts = false;
    final decoder = Decode();
    decoder.onMesg = (Mesg mesg) {
      if (mesg.num != MesgNum.record) {
        return;
      }
      if (mesg.getFieldValue(4) != null) {
        sawCadence = true;
      }
      if (mesg.getFieldValue(7) != null) {
        sawWatts = true;
      }
    };
    decoder.read(Uint8List.fromList(bytes));
    expect(sawCadence, isFalse);
    expect(sawWatts, isFalse);
  });

  test('buildFitFromStravaStreams accepts time-only samples', () {
    // Sync gates on stravaSamplesWorthTrack (isUseful); the builder itself
    // still accepts timestamp-only rows so callers stay consistent.
    final bytes = buildFitFromStravaStreams(
      startDate: DateTime.utc(2026, 9, 5, 10),
      samples: const [
        StravaStreamSample(timeSeconds: 0),
        StravaStreamSample(timeSeconds: 30),
        StravaStreamSample(timeSeconds: 60),
      ],
    );

    var recordCount = 0;
    var sawSensor = false;
    final decoder = Decode();
    decoder.onMesg = (Mesg mesg) {
      if (mesg.num != MesgNum.record) {
        return;
      }
      recordCount++;
      for (final field in [0, 1, 2, 3, 4, 5, 6, 7, 9, 13]) {
        if (mesg.getFieldValue(field) != null) {
          sawSensor = true;
        }
      }
    };
    decoder.read(Uint8List.fromList(bytes));
    expect(recordCount, 3);
    expect(sawSensor, isFalse);
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

  test('stravaSamplesWorthTrack rejects time-only samples', () {
    expect(
      stravaSamplesWorthTrack(const [
        StravaStreamSample(timeSeconds: 0),
        StravaStreamSample(timeSeconds: 30),
        StravaStreamSample(timeSeconds: 60),
      ]),
      isFalse,
    );
  });

  test('stravaSamplesWorthTrack accepts GPS-only samples', () {
    expect(
      stravaSamplesWorthTrack(const [
        StravaStreamSample(timeSeconds: 0, lat: 55.75, lon: 37.61),
        StravaStreamSample(timeSeconds: 60, lat: 55.76, lon: 37.62),
      ]),
      isTrue,
    );
  });
}
