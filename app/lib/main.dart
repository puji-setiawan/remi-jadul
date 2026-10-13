import 'package:flutter/material.dart';

void main() {
  runApp(const RemiJadulApp());
}

class RemiJadulApp extends StatelessWidget {
  const RemiJadulApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Remi Jadul',
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: const Color(0xFF123C35)),
        useMaterial3: true,
      ),
      home: const Scaffold(
        body: Center(child: Text('Remi Jadul')),
      ),
    );
  }
}
