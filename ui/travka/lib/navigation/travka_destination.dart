enum TravkaDestination {
  home,
  userSearch,
  profile,
  equipment,
  login,
  register,
  settings,
}

extension TravkaDestinationNavigation on TravkaDestination {
  bool get isHome => this == TravkaDestination.home;
}
