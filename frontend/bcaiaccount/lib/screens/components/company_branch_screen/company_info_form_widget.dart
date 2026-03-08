import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/global_model.dart';

class CompanyInfoFormWidget extends StatelessWidget {
  final List<LanguageModel> languageList;
  final List<LanguageDataModel> companyNames;
  final bool isEditMode;
  final bool Function() disableBranch;
  final Function(int languageIndex, String value) onCompanyNameChanged;
  final Function(int focusNodeIndex) findFocusNext;
  final List<global.FieldFocusModel> fieldFocusNodes;
  final int focusNodeMax;

  const CompanyInfoFormWidget({
    super.key,
    required this.languageList,
    required this.companyNames,
    required this.isEditMode,
    required this.disableBranch,
    required this.onCompanyNameChanged,
    required this.findFocusNext,
    required this.fieldFocusNodes,
    required this.focusNodeMax,
  });

  String getLangName(String? code) {
    LanguageModel name = languageList.firstWhere(
      (element) => element.code == (code ?? ''),
      orElse: () =>
          LanguageModel(code: '', codeTranslator: '', name: '', isuse: false),
    );
    return name.name!;
  }

  @override
  Widget build(BuildContext context) {
    List<Widget> widgets = [];
    int currentFocusIndex = focusNodeMax;

    widgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: Text(
          global.language("company_name"),
          style: const TextStyle(fontWeight: FontWeight.bold),
        ),
      ),
    );

    for (
      int languageIndex = 0;
      languageIndex < languageList.length;
      languageIndex++
    ) {
      // Ensure companyNames has enough entries
      while (companyNames.length <= languageIndex) {
        companyNames.add(
          LanguageDataModel(
            code: languageList[companyNames.length].code!,
            name: '',
          ),
        );
      }

      LanguageDataModel companyName = companyNames[languageIndex];

      // Make sure the code matches the language
      if (companyName.code != languageList[languageIndex].code) {
        companyName = companyNames.firstWhere(
          (element) => element.code == languageList[languageIndex].code,
          orElse: () => LanguageDataModel(
            code: languageList[languageIndex].code!,
            name: '',
          ),
        );
      }

      widgets.add(
        Padding(
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
          child: TextField(
            enabled: disableBranch(),
            readOnly: !isEditMode,
            onChanged: (value) => onCompanyNameChanged(languageIndex, value),
            onSubmitted: (value) => findFocusNext(currentFocusIndex),
            focusNode: fieldFocusNodes[++currentFocusIndex].focusNode,
            textAlign: TextAlign.left,
            controller: TextEditingController(text: companyName.name),
            decoration: InputDecoration(
              floatingLabelBehavior: FloatingLabelBehavior.always,
              border: const OutlineInputBorder(),
              labelText:
                  "${global.language("company_name")} (${getLangName(languageList[languageIndex].code)})",
            ),
          ),
        ),
      );
    }

    return Column(children: widgets);
  }
}
