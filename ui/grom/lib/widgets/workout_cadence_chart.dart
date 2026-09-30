import 'dart:math' as math;

import 'package:fl_chart/fl_chart.dart';
import 'package:flutter/material.dart';
import 'package:grom/l10n/app_localizations.dart';
import 'package:grom/l10n/sport_type_localizations.dart';
import 'package:grom/models/sport_types.dart';
import 'package:grom/models/workout_cadence.dart';

import 'chart_axis.dart';
import 'workout_map_preview.dart';

const Color kWorkoutCadenceChartColor = Color(0xFFB388FF);
const double kWorkoutCadenceChartHeight = 180;

class WorkoutCadenceChart extends StatelessWidget {
  const WorkoutCadenceChart({
    super.key,
    required this.samples,
    required this.hasGps,
    required this.sportType,
    this.cadenceAvg,
    this.cadenceMax,
  });

  final List<WorkoutCadenceSample> samples;
  final bool hasGps;
  final String sportType;
  final double? cadenceAvg;
  final double? cadenceMax;

  @override
  Widget build(BuildContext context) {
    if (!hasCadenceChart(sportType: sportType, samples: samples)) {
      return const SizedBox.shrink();
    }

    final l10n = AppLocalizations.of(context)!;
    final theme = Theme.of(context);
    final footUnits =
        sportTypeById(sportType)?.category == SportCategory.foot;
    final scale = footUnits ? 2.0 : 1.0;
    final spots = <FlSpot>[
      for (final s in samples)
        FlSpot(
          hasGps
              ? (s.distanceKm ?? 0)
              : cadenceMinutesFromSeriesStart(samples, s.time),
          s.cadence * scale,
        ),
    ];

    var minX = spots.first.x;
    var maxX = spots.first.x;
    var minSeriesY = spots.first.y;
    var maxY = spots.first.y;
    for (final spot in spots) {
      if (spot.x < minX) minX = spot.x;
      if (spot.x > maxX) maxX = spot.x;
      if (spot.y < minSeriesY) minSeriesY = spot.y;
      if (spot.y > maxY) maxY = spot.y;
    }
    if (maxX <= minX) {
      maxX = minX + 0.1;
    }
    final yAxis = computeChartYAxisBounds(
      minSeriesY: minSeriesY,
      maxSeriesY: maxY,
    );
    final yBottom = yAxis.bottom;
    final yTop = yAxis.top;
    final yInterval = (yTop - yBottom) / 4;

    final displayAvg = displayCadence(cadenceAvg, sportType);
    final displayMax = displayCadence(cadenceMax, sportType);

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
                  l10n.workoutCadenceChartTitle,
                  style: theme.textTheme.titleMedium,
                ),
                const SizedBox(height: 8),
                SizedBox(
                  height: kWorkoutCadenceChartHeight,
                  child: LineChart(
                    LineChartData(
                      minX: minX,
                      maxX: maxX,
                      minY: yBottom,
                      maxY: yTop,
                      clipData: const FlClipData.all(),
                      gridData: FlGridData(
                        show: true,
                        drawVerticalLine: false,
                        horizontalInterval: yInterval,
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
                            interval: yInterval,
                            getTitlesWidget: (value, meta) {
                              if (value <= yBottom || value >= yTop) {
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
                              final cadText = formatCadence(
                                l10n,
                                spot.y,
                                footUnits: footUnits,
                              );
                              final secondLine = hasGps
                                  ? formatDistanceKm(l10n, spot.x * 1000)
                                  : _formatMinutes(l10n, spot.x);
                              return LineTooltipItem(
                                '$cadText\n$secondLine',
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
                                color: kWorkoutCadenceChartColor
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
                          color: kWorkoutCadenceChartColor,
                          barWidth: 2,
                          isStrokeCapRound: true,
                          dotData: const FlDotData(show: false),
                          belowBarData: BarAreaData(
                            show: true,
                            color: kWorkoutCadenceChartColor,
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
                if (displayAvg != null) ...[
                  const SizedBox(height: 12),
                  _CadenceStatRow(
                    label: l10n.workoutCadenceAvg,
                    value: formatCadence(
                      l10n,
                      displayAvg,
                      footUnits: footUnits,
                    ),
                  ),
                ],
                if (displayMax != null) ...[
                  const SizedBox(height: 4),
                  _CadenceStatRow(
                    label: l10n.workoutCadenceMax,
                    value: formatCadence(
                      l10n,
                      displayMax,
                      footUnits: footUnits,
                    ),
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

class _CadenceStatRow extends StatelessWidget {
  const _CadenceStatRow({
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
