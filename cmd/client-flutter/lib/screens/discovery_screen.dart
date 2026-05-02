import 'package:flutter/material.dart';

class DiscoveryScreen extends StatefulWidget {
  @override
  _DiscoveryScreenState createState() => _DiscoveryScreenState();
}

class _DiscoveryScreenState extends State<DiscoveryScreen> {
  List<String> _hosts = [];

  @override
  void initState() {
    super.initState();
    _startDiscovery();
  }

  void _startDiscovery() {
    setState(() {
      _hosts = ['Host A', 'Host B'];
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text('Discover Hosts')),
      body: ListView.builder(
        itemCount: _hosts.length,
        itemBuilder: (context, index) {
          return ListTile(
            title: Text(_hosts[index]),
            onTap: () {
              Navigator.pushNamed(context, '/stream');
            },
          );
        },
      ),
    );
  }
}
