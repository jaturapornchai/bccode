import 'package:flutter/material.dart';
import 'package:url_launcher/url_launcher.dart';
import 'package:smlaicloud/global.dart' as global;

const String _manualBaseUrl = 'https://www.bcaccount.com/manual';

class ManualButton extends StatelessWidget {
  final String? path;

  const ManualButton({super.key, this.path});

  @override
  Widget build(BuildContext context) {
    return IconButton(
      focusNode: FocusNode(skipTraversal: true),
      icon: const Icon(Icons.menu_book, size: 26.0),
      tooltip: global.language('manual'),
      onPressed: () {
        final url = path != null ? '$_manualBaseUrl/$path' : _manualBaseUrl;
        launchUrl(Uri.parse(url), mode: LaunchMode.externalApplication);
      },
    );
  }
}
