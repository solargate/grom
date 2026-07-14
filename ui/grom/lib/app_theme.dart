import 'package:flutter/material.dart';

const kSeedColor = Color(0xFF2682B9);
const kAutoPauseColor = Color(0xFFB8860B);

ThemeData buildAppTheme() => ThemeData(
      colorScheme: ColorScheme.fromSeed(seedColor: kSeedColor),
      useMaterial3: true,
    );
