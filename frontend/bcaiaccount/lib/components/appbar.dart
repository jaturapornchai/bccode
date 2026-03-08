import 'package:flutter/material.dart';

class BaseAppBar extends StatelessWidget implements PreferredSizeWidget {
  final Text title;
  final AppBar appBar;
  final List<Widget> widgets;
  final IconButton? leading;

  const BaseAppBar({
    super.key,
    required this.title,
    required this.appBar,
    required this.widgets,
    this.leading,
  });

  @override
  Widget build(BuildContext context) {
    return (leading == null)
        ? AppBar(
            title: title,
            actions: const [
              Padding(
                padding: EdgeInsets.all(8.0),
                child: CircleAvatar(
                  radius: 25.0,
                  backgroundImage: AssetImage('assets/img/vficon.png'),
                ),
              ),
            ],
            centerTitle: true,
            foregroundColor: Colors.black,
            backgroundColor: Colors.white,
          )
        : AppBar(
            title: title,
            leading: leading,
            actions: const [
              Padding(
                padding: EdgeInsets.all(8.0),
                child: CircleAvatar(
                  radius: 25.0,
                  backgroundImage: AssetImage('assets/img/vficon.png'),
                ),
              ),
            ],
            centerTitle: true,
            foregroundColor: Colors.black,
            backgroundColor: Colors.white,
          );
  }

  @override
  Size get preferredSize => Size.fromHeight(appBar.preferredSize.height);
}
