import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/contact_model.dart';

class BranchInfoFormWidget extends StatelessWidget {
  final List<LanguageModel> languageList;
  final TextEditingController branchCodeController;
  final List<LanguageDataModel> branchNames;
  final ContactModel? contact;
  final bool isEditMode;
  final bool Function() disableBranch;
  final Function(String value) onBranchCodeChanged;
  final Function(int languageIndex, String value) onBranchNameChanged;
  final Function(int languageIndex, String value) onAddressChanged;
  final Function(String value) onLatitudeChanged;
  final Function(String value) onLongitudeChanged;
  final Function(String value) onPhoneChanged;
  final VoidCallback onMapLocationPressed;
  final Function(int focusNodeIndex) findFocusNext;
  final List<global.FieldFocusModel> fieldFocusNodes;
  final int initialFocusNodeIndex;

  const BranchInfoFormWidget({
    super.key,
    required this.languageList,
    required this.branchCodeController,
    required this.branchNames,
    required this.contact,
    required this.isEditMode,
    required this.disableBranch,
    required this.onBranchCodeChanged,
    required this.onBranchNameChanged,
    required this.onAddressChanged,
    required this.onLatitudeChanged,
    required this.onLongitudeChanged,
    required this.onPhoneChanged,
    required this.onMapLocationPressed,
    required this.findFocusNext,
    required this.fieldFocusNodes,
    required this.initialFocusNodeIndex,
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
    int currentFocusIndex = initialFocusNodeIndex;

    widgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: Text(
          global.language("company_branch_data"),
          style: const TextStyle(fontWeight: FontWeight.bold),
        ),
      ),
    );

    // Branch Code
    widgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
        child: TextField(
          enabled: disableBranch(),
          onSubmitted: (value) => findFocusNext(currentFocusIndex),
          focusNode: fieldFocusNodes[++currentFocusIndex].focusNode,
          textAlign: TextAlign.left,
          controller: branchCodeController,
          textCapitalization: TextCapitalization.characters,
          inputFormatters: <TextInputFormatter>[
            FilteringTextInputFormatter.allow(RegExp('[a-z A-Z 0-9 -]')),
          ],
          onChanged: onBranchCodeChanged,
          decoration: InputDecoration(
            floatingLabelBehavior: FloatingLabelBehavior.always,
            border: OutlineInputBorder(),
            labelText: global.language("company_branch_code"),
          ),
        ),
      ),
    );

    // Branch Names
    for (
      int languageIndex = 0;
      languageIndex < languageList.length;
      languageIndex++
    ) {
      // Ensure branchNames has enough entries
      while (branchNames.length <= languageIndex) {
        branchNames.add(
          LanguageDataModel(
            code: languageList[branchNames.length].code!,
            name: '',
          ),
        );
      }

      LanguageDataModel branchName = branchNames[languageIndex];

      // Make sure the code matches the language
      if (branchName.code != languageList[languageIndex].code) {
        branchName = branchNames.firstWhere(
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
            onChanged: (value) => onBranchNameChanged(languageIndex, value),
            onSubmitted: (value) => findFocusNext(currentFocusIndex),
            focusNode: fieldFocusNodes[++currentFocusIndex].focusNode,
            textAlign: TextAlign.left,
            controller: TextEditingController(text: branchName.name),
            decoration: InputDecoration(
              floatingLabelBehavior: FloatingLabelBehavior.always,
              border: const OutlineInputBorder(),
              labelText:
                  "${global.language("company_branch_name")} (${getLangName(languageList[languageIndex].code)})",
            ),
          ),
        ),
      );
    }

    // Addresses
    for (
      int languageIndex = 0;
      languageIndex < languageList.length;
      languageIndex++
    ) {
      // Ensure addresses has enough entries
      while (contact!.address!.length <= languageIndex) {
        contact!.address!.add(
          LanguageDataModel(
            code: languageList[contact!.address!.length].code!,
            name: '',
          ),
        );
      }

      LanguageDataModel address = contact!.address![languageIndex];

      // Make sure the code matches the language
      if (address.code != languageList[languageIndex].code) {
        address = contact!.address!.firstWhere(
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
            onChanged: (value) => onAddressChanged(languageIndex, value),
            onSubmitted: (value) => findFocusNext(currentFocusIndex),
            focusNode: fieldFocusNodes[++currentFocusIndex].focusNode,
            textAlign: TextAlign.left,
            controller: TextEditingController(text: address.name),
            decoration: InputDecoration(
              floatingLabelBehavior: FloatingLabelBehavior.always,
              border: const OutlineInputBorder(),
              labelText:
                  "${global.language("company_branch_address")} (${getLangName(languageList[languageIndex].code)})",
            ),
          ),
        ),
      );
    }

    // Location (Latitude & Longitude)
    widgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.start,
          children: [
            Expanded(
              child: TextField(
                key: ValueKey('lat_${contact?.latitude}'),
                enabled: disableBranch(),
                readOnly: !isEditMode,
                onChanged: onLatitudeChanged,
                onSubmitted: (value) => findFocusNext(currentFocusIndex),
                focusNode: fieldFocusNodes[++currentFocusIndex].focusNode,
                textAlign: TextAlign.left,
                controller: TextEditingController(
                  text: contact?.latitude.toString(),
                ),
                decoration: InputDecoration(
                  floatingLabelBehavior: FloatingLabelBehavior.always,
                  border: OutlineInputBorder(),
                  labelText: global.language("company_latitude"),
                ),
              ),
            ),
            SizedBox(width: 10),
            Expanded(
              child: TextField(
                key: ValueKey('lng_${contact?.longitude}'),
                enabled: disableBranch(),
                readOnly: !isEditMode,
                onChanged: onLongitudeChanged,
                onSubmitted: (value) => findFocusNext(currentFocusIndex),
                focusNode: fieldFocusNodes[++currentFocusIndex].focusNode,
                textAlign: TextAlign.left,
                controller: TextEditingController(
                  text: contact?.longitude.toString(),
                ),
                decoration: InputDecoration(
                  floatingLabelBehavior: FloatingLabelBehavior.always,
                  border: OutlineInputBorder(),
                  labelText: global.language("company_longitude"),
                ),
              ),
            ),
            SizedBox(
              child: IconButton(
                focusNode: FocusNode(skipTraversal: true),
                icon: Icon(
                  Icons.location_on,
                  color: Colors.blue.shade700,
                ),
                tooltip: global.language("select_map_location"),
                onPressed: onMapLocationPressed,
              ),
            ),
          ],
        ),
      ),
    );

    // Phone Number
    widgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
        child: TextField(
          enabled: disableBranch(),
          readOnly: !isEditMode,
          onChanged: onPhoneChanged,
          onSubmitted: (value) => findFocusNext(currentFocusIndex),
          focusNode: fieldFocusNodes[++currentFocusIndex].focusNode,
          textAlign: TextAlign.left,
          controller: TextEditingController(text: contact?.phonenumber),
          decoration: InputDecoration(
            floatingLabelBehavior: FloatingLabelBehavior.always,
            border: OutlineInputBorder(),
            labelText: global.language("company_branch_phone"),
          ),
        ),
      ),
    );

    return Column(children: widgets);
  }
}
