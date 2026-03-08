import 'package:smlaicloud/imports_repositories.dart';
import 'package:smlaicloud/usersystem/registration.dart';
import 'package:flutter/material.dart';
import 'package:smlaicloud/components/otp_numeric_keyboard.dart';
import 'package:smlaicloud/global.dart' as global;

class OtpScreen extends StatefulWidget {
  final String ref;
  final String phoneNumber;
  final String tokenOTP;
  const OtpScreen({
    super.key,
    required this.ref,
    required this.phoneNumber,
    required this.tokenOTP,
  });

  @override
  State<OtpScreen> createState() => _OtpScreenState();
}

class _OtpScreenState extends State<OtpScreen> {
  final UserRepository _userRepository = UserRepository();
  List<int> number = [0, 0, 0, 0, 0, 0];
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
        title: Text(global.language('otp.verify_identity')),
      ), //
      body: SafeArea(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          crossAxisAlignment: CrossAxisAlignment.center,
          children: <Widget>[
            Expanded(
              child: Column(
                children: <Widget>[
                  SizedBox(height: 24),
                  Expanded(
                    child: Column(
                      children: <Widget>[
                        Text(
                          '${global.language('otp.message_sent')} ${widget.phoneNumber} ref: ${widget.ref} ${global.language('otp.verify_identity')}',
                          style: const TextStyle(
                            color: Colors.black,
                            fontSize: 18,
                            fontWeight: FontWeight.w500,
                          ),
                        ),
                        Padding(
                          padding: const EdgeInsets.all(20),
                          child: Row(
                            children: <Widget>[
                              for (int i = 0; i < number.length; i++)
                                Expanded(child: numberWidget(i)),
                            ],
                          ),
                        ),
                      ],
                    ),
                  ),
                  Container(
                    padding: const EdgeInsets.only(left: 20, right: 20),
                    width: double.infinity,
                    child: ElevatedButton(
                      onPressed: () async {
                        var pin = '';

                        for (var element in number) {
                          pin += element.toString();
                        }
                        final result = await _userRepository.verifyOPT(
                          widget.tokenOTP,
                          pin,
                        );
                        // ignore: avoid_print
                        // print(result);
                        if (result) {
                          // ignore: use_build_context_synchronously
                          Navigator.push(
                            context,
                            MaterialPageRoute(
                              builder: (context) => RegistrationScreen(
                                phoneNumber: widget.phoneNumber,
                              ),
                            ),
                          );
                        }
                      },
                      child: Padding(
                        padding: EdgeInsets.only(top: 10, bottom: 10),
                        child: Text(
                          global.language('otp.verify_button'),
                          style: const TextStyle(
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
