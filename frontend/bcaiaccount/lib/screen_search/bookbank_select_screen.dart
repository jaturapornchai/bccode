import 'package:smlaicloud/bloc/book_bank/book_bank_bloc.dart';
import 'package:smlaicloud/model/book_bank_model.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/global.dart' as global;

class BookBankSelectScreen extends StatefulWidget {
  const BookBankSelectScreen({super.key});

  @override
  State<BookBankSelectScreen> createState() => BookBankSelectScreenState();
}

class BookBankSelectScreenState extends State<BookBankSelectScreen>
    with SingleTickerProviderStateMixin {
  TextEditingController searchController = TextEditingController();
  FocusNode searchFocusNode = FocusNode(skipTraversal: true);
  ScrollController listScrollController = ScrollController();
  String searchText = "";
  List<BookBankModel> bookbankListData = [];
  bool isKeyUp = false;
  bool isKeyDown = false;
  String selectGuid = "";
  int currentListIndex = 0;

  void setSystemLanguageList() async {
    await global.setSystemLanguage(context);

    loadDataList("");
  }

  @override
  void initState() {
    setSystemLanguageList();

    super.initState();
  }

  void loadDataList(String search) {
    context.read<BookBankBloc>().add(
      BookBankLoadList(
        offset: (bookbankListData.isEmpty) ? 0 : bookbankListData.length,
        limit: global.loadDataPerPage,
        search: search,
      ),
    );
  }

  void onScrollList() {
    if (listScrollController.offset >=
            listScrollController.position.maxScrollExtent &&
        !listScrollController.position.outOfRange) {
      loadDataList(searchText);
    }
  }

  @override
  void dispose() {
    listScrollController.dispose();
    searchController.dispose();
    super.dispose();
  }

  Widget listScreen({bool mobileScreen = false}) {
    return Scaffold(
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        automaticallyImplyLeading: false,
        title: Text(global.language('bookbank')),
        leading: IconButton(
          focusNode: FocusNode(skipTraversal: true),
          icon: const Icon(Icons.arrow_back),
          onPressed: () {
            Navigator.pop(context);
          },
        ),
      ),
      body: Focus(
        focusNode: FocusNode(skipTraversal: true, canRequestFocus: true),
        onKey: (node, event) {
          if (kIsWeb) {
            if (event is RawKeyDownEvent) {
              if (event.logicalKey == LogicalKeyboardKey.arrowUp) {
                isKeyDown = false;
                int index = bookbankListData.indexOf(
                  bookbankListData.firstWhere(
                    (element) => element.guidfixed == selectGuid,
                  ),
                );
                if (index > 0) {
                  selectGuid = bookbankListData[index - 1].guidfixed!;
                  currentListIndex = index + 1;
                  isKeyUp = true;
                }
              }
              if (event.logicalKey == LogicalKeyboardKey.arrowDown) {
                isKeyUp = false;
                int index = bookbankListData.indexOf(
                  bookbankListData.firstWhere(
                    (element) => element.guidfixed == selectGuid,
                  ),
                );
                selectGuid = bookbankListData[index + 1].guidfixed!;
                currentListIndex = index + 1;
                isKeyDown = true;
              }
            }
          }
          return KeyEventResult.ignored;
        },
        child: GridView.count(
          crossAxisCount: (mobileScreen) ? 2 : 4,
          children: bookbankListData.map((value) => listObject(value)).toList(),
        ),
      ),
    );
  }

  Widget listObject(BookBankModel value) {
    return InkWell(
      onTap: () {
        Navigator.pop(context, value);
      },
      child: Card(
        elevation: 0,
        color: Colors.transparent,
        child: Column(
          mainAxisSize: MainAxisSize.max,
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            (value.images!.isNotEmpty)
                ? Image.network(
                    value.images![0].uri,
                    fit: BoxFit.fill,
                    height: 150,
                    width: 150,
                  )
                : Image.asset(
                    'assets/img/noimage.png',
                    fit: BoxFit.fill,
                    height: 150,
                    width: 150,
                  ),
            const SizedBox(height: 8),
            Text(value.bookcode!),
            Text(value.passbook!),
            Text(global.packName(value.banknames!)),
            Text(global.packName(value.names!)),
          ],
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      resizeToAvoidBottomInset: true,
      body: LayoutBuilder(
        builder: (context, constraints) {
          return BlocListener<BookBankBloc, BookBankState>(
            listener: (context, state) {
              // Load
              if (state is BookBankLoadSuccess) {
                setState(() {
                  if (state.bookBanks.isNotEmpty) {
                    bookbankListData.addAll(state.bookBanks);
                    if (bookbankListData.isNotEmpty) {
                      selectGuid = bookbankListData[0].guidfixed!;
                    } else {
                      selectGuid = "";
                    }
                  }
                });
              }
            },
            child: (constraints.maxWidth > 800)
                ? listScreen(mobileScreen: false)
                : listScreen(mobileScreen: true),
          );
        },
      ),
    );
  }
}
