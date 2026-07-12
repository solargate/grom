enum GromDestination {
  home,
  userSearch,
  profile,
  equipment,
  login,
  register,
  about,
  settings,
}

extension GromDestinationNavigation on GromDestination {
  bool get isHome => this == GromDestination.home;
}
