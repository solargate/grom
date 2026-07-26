import 'package:flutter/material.dart';

const kSeedColor = Color(0xFF2682B9);
const kAutoPauseColor = Color(0xFFB8860B);
const kWorkoutTrackColor = Color(0xFFF45E1E);

ThemeData buildAppTheme() => ThemeData(
      colorScheme: ColorScheme.fromSeed(seedColor: kSeedColor),
      useMaterial3: true,
    );
