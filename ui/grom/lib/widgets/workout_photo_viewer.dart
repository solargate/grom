import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';
import 'package:photo_view/photo_view.dart';
import 'package:photo_view/photo_view_gallery.dart';

import '../api_request.dart';
import '../models/workout.dart';

class WorkoutPhotoViewer extends StatefulWidget {
  const WorkoutPhotoViewer({
    super.key,
    required this.workout,
    required this.authToken,
    required this.initialIndex,
  });

  final Workout workout;
  final String authToken;
  final int initialIndex;

  @override
  State<WorkoutPhotoViewer> createState() => _WorkoutPhotoViewerState();
}

class _WorkoutPhotoViewerState extends State<WorkoutPhotoViewer> {
  late final PageController _pageController;
  late int _currentIndex;
  final _api = ApiRequest();

  @override
  void initState() {
    super.initState();
    _currentIndex = widget.initialIndex;
    _pageController = PageController(initialPage: widget.initialIndex);
  }

  @override
  void dispose() {
    _pageController.dispose();
    super.dispose();
  }

  Map<String, String> get _headers => {
        'Authorization': 'Bearer ${widget.authToken}',
      };

  String _originalUrl(String filename) {
    final owner = widget.workout.ownerNickname;
    return _api.mediaOriginalUrl(
      widget.workout.id,
      filename,
      owner: owner.isNotEmpty ? owner : null,
    );
  }

  void _handlePointerSignal(PointerSignalEvent event) {
    if (event is! PointerScrollEvent || widget.workout.mediaFiles.length <= 1) {
      return;
    }
    final delta = event.scrollDelta.dy + event.scrollDelta.dx;
    if (delta > 0 && _currentIndex < widget.workout.mediaFiles.length - 1) {
      _pageController.nextPage(
        duration: const Duration(milliseconds: 200),
        curve: Curves.easeOut,
      );
    } else if (delta < 0 && _currentIndex > 0) {
      _pageController.previousPage(
        duration: const Duration(milliseconds: 200),
        curve: Curves.easeOut,
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final files = widget.workout.mediaFiles;

    return Material(
      color: Colors.black,
      child: SafeArea(
        child: Listener(
          onPointerSignal: _handlePointerSignal,
          child: Stack(
            children: [
              PhotoViewGallery.builder(
                pageController: _pageController,
                itemCount: files.length,
                onPageChanged: (index) => setState(() => _currentIndex = index),
                builder: (context, index) {
                  return PhotoViewGalleryPageOptions(
                    imageProvider: NetworkImage(
                      _originalUrl(files[index]),
                      headers: _headers,
                    ),
                    minScale: PhotoViewComputedScale.contained,
                    initialScale: PhotoViewComputedScale.contained,
                    heroAttributes: PhotoViewHeroAttributes(tag: files[index]),
                  );
                },
                scrollPhysics: const BouncingScrollPhysics(),
                backgroundDecoration: const BoxDecoration(color: Colors.black),
              ),
              if (files.length > 1)
                Positioned(
                  bottom: 16,
                  left: 0,
                  right: 0,
                  child: Center(
                    child: Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 12,
                        vertical: 6,
                      ),
                      decoration: BoxDecoration(
                        color: Colors.black54,
                        borderRadius: BorderRadius.circular(16),
                      ),
                      child: Text(
                        '${_currentIndex + 1} / ${files.length}',
                        style: const TextStyle(color: Colors.white),
                      ),
                    ),
                  ),
                ),
            ],
          ),
        ),
      ),
    );
  }
}
