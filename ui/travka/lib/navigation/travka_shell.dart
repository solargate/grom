import 'package:flutter/material.dart';
import 'package:travka/l10n/app_localizations.dart';

import '../api_request.dart';
import '../auth_storage.dart';
import '../login.dart';
import '../pages/home_page.dart';
import '../registration.dart';
import 'travka_destination.dart';
import 'travka_side_menu.dart';

const kWideLayoutBreakpoint = 600.0;

class TravkaShell extends StatefulWidget {
  const TravkaShell({
    super.key,
    required this.locale,
    required this.onLocaleChanged,
  });

  final Locale locale;
  final ValueChanged<Locale> onLocaleChanged;

  @override
  State<TravkaShell> createState() => _TravkaShellState();
}

class _TravkaShellState extends State<TravkaShell> {
  final ApiRequest _api = ApiRequest();

  String _title = 'Travka Home';
  String? _nickname;
  TravkaDestination _selectedDestination = TravkaDestination.home;

  bool get _isLoggedIn => _nickname != null;

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

  void _onDestinationSelected(TravkaDestination destination) {
    setState(() => _selectedDestination = destination);
  }

  void _onMenuDestinationSelected(TravkaDestination destination) {
    _onDestinationSelected(destination);
    if (Scaffold.maybeOf(context)?.isDrawerOpen ?? false) {
      Navigator.pop(context);
    }
  }

  Future<void> _onLoggedIn(UserInfo user) async {
    if (!mounted) return;
    final l10n = AppLocalizations.of(context)!;
    setState(() {
      _nickname = user.nickname;
      _selectedDestination = TravkaDestination.home;
    });
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(l10n.welcomeUser(user.nickname))),
    );
  }

  void _onRegistered() {
    if (!mounted) return;
    final l10n = AppLocalizations.of(context)!;
    setState(() => _selectedDestination = TravkaDestination.login);
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(l10n.registrationSuccessful)),
    );
  }

  Future<void> _logout() async {
    await AuthStorage.clear();
    if (!mounted) return;
    final l10n = AppLocalizations.of(context)!;
    setState(() {
      _nickname = null;
      _selectedDestination = TravkaDestination.home;
    });
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(l10n.signedOut)),
    );
  }

  String _sectionTitle(AppLocalizations l10n) {
    switch (_selectedDestination) {
      case TravkaDestination.home:
        return l10n.home;
      case TravkaDestination.login:
        return l10n.signIn;
      case TravkaDestination.register:
        return l10n.register;
    }
  }

  String _appBarTitle() {
    if (_nickname != null) {
      return '$_title · $_nickname';
    }
    return _title;
  }

  Widget _buildSideMenu() {
    return TravkaSideMenu(
      selectedDestination: _selectedDestination,
      onDestinationSelected: _onMenuDestinationSelected,
      serverTitle: _title,
      nickname: _nickname,
      isLoggedIn: _isLoggedIn,
      onLogout: _logout,
      locale: widget.locale,
      onLocaleChanged: widget.onLocaleChanged,
    );
  }

  Widget _buildContent() {
    switch (_selectedDestination) {
      case TravkaDestination.home:
        return HomePage(nickname: _nickname);
      case TravkaDestination.login:
        return SingleChildScrollView(
          padding: const EdgeInsets.all(24),
          child: LoginForm(onLoggedIn: _onLoggedIn),
        );
      case TravkaDestination.register:
        return SingleChildScrollView(
          padding: const EdgeInsets.all(24),
          child: RegistrationForm(onRegistered: _onRegistered),
        );
    }
  }

  Widget? _buildFab(AppLocalizations l10n) {
    if (_selectedDestination != TravkaDestination.home) {
      return null;
    }

    return FloatingActionButton(
      onPressed: () {},
      tooltip: l10n.add,
      child: const Icon(Icons.add),
    );
  }

  Widget _buildNarrowLayout(AppLocalizations l10n) {
    return Scaffold(
      appBar: AppBar(
        backgroundColor: Theme.of(context).colorScheme.inversePrimary,
        title: Text(_appBarTitle()),
      ),
      drawer: _buildSideMenu(),
      body: _buildContent(),
      floatingActionButton: _buildFab(l10n),
    );
  }

  Widget _buildWideLayout(AppLocalizations l10n) {
    return Scaffold(
      body: Row(
        children: [
          SizedBox(
            width: kSideMenuWidth,
            child: _buildSideMenu(),
          ),
          const VerticalDivider(width: 1),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                Material(
                  color: Theme.of(context).colorScheme.inversePrimary,
                  child: SafeArea(
                    bottom: false,
                    child: SizedBox(
                      height: kToolbarHeight,
                      child: Align(
                        alignment: Alignment.centerLeft,
                        child: Padding(
                          padding: const EdgeInsets.symmetric(horizontal: 16),
                          child: Text(
                            _sectionTitle(l10n),
                            style: Theme.of(context).textTheme.titleLarge,
                          ),
                        ),
                      ),
                    ),
                  ),
                ),
                Expanded(child: _buildContent()),
              ],
            ),
          ),
        ],
      ),
      floatingActionButton: _buildFab(l10n),
    );
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return LayoutBuilder(
      builder: (context, constraints) {
        final isWide = constraints.maxWidth >= kWideLayoutBreakpoint;
        if (isWide) {
          return _buildWideLayout(l10n);
        }
        return _buildNarrowLayout(l10n);
      },
    );
  }
}
