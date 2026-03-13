// ignore_for_file: prefer_const_constructors

import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/screens/product/product_add.dart';
import 'package:smlaicloud/bloc/inventory/inventory_bloc.dart';
import 'package:smlaicloud/components/background_main.dart';
import 'package:smlaicloud/components/textfield_search.dart';
import 'package:smlaicloud/utils/util.dart';
import '../../../global.dart' as global;

class Body extends StatefulWidget {
  const Body({super.key});

  @override
  State<Body> createState() => _BodyState();
}

class _BodyState extends State<Body> with global.ThemeRefreshMixin {
  final ScrollController _scrollController = ScrollController();
  double currentScroll = 0.0;

  int perPage = 8;
  String search = "";

  @override
  initState() {
    context.read<InventoryBloc>().add(
      ListInventoryLoad(
        page: 0,
        perPage: perPage,
        search: search,
        nextPage: true,
      ),
    );
    _scrollController.addListener(_onScroll);
    super.initState();
  }

  void _searchText(String enteredKeyword) {
  }

  void _onScroll() {
    if (_scrollController.offset >=
            _scrollController.position.maxScrollExtent &&
        !_scrollController.position.outOfRange) {
      InventoryBloc inventoryBloc = context.read<InventoryBloc>();
      InventoryLoadSuccess inventoryLoadSuccess;

      setState(() {
        currentScroll = _scrollController.position.maxScrollExtent;
      });

      if (inventoryBloc.state is InventoryLoadSuccess) {
        inventoryLoadSuccess = (inventoryBloc.state as InventoryLoadSuccess);
        if (inventoryLoadSuccess.page!.page <
            inventoryLoadSuccess.page!.totalPage) {
          inventoryBloc.add(
            ListInventoryLoad(
              page: inventoryLoadSuccess.page!.page + 1,
              perPage: perPage,
              search: search,
              nextPage: false,
            ),
          );
        }
      }
    }
  }

  @override
  void dispose() {
    _scrollController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return MultiBlocListener(
      listeners: [
        BlocListener<InventoryBloc, InventoryState>(
          listener: (context, state) {
            if (state is InventoryLoadSuccess) {
              if (currentScroll > 0) {
                Future.delayed(const Duration(milliseconds: 100), () {
                  setState(() {
                    _scrollController.jumpTo(currentScroll);
                  });
                });
              }
            }
          },
        ),
      ],
      child: BackgroundMain(
        child: Padding(
          padding: const EdgeInsets.all(8.0),
          child: Column(
            children: [
              Card(
                child: TextFieldSearch(
                  onChange: (value) {
                    _searchText(value);
                  },
                ),
              ),
              Expanded(
                child: SizedBox(
                  width: double.infinity,
                  child: Column(
                    children: [
                      Expanded(
                        child: BlocBuilder<InventoryBloc, InventoryState>(
                          builder: (context, state) {
                            // แสดง Loading
                            if (state is InventoryInProgress) {
                              return Center(
                                child: Column(
                                  mainAxisAlignment: MainAxisAlignment.center,
                                  children: [
                                    CircularProgressIndicator(),
                                    SizedBox(height: 16),
                                    Text(
                                      'กำลังโหลดข้อมูลสินค้า...',
                                      style: TextStyle(fontSize: 16),
                                    ),
                                  ],
                                ),
                              );
                            }

                            // แสดง Error
                            if (state is InventoryLoadFailed) {
                              return Center(
                                child: Column(
                                  mainAxisAlignment: MainAxisAlignment.center,
                                  children: [
                                    Icon(
                                      Icons.error_outline,
                                      size: 64,
                                      color: global.theme.negativeHighlightTextColor,
                                    ),
                                    SizedBox(height: 16),
                                    Text(
                                      global.language('error_occurred'),
                                      style: TextStyle(
                                        fontSize: 20,
                                        fontWeight: FontWeight.bold,
                                      ),
                                    ),
                                    SizedBox(height: 8),
                                    Text(
                                      state.message,
                                      style: TextStyle(
                                        fontSize: 14,
                                        color: global.theme.textSecondaryColor,
                                      ),
                                      textAlign: TextAlign.center,
                                    ),
                                    SizedBox(height: 16),
                                    ElevatedButton.icon(
                                      onPressed: () {
                                        context.read<InventoryBloc>().add(
                                          ListInventoryLoad(
                                            page: 0,
                                            perPage: perPage,
                                            search: search,
                                            nextPage: true,
                                          ),
                                        );
                                      },
                                      icon: Icon(Icons.refresh),
                                      label: Text(global.language('try_again')),
                                    ),
                                  ],
                                ),
                              );
                            }

                            // แสดง Success
                            if (state is InventoryLoadSuccess) {
                              if (state.inventory.isEmpty) {
                                return Center(
                                  child: Text(
                                    'ไม่มีข้อมูลสินค้า',
                                    style: TextStyle(fontSize: 18),
                                  ),
                                );
                              }

                              return GridView.builder(
                                controller: _scrollController,
                                shrinkWrap: true,
                                physics: BouncingScrollPhysics(),
                                itemCount: state.inventory.length,
                                gridDelegate:
                                    SliverGridDelegateWithFixedCrossAxisCount(
                                      childAspectRatio: 1,
                                      crossAxisCount:
                                          (Util.isLandscape(context)) ? 5 : 2,
                                      crossAxisSpacing: 4.0,
                                      mainAxisSpacing: 4.0,
                                      mainAxisExtent: 208,
                                    ),
                                itemBuilder: (context, index) {
                                  return Card(
                                    elevation: 3,
                                    shape: RoundedRectangleBorder(
                                      borderRadius: BorderRadius.circular(8.0),
                                    ),
                                    child: Padding(
                                      padding: EdgeInsets.all(8.0),
                                      child: InkWell(
                                        onTap: () {
                                          Navigator.pushReplacement(
                                            context,
                                            MaterialPageRoute(
                                              builder: (context) => ProductAdd(
                                                guidfixed: state
                                                    .inventory[index]
                                                    .guidfixed,
                                              ),
                                            ),
                                          );
                                        },
                                        child: Column(
                                          children: <Widget>[
                                            SizedBox(
                                              height: 100,
                                              child: Image.asset(
                                                "assets/img/vficon.png",
                                                fit: BoxFit.fill,
                                              ),
                                            ),
                                            Expanded(
                                              child: Column(
                                                children: [
                                                  Align(
                                                    alignment:
                                                        Alignment.topLeft,
                                                    child: Text(
                                                      state
                                                          .inventory[index]
                                                          .name1,
                                                      overflow:
                                                          TextOverflow.ellipsis,
                                                      style: TextStyle(
                                                        fontSize: 18,
                                                        color: global.theme.textColor,
                                                        fontWeight:
                                                            FontWeight.bold,
                                                      ),
                                                    ),
                                                  ),
                                                  Align(
                                                    alignment:
                                                        Alignment.topLeft,
                                                    child: Text(
                                                      state
                                                          .inventory[index]
                                                          .barcode,
                                                      overflow:
                                                          TextOverflow.ellipsis,
                                                      style: TextStyle(
                                                        color: global.theme.textSecondaryColor,
                                                      ),
                                                    ),
                                                  ),
                                                  Align(
                                                    alignment:
                                                        Alignment.topRight,
                                                    child: Text(
                                                      state
                                                          .inventory[index]
                                                          .price
                                                          .toString(),
                                                      style: TextStyle(
                                                        color: global.theme.warningHighlightTextColor,
                                                        fontSize: 24,
                                                        fontWeight:
                                                            FontWeight.bold,
                                                      ),
                                                    ),
                                                  ),
                                                ],
                                              ),
                                            ),
                                          ],
                                        ),
                                      ),
                                    ),
                                  );
                                },
                              );
                            }

                            // Default: แสดง Container ว่าง
                            return Container();
                          },
                        ),
                      ),
                    ],
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
