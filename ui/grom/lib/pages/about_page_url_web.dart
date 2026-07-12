import 'package:web/web.dart' as web;

void openCopyrightUrlInBrowser(String url) {
  web.window.open(url, '_blank');
}
