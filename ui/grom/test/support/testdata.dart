import 'dart:io';

File testdataFile(String relative) {
  final cwd = Directory.current.path;
  final root = cwd.contains('ui/grom')
      ? Directory('$cwd/../..').absolute.path
      : Directory('$cwd/../../..').absolute.path;
  return File('$root/testdata/$relative');
}
