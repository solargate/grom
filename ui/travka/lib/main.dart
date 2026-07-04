import 'package:flutter/material.dart';
import 'api_request.dart';
import 'registration.dart';

void main() => runApp(const TravkaApp());

class TravkaApp extends StatelessWidget {
  const TravkaApp({super.key});

  @override
  Widget build(BuildContext context) {

    //late ApiRequest apiRequest;

    return MaterialApp(
      title: 'Travka',
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: const Color.fromARGB(255, 45, 148, 49)),
        useMaterial3: true,
      ),
      home: const TravkaHomePage(),
    );
  }
}

class TravkaHomePage extends StatefulWidget {
  const TravkaHomePage({super.key});

  @override
  State<TravkaHomePage> createState() => _TravkaHomePageState();
}

class _TravkaHomePageState extends State<TravkaHomePage> {
  final ApiRequest _api = ApiRequest();
  String _title = 'Travka Home';
  int _counter = 0;

  @override
  void initState() {
    super.initState();
    _loadServerName();
  }

  Future<void> _loadServerName() async {
    final name = await _api.getServerInfo();
    if (!mounted) return;
    setState(() => _title = name);
  }

  void _incrementCounter() {
    setState(() {
      _counter++;
    });
  }

  void _onRegisterButtonPressed() {
    Navigator.push(
      context,
      MaterialPageRoute(builder: (context) => const RegistrationPage()),
    );
  }

  @override
  Widget build(BuildContext context) {
    // This method is rerun every time setState is called, for instance as done
    // by the _incrementCounter method above.
    //
    // The Flutter framework has been optimized to make rerunning build methods
    // fast, so that you can just rebuild anything that needs updating rather
    // than having to individually change instances of widgets.
    return Scaffold(
      appBar: AppBar(
        // TRY THIS: Try changing the color here to a specific color (to
        // Colors.amber, perhaps?) and trigger a hot reload to see the AppBar
        // change color while the other colors stay the same.
        backgroundColor: Theme.of(context).colorScheme.inversePrimary,
        // Here we take the value from the MyHomePage object that was created by
        // the App.build method, and use it to set our appbar title.
        title: Text(_title),
      ),
      body: Center(
        // Center is a layout widget. It takes a single child and positions it
        // in the middle of the parent.
        child: Column(
          // Column is also a layout widget. It takes a list of children and
          // arranges them vertically. By default, it sizes itself to fit its
          // children horizontally, and tries to be as tall as its parent.
          //
          // Column has various properties to control how it sizes itself and
          // how it positions its children. Here we use mainAxisAlignment to
          // center the children vertically; the main axis here is the vertical
          // axis because Columns are vertical (the cross axis would be
          // horizontal).
          //
          // TRY THIS: Invoke "debug painting" (choose the "Toggle Debug Paint"
          // action in the IDE, or press "p" in the console), to see the
          // wireframe for each widget.
          mainAxisAlignment: MainAxisAlignment.center,
          children: <Widget>[
            const Text(
              'You have pushed the button this many times:',
            ),
            Text(
              '$_counter',
              style: Theme.of(context).textTheme.headlineMedium,
            ),
            ElevatedButton(
              onPressed: _onRegisterButtonPressed,
              child: const Text('Регистрация')),
          ],
        ),
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: _incrementCounter,
        tooltip: 'Increment',
        child: const Icon(Icons.add),
      ), // This trailing comma makes auto-formatting nicer for build methods.
    );
  }
}
