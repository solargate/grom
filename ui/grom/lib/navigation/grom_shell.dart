import 'dart:async';

import 'package:flutter/foundation.dart' show kIsWeb, defaultTargetPlatform, TargetPlatform;
import 'package:flutter/material.dart';
import 'package:grom/l10n/app_localizations.dart';
import 'package:material_symbols_icons/symbols.dart';

import '../api_request.dart';
import '../auth_storage.dart';
import '../login.dart';
import '../models/workout.dart';
import '../pages/about_page.dart';
import '../pages/equipment_page.dart';
import '../pages/home_page.dart';
import '../pages/integration_page.dart';
import '../pages/profile_page.dart';
import '../pages/settings_page.dart';
import '../pages/user_search_page.dart';
import '../platform/is_mobile_client.dart';
import '../platform/file_download.dart';
import '../platform/shared_track_intent.dart';
import '../registration.dart';
import '../server_storage.dart';
import '../services/health_sync_service.dart';
import '../services/track_recording_service.dart';
import '../widgets/add_workout_sheet.dart';
import '../widgets/profile_menu.dart';
import '../widgets/track_recording_recovery_dialog.dart';
import '../widgets/workout_detail_menu.dart';
import 'grom_destination.dart';
import 'grom_side_menu.dart';

const kWideLayoutBreakpoint = 600.0;

class GromShell extends StatefulWidget {
  const GromShell({
    super.key,
    required this.locale,
    required this.onLocaleChanged,
  });

  final Locale locale;
  final ValueChanged<Locale> onLocaleChanged;

  @override
  State<GromShell> createState() => _GromShellState();
}

class _GromShellState extends State<GromShell> {
  final _scaffoldKey = GlobalKey<ScaffoldState>();
  final _homePageKey = GlobalKey<HomePageState>();
  final _profilePageKey = GlobalKey<ProfilePageState>();
  final ApiRequest _api = ApiRequest();

  String _title = 'Grom Home';
  bool _federationEnabled = false;
  String? _nickname;
  GromDestination _selectedDestination = GromDestination.home;
  int _workoutRefreshToken = 0;
  Workout? _viewingWorkout;
  bool _isWorkoutMapExpanded = false;
  int? _workoutPhotoViewerIndex;
  Workout? _feedPhotoViewerWorkout;

  bool _isShellReady = false;
  bool _healthSyncEnabled = false;
  StreamSubscription<SharedTrackPayload>? _sharedTrackSub;
  SharedTrackPayload? _pendingSharedTrack;
  bool _isProcessingSharedTrack = false;

  bool get _isLoggedIn => _nickname != null;
  bool get _isAndroid =>
      !kIsWeb && defaultTargetPlatform == TargetPlatform.android;
  bool get _showHealthSyncButton =>
      _isAndroid &&
      _healthSyncEnabled &&
      _selectedDestination == GromDestination.home &&
      !_isViewingWorkout &&
      !_isViewingFeedPhoto &&
      _isLoggedIn;
  bool get _isViewingWorkout =>
      _selectedDestination == GromDestination.home && _viewingWorkout != null;
  bool get _isViewingProfile =>
      _selectedDestination == GromDestination.profile && !_isViewingWorkout;
  bool get _isViewingFeedPhoto =>
      _selectedDestination == GromDestination.home &&
      _viewingWorkout == null &&
      _feedPhotoViewerWorkout != null &&
      _workoutPhotoViewerIndex != null;
  bool get _shouldShowHeaderBackButton =>
      _isViewingWorkout || _isViewingFeedPhoto;
  bool get _shouldInterceptPop =>
      _isViewingWorkout || _isViewingFeedPhoto || !_selectedDestination.isHome;

  @override
  void initState() {
    super.initState();
    HealthSyncService.instance.addListener(_onHealthSyncChanged);
    unawaited(_loadHealthSyncState());
    if (isMobileClient) {
      _sharedTrackSub = watchSharedTracks().listen(_handleIncomingSharedTrack);
    }
    _loadInitialData();
  }

  @override
  void dispose() {
    HealthSyncService.instance.removeListener(_onHealthSyncChanged);
    _sharedTrackSub?.cancel();
    super.dispose();
  }

  Future<void> _loadHealthSyncState() async {
    await HealthSyncService.instance.loadFromStorage();
    if (!mounted) {
      return;
    }
    setState(() => _healthSyncEnabled = HealthSyncService.instance.enabled);
  }

  void _onHealthSyncChanged() {
    if (!mounted) {
      return;
    }
    setState(() => _healthSyncEnabled = HealthSyncService.instance.enabled);
  }

  Future<void> _runHealthSync() async {
    final l10n = AppLocalizations.of(context)!;
    final service = HealthSyncService.instance;

    if (service.folderName.trim().isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(l10n.healthSyncFolderNameRequired)),
      );
      return;
    }

    if (!mounted) {
      return;
    }

    showDialog<void>(
      context: context,
      barrierDismissible: false,
      builder: (context) => AlertDialog(
        content: Row(
          children: [
            const CircularProgressIndicator(),
            const SizedBox(width: 24),
            Expanded(child: Text(l10n.healthSyncSynchronizing)),
          ],
        ),
      ),
    );

    final result = await service.syncWorkouts();

    if (!mounted) {
      return;
    }
    Navigator.of(context, rootNavigator: true).pop();

    final message = healthSyncResultSnackBarMessage(
      result,
      imported: l10n.healthSyncImported(result.importedCount),
      noNewWorkouts: l10n.healthSyncNoNewWorkouts,
      folderNotFound: l10n.healthSyncFolderNotFound,
      folderEmpty: l10n.healthSyncFolderEmpty,
      signInCancelled: l10n.healthSyncGoogleSignInCancelled,
      signInFailed: l10n.healthSyncGoogleSignInFailed,
      accessDenied: l10n.healthSyncDriveAccessDenied,
      syncError: l10n.healthSyncSyncError,
    );
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(message)),
    );

    if (result.kind == HealthSyncResultKind.imported) {
      setState(() => _workoutRefreshToken++);
    }
  }

  Widget? _buildHealthSyncHeaderButton() {
    if (!_showHealthSyncButton) {
      return null;
    }

    final l10n = AppLocalizations.of(context)!;
    return IconButton(
      tooltip: l10n.healthSyncSync,
      onPressed: HealthSyncService.instance.syncing ? null : _runHealthSync,
      icon: const Icon(Symbols.directory_sync),
    );
  }

  Future<({String name, bool federationEnabled})> _fetchServerInfo() async {
    var name = 'Grom';
    var federationEnabled = false;
    if (!isMobileClient || ServerStorage.cachedBaseUrl != null) {
      try {
        final serverInfo = await _api.getServerInfo();
        name = serverInfo.name;
        federationEnabled = serverInfo.federationEnabled;
      } catch (_) {
        // Network or server errors: keep default title.
      }
    }
    return (name: name, federationEnabled: federationEnabled);
  }

  Future<void> _loadInitialData() async {
    final serverInfo = await _fetchServerInfo();

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

    if (isMobileClient) {
      await _maybeRecoverRecording();
      final initialTrackResult = await takePendingSharedTrack();
      if (initialTrackResult.readFailed && mounted) {
        final l10n = AppLocalizations.of(context)!;
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(l10n.shareTrackReadFailed)),
        );
      } else if (initialTrackResult.unsupportedFormat && mounted) {
        final l10n = AppLocalizations.of(context)!;
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(l10n.invalidTrackFormat)),
        );
      }
      final initialTrack = initialTrackResult.payload;
      if (initialTrack != null) {
        _pendingSharedTrack = initialTrack;
      }
    }

    if (!mounted) return;
    setState(() {
      _title = serverInfo.name;
      _federationEnabled = serverInfo.federationEnabled;
      _nickname = nickname;
      _isShellReady = true;
    });

    await _processPendingSharedTrack();
  }

  void _handleIncomingSharedTrack(SharedTrackPayload payload) {
    _pendingSharedTrack = payload;
    if (_isShellReady) {
      unawaited(_processPendingSharedTrack());
    }
  }

  Future<void> _processPendingSharedTrack() async {
    if (_isProcessingSharedTrack || !_isShellReady || !mounted) {
      return;
    }

    final payload = _pendingSharedTrack;
    if (payload == null) {
      return;
    }

    _isProcessingSharedTrack = true;
    try {
      if (!_isLoggedIn) {
        final l10n = AppLocalizations.of(context)!;
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(l10n.shareTrackLoginRequired)),
        );
        setState(() => _selectedDestination = GromDestination.login);
        return;
      }

      if (isMobileClient && TrackRecordingService.instance.isActive) {
        final l10n = AppLocalizations.of(context)!;
        final discard = await showDialog<bool>(
          context: context,
          builder: (context) => AlertDialog(
            title: Text(l10n.discardRecordingTitle),
            content: Text(l10n.discardRecordingMessage),
            actions: [
              TextButton(
                onPressed: () => Navigator.pop(context, false),
                child: Text(l10n.cancel),
              ),
              FilledButton(
                onPressed: () => Navigator.pop(context, true),
                child: Text(l10n.discardRecordingConfirm),
              ),
            ],
          ),
        );
        if (discard != true) {
          return;
        }
        await TrackRecordingService.instance.discardRecording();
      }

      _pendingSharedTrack = null;

      setState(() {
        _selectedDestination = GromDestination.home;
        _viewingWorkout = null;
        _isWorkoutMapExpanded = false;
        _workoutPhotoViewerIndex = null;
        _feedPhotoViewerWorkout = null;
      });

      if (!mounted) {
        return;
      }

      final saved = await showAddWorkoutSheet(context, initialTrack: payload);
      if (saved != null && mounted) {
        setState(() => _workoutRefreshToken++);
      }
    } finally {
      _isProcessingSharedTrack = false;
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

  void _onDestinationSelected(GromDestination destination) {
    setState(() {
      if (destination == GromDestination.home &&
          _selectedDestination == GromDestination.home) {
        if (_viewingWorkout != null) {
          _viewingWorkout = null;
          _isWorkoutMapExpanded = false;
          _workoutPhotoViewerIndex = null;
          _feedPhotoViewerWorkout = null;
          return;
        }
        if (_isViewingFeedPhoto) {
          _workoutPhotoViewerIndex = null;
          _feedPhotoViewerWorkout = null;
          return;
        }
        _homePageKey.currentState?.scrollActiveTabToTop();
        return;
      }
      _selectedDestination = destination;
      if (destination != GromDestination.home) {
        _viewingWorkout = null;
        _isWorkoutMapExpanded = false;
        _workoutPhotoViewerIndex = null;
        _feedPhotoViewerWorkout = null;
      }
    });
  }

  void _closeWorkoutDetail() {
    setState(() {
      _viewingWorkout = null;
      _isWorkoutMapExpanded = false;
      _workoutPhotoViewerIndex = null;
      _feedPhotoViewerWorkout = null;
    });
  }

  void _closeFeedPhotoViewer() {
    setState(() {
      _workoutPhotoViewerIndex = null;
      _feedPhotoViewerWorkout = null;
    });
  }

  void _handleWorkoutDetailBack() {
    if (_workoutPhotoViewerIndex != null) {
      setState(() => _workoutPhotoViewerIndex = null);
    } else if (_isWorkoutMapExpanded) {
      setState(() => _isWorkoutMapExpanded = false);
    } else {
      _closeWorkoutDetail();
    }
  }

  void _handleShellBack() {
    if (_isViewingWorkout) {
      _handleWorkoutDetailBack();
      return;
    }
    if (_isViewingFeedPhoto) {
      _closeFeedPhotoViewer();
      return;
    }
    if (!_selectedDestination.isHome) {
      _onDestinationSelected(GromDestination.home);
    }
  }

  bool _isOwnWorkout(Workout workout) {
    final nickname = _nickname;
    if (nickname == null) {
      return false;
    }
    return workout.ownerNickname == nickname;
  }

  Future<void> _handleWorkoutMenuAction(WorkoutDetailMenuAction action) async {
    final workout = _viewingWorkout;
    if (workout == null) {
      return;
    }

    switch (action) {
      case WorkoutDetailMenuAction.downloadGpx:
        await _downloadWorkoutTrack(workout: workout, format: 'gpx');
      case WorkoutDetailMenuAction.downloadOriginal:
        await _downloadWorkoutTrack(workout: workout);
      case WorkoutDetailMenuAction.delete:
        await _confirmDeleteWorkout(workout);
      case WorkoutDetailMenuAction.edit:
        await _editWorkout(workout);
    }
  }

  Future<void> _editWorkout(Workout workout) async {
    final updated = await showAddWorkoutSheet(context, workout: workout);
    if (updated == null || !mounted) {
      return;
    }
    setState(() {
      _viewingWorkout = Workout(
        id: updated.id,
        name: updated.name,
        description: updated.description,
        sportType: updated.sportType,
        startDate: updated.startDate,
        durationSeconds: updated.durationSeconds,
        distance: updated.distance,
        durationTotalSeconds: updated.durationTotalSeconds,
        tempAvgKmm: updated.tempAvgKmm,
        speedMaxKmh: updated.speedMaxKmh,
        speedAvgKmh: updated.speedAvgKmh,
        elevationGain: updated.elevationGain,
        heartRateAvg: updated.heartRateAvg,
        heartRateMax: updated.heartRateMax,
        stepsTotal: updated.stepsTotal,
        calories: updated.calories,
        owner: workout.owner.isNotEmpty ? workout.owner : updated.owner,
        device: updated.device,
        track: updated.track,
        hasMapPreview: updated.hasMapPreview,
        hasMedia: updated.hasMedia,
        mediaFiles: updated.mediaFiles,
        author: workout.author ?? updated.author,
        equipment: updated.equipment,
      );
      _workoutRefreshToken++;
    });
  }

  Future<void> _confirmDeleteWorkout(Workout workout) async {
    final l10n = AppLocalizations.of(context)!;
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(l10n.deleteWorkout),
        content: Text(l10n.deleteWorkoutConfirm),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: Text(l10n.cancel),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: Text(l10n.deleteWorkout),
          ),
        ],
      ),
    );
    if (confirmed != true || !mounted) {
      return;
    }

    final messenger = ScaffoldMessenger.of(context);
    try {
      final token = await AuthStorage.getToken();
      if (token == null) {
        if (!mounted) {
          return;
        }
        messenger.showSnackBar(
          SnackBar(content: Text(l10n.failedToDeleteWorkout)),
        );
        return;
      }

      await _api.deleteWorkout(token: token, workoutId: workout.id);
      if (!mounted) {
        return;
      }
      _closeWorkoutDetail();
      setState(() => _workoutRefreshToken++);
      messenger.showSnackBar(
        SnackBar(content: Text(l10n.workoutDeleted)),
      );
    } on ApiException catch (e) {
      if (!mounted) {
        return;
      }
      messenger.showSnackBar(
        SnackBar(content: Text(e.message)),
      );
    } catch (_) {
      if (!mounted) {
        return;
      }
      messenger.showSnackBar(
        SnackBar(content: Text(l10n.failedToDeleteWorkout)),
      );
    }
  }

  Future<void> _downloadWorkoutTrack({
    required Workout workout,
    String? format,
  }) async {
    final l10n = AppLocalizations.of(context)!;
    final messenger = ScaffoldMessenger.of(context);
    messenger.showSnackBar(SnackBar(content: Text(l10n.downloadingTrack)));

    try {
      final token = await AuthStorage.getToken();
      if (token == null) {
        if (!mounted) {
          return;
        }
        messenger.hideCurrentSnackBar();
        messenger.showSnackBar(
          SnackBar(content: Text(l10n.failedToDownloadTrack)),
        );
        return;
      }

      final owner = _isOwnWorkout(workout) ? null : workout.ownerNickname;
      final downloaded = await _api.downloadWorkoutTrack(
        token: token,
        workoutId: workout.id,
        fallbackFilename: workout.track,
        owner: owner != null && owner.isNotEmpty ? owner : null,
        format: format,
      );
      await saveDownloadedFile(
        bytes: downloaded.bytes,
        filename: downloaded.filename,
      );
      if (!mounted) {
        return;
      }
      messenger.hideCurrentSnackBar();
      messenger.showSnackBar(
        SnackBar(content: Text(l10n.trackSaved)),
      );
    } on SaveDownloadedFileCancelled {
      if (!mounted) {
        return;
      }
      messenger.hideCurrentSnackBar();
    } on ApiException catch (e) {
      if (!mounted) {
        return;
      }
      messenger.hideCurrentSnackBar();
      messenger.showSnackBar(
        SnackBar(content: Text(e.message)),
      );
    } catch (_) {
      if (!mounted) {
        return;
      }
      messenger.hideCurrentSnackBar();
      messenger.showSnackBar(
        SnackBar(content: Text(l10n.failedToDownloadTrack)),
      );
    }
  }

  Widget _buildWorkoutDetailMenu() {
    final workout = _viewingWorkout;
    if (workout == null) {
      return const SizedBox.shrink();
    }

    return WorkoutDetailMenu(
      hasTrack: workout.track.isNotEmpty,
      canDownloadOriginal: _isOwnWorkout(workout),
      canEdit: _isOwnWorkout(workout),
      canDelete: _isOwnWorkout(workout),
      onSelected: _handleWorkoutMenuAction,
    );
  }

  Widget _buildProfileMenu() {
    return ProfileMenu(
      onSelected: _handleProfileMenuAction,
    );
  }

  void _handleProfileMenuAction(ProfileMenuAction action) {
    switch (action) {
      case ProfileMenuAction.edit:
        unawaited(_profilePageKey.currentState?.openEditProfile());
      case ProfileMenuAction.deleteAccount:
        break;
    }
  }

  String _contentHeaderTitle(AppLocalizations l10n) {
    if (_isViewingWorkout) {
      return _viewingWorkout!.name;
    }
    return _sectionTitle(l10n);
  }

  void _onMenuDestinationSelected(GromDestination destination) {
    _onDestinationSelected(destination);
    _scaffoldKey.currentState?.closeDrawer();
  }

  Future<void> _onLoggedIn(UserInfo user) async {
    final serverInfo = await _fetchServerInfo();
    if (!mounted) return;
    final l10n = AppLocalizations.of(context)!;
    setState(() {
      _title = serverInfo.name;
      _federationEnabled = serverInfo.federationEnabled;
      _nickname = user.nickname;
      _selectedDestination = GromDestination.home;
    });
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(l10n.welcomeUser(user.nickname))),
    );
    await _processPendingSharedTrack();
  }

  void _onRegistered() {
    if (!mounted) return;
    final l10n = AppLocalizations.of(context)!;
    setState(() => _selectedDestination = GromDestination.login);
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
      _selectedDestination = GromDestination.home;
      _viewingWorkout = null;
      _isWorkoutMapExpanded = false;
      _workoutPhotoViewerIndex = null;
      _feedPhotoViewerWorkout = null;
    });
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(l10n.signedOut)),
    );
  }

  String _sectionTitle(AppLocalizations l10n) {
    switch (_selectedDestination) {
      case GromDestination.home:
        return l10n.home;
      case GromDestination.userSearch:
        return l10n.userSearch;
      case GromDestination.profile:
        return l10n.profile;
      case GromDestination.equipment:
        return l10n.equipment;
      case GromDestination.integration:
        return l10n.integration;
      case GromDestination.login:
        return l10n.signIn;
      case GromDestination.register:
        return l10n.register;
      case GromDestination.about:
        return l10n.about;
      case GromDestination.settings:
        return l10n.settings;
    }
  }

  Widget _buildSideMenu() {
    return GromSideMenu(
      selectedDestination: _selectedDestination,
      onDestinationSelected: _onMenuDestinationSelected,
      serverTitle: _title,
      nickname: _nickname,
      isLoggedIn: _isLoggedIn,
      onLogout: _logout,
    );
  }

  Widget _buildContent() {
    switch (_selectedDestination) {
      case GromDestination.home:
        return HomePage(
          key: _homePageKey,
          nickname: _nickname,
          federationEnabled: _federationEnabled,
          refreshToken: _workoutRefreshToken,
          viewingWorkout: _viewingWorkout,
          isMapExpanded: _isWorkoutMapExpanded,
          onViewingWorkoutChanged: (workout) {
            setState(() {
              _viewingWorkout = workout;
              _isWorkoutMapExpanded = false;
              if (workout == null) {
                _workoutPhotoViewerIndex = null;
              } else {
                _feedPhotoViewerWorkout = null;
              }
            });
          },
          onMapExpandedChanged: (expanded) {
            setState(() => _isWorkoutMapExpanded = expanded);
          },
          photoViewerIndex: _workoutPhotoViewerIndex,
          onPhotoViewerIndexChanged: (index) {
            setState(() {
              _workoutPhotoViewerIndex = index;
              if (index == null && _viewingWorkout == null) {
                _feedPhotoViewerWorkout = null;
              }
            });
          },
          feedPhotoViewerWorkout: _feedPhotoViewerWorkout,
          onFeedPhotoViewerWorkoutChanged: (workout) {
            setState(() => _feedPhotoViewerWorkout = workout);
          },
        );
      case GromDestination.userSearch:
        return const UserSearchPage();
      case GromDestination.profile:
        return ProfilePage(
          key: _profilePageKey,
          nickname: _nickname!,
        );
      case GromDestination.equipment:
        return const EquipmentPage();
      case GromDestination.integration:
        return const IntegrationPage();
      case GromDestination.login:
        return SingleChildScrollView(
          padding: const EdgeInsets.all(24),
          child: LoginForm(onLoggedIn: _onLoggedIn),
        );
      case GromDestination.register:
        return SingleChildScrollView(
          padding: const EdgeInsets.all(24),
          child: RegistrationForm(onRegistered: _onRegistered),
        );
      case GromDestination.about:
        return const AboutPage();
      case GromDestination.settings:
        return SettingsPage(
          locale: widget.locale,
          onLocaleChanged: widget.onLocaleChanged,
        );
    }
  }

  Widget? _buildFab(AppLocalizations l10n) {
    if (_selectedDestination != GromDestination.home ||
        !_isLoggedIn ||
        _isViewingWorkout ||
        _isViewingFeedPhoto) {
      return null;
    }

    return FloatingActionButton(
      onPressed: () async {
        final saved = await showAddWorkoutSheet(context);
        if (saved != null && mounted) {
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
        leading: _shouldShowHeaderBackButton
            ? BackButton(onPressed: _handleShellBack)
            : null,
        title: Text(_contentHeaderTitle(l10n)),
        actions: [
          if (_showHealthSyncButton) _buildHealthSyncHeaderButton()!,
          if (_isViewingWorkout) _buildWorkoutDetailMenu(),
          if (_isViewingProfile) _buildProfileMenu(),
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
                          if (_shouldShowHeaderBackButton)
                            BackButton(onPressed: _handleShellBack),
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
                          if (_showHealthSyncButton) _buildHealthSyncHeaderButton()!,
                          if (_isViewingWorkout) _buildWorkoutDetailMenu(),
                          if (_isViewingProfile) _buildProfileMenu(),
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
      canPop: !_shouldInterceptPop,
      onPopInvokedWithResult: (didPop, result) {
        if (!didPop) {
          _handleShellBack();
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
