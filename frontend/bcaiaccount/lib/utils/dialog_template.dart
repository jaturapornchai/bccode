import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;

class DialogTemplate {
  static Future<List> showDataListTemplateDialog(
      BuildContext context, dynamic jsonData, String menu) async {
    List<dynamic>? selectedData = await showDialog(
      context: context,
      builder: (BuildContext context) {
        List<dynamic> selectedItemTemp = [];
        bool selectItemTempAll = false;
        String headerMenu = "";
        if (menu == 'bank') {
          headerMenu = global.language('select_bank_template');
        } else if (menu == 'transportchannel') {
          headerMenu = global.language('select_transportchannel_template');
        } else if (menu == 'salechannel') {
          headerMenu = global.language('select_salechannel_template');
        } else if (menu == 'unit') {
          headerMenu = global.language('select_unit_template');
        } else if (menu == 'qrcode') {
          headerMenu = global.language('select_qrcode_template');
        }
        return StatefulBuilder(
            builder: (BuildContext context, StateSetter setState) {
          return Dialog(
            child: Container(
              width: 600,
              padding: const EdgeInsets.all(16.0),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Row(
                    children: [
                      Text(
                        headerMenu,
                        style: const TextStyle(
                          fontWeight: FontWeight.bold,
                          fontSize: 18.0,
                        ),
                      ),
                      const Spacer(),
                      Checkbox(
                        value: selectItemTempAll,
                        onChanged: (value) {
                          for (int i = 0; i < jsonData.length; i++) {
                            selectedItemTemp[i] = value!;
                          }
                          setState(() {
                            selectItemTempAll = value!;
                          });
                        },
                      ),
                      Text(global.language('select_all')),
                    ],
                  ),
                  const SizedBox(height: 16.0),
                  (jsonData.isNotEmpty)
                      ? Expanded(
                          child: SingleChildScrollView(
                            child: ListView.builder(
                              shrinkWrap: true,
                              itemCount: jsonData.length,
                              itemBuilder: (context, index) {
                                if (selectedItemTemp.length <= index) {
                                  selectedItemTemp.add(false);
                                }
                                var packName = '';
                                var logo = '';
                                if (menu == 'bank') {
                                  packName =
                                      global.packName(jsonData[index].names);
                                  logo = jsonData[index].logo;
                                } else if (menu == 'transportchannel' ||
                                    menu == 'salechannel') {
                                  packName = jsonData[index].name;
                                  logo = jsonData[index].imageuri;
                                } else if (menu == 'unit') {
                                  packName =
                                      global.packName(jsonData[index].names);
                                  logo = '';
                                } else if (menu == 'qrcode') {
                                  packName =
                                      global.packName(jsonData[index].qrnames);
                                  logo = jsonData[index].logo;
                                }

                                return Card(
                                  shape: RoundedRectangleBorder(
                                    side: BorderSide(
                                      color: Colors.grey.withValues(alpha: 0.2),
                                    ),
                                    borderRadius: BorderRadius.circular(10.0),
                                  ),
                                  child: Padding(
                                    padding: const EdgeInsets.all(8.0),
                                    child: ListTile(
                                      title: Text(packName),
                                      leading: (logo.isNotEmpty)
                                          ? SizedBox(
                                              width: 50,
                                              height: 50,
                                              child: Image.network(logo),
                                            )
                                          : null,
                                      trailing: Checkbox(
                                        value: selectedItemTemp[index],
                                        onChanged: (value) {
                                          setState(() {
                                            selectedItemTemp[index] = value!;
                                          });
                                        },
                                      ),
                                      onTap: () {
                                        setState(() {
                                          selectedItemTemp[index] =
                                              !selectedItemTemp[index];
                                        });
                                      },
                                    ),
                                  ),
                                );
                              },
                            ),
                          ),
                        )
                      : Text(global.language('not_found')),
                  SizedBox(height: 16.0), // Space before actions
                  Row(
                    mainAxisAlignment: MainAxisAlignment.end,
                    children: [
                      ElevatedButton(
                        style: ElevatedButton.styleFrom(
                            backgroundColor: Colors.red),
                        onPressed: () {
                          selectedItemTemp = [];
                          selectItemTempAll = false;
                          Navigator.pop(context);
                        },
                        child: Text(global.language('cancel')),
                      ),
                      SizedBox(width: 16.0), // Space between buttons
                      ElevatedButton(
                        onPressed: () {
                          Navigator.pop(context, selectedItemTemp);
                        },
                        child: Text(global.language('confirm')),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          );
        });
      },
    );
    return selectedData ?? [];
  }
}
