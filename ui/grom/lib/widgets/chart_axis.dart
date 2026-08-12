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
