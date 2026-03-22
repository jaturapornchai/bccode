import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/global_model.dart';

class LanguageSelectorWidget extends StatelessWidget {
  final List<String> selectedLanguages;
  final List<LanguageModel> defaultLanguageList;
  final bool isSaveAllow;
  final ValueChanged<List<String>> onLanguagesChanged;

  const LanguageSelectorWidget({
    super.key,
    required this.selectedLanguages,
    required this.defaultLanguageList,
    required this.isSaveAllow,
    required this.onLanguagesChanged,
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: EdgeInsets.only(left: 10, right: 10),
      child: Center(
        child: Container(
          margin: const EdgeInsets.only(top: 10, bottom: 10),
          padding: EdgeInsets.all(10),
          decoration: BoxDecoration(
            border: Border.all(color: global.theme.textSecondaryColor),
            borderRadius: BorderRadius.circular(5),
            color: global.theme.cardColor,
          ),
          child: Column(
            children: [
              Text(
                global.language('select_data_language'),
                style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold),
              ),
              const SizedBox(height: 10),
              for (var i = 0; i < selectedLanguages.length; i++)
                Container(
                  margin: const EdgeInsets.only(bottom: 10),
                  child: Row(
                    children: [
                      Expanded(
                        child: Row(
                          children: [
                            SizedBox(
                              width: 30,
                              child: Text((i + 1).toString()),
                            ),
                            Expanded(
                              child: ElevatedButton(
                                focusNode: FocusNode(skipTraversal: true),
                                onPressed: () {
                                  if (isSaveAllow) {
                                    _showLanguageDialog(context, i);
                                  }
                                },
                                child: Row(
                                  children: [
                                    if (selectedLanguages[i].isNotEmpty)
                                      Image.asset(
                                        'assets/flags/${selectedLanguages[i]}.png',
                                        width: 30,
                                        height: 30,
                                      ),
                                    const SizedBox(width: 10),
                                    if (selectedLanguages[i].isNotEmpty)
                                      Text(
                                        defaultLanguageList[defaultLanguageList
                                                .indexWhere(
                                                  (element) =>
                                                      element.code ==
                                                      selectedLanguages[i],
                                                )]
                                            .name!,
                                      )
                                    else
                                      Text(global.language('select_language')),
                                  ],
                                ),
                              ),
                            ),
                          ],
                        ),
                      ),
                      if (i == 0)
                        const SizedBox(width: 50)
                      else
                        SizedBox(
                          width: 50,
                          child: IconButton(
                            focusNode: FocusNode(skipTraversal: true),
                            onPressed: () {
                              List<String> newList = List.from(
                                selectedLanguages,
                              );
                              newList.removeAt(i);
                              onLanguagesChanged(newList);
                            },
                            color: global.theme.negativeHighlightTextColor,
                            icon: Icon(Icons.delete),
                          ),
                        ),
                    ],
                  ),
                ),
              const SizedBox(height: 10),
              if (isSaveAllow)
                Row(
                  children: [
                    ElevatedButton(
                      focusNode: FocusNode(skipTraversal: true),
                      onPressed: () {
                        List<String> newList = List.from(selectedLanguages);
                        newList.add("");
                        onLanguagesChanged(newList);
                      },
                      child: Text(global.language('add_language')),
                    ),
                  ],
                ),
            ],
          ),
        ),
      ),
    );
  }

  void _showLanguageDialog(BuildContext context, int index) {
    List<LanguageModel> languagesSelectList = [];
    languagesSelectList.addAll(defaultLanguageList);
    for (var selected in selectedLanguages) {
      languagesSelectList.removeWhere((ele) => ele.code == selected);
    }

    showDialog(
      context: context,
      builder: (BuildContext context) {
        return AlertDialog(
          title: Text(global.language('select_language')),
          content: SizedBox(
            width: 300,
            height: 400,
            child: ListView.builder(
              itemCount: languagesSelectList.length,
              itemBuilder: (context, idx) {
                return ListTile(
                  title: Row(
                    children: [
                      Text(languagesSelectList[idx].name!),
                      const Spacer(),
                      Image.asset(
                        'assets/flags/${languagesSelectList[idx].code}.png',
                        width: 30,
                        height: 30,
                      ),
                    ],
                  ),
                  onTap: () {
                    List<String> newList = List.from(selectedLanguages);
                    newList[index] = languagesSelectList[idx].code!;
                    onLanguagesChanged(newList);
                    Navigator.of(context).pop();
                  },
                );
              },
            ),
          ),
        );
      },
    );
  }
}
