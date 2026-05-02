import 'package:flutter/material.dart';
import 'screens/discovery_screen.dart';
import 'screens/streaming_screen.dart';

void main() {
  runApp(HelixPlayApp());
}

class HelixPlayApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'HelixPlay',
      theme: ThemeData.dark(),
      initialRoute: '/',
      routes: {
        '/': (context) => DiscoveryScreen(),
        '/stream': (context) => StreamingScreen(),
      },
    );
  }
}
