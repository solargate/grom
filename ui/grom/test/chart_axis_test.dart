import 'package:flutter_test/flutter_test.dart';
import 'package:grom/widgets/chart_axis.dart';

void main() {
  group('computeChartYAxisBounds', () {
    test('uses min minus padding when above zero', () {
      final bounds = computeChartYAxisBounds(minSeriesY: 120, maxSeriesY: 160);
      expect(bounds.bottom, 115);
      expect(bounds.top, closeTo(176, 0.001));
    });

    test('clamps bottom at zero when min is below padding', () {
      final bounds = computeChartYAxisBounds(minSeriesY: 3, maxSeriesY: 20);
      expect(bounds.bottom, 0);
      expect(bounds.top, closeTo(22, 0.001));
    });

    test('uses minimum top when max is zero or negative', () {
      final zeroMax = computeChartYAxisBounds(minSeriesY: 0, maxSeriesY: 0);
      expect(zeroMax.bottom, 0);
      expect(zeroMax.top, 1);

      final flat = computeChartYAxisBounds(minSeriesY: 10, maxSeriesY: 10);
      expect(flat.bottom, 5);
      expect(flat.top, closeTo(11, 0.001));
    });

    test('extends top when it would not exceed bottom', () {
      final bounds = computeChartYAxisBounds(minSeriesY: 0.5, maxSeriesY: 0.5);
      expect(bounds.bottom, 0);
      expect(bounds.top, closeTo(0.55, 0.001));
    });

    test('supports custom padding', () {
      final bounds = computeChartYAxisBounds(
        minSeriesY: 50,
        maxSeriesY: 80,
        padding: 10,
      );
      expect(bounds.bottom, 40);
    });
  });
}
