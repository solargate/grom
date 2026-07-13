import 'package:flutter/material.dart';
import 'package:grom/l10n/app_localizations.dart';

import '../api_request.dart';
import '../auth_storage.dart';
import '../models/workout.dart';
import '../pages/workout_detail_page.dart';
import '../widgets/workout_feed_list.dart';
import '../widgets/workout_photo_viewer.dart';

enum HomeFeedTab { feed, myWorkouts }

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
    _loadTabState(resetToDefaultTab: true);
  }

  @override
  void didUpdateWidget(covariant HomePage oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (widget.nickname != oldWidget.nickname) {
      _loadTabState(resetToDefaultTab: true);
    } else if (widget.refreshToken != oldWidget.refreshToken) {
      _loadTabState();
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
      });
      _syncTabController(showFeedTab: false, preferredTab: HomeFeedTab.myWorkouts);
      return;
    }

    setState(() => _isLoadingTabs = true);

    try {
      final token = await AuthStorage.getToken();
      if (token == null) {
        throw ApiException('Not authenticated');
      }

      final following = await _api.listFollowing(token);
      if (!mounted) return;

      final showFeedTab =
          following.any((follow) => follow.status == 'active');
      final preferredTab = showFeedTab && resetToDefaultTab
          ? HomeFeedTab.feed
          : _activeTab == HomeFeedTab.feed && !showFeedTab
              ? HomeFeedTab.myWorkouts
              : _activeTab;

      setState(() {
        _showFeedTab = showFeedTab;
        _activeTab = preferredTab;
        _isLoadingTabs = false;
      });
      _syncTabController(showFeedTab: showFeedTab, preferredTab: preferredTab);
    } catch (_) {
      if (!mounted) return;
      setState(() {
        _showFeedTab = false;
        _isLoadingTabs = false;
      });
      _syncTabController(showFeedTab: false, preferredTab: HomeFeedTab.myWorkouts);
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
      setState(() => _activeTab = tab);
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


  @override
  Widget build(BuildContext context) {
    if (widget.nickname == null) {
      return const SizedBox.shrink();
    }

    final viewingWorkout = widget.viewingWorkout;
    if (viewingWorkout != null && _authToken != null) {
      return WorkoutDetailView(
        workout: viewingWorkout,
        authToken: _authToken!,
        federationEnabled: widget.federationEnabled,
        isMapExpanded: widget.isMapExpanded,
        onMapExpandedChanged: widget.onMapExpandedChanged,
        photoViewerIndex: widget.photoViewerIndex,
        onPhotoViewerIndexChanged: widget.onPhotoViewerIndexChanged,
      );
    }

    final l10n = AppLocalizations.of(context)!;
    final tabController = _tabController;

    if (_isLoadingTabs || tabController == null) {
      return const Center(child: CircularProgressIndicator());
    }

    final feedPhotoWorkout = widget.feedPhotoViewerWorkout;
    final photoViewerIndex = widget.photoViewerIndex;
    final showFeedPhotoViewer = feedPhotoWorkout != null &&
        photoViewerIndex != null &&
        feedPhotoWorkout.mediaFiles.isNotEmpty &&
        _authToken != null;

    final tabs = <Widget>[
      if (_showFeedTab) Tab(text: l10n.homeTabFeed),
      Tab(text: l10n.homeTabMyWorkouts),
    ];

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
        onWorkoutTap: _openWorkout,
        onPhotoTap: _openWorkoutPhoto,
        onAuthTokenLoaded: (token) {
          if (_authToken == null) {
            setState(() => _authToken = token);
          }
        },
      ),
    ];

    return Stack(
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
    );
  }
}
