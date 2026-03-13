import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/bloc/login_bloc/login_bloc.dart';
import 'package:smlaicloud/model/business_type_model.dart';
import 'package:smlaicloud/model/create_shop_model.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

class CreateSubShopDialog {
  static Future<void> show(BuildContext context) {
    final TextEditingController shopNameController = TextEditingController();
    final formKey = GlobalKey<FormState>();

    // สีตามธีมของแอป
    final Color primaryColor = global.theme.primaryColor;
    final Color primaryDarkColor = global.theme.primaryColor;
    final Color primaryLightColor = global.theme.primaryLightColor;

    return showDialog<void>(
      context: context,
      barrierDismissible: false,
      builder: (BuildContext dialogContext) {
        return BlocListener<LoginBloc, LoginState>(
          listener: (context, state) {
            if (state is CreateShopInProgress) {
              // แสดง loading dialog
              showDialog(
                context: dialogContext,
                barrierDismissible: false,
                builder: (BuildContext loadingContext) {
                  return Dialog(
                    backgroundColor: global.theme.dialogColor,
                    child: Padding(
                      padding: const EdgeInsets.all(20.0),
                      child: Column(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          CircularProgressIndicator(color: primaryColor),
                          const SizedBox(height: 20),
                          Text(
                            global.language("creating_sub_shop"),
                            style: const TextStyle(fontSize: 16),
                          ),
                        ],
                      ),
                    ),
                  );
                },
              );
            } else if (state is CreateShopSuccess) {
              // ปิด loading dialog
              Navigator.of(dialogContext).pop();
              // ปิด create shop dialog
              Navigator.of(dialogContext).pop();

              // แสดงผลสำเร็จ
              global.showSnackBar(
                dialogContext,
                Icon(Icons.check_circle, color: global.theme.onPrimaryColor),
                'สร้างร้านย่อย "${shopNameController.text.trim()}" สำเร็จ',
                global.theme.positiveHighlightTextColor, // successColor
              );

              // TODO: อัพเดทข้อมูลหรือ refresh หน้าตามต้องการ
            } else if (state is CreateShopFailed) {
              // ปิด loading dialog
              Navigator.of(dialogContext).pop();

              // แสดงข้อผิดพลาด
              global.showSnackBar(
                dialogContext,
                Icon(Icons.error, color: global.theme.onPrimaryColor),
                '${global.language("error_occurred")}: ${state.message}',
                global.theme.negativeHighlightTextColor, // errorColor
              );
            }
          },
          child: AlertDialog(
            title: Row(
              children: [
                Icon(Icons.store, color: primaryColor),
                SizedBox(width: 8),
                Text(
                  global.language('create_sub_shop'),
                  style: TextStyle(
                    color: primaryDarkColor,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ],
            ),
            content: Form(
              key: formKey,
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // แสดง Shop ID ปัจจุบัน
                  Container(
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      color: primaryLightColor.withValues(alpha: 0.1),
                      borderRadius: BorderRadius.circular(8),
                      border: Border.all(
                        color: primaryColor.withValues(alpha: 0.3),
                        width: 1,
                      ),
                    ),
                    child: Row(
                      children: [
                        Icon(Icons.info_outline, color: primaryColor, size: 20),
                        const SizedBox(width: 8),
                        Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              global.language("current_shop_id"),
                              style: TextStyle(
                                fontSize: 12,
                                color: primaryDarkColor,
                                fontWeight: FontWeight.w500,
                              ),
                            ),
                            Text(
                              global.getShopId(),
                              style: TextStyle(
                                fontSize: 16,
                                color: primaryDarkColor,
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                          ],
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(height: 16),

                  // ช่องกรอกชื่อร้านย่อย
                  TextFormField(
                    controller: shopNameController,
                    decoration: InputDecoration(
                      labelText: global.language("sub_shop_name"),
                      hintText: global.language('please_enter_sub_shop_name'),
                      prefixIcon:
                          Icon(Icons.store_outlined, color: primaryColor),
                      border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(8),
                      ),
                      focusedBorder: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(8),
                        borderSide: BorderSide(color: primaryColor, width: 2),
                      ),
                    ),
                    validator: (value) {
                      if (value == null || value.trim().isEmpty) {
                        return global.language('please_enter_sub_shop_name');
                      }
                      if (value.trim().length < 2) {
                        return 'ชื่อร้านย่อยต้องมีอย่างน้อย 2 ตัวอักษร';
                      }
                      return null;
                    },
                    maxLength: 100,
                    textInputAction: TextInputAction.done,
                  ),
                ],
              ),
            ),
            actions: [
              TextButton(
                onPressed: () {
                  Navigator.of(dialogContext).pop();
                },
                child: Text(
                  global.language('cancel'),
                  style: TextStyle(color: global.theme.iconSecondaryColor),
                ),
              ),
              ElevatedButton(
                onPressed: () {
                  if (formKey.currentState!.validate()) {
                    // สร้าง CreateShopModel
                    final createShopModel = CreateShopModel(
                      address: [],
                      branchcode: '',
                      images: [],
                      logo: '',
                      name1: '',
                      names: [
                        LanguageDataModel(
                          code: "th",
                          name: shopNameController.text.trim(),
                        ),
                      ],
                      profilepicture: '',
                      settings: Settings(
                        emailowners: [],
                        emailstaffs: [],
                        isusebranch: false,
                        isusedepartment: false,
                        languageconfigs: [
                          LanguageModel(
                            code: "th",
                            codeTranslator: "th",
                            name: "Thai",
                            isuse: false,
                          ),
                        ],
                        latitude: 0,
                        longitude: 0,
                        taxid: '',
                        vatrate: 7,
                        vattypesale: 0,
                        vattypepurchase: 0,
                        inquirytypesale: 0,
                        inquirytypepurchase: 0,
                      ),
                      telephone: '',
                      businesstype: BusinessTypeModel(
                        guidfixed: "",
                        code: "003",
                        isdefault: false,
                        names: [
                          LanguageDataModel(
                            code: "th",
                            name:
                                "ธุรกิจก่อสร้าง วัสดุก่อสร้าง และพัฒนาอสังหาริมทรัพย์",
                          ),
                          LanguageDataModel(
                            code: "en",
                            name:
                                "construction business construction materials and real estate development",
                          ),
                        ],
                      ),
                      mainshopid: global.getShopId(), // ใช้ Shop ID ปัจจุบัน
                    );

                    if (kDebugMode) {
                      AppLogger.debug("CreateShopModel: ${createShopModel.toJson()}");
                    }

                    // เรียกใช้ BLoC event
                    context
                        .read<LoginBloc>()
                        .add(CreateShop(createShop: createShopModel));
                  }
                },
                style: ElevatedButton.styleFrom(
                  backgroundColor: primaryColor,
                  foregroundColor: global.theme.onPrimaryColor,
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(8),
                  ),
                  padding:
                      const EdgeInsets.symmetric(horizontal: 20, vertical: 12),
                ),
                child: Text(
                  global.language('create_sub_shop'),
                  style: TextStyle(fontWeight: FontWeight.bold),
                ),
              ),
            ],
          ),
        );
      },
    );
  }
}
