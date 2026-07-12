enum GromDestination {
  home,
  userSearch,
  profile,
  equipment,
  integration,
  login,
  register,
  about,
  settings,
}

extension GromDestinationNavigation on GromDestination {
  bool get isHome => this == GromDestination.home;
}
