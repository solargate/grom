import 'dart:math' as math;

/// Y-axis bounds for workout speed / heart-rate charts.
class ChartYAxisBounds {
  const ChartYAxisBounds({required this.bottom, required this.top});

  final double bottom;
  final double top;
}

/// Computes chart Y bounds: bottom at [max(0, minSeriesY - padding)], top at
/// maxSeriesY * 1.1 (minimum 1 when max ≤ 0).
ChartYAxisBounds computeChartYAxisBounds({
  required double minSeriesY,
  required double maxSeriesY,
  double padding = 5,
}) {
  final yBottom = math.max(0.0, minSeriesY - padding);
  var yTop = maxSeriesY <= 0 ? 1.0 : maxSeriesY * 1.1;
  if (yTop <= yBottom) {
    yTop = yBottom + 1;
  }
  return ChartYAxisBounds(bottom: yBottom, top: yTop);
}

/// Y bounds for a negated pace series (faster / smaller sec/km plots higher).
///
/// [minSeriesY] / [maxSeriesY] are negated pace values (e.g. −360…−300).
/// Expands the slow end downward and the fast end upward by [paddingSec].
ChartYAxisBounds computeInvertedPaceYAxisBounds({
  required double minSeriesY,
  required double maxSeriesY,
  double paddingSec = 30,
}) {
  final yBottom = minSeriesY - paddingSec;
  var yTop = maxSeriesY + paddingSec;
  if (yTop >= 0) {
    yTop = math.min(-1.0, maxSeriesY / 2);
  }
  if (yTop <= yBottom) {
    yTop = yBottom + 1;
  }
  return ChartYAxisBounds(bottom: yBottom, top: yTop);
}
