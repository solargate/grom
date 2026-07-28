import 'dart:math' as math;

import 'package:fl_chart/fl_chart.dart';
import 'package:flutter/material.dart';
import 'package:grom/l10n/app_localizations.dart';
import 'package:grom/l10n/sport_type_localizations.dart';
import 'package:grom/models/workout_heartrate.dart';
import 'package:grom/models/workout_stats.dart';

import 'workout_map_preview.dart';

const Color kWorkoutHeartRateChartColor = Color(0xFFB85C5C);
const double kWorkoutHeartRateChartHeight = 180;

class WorkoutHeartRateChart extends StatelessWidget {
  const WorkoutHeartRateChart({
    super.key,
    required this.samples,
    required this.hasGps,
    this.heartRateAvg,
    this.heartRateMax,
  });

  final List<WorkoutHeartRateSample> samples;
  final bool hasGps;
  final double? heartRateAvg;
  final double? heartRateMax;

  @override
  Widget build(BuildContext context) {
    if (samples.length < 2) {
      return const SizedBox.shrink();
    }

    final l10n = AppLocalizations.of(context)!;
    final theme = Theme.of(context);
    final spots = <FlSpot>[
      for (final s in samples)
        FlSpot(
          hasGps
              ? (s.distanceKm ?? 0)
              : minutesFromSeriesStart(samples, s.time),
          s.heartRateBpm,
        ),
    ];

    var minX = spots.first.x;
    var maxX = spots.first.x;
    var maxY = spots.first.y;
    for (final spot in spots) {
      if (spot.x < minX) minX = spot.x;
      if (spot.x > maxX) maxX = spot.x;
      if (spot.y > maxY) maxY = spot.y;
    }
    if (maxX <= minX) {
      maxX = minX + 0.1;
    }
    final yTop = maxY <= 0 ? 1.0 : maxY * 1.1;

    return LayoutBuilder(
      builder: (context, constraints) {
        final displayWidth = math.min(
          kWorkoutMapPreviewMaxWidth,
          constraints.maxWidth,
        );
        return Align(
          alignment: Alignment.centerLeft,
          child: SizedBox(
            width: displayWidth,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                Text(
                  l10n.workoutHeartRateChartTitle,
                  style: theme.textTheme.titleMedium,
                ),
                const SizedBox(height: 8),
                SizedBox(
                  height: kWorkoutHeartRateChartHeight,
                  child: LineChart(
                    LineChartData(
                      minX: minX,
                      maxX: maxX,
                      minY: 0,
                      maxY: yTop,
                      clipData: const FlClipData.all(),
                      gridData: FlGridData(
                        show: true,
                        drawVerticalLine: false,
                        horizontalInterval: yTop / 4,
                        getDrawingHorizontalLine: (value) => FlLine(
                          color: theme.dividerColor.withValues(alpha: 0.4),
                          strokeWidth: 1,
                        ),
                      ),
                      borderData: FlBorderData(
                        show: true,
                        border: Border(
                          bottom: BorderSide(color: theme.dividerColor),
                          left: BorderSide(color: theme.dividerColor),
                        ),
                      ),
                      titlesData: FlTitlesData(
                        topTitles: const AxisTitles(
                          sideTitles: SideTitles(showTitles: false),
                        ),
                        rightTitles: const AxisTitles(
                          sideTitles: SideTitles(showTitles: false),
                        ),
                        leftTitles: AxisTitles(
                          sideTitles: SideTitles(
                            showTitles: true,
                            reservedSize: 36,
                            interval: yTop / 4,
                            getTitlesWidget: (value, meta) {
                              if (value <= 0 || value >= yTop) {
                                return const SizedBox.shrink();
                              }
                              return Text(
                                value.toStringAsFixed(0),
                                style: theme.textTheme.bodySmall,
                              );
                            },
                          ),
                        ),
                        bottomTitles: AxisTitles(
                          sideTitles: SideTitles(
                            showTitles: true,
                            reservedSize: 24,
                            interval: (maxX - minX) / 4,
                            getTitlesWidget: (value, meta) {
                              if (value <= minX || value >= maxX) {
                                return const SizedBox.shrink();
                              }
                              final text = hasGps
                                  ? (value >= 10
                                      ? value.toStringAsFixed(1)
                                      : value.toStringAsFixed(2))
                                  : (value >= 10
                                      ? value.toStringAsFixed(0)
                                      : value.toStringAsFixed(1));
                              return Padding(
                                padding: const EdgeInsets.only(top: 4),
                                child: Text(
                                  text,
                                  style: theme.textTheme.bodySmall,
                                ),
                              );
                            },
                          ),
                        ),
                      ),
                      lineTouchData: LineTouchData(
                        handleBuiltInTouches: true,
                        touchTooltipData: LineTouchTooltipData(
                          fitInsideHorizontally: true,
                          fitInsideVertically: true,
                          getTooltipColor: (_) =>
                              theme.colorScheme.inverseSurface,
                          getTooltipItems: (touchedSpots) {
                            return touchedSpots.map((spot) {
                              final hrText = formatHeartRate(l10n, spot.y);
                              final secondLine = hasGps
                                  ? formatDistanceKm(l10n, spot.x * 1000)
                                  : _formatMinutes(l10n, spot.x);
                              return LineTooltipItem(
                                '$hrText\n$secondLine',
                                TextStyle(
                                  color: theme.colorScheme.onInverseSurface,
                                  fontWeight: FontWeight.w600,
                                  fontSize: 12,
                                ),
                              );
                            }).toList();
                          },
                        ),
                        getTouchedSpotIndicator: (barData, spotIndexes) {
                          return spotIndexes.map((index) {
                            return TouchedSpotIndicatorData(
                              FlLine(
                                color: kWorkoutHeartRateChartColor
                                    .withValues(alpha: 0.6),
                                strokeWidth: 1,
                              ),
                              const FlDotData(show: true),
                            );
                          }).toList();
                        },
                      ),
                      lineBarsData: [
                        LineChartBarData(
                          spots: spots,
                          isCurved: false,
                          color: kWorkoutHeartRateChartColor,
                          barWidth: 2,
                          isStrokeCapRound: true,
                          dotData: const FlDotData(show: false),
                          belowBarData: BarAreaData(
                            show: true,
                            color: kWorkoutHeartRateChartColor,
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
                if (heartRateAvg != null && heartRateAvg! > 0) ...[
                  const SizedBox(height: 12),
                  _HeartRateStatRow(
                    label: l10n.workoutHeartRateAvg,
                    value: formatHeartRate(l10n, heartRateAvg!),
                  ),
                ],
                if (heartRateMax != null && heartRateMax! > 0) ...[
                  const SizedBox(height: 4),
                  _HeartRateStatRow(
                    label: l10n.workoutHeartRateMax,
                    value: formatHeartRate(l10n, heartRateMax!),
                  ),
                ],
              ],
            ),
          ),
        );
      },
    );
  }
}

String _formatMinutes(AppLocalizations l10n, double minutes) {
  final text = minutes >= 10
      ? minutes.toStringAsFixed(0)
      : minutes.toStringAsFixed(1);
  return l10n.chartMinutes(text);
}

class _HeartRateStatRow extends StatelessWidget {
  const _HeartRateStatRow({
    required this.label,
    required this.value,
  });

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Row(
      children: [
        Expanded(
          child: Text(
            label,
            style: theme.textTheme.bodyMedium,
          ),
        ),
        Text(
          value,
          style: theme.textTheme.bodyMedium?.copyWith(
            fontWeight: FontWeight.bold,
          ),
        ),
      ],
    );
  }
}
