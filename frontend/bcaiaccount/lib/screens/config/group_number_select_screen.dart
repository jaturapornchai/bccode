import 'package:smlaicloud/screens/config/kitchen_screen.dart';
import 'package:smlaicloud/screens/config/product_category_screen.dart';
import 'package:smlaicloud/screens/config/qrcode_order_screen.dart';
import 'package:smlaicloud/screens/config/table_order_screen.dart';
import 'package:smlaicloud/screens/config/table_screen.dart';
import 'package:smlaicloud/screens/config/zone_screen.dart';
import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;

class GroupNumberSelectScreen extends StatefulWidget {
  final global.SelectGroupNumberEnum type;
  const GroupNumberSelectScreen({super.key, required this.type});

  @override
  State<GroupNumberSelectScreen> createState() =>
      GroupNumberSelectScreenState();
}

class GroupNumberSelectScreenState extends State<GroupNumberSelectScreen>
    with SingleTickerProviderStateMixin, global.ThemeRefreshMixin {
  @override
  void initState() {
    super.initState();
  }

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (context, constraints) {
        return Scaffold(
          resizeToAvoidBottomInset: true,
          appBar: AppBar(
            backgroundColor: global.theme.appBarColor,
            automaticallyImplyLeading: false,
            title: Text(global.language('please_select_group_number')),
            leading: IconButton(
              focusNode: FocusNode(skipTraversal: true),
              icon: Icon(Icons.arrow_back),
              onPressed: () {
                global.gotoMainMenu(context);
              },
            ),
          ),
          body: Column(
            children: [
              Expanded(
                child: GridView.builder(
                  padding: const EdgeInsets.all(10),
                  gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
                    crossAxisCount: (constraints.maxWidth > 800) ? 6 : 3,
                    crossAxisSpacing: 10,
                    mainAxisSpacing: 10,
                    childAspectRatio: 1,
                  ),
                  itemCount: global.groupNumber.length,
                  itemBuilder: (context, index) {
                    return InkWell(
                      onTap: () {
                        if (widget.type == global.SelectGroupNumberEnum.table) {
                          Navigator.push(
                            context,
                            MaterialPageRoute(
                              builder: (context) =>
                                  TableScreen(groupnumber: (index + 1)),
                            ),
                          );
                        } else if (widget.type ==
                            global.SelectGroupNumberEnum.tableOrder) {
                          Navigator.push(
                            context,
                            MaterialPageRoute(
                              builder: (context) =>
                                  TableOrderScreen(groupnumber: (index + 1)),
                            ),
                          );
                        } else if (widget.type ==
                            global.SelectGroupNumberEnum.zone) {
                          Navigator.push(
                            context,
                            MaterialPageRoute(
                              builder: (context) =>
                                  ZoneScreen(groupnumber: (index + 1)),
                            ),
                          );
                        } else if (widget.type ==
                            global.SelectGroupNumberEnum.category) {
                          Navigator.push(
                            context,
                            MaterialPageRoute(
                              builder: (context) => ProductCategoryScreen(
                                groupnumber: (index + 1),
                              ),
                            ),
                          );
                        } else if (widget.type ==
                            global.SelectGroupNumberEnum.kitchen) {
                          Navigator.push(
                            context,
                            MaterialPageRoute(
                              builder: (context) =>
                                  KitchenScreen(groupnumber: (index + 1)),
                            ),
                          );
                        } else if (widget.type ==
                            global.SelectGroupNumberEnum.genQrcode) {
                          Navigator.push(
                            context,
                            MaterialPageRoute(
                              builder: (context) =>
                                  QrcodeOrderScreen(groupnumber: (index + 1)),
                            ),
                          );
                        }
                      },
                      child: Container(
                        decoration: BoxDecoration(
                          color: global.theme.cardColor,
                          border: Border.all(color: global.theme.dividerBorderColor),
                        ),
                        child: Column(
                          children: [
                            Expanded(
                              child: Padding(
                                padding: const EdgeInsets.only(
                                  top: 10,
                                  left: 10,
                                  right: 10,
                                ),
                                child: Column(
                                  mainAxisAlignment: MainAxisAlignment.center,
                                  children: [
                                    Text(
                                      global.language('select_group_number'),
                                      maxLines: 2,
                                      overflow: TextOverflow.ellipsis,
                                      style: TextStyle(
                                        fontSize: 14,
                                        fontWeight: FontWeight.bold,
                                      ),
                                    ),
                                    const SizedBox(height: 10),
                                    Text(
                                      "${global.groupNumber[index]}",
                                      maxLines: 2,
                                      overflow: TextOverflow.ellipsis,
                                      style: TextStyle(
                                        fontSize: 20,
                                        fontWeight: FontWeight.bold,
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                            ),
                          ],
                        ),
                      ),
                    );
                  },
                ),
              ),
            ],
          ),
        );
      },
    );
  }
}
