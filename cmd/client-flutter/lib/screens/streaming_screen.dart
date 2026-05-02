import 'package:flutter/material.dart';

class StreamingScreen extends StatefulWidget {
  @override
  _StreamingScreenState createState() => _StreamingScreenState();
}

class _StreamingScreenState extends State<StreamingScreen> {
  bool _streaming = false;

  void _toggleStream() {
    setState(() {
      _streaming = !_streaming;
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text('Streaming')),
      body: Center(
        child: _streaming
            ? AspectRatio(
                aspectRatio: 16 / 9,
                child: Container(
                  color: Colors.black,
                  child: Center(
                    child: Text(
                      'Video Feed',
                      style: TextStyle(color: Colors.green),
                    ),
                  ),
                ),
              )
            : ElevatedButton(
                onPressed: _toggleStream,
                child: Text('Start Stream'),
              ),
      ),
    );
  }
}
