import 'package:flutter/material.dart';
import 'package:travka/l10n/app_localizations.dart';

import '../api_request.dart';
import '../auth_storage.dart';
import '../login.dart';
import '../models/workout.dart';
import '../pages/home_page.dart';
import '../pages/profile_page.dart';
import '../pages/user_search_page.dart';
import '../platform/is_mobile_client.dart';
import '../registration.dart';
import '../server_storage.dart';
import '../services/track_recording_service.dart';
import '../widgets/add_workout_sheet.dart';
import '../widgets/settings_dialog.dart';
import '../widgets/track_recording_recovery_dialog.dart';
import '../widgets/workout_detail_menu.dart';
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
  final _scaffoldKey = GlobalKey<ScaffoldState>();
  final ApiRequest _api = ApiRequest();

  String _title = 'Travka Home';
  String? _nickname;
  TravkaDestination _selectedDestination = TravkaDestination.home;
  int _workoutRefreshToken = 0;
  Workout? _viewingWorkout;
  bool _isWorkoutMapExpanded = false;

  bool get _isLoggedIn => _nickname != null;
  bool get _isViewingWorkout =>
      _selectedDestination == TravkaDestination.home && _viewingWorkout != null;

  @override
  void initState() {
    super.initState();
    _loadInitialData();
  }

  Future<void> _loadInitialData() async {
    String name = 'Travka';
    if (!isMobileClient || ServerStorage.cachedBaseUrl != null) {
      try {
        name = await _api.getServerInfo();
      } catch (_) {
        // Network or server errors: keep default title.
      }
    }

    final token = await AuthStorage.getToken();

    String? nickname;
    if (token != null) {
      try {
        final user = await _api.getMe(token);
        nickname = user.nickname;
      } on ApiException catch (e) {
        if (e.statusCode == 401) {
          await AuthStorage.clear();
        }
      } catch (_) {
        // Network or server errors: keep token, show logged-out UI.
      }
    }

    if (!mounted) return;
    setState(() {
      _title = name;
      _nickname = nickname;
    });

    if (isMobileClient) {
      await _maybeRecoverRecording();
    }
  }

  Future<void> _maybeRecoverRecording() async {
    final needsRecovery = await TrackRecordingService.instance.needsRecovery();
    if (!needsRecovery || !mounted) {
      return;
    }
    await showTrackRecordingRecoveryDialog(
      context,
      TrackRecordingService.instance,
    );
  }

  void _onDestinationSelected(TravkaDestination destination) {
    setState(() {
      if (destination == TravkaDestination.home &&
          _selectedDestination == TravkaDestination.home &&
          _viewingWorkout != null) {
        _viewingWorkout = null;
        _isWorkoutMapExpanded = false;
        return;
      }
      _selectedDestination = destination;
      if (destination != TravkaDestination.home) {
        _viewingWorkout = null;
        _isWorkoutMapExpanded = false;
      }
    });
  }

  void _closeWorkoutDetail() {
    setState(() {
      _viewingWorkout = null;
      _isWorkoutMapExpanded = false;
    });
  }

  void _handleWorkoutDetailBack() {
    if (_isWorkoutMapExpanded) {
      setState(() => _isWorkoutMapExpanded = false);
    } else {
      _closeWorkoutDetail();
    }
  }

  String _contentHeaderTitle(AppLocalizations l10n) {
    if (_isViewingWorkout) {
      return _viewingWorkout!.name;
    }
    return _sectionTitle(l10n);
  }

  void _onMenuDestinationSelected(TravkaDestination destination) {
    _onDestinationSelected(destination);
    _scaffoldKey.currentState?.closeDrawer();
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
      _viewingWorkout = null;
      _isWorkoutMapExpanded = false;
    });
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(l10n.signedOut)),
    );
  }

  String _sectionTitle(AppLocalizations l10n) {
    switch (_selectedDestination) {
      case TravkaDestination.home:
        return l10n.home;
      case TravkaDestination.userSearch:
        return l10n.userSearch;
      case TravkaDestination.profile:
        return l10n.profile;
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

  void _onOpenSettings() {
    _scaffoldKey.currentState?.closeDrawer();
    showSettingsDialog(
      context,
      locale: widget.locale,
      onLocaleChanged: widget.onLocaleChanged,
    );
  }

  Widget _buildSideMenu() {
    return TravkaSideMenu(
      selectedDestination: _selectedDestination,
      onDestinationSelected: _onMenuDestinationSelected,
      serverTitle: _title,
      nickname: _nickname,
      isLoggedIn: _isLoggedIn,
      onLogout: _logout,
      onOpenSettings: _onOpenSettings,
    );
  }

  Widget _buildContent() {
    switch (_selectedDestination) {
      case TravkaDestination.home:
        return HomePage(
          nickname: _nickname,
          refreshToken: _workoutRefreshToken,
          viewingWorkout: _viewingWorkout,
          isMapExpanded: _isWorkoutMapExpanded,
          onViewingWorkoutChanged: (workout) {
            setState(() {
              _viewingWorkout = workout;
              _isWorkoutMapExpanded = false;
            });
          },
          onMapExpandedChanged: (expanded) {
            setState(() => _isWorkoutMapExpanded = expanded);
          },
        );
      case TravkaDestination.userSearch:
        return const UserSearchPage();
      case TravkaDestination.profile:
        return ProfilePage(nickname: _nickname!);
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
    if (_selectedDestination != TravkaDestination.home ||
        !_isLoggedIn ||
        _isViewingWorkout) {
      return null;
    }

    return FloatingActionButton(
      onPressed: () async {
        final saved = await showAddWorkoutSheet(context);
        if (saved == true && mounted) {
          setState(() => _workoutRefreshToken++);
        }
      },
      tooltip: l10n.add,
      child: const Icon(Icons.add),
    );
  }

  Widget _buildNarrowLayout(AppLocalizations l10n) {
    return Scaffold(
      key: _scaffoldKey,
      appBar: AppBar(
        backgroundColor: Theme.of(context).colorScheme.inversePrimary,
        leading: _isViewingWorkout
            ? BackButton(onPressed: _handleWorkoutDetailBack)
            : null,
        title: Text(
          _isViewingWorkout ? _viewingWorkout!.name : _appBarTitle(),
        ),
        actions: [
          if (_isViewingWorkout) const WorkoutDetailMenu(),
        ],
      ),
      drawer: _isViewingWorkout ? null : _buildSideMenu(),
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
                      child: Row(
                        children: [
                          if (_isViewingWorkout)
                            BackButton(onPressed: _handleWorkoutDetailBack),
                          Expanded(
                            child: Align(
                              alignment: Alignment.centerLeft,
                              child: Padding(
                                padding: const EdgeInsets.symmetric(
                                  horizontal: 16,
                                ),
                                child: Text(
                                  _contentHeaderTitle(l10n),
                                  style: Theme.of(context).textTheme.titleLarge,
                                ),
                              ),
                            ),
                          ),
                          if (_isViewingWorkout) const WorkoutDetailMenu(),
                        ],
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

    return PopScope(
      canPop: !_isViewingWorkout,
      onPopInvokedWithResult: (didPop, result) {
        if (!didPop && _isViewingWorkout) {
          _handleWorkoutDetailBack();
        }
      },
      child: LayoutBuilder(
        builder: (context, constraints) {
          final isWide = constraints.maxWidth >= kWideLayoutBreakpoint;
          if (isWide) {
            return _buildWideLayout(l10n);
          }
          return _buildNarrowLayout(l10n);
        },
      ),
    );
  }
}
