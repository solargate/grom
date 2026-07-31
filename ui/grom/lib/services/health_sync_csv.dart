import 'health_sync_sport_map.dart';

class HealthSyncCsvRow {
  const HealthSyncCsvRow({
    required this.sourceApp,
    required this.sport,
    required this.name,
    required this.startDate,
    required this.durationTotalSeconds,
    required this.durationMovingSeconds,
    required this.distanceKm,
  });

  final String sourceApp;
  final String sport;
  final String name;
  final DateTime startDate;
  final int durationTotalSeconds;
  final int durationMovingSeconds;
  final double distanceKm;

  String get externalIdName => 'health-sync/${sourceApp.trim().toLowerCase()}';

  String get sportTypeId => mapHealthSyncSportType(sport);
}

List<String> parseCsvLine(String line) {
  final fields = <String>[];
  final buffer = StringBuffer();
  var inQuotes = false;

  for (var i = 0; i < line.length; i++) {
    final char = line[i];
    if (char == '"') {
      if (inQuotes && i + 1 < line.length && line[i + 1] == '"') {
        buffer.write('"');
        i++;
      } else {
        inQuotes = !inQuotes;
      }
      continue;
    }
    if (char == ',' && !inQuotes) {
      fields.add(buffer.toString());
      buffer.clear();
      continue;
    }
    buffer.write(char);
  }
  fields.add(buffer.toString());
  return fields;
}

DateTime parseHealthSyncStartDate(String raw) {
  final trimmed = raw.trim();
  final parts = trimmed.split(RegExp(r'\s+'));
  if (parts.length < 2) {
    throw FormatException('invalid health sync date: $raw');
  }

  final dateParts = parts[0].split('.');
  if (dateParts.length != 3) {
    throw FormatException('invalid health sync date: $raw');
  }

  final timeParts = parts[1].split(':');
  if (timeParts.length < 2) {
    throw FormatException('invalid health sync time: $raw');
  }

  return DateTime(
    int.parse(dateParts[0]),
    int.parse(dateParts[1]),
    int.parse(dateParts[2]),
    int.parse(timeParts[0]),
    int.parse(timeParts[1]),
    timeParts.length > 2 ? int.parse(timeParts[2]) : 0,
  );
}

int _parseIntField(String raw, {required String fieldName}) {
  final trimmed = raw.trim();
  if (trimmed.isEmpty) {
    throw FormatException('missing $fieldName');
  }
  final value = int.tryParse(trimmed);
  if (value == null) {
    throw FormatException('invalid $fieldName: $raw');
  }
  return value;
}

double _parseDoubleField(String raw, {required String fieldName}) {
  final trimmed = raw.trim();
  if (trimmed.isEmpty) {
    throw FormatException('missing $fieldName');
  }
  final normalized = trimmed.replaceAll(',', '.');
  final value = double.tryParse(normalized);
  if (value == null) {
    throw FormatException('invalid $fieldName: $raw');
  }
  return value;
}

HealthSyncCsvRow parseHealthSyncCsvRow(List<String> fields) {
  if (fields.length < 8) {
    throw FormatException('expected at least 8 columns, got ${fields.length}');
  }

  return HealthSyncCsvRow(
    sourceApp: fields[0].trim(),
    sport: fields[1].trim(),
    name: fields[2].trim(),
    startDate: parseHealthSyncStartDate(fields[3]),
    durationTotalSeconds: _parseIntField(fields[5], fieldName: 'duration total'),
    durationMovingSeconds: _parseIntField(fields[6], fieldName: 'duration moving'),
    distanceKm: _parseDoubleField(fields[7], fieldName: 'distance km'),
  );
}

HealthSyncCsvRow? parseHealthSyncCsvContent(String content) {
  final lines = content
      .split('\n')
      .map((line) => line.trim())
      .where((line) => line.isNotEmpty)
      .toList();
  if (lines.length < 2) {
    return null;
  }

  final fields = parseCsvLine(lines[1]);
  return parseHealthSyncCsvRow(fields);
}

final _csvFilenamePattern = RegExp(
  r'^(\S+)\s+(\d{4}\.\d{2}\.\d{2})\s+(\d{2}\.\d{2})\s+(.+)\.csv$',
  caseSensitive: false,
);

String? matchHealthSyncTrackFilename(
  String csvFilename,
  Iterable<String> availableFilenames,
) {
  final match = _csvFilenamePattern.firstMatch(csvFilename.trim());
  if (match == null) {
    return null;
  }

  final sport = match.group(1)!;
  final date = match.group(2)!;
  final time = match.group(3)!;
  final base = '$date $time-$sport';

  final available = availableFilenames.toSet();
  final fitName = '$base.fit';
  if (available.contains(fitName)) {
    return fitName;
  }

  final gpxName = '$base.gpx';
  if (available.contains(gpxName)) {
    return gpxName;
  }

  return null;
}

bool isHealthSyncCsvFilename(String filename) {
  return _csvFilenamePattern.hasMatch(filename.trim());
}
