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

  group('computeInvertedPaceYAxisBounds', () {
    test('pads slow and fast ends of negated pace series', () {
      // -360 (6:00) … -300 (5:00)
      final bounds = computeInvertedPaceYAxisBounds(
        minSeriesY: -360,
        maxSeriesY: -300,
      );
      expect(bounds.bottom, -390);
      expect(bounds.top, -270);
    });

    test('clamps fast end when padding would reach or cross zero', () {
      final bounds = computeInvertedPaceYAxisBounds(
        minSeriesY: -40,
        maxSeriesY: -10,
      );
      expect(bounds.bottom, -70);
      // maxSeriesY + padding = 20 → clamp to min(-1, -5) = -5
      expect(bounds.top, -5);
    });

    test('extends top when flat series would collapse', () {
      final bounds = computeInvertedPaceYAxisBounds(
        minSeriesY: -10,
        maxSeriesY: -10,
        paddingSec: 0,
      );
      expect(bounds.bottom, -10);
      expect(bounds.top, -9);
    });

    test('supports custom padding', () {
      final bounds = computeInvertedPaceYAxisBounds(
        minSeriesY: -360,
        maxSeriesY: -300,
        paddingSec: 10,
      );
      expect(bounds.bottom, -370);
      expect(bounds.top, -290);
    });
  });
}
