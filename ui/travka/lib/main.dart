import 'package:flutter/material.dart';
import 'api_request.dart';
import 'auth_storage.dart';
import 'login.dart';
import 'registration.dart';

void main() => runApp(const TravkaApp());

class TravkaApp extends StatelessWidget {
  const TravkaApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Travka',
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(
            seedColor: const Color.fromARGB(255, 45, 148, 49)),
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
  String? _nickname;
  int _counter = 0;

  @override
  void initState() {
    super.initState();
    _loadInitialData();
  }

  Future<void> _loadInitialData() async {
    final name = await _api.getServerInfo();
    final nickname = await AuthStorage.getNickname();
    if (!mounted) return;
    setState(() {
      _title = name;
      _nickname = nickname;
    });
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

  Future<void> _onLoginButtonPressed() async {
    final user = await Navigator.push<UserInfo>(
      context,
      MaterialPageRoute(
        builder: (context) => LoginPage(
          onLoggedIn: () {
            _loadInitialData();
          },
        ),
      ),
    );

    if (user != null && mounted) {
      setState(() => _nickname = user.nickname);
    }
  }

  Future<void> _logout() async {
    await AuthStorage.clear();
    if (!mounted) return;
    setState(() => _nickname = null);
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(content: Text('Вы вышли из аккаунта')),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        backgroundColor: Theme.of(context).colorScheme.inversePrimary,
        title: Text(_nickname != null ? '$_title · $_nickname' : _title),
        actions: [
          if (_nickname != null)
            IconButton(
              tooltip: 'Выйти',
              onPressed: _logout,
              icon: const Icon(Icons.logout),
            ),
        ],
      ),
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: <Widget>[
            const Text(
              'You have pushed the button this many times:',
            ),
            Text(
              '$_counter',
              style: Theme.of(context).textTheme.headlineMedium,
            ),
            const SizedBox(height: 24),
            ElevatedButton(
              onPressed: _onRegisterButtonPressed,
              child: const Text('Регистрация'),
            ),
            const SizedBox(height: 12),
            if (_nickname == null)
              ElevatedButton(
                onPressed: _onLoginButtonPressed,
                child: const Text('Вход'),
              )
            else ...[
              Text(
                'Вы вошли как $_nickname',
                style: Theme.of(context).textTheme.bodyLarge,
              ),
              const SizedBox(height: 12),
              OutlinedButton.icon(
                onPressed: _logout,
                icon: const Icon(Icons.logout),
                label: const Text('Выйти'),
              ),
            ],
          ],
        ),
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: _incrementCounter,
        tooltip: 'Increment',
        child: const Icon(Icons.add),
      ),
    );
  }
}
