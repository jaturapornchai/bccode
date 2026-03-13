// ignore_for_file: use_build_context_synchronously

import 'package:flutter/material.dart';
import 'package:smlaicloud/components/otp_numeric_keyboard.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/usersystem/otp/otp_screen.dart';
import 'package:smlaicloud/repositories/user_repository.dart';

class TelephoneScreen extends StatefulWidget {
  const TelephoneScreen({super.key});

  @override
  State<TelephoneScreen> createState() => _TelephoneScreenState();
}

class _TelephoneScreenState extends State<TelephoneScreen> with global.ThemeRefreshMixin {
  final UserRepository _userRepository = UserRepository();
  List<int> number = [0, 0, 0, 0, 0, 0, 0, 0, 0, 0];
  int numberIndex = 0;

  void onKeyboardTap(String value) {
    setState(() {
      if (numberIndex < number.length) {
        number[numberIndex] = int.parse(value);
        numberIndex++;
      }
    });
  }

  Widget numberWidget(int position) {
    return Padding(
      padding: const EdgeInsets.only(left: 2, right: 2),
      child: Container(
        height: 50,
        decoration: BoxDecoration(
          border: Border.all(color: Colors.black, width: 0),
          borderRadius: const BorderRadius.all(Radius.circular(8)),
        ),
        child: (position < numberIndex)
            ? Center(
                child: Text(
                  number[position].toString(),
                  style: const TextStyle(
                    color: Colors.black,
                    fontSize: 20,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              )
            : const Center(),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        leading: IconButton(
          icon: const Icon(Icons.arrow_back),
          onPressed: () => Navigator.of(context).pop(),
        ),
        title: const Text('กรุณาบันทึกหมายเลขโทรศัพท์'),
      ), //
      body: SafeArea(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          crossAxisAlignment: CrossAxisAlignment.center,
          children: <Widget>[
            Expanded(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                crossAxisAlignment: CrossAxisAlignment.center,
                children: <Widget>[
                  const Expanded(child: SizedBox()),
                  Column(
                    children: <Widget>[
                      Container(
                        margin: const EdgeInsets.only(top: 50),
                        child: const Text(
                          'หมายเลขโทรศัพท์ เพื่อยืนยันตัวตน',
                          style: TextStyle(
                            color: Colors.black,
                            fontSize: 26,
                            fontWeight: FontWeight.w500,
                          ),
                        ),
                      ),
                      Padding(
                        padding: const EdgeInsets.all(5),
                        child: Row(
                          children: <Widget>[
                            for (int i = 0; i < number.length; i++)
                              Expanded(child: numberWidget(i)),
                          ],
                        ),
                      ),
                    ],
                  ),
                  Container(
                    padding: const EdgeInsets.only(left: 20, right: 20),
                    width: double.infinity,
                    child: ElevatedButton(
                      onPressed: () async {
                        var phoneNum = '66';
                        var phoneNumberInput = '0';
                        int idx = 0;
                        for (var element in number) {
                          if (idx > 0) {
                            phoneNum += element.toString();
                            phoneNumberInput += element.toString();
                          }
                          idx++;
                        }

                        final result = await _userRepository.requestOPT(
                          phoneNum,
                        );
                        // print(result);

                        if (result != null) {
                          Navigator.push(
                            context,
                            MaterialPageRoute(
                              builder: (context) => OtpScreen(
                                ref: result['ref'],
                                phoneNumber: phoneNumberInput,
                                tokenOTP: result['token'],
                              ),
                            ),
                          );
                        }
                      },
                      child: Padding(
                        padding: EdgeInsets.only(top: 10, bottom: 10),
                        child: Text(
                          global.language('confirm'),
                          style: TextStyle(
                            color: Colors.white,
                            fontSize: 24,
                            fontWeight: FontWeight.w500,
                          ),
                        ),
                      ),
                    ),
                  ),
                  OtpNumericKeyboard(
                    onKeyboardTap: onKeyboardTap,
                    rightIcon: const Icon(Icons.backspace),
                    rightButtonFn: () {
                      setState(() {
                        if (numberIndex > 0) {
                          number[numberIndex - 1] = 0;
                          numberIndex--;
                        }
                      });
                    },
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}
