import 'dart:io' as io;

import 'package:fit_sdk/fit_sdk.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:latlong2/latlong.dart';
import 'package:grom/services/track_parser.dart';

void main() {
  test('parseTrackPoints reads GPX track points', () {
    final samplePath = io.Directory.current.path.contains('ui/grom')
        ? '../../internal/tracks/testdata/sample.gpx'
        : '../../../internal/tracks/testdata/sample.gpx';
    final bytes = io.File(samplePath).readAsBytesSync();
    final points = parseTrackPoints(bytes, 'sample.gpx');

    expect(points.length, 3);
    expect(points.first.latitude, closeTo(55.7558, 0.0001));
    expect(points.first.longitude, closeTo(37.6173, 0.0001));
  });

  test('parseTrackPoints reads FIT record messages', () {
    final bytes = _sampleFitBytes();
    final points = parseTrackPoints(bytes, 'track.fit');

    expect(points.length, 2);
    expect(points.first.latitude, closeTo(55.7558, 0.0001));
    expect(points.first.longitude, closeTo(37.6173, 0.0001));
  });

  test('simplifyForRender keeps endpoints for large tracks', () {
    final points = List<LatLng>.generate(
      1000,
      (index) => LatLng(55.0 + index * 0.0001, 37.0),
    );

    final simplified = simplifyForRender(points);

    expect(simplified.length, 500);
    expect(simplified.first, points.first);
    expect(simplified.last, points.last);
  });
}

List<int> _sampleFitBytes() {
  const fitEpochOffset = 631065600;
  final encoder = Encode();
  encoder.open();

  final fileIdMesg = Mesg.fromMesgNum(MesgNum.fileId);
  fileIdMesg.setFieldValue(0, 4);
  fileIdMesg.setFieldValue(1, 1);
  fileIdMesg.setFieldValue(
    4,
    (DateTime.utc(2026, 7, 6).millisecondsSinceEpoch ~/ 1000) - fitEpochOffset,
  );

  final fileIdDef = MesgDefinition.fromMesg(fileIdMesg);
  encoder.writeMesgDefinition(fileIdDef);
  encoder.writeMesg(fileIdMesg);

  final coordsList = [
    (55.7558, 37.6173),
    (55.7578, 37.6373),
  ];

  for (var i = 0; i < coordsList.length; i++) {
    final coords = coordsList[i];
    final recordMesg = Mesg.fromMesgNum(MesgNum.record);
    recordMesg.setFieldValue(
      253,
      (DateTime.utc(2026, 7, 6, 8, 40 + i).millisecondsSinceEpoch ~/ 1000) -
          fitEpochOffset,
    );
    recordMesg.setFieldValue(0, (coords.$1 * 11930464.7111).round());
    recordMesg.setFieldValue(1, (coords.$2 * 11930464.7111).round());
    recordMesg.setFieldValue(5, i * 100);

    if (i == 0) {
      encoder.writeMesgDefinition(MesgDefinition.fromMesg(recordMesg));
    }
    encoder.writeMesg(recordMesg);
  }

  return encoder.close();
}
