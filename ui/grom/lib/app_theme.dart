import 'package:flutter/material.dart';

const kSeedColor = Color(0xFF2682B9);
const kAutoPauseColor = Color(0xFFB8860B);
const kWorkoutTrackColor = Color(0xFF1976D2); // Material Blue 700

ThemeData buildAppTheme() => ThemeData(
      colorScheme: ColorScheme.fromSeed(seedColor: kSeedColor),
      useMaterial3: true,
    );
