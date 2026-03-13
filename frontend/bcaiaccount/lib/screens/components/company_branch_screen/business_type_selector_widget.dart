import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/business_type_model.dart';
import 'package:smlaicloud/screen_search/business_type_search_screen.dart';

class BusinessTypeSelectorWidget extends StatelessWidget {
  final BusinessTypeModel? businessType;
  final bool businessTypeNull;
  final ValueChanged<BusinessTypeModel> onBusinessTypeSelected;
  final VoidCallback onBusinessTypeCleared;

  const BusinessTypeSelectorWidget({
    super.key,
    required this.businessType,
    required this.businessTypeNull,
    required this.onBusinessTypeSelected,
    required this.onBusinessTypeCleared,
  });

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Padding(
          padding: EdgeInsets.only(left: 10, right: 10),
          child: Row(
            children: [
              Expanded(
                child: SizedBox(
                  height: 50,
                  child: ElevatedButton(
                    style: ButtonStyle(
                      backgroundColor: WidgetStateProperty.all<Color>(
                        global.theme.searchBarColor,
                      ),
                      foregroundColor: WidgetStateProperty.all<Color>(
                        global.theme.textColor,
                      ),
                    ),
                    onPressed: () {
                      Navigator.push(
                        context,
                        MaterialPageRoute(
                          builder: (context) =>
                              const BusinessTypeSearchScreen(word: ''),
                        ),
                      ).then((value) {
                        if (value != null) {
                          BusinessTypeModel result = value;
                          if (result.guidfixed!.isNotEmpty) {
                            onBusinessTypeSelected(result);
                          }
                        }
                      });
                    },
                    child: Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Expanded(
                          child: Text(
                            (businessType?.guidfixed?.isEmpty ?? true)
                                ? "${global.language("business_code")} ~ ${global.language("business_name")} "
                                : "${global.language("product_type_name")} : ${businessType!.code} ~ ${global.packName(businessType!.names!)}  ",
                            overflow: TextOverflow.ellipsis,
                            maxLines: 2,
                          ),
                        ),
                        if (businessType?.guidfixed?.isNotEmpty ?? false)
                          IconButton(
                            onPressed: onBusinessTypeCleared,
                            icon: Icon(Icons.delete),
                          )
                        else
                          Container(),
                        Icon(Icons.search),
                      ],
                    ),
                  ),
                ),
              ),
            ],
          ),
        ),
        if (businessTypeNull)
          Padding(
            padding: EdgeInsets.only(left: 10, right: 10),
            child: SizedBox(
              width: double.infinity,
              child: Text(
                "**${global.language("please_select_business_type")}**",
                style: TextStyle(color: global.theme.negativeHighlightTextColor),
              ),
            ),
          )
        else
          Container(),
      ],
    );
  }
}
