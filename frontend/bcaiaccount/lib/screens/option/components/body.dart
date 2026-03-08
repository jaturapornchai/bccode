
import 'package:smlaicloud/screens/Option/option_add.dart';
import 'package:smlaicloud/bloc/option/option_bloc.dart';
import 'package:smlaicloud/model/option_model.dart';
import 'package:flutter/material.dart';
import 'package:smlaicloud/components/background_main.dart';
import 'package:smlaicloud/components/textfield_search.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

class Body extends StatefulWidget {
  const Body({super.key});

  @override
  State<Body> createState() => _BodyState();
}

class _BodyState extends State<Body> {
  final List<OptionModel> _option = [];

  final ScrollController _scrollController = ScrollController();

  final int _perPage = 10;
  final String _search = "";

  @override
  initState() {
    super.initState();
    context.read<OptionBloc>().add(
      ListOptionLoad(
        page: 0,
        perPage: _perPage,
        search: _search,
        nextPage: true,
      ),
    );
    _scrollController.addListener(_onScroll);
  }

  void _onScroll() {
    if (_scrollController.offset >=
            _scrollController.position.maxScrollExtent &&
        !_scrollController.position.outOfRange) {
      OptionBloc optionBloc = context.read<OptionBloc>();
      OptionLoadSuccess optionLoadSuccess;

      if (optionBloc.state is OptionLoadSuccess) {
        optionLoadSuccess = (optionBloc.state as OptionLoadSuccess);
        if (optionLoadSuccess.page!.page < optionLoadSuccess.page!.totalPage) {
          optionBloc.add(
            ListOptionLoad(
              page: optionLoadSuccess.page!.page + 1,
              perPage: _perPage,
              search: _search,
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
    Size size = MediaQuery.of(context).size;

    return MultiBlocListener(
      listeners: [
        BlocListener<OptionBloc, OptionState>(
          listener: (context, state) {
            if (state is OptionLoadSuccess) {
              if (state.option.isNotEmpty) {
                //_option = state.option;
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
              Card(child: TextFieldSearch(onChange: (value) {})),
              Expanded(
                child: SizedBox(
                  width: double.infinity,
                  child: Column(
                    children: [
                      Expanded(
                        child: BlocBuilder<OptionBloc, OptionState>(
                          builder: (context, state) {
                            return (state is OptionInProgress)
                                ? Container(
                                    color: Colors.transparent,
                                    alignment: Alignment.center,
                                    child: const CircularProgressIndicator(),
                                  )
                                : (state is OptionLoadSuccess)
                                ? ListView.builder(
                                    controller: _scrollController,
                                    itemCount: _option.length,
                                    itemBuilder: (context, index) {
                                      return Card(
                                        elevation: 1,
                                        child: ListTile(
                                          onTap: () {
                                            Navigator.pushReplacement(
                                              context,
                                              MaterialPageRoute(
                                                builder: (context) => OptionAdd(
                                                  guidfixed: state
                                                      .option[index]
                                                      .guidfixed,
                                                ),
                                              ),
                                            );
                                          },
                                          title: Text(
                                            _option[index].name1.toString(),
                                          ),
                                          subtitle: Text(
                                            'ตัวเลือกย่อย: ${_option[index].choices.length}',
                                            style: const TextStyle(
                                              color: Colors.black,
                                            ),
                                          ),
                                          trailing: const Icon(
                                            Icons.keyboard_arrow_right_sharp,
                                          ),
                                        ),
                                      );
                                    },
                                  )
                                : (state is OptionLoadFailed)
                                ? const Text(
                                    'ไม่เจอข้อมูล',
                                    style: TextStyle(fontSize: 24),
                                  )
                                : Container();
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
