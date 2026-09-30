import 'dart:math' as math;

import 'package:fl_chart/fl_chart.dart';
import 'package:flutter/material.dart';
import 'package:grom/l10n/app_localizations.dart';
import 'package:grom/l10n/sport_type_localizations.dart';
import 'package:grom/models/workout_pace.dart';

import 'chart_axis.dart';
import 'workout_map_preview.dart';
import 'workout_speed_chart.dart';

const double kWorkoutPaceChartHeight = kWorkoutSpeedChartHeight;

/// Pace chart uses the same accent as [kWorkoutSpeedChartColor].
const Color kWorkoutPaceChartColor = kWorkoutSpeedChartColor;

class WorkoutPaceChart extends StatelessWidget {
  const WorkoutPaceChart({
    super.key,
    required this.samples,
    this.paceAvgSec,
    this.paceBestSec,
  });

  final List<WorkoutPaceSample> samples;
  final double? paceAvgSec;
  final double? paceBestSec;

  @override
  Widget build(BuildContext context) {
    if (samples.length < 2) {
      return const SizedBox.shrink();
    }

    final l10n = AppLocalizations.of(context)!;
    final theme = Theme.of(context);

    // Negate pace so faster (smaller sec/km) plots higher on the Y axis.
    final spots = <FlSpot>[
      for (final s in samples) FlSpot(s.distanceKm, -s.paceSecPerKm),
    ];

    var minX = spots.first.x;
    var maxX = spots.first.x;
    var minSeriesY = spots.first.y;
    var maxSeriesY = spots.first.y;
    for (final spot in spots) {
      if (spot.x < minX) minX = spot.x;
      if (spot.x > maxX) maxX = spot.x;
      if (spot.y < minSeriesY) minSeriesY = spot.y;
      if (spot.y > maxSeriesY) maxSeriesY = spot.y;
    }
    if (maxX <= minX) {
      maxX = minX + 0.1;
    }

    final yAxis = computeInvertedPaceYAxisBounds(
      minSeriesY: minSeriesY,
      maxSeriesY: maxSeriesY,
    );
    final yBottom = yAxis.bottom;
    final yTop = yAxis.top;
    final yInterval = (yTop - yBottom) / 4;

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
                  l10n.workoutPace,
                  style: theme.textTheme.titleMedium,
                ),
                const SizedBox(height: 8),
                SizedBox(
                  height: kWorkoutPaceChartHeight,
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
                            reservedSize: 44,
                            interval: yInterval,
                            getTitlesWidget: (value, meta) {
                              if (value <= yBottom || value >= yTop) {
                                return const SizedBox.shrink();
                              }
                              return Text(
                                formatPaceMmSs(-value),
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
                              final text = value >= 10
                                  ? value.toStringAsFixed(1)
                                  : value.toStringAsFixed(2);
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
                              final paceText = formatPaceWithUnit(
                                l10n,
                                -spot.y,
                              );
                              final distanceText = formatDistanceKm(
                                l10n,
                                spot.x * 1000,
                              );
                              return LineTooltipItem(
                                '$paceText\n$distanceText',
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
                                color: kWorkoutPaceChartColor.withValues(
                                  alpha: 0.6,
                                ),
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
                          color: kWorkoutPaceChartColor,
                          barWidth: 2,
                          isStrokeCapRound: true,
                          dotData: const FlDotData(show: false),
                          belowBarData: BarAreaData(
                            show: true,
                            color: kWorkoutPaceChartColor,
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
                if (paceAvgSec != null && paceAvgSec! > 0) ...[
                  const SizedBox(height: 12),
                  _PaceStatRow(
                    label: l10n.workoutPaceAvg,
                    value: formatPaceWithUnit(l10n, paceAvgSec!),
                  ),
                ],
                if (paceBestSec != null && paceBestSec! > 0) ...[
                  const SizedBox(height: 4),
                  _PaceStatRow(
                    label: l10n.workoutPaceBest,
                    value: formatPaceWithUnit(l10n, paceBestSec!),
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

class _PaceStatRow extends StatelessWidget {
  const _PaceStatRow({
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
