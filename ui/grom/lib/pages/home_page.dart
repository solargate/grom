import 'dart:async';

import 'package:flutter/material.dart';
import 'package:grom/l10n/app_localizations.dart';

import '../api_request.dart';
import '../auth_storage.dart';
import '../models/my_workouts_layout.dart';
import '../models/social.dart';
import '../models/workout.dart';
import '../pages/workout_detail_page.dart';
import '../platform/is_mobile_client.dart';
import '../services/my_workouts_layout_storage.dart';
import '../widgets/sport_type_toggle.dart';
import '../widgets/welcome_guest_view.dart';
import '../widgets/workout_feed_list.dart';
import '../widgets/workout_photo_viewer.dart';

enum HomeFeedTab { feed, myWorkouts }

/// AppBar chrome for the My workouts sport filter (owned by [HomePage]).
class SportFilterChrome {
  const SportFilterChrome({
    required this.visible,
    required this.expanded,
    required this.onToggle,
  });

  final bool visible;
  final bool expanded;
  final VoidCallback onToggle;
}

/// AppBar chrome for cards/list layout toggle (owned by [HomePage]).
class MyWorkoutsLayoutChrome {
  const MyWorkoutsLayoutChrome({
    required this.visible,
    required this.layout,
    required this.onToggle,
  });

  final bool visible;
  final MyWorkoutsLayout layout;
  final VoidCallback onToggle;
}

class HomePage extends StatefulWidget {
  const HomePage({
    super.key,
    this.nickname,
    this.federationEnabled = false,
    this.refreshToken = 0,
    this.viewingWorkout,
    this.isMapExpanded = false,
    this.onViewingWorkoutChanged,
    this.onMapExpandedChanged,
    this.photoViewerIndex,
    this.onPhotoViewerIndexChanged,
    this.feedPhotoViewerWorkout,
    this.onFeedPhotoViewerWorkoutChanged,
    this.onSignIn,
    this.onRegister,
    this.onSportFilterChromeChanged,
    this.onMyWorkoutsLayoutChromeChanged,
  });

  final String? nickname;
  final bool federationEnabled;
  final int refreshToken;
  final Workout? viewingWorkout;
  final bool isMapExpanded;
  final ValueChanged<Workout?>? onViewingWorkoutChanged;
  final ValueChanged<bool>? onMapExpandedChanged;
  final int? photoViewerIndex;
  final ValueChanged<int?>? onPhotoViewerIndexChanged;
  final Workout? feedPhotoViewerWorkout;
  final ValueChanged<Workout?>? onFeedPhotoViewerWorkoutChanged;
  final VoidCallback? onSignIn;
  final VoidCallback? onRegister;
  final ValueChanged<SportFilterChrome>? onSportFilterChromeChanged;
  final ValueChanged<MyWorkoutsLayoutChrome>? onMyWorkoutsLayoutChromeChanged;

  @override
  State<HomePage> createState() => HomePageState();
}

class HomePageState extends State<HomePage> with TickerProviderStateMixin {
  final ApiRequest _api = ApiRequest();

  TabController? _tabController;
  bool _showFeedTab = false;
  HomeFeedTab _activeTab = HomeFeedTab.myWorkouts;
  String? _authToken;
  bool _isLoadingTabs = true;

  List<String> _usedSportTypes = const [];
  Set<String> _includedSportTypes = {};
  bool _filterExpanded = false;
  MyWorkoutsLayout _myWorkoutsLayout = MyWorkoutsLayout.cards;

  final Map<HomeFeedTab, ScrollController> _scrollControllers = {
    HomeFeedTab.feed: ScrollController(),
    HomeFeedTab.myWorkouts: ScrollController(),
  };

  final Map<HomeFeedTab, GlobalKey<WorkoutFeedListState>> _feedListKeys = {
    HomeFeedTab.feed: GlobalKey<WorkoutFeedListState>(),
    HomeFeedTab.myWorkouts: GlobalKey<WorkoutFeedListState>(),
  };

  @override
  void initState() {
    super.initState();
    unawaited(_loadMyWorkoutsLayout());
    _loadTabState(resetToDefaultTab: true);
  }

  Future<void> _loadMyWorkoutsLayout() async {
    final layout = await MyWorkoutsLayoutStorage.getLayout();
    if (!mounted) {
      return;
    }
    setState(() => _myWorkoutsLayout = layout);
    _notifyMyWorkoutsLayoutChrome();
  }

  @override
  void didUpdateWidget(covariant HomePage oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (widget.nickname != oldWidget.nickname) {
      _usedSportTypes = const [];
      _includedSportTypes = {};
      _filterExpanded = false;
      _loadTabState(resetToDefaultTab: true);
    } else if (widget.refreshToken != oldWidget.refreshToken) {
      _loadTabState();
    } else if (widget.viewingWorkout != oldWidget.viewingWorkout ||
        widget.feedPhotoViewerWorkout != oldWidget.feedPhotoViewerWorkout ||
        widget.onSportFilterChromeChanged !=
            oldWidget.onSportFilterChromeChanged ||
        widget.onMyWorkoutsLayoutChromeChanged !=
            oldWidget.onMyWorkoutsLayoutChromeChanged) {
      _notifySportFilterChrome();
      _notifyMyWorkoutsLayoutChrome();
    }
  }

  @override
  void dispose() {
    _tabController?.dispose();
    for (final controller in _scrollControllers.values) {
      controller.dispose();
    }
    super.dispose();
  }

  Future<void> _loadTabState({bool resetToDefaultTab = false}) async {
    if (widget.nickname == null) {
      setState(() {
        _showFeedTab = false;
        _isLoadingTabs = false;
        _usedSportTypes = const [];
        _includedSportTypes = {};
        _filterExpanded = false;
      });
      _syncTabController(showFeedTab: false, preferredTab: HomeFeedTab.myWorkouts);
      _notifySportFilterChrome();
      _notifyMyWorkoutsLayoutChrome();
      return;
    }

    setState(() => _isLoadingTabs = true);

    try {
      final token = await AuthStorage.getToken();
      if (token == null) {
        throw ApiException('Not authenticated');
      }

      final results = await Future.wait<Object>([
        _api.listFollowing(token),
        _api.getProfile(token),
      ]);
      if (!mounted) return;

      final following = results[0] as List<FollowInfo>;
      final profile = results[1] as UserProfile;
      final showFeedTab =
          following.any((follow) => follow.status == 'active');
      final preferredTab = showFeedTab && resetToDefaultTab
          ? HomeFeedTab.feed
          : _activeTab == HomeFeedTab.feed && !showFeedTab
              ? HomeFeedTab.myWorkouts
              : _activeTab;

      final used = List<String>.from(profile.usedSportTypes);
      final previousUsed = _usedSportTypes;
      final included = <String>{};
      if (previousUsed.isEmpty && _includedSportTypes.isEmpty) {
        included.addAll(used);
      } else {
        included.addAll(
          _includedSportTypes.where((id) => used.contains(id)),
        );
        for (final id in used) {
          if (!previousUsed.contains(id)) {
            included.add(id);
          }
        }
      }

      setState(() {
        _showFeedTab = showFeedTab;
        _activeTab = preferredTab;
        _isLoadingTabs = false;
        _usedSportTypes = used;
        _includedSportTypes = included;
        if (used.isEmpty) {
          _filterExpanded = false;
        }
      });
      _syncTabController(showFeedTab: showFeedTab, preferredTab: preferredTab);
      _notifySportFilterChrome();
      _notifyMyWorkoutsLayoutChrome();
    } catch (_) {
      if (!mounted) return;
      setState(() {
        _showFeedTab = false;
        _isLoadingTabs = false;
      });
      _syncTabController(showFeedTab: false, preferredTab: HomeFeedTab.myWorkouts);
      _notifySportFilterChrome();
      _notifyMyWorkoutsLayoutChrome();
    }
  }

  void _syncTabController({
    required bool showFeedTab,
    required HomeFeedTab preferredTab,
  }) {
    final tabCount = showFeedTab ? 2 : 1;
    final initialIndex = _tabIndexFor(preferredTab, showFeedTab: showFeedTab);

    if (_tabController != null &&
        _tabController!.length == tabCount &&
        _tabController!.index == initialIndex) {
      return;
    }

    _tabController?.dispose();
    _tabController = TabController(
      length: tabCount,
      vsync: this,
      initialIndex: initialIndex,
    )..addListener(_handleTabChanged);
  }

  void _handleTabChanged() {
    if (_tabController == null || _tabController!.indexIsChanging) {
      return;
    }
    final tab = _tabForIndex(_tabController!.index);
    if (tab != _activeTab) {
      setState(() {
        _activeTab = tab;
        if (tab != HomeFeedTab.myWorkouts) {
          _filterExpanded = false;
        }
      });
      _notifySportFilterChrome();
      _notifyMyWorkoutsLayoutChrome();
    }
  }

  int _tabIndexFor(HomeFeedTab tab, {required bool showFeedTab}) {
    if (!showFeedTab) {
      return 0;
    }
    return tab == HomeFeedTab.feed ? 0 : 1;
  }

  HomeFeedTab _tabForIndex(int index) {
    if (!_showFeedTab) {
      return HomeFeedTab.myWorkouts;
    }
    return index == 0 ? HomeFeedTab.feed : HomeFeedTab.myWorkouts;
  }

  void scrollActiveTabToTop() {
    _scrollTabToTop(_activeTab);
  }

  void _scrollTabToTop(HomeFeedTab tab) {
    final controller = _scrollControllers[tab];
    if (controller == null || !controller.hasClients) {
      return;
    }
    controller.animateTo(
      0,
      duration: const Duration(milliseconds: 300),
      curve: Curves.easeOut,
    );
  }

  void _onTabTapped(int index) {
    final tab = _tabForIndex(index);
    if (index == _tabController?.index) {
      _scrollTabToTop(tab);
    }
  }

  void _openWorkout(Workout workout) {
    widget.onFeedPhotoViewerWorkoutChanged?.call(null);
    widget.onPhotoViewerIndexChanged?.call(null);
    widget.onViewingWorkoutChanged?.call(workout);
  }

  void _openWorkoutPhoto(Workout workout, int index) {
    widget.onFeedPhotoViewerWorkoutChanged?.call(workout);
    widget.onPhotoViewerIndexChanged?.call(index);
  }

  bool get _filterButtonVisible =>
      widget.nickname != null &&
      _activeTab == HomeFeedTab.myWorkouts &&
      _usedSportTypes.isNotEmpty &&
      widget.viewingWorkout == null &&
      widget.feedPhotoViewerWorkout == null;

  bool get _layoutButtonVisible =>
      widget.nickname != null &&
      _activeTab == HomeFeedTab.myWorkouts &&
      widget.viewingWorkout == null &&
      widget.feedPhotoViewerWorkout == null;

  void _notifySportFilterChrome() {
    widget.onSportFilterChromeChanged?.call(
      SportFilterChrome(
        visible: _filterButtonVisible,
        expanded: _filterExpanded && _filterButtonVisible,
        onToggle: _toggleSportFilter,
      ),
    );
  }

  void _notifyMyWorkoutsLayoutChrome() {
    widget.onMyWorkoutsLayoutChromeChanged?.call(
      MyWorkoutsLayoutChrome(
        visible: _layoutButtonVisible,
        layout: _myWorkoutsLayout,
        onToggle: _toggleMyWorkoutsLayout,
      ),
    );
  }

  void _toggleSportFilter() {
    if (!_filterButtonVisible) {
      return;
    }
    setState(() => _filterExpanded = !_filterExpanded);
    _notifySportFilterChrome();
  }

  void _toggleMyWorkoutsLayout() {
    if (!_layoutButtonVisible) {
      return;
    }
    final next = _myWorkoutsLayout == MyWorkoutsLayout.cards
        ? MyWorkoutsLayout.list
        : MyWorkoutsLayout.cards;
    setState(() => _myWorkoutsLayout = next);
    _notifyMyWorkoutsLayoutChrome();
    unawaited(MyWorkoutsLayoutStorage.setLayout(next));
  }

  void _onSportTypeToggled(String id) {
    setState(() {
      if (_includedSportTypes.contains(id)) {
        _includedSportTypes = Set<String>.from(_includedSportTypes)..remove(id);
      } else {
        _includedSportTypes = Set<String>.from(_includedSportTypes)..add(id);
      }
    });
  }

  /// Null when all used sports are included (no query param).
  /// Empty list when none included (`sport_types=`).
  List<String>? get _ownListSportTypes {
    if (_usedSportTypes.isEmpty) {
      return null;
    }
    if (_includedSportTypes.length == _usedSportTypes.length &&
        _usedSportTypes.every(_includedSportTypes.contains)) {
      return null;
    }
    return _usedSportTypes
        .where(_includedSportTypes.contains)
        .toList(growable: false);
  }

  @override
  Widget build(BuildContext context) {
    if (widget.nickname == null) {
      return WelcomeGuestView(
        onSignIn: widget.onSignIn ?? () {},
        onRegister: widget.onRegister ?? () {},
        showMobileServerHint: isMobileClient,
      );
    }

    final l10n = AppLocalizations.of(context)!;
    final tabController = _tabController;

    if (_isLoadingTabs || tabController == null) {
      return const Center(child: CircularProgressIndicator());
    }

    final viewingWorkout = widget.viewingWorkout;
    final showWorkoutDetail = viewingWorkout != null && _authToken != null;
    final feedPhotoWorkout = widget.feedPhotoViewerWorkout;
    final photoViewerIndex = widget.photoViewerIndex;
    final showFeedPhotoViewer = !showWorkoutDetail &&
        feedPhotoWorkout != null &&
        photoViewerIndex != null &&
        feedPhotoWorkout.mediaFiles.isNotEmpty &&
        _authToken != null;

    final tabs = <Widget>[
      if (_showFeedTab) Tab(text: l10n.homeTabFeed),
      Tab(text: l10n.homeTabMyWorkouts),
    ];

    final ownSportTypes = _ownListSportTypes;
    final ownEmptyMessage = ownSportTypes != null
        ? l10n.noWorkoutsMatchSportFilter
        : null;

    final tabViews = <Widget>[
      if (_showFeedTab)
        WorkoutFeedList(
          key: _feedListKeys[HomeFeedTab.feed],
          nickname: widget.nickname!,
          scope: 'feed',
          scrollController: _scrollControllers[HomeFeedTab.feed]!,
          refreshToken: widget.refreshToken,
          federationEnabled: widget.federationEnabled,
          onWorkoutTap: _openWorkout,
          onPhotoTap: _openWorkoutPhoto,
          onAuthTokenLoaded: (token) {
            if (_authToken != token) {
              setState(() => _authToken = token);
            }
          },
        ),
      WorkoutFeedList(
        key: _feedListKeys[HomeFeedTab.myWorkouts],
        nickname: widget.nickname!,
        scope: 'own',
        scrollController: _scrollControllers[HomeFeedTab.myWorkouts]!,
        refreshToken: widget.refreshToken,
        federationEnabled: widget.federationEnabled,
        sportTypes: ownSportTypes,
        emptyMessage: ownEmptyMessage,
        layout: _myWorkoutsLayout,
        onWorkoutTap: _openWorkout,
        onPhotoTap: _openWorkoutPhoto,
        onAuthTokenLoaded: (token) {
          if (_authToken == null) {
            setState(() => _authToken = token);
          }
        },
      ),
    ];

    final showFilterPanel =
        _filterExpanded && _activeTab == HomeFeedTab.myWorkouts;

    // Keep the feed list mounted while viewing a workout so scroll position
    // and loaded pages survive back navigation.
    return Stack(
      children: [
        Positioned.fill(
          child: Offstage(
            offstage: showWorkoutDetail,
            child: TickerMode(
              enabled: !showWorkoutDetail,
              child: Stack(
                children: [
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: [
                      Material(
                        color: Theme.of(context).colorScheme.surface,
                        child: TabBar(
                          controller: tabController,
                          onTap: _onTabTapped,
                          tabs: tabs,
                        ),
                      ),
                      AnimatedSize(
                        duration: const Duration(milliseconds: 200),
                        curve: Curves.easeInOut,
                        alignment: Alignment.topCenter,
                        child: showFilterPanel
                            ? Material(
                                color: Theme.of(context).colorScheme.surface,
                                child: SportTypeToggleWrap(
                                  sportTypeIds: _usedSportTypes,
                                  selectedIds: _includedSportTypes,
                                  onToggle: _onSportTypeToggled,
                                ),
                              )
                            : const SizedBox(width: double.infinity),
                      ),
                      Expanded(
                        child: TabBarView(
                          controller: tabController,
                          children: tabViews,
                        ),
                      ),
                    ],
                  ),
                  if (showFeedPhotoViewer)
                    Positioned.fill(
                      child: WorkoutPhotoViewer(
                        workout: feedPhotoWorkout,
                        authToken: _authToken!,
                        initialIndex: photoViewerIndex.clamp(
                          0,
                          feedPhotoWorkout.mediaFiles.length - 1,
                        ),
                      ),
                    ),
                ],
              ),
            ),
          ),
        ),
        if (showWorkoutDetail)
          Positioned.fill(
            child: WorkoutDetailView(
              workout: viewingWorkout,
              authToken: _authToken!,
              federationEnabled: widget.federationEnabled,
              isMapExpanded: widget.isMapExpanded,
              onMapExpandedChanged: widget.onMapExpandedChanged,
              photoViewerIndex: widget.photoViewerIndex,
              onPhotoViewerIndexChanged: widget.onPhotoViewerIndexChanged,
            ),
          ),
      ],
    );
  }
}
