import 'package:smlaicloud/bloc/login_bloc/login_bloc.dart';
import 'package:smlaicloud/usersystem/login_shop_screen.dart';
import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_html/flutter_html.dart';

class RegistrationScreen extends StatefulWidget {
  final String phoneNumber;
  const RegistrationScreen({super.key, required this.phoneNumber});

  @override
  State<RegistrationScreen> createState() => _RegistrationScreenState();
}

class _RegistrationScreenState extends State<RegistrationScreen> with global.ThemeRefreshMixin {
  // final GoogleSignIn _googleSignIn = GoogleSignIn();
  // final FirebaseAuth _auth = FirebaseAuth.instance;
  final TextEditingController _userControl = TextEditingController();
  final TextEditingController _passControl = TextEditingController();
  final TextEditingController _confirmpassControl = TextEditingController();
  global.LoginEnum loginType = global.LoginEnum.none;
  String messages =
      """<b>เจ้าของกิจการหมายถึง</b> ผู้ที่สามารถสร้างกิจการ โดยสามารถสร้างบริษัทภายใต้กิจการได้หลายบริษัท และหลายสาขา พร้อมทั้งมีสิทธิ์ในการจัดการกิจการได้ 
          เช่น เพิ่มผู้ดูแลกิจการเสริม เพิ่มผู้ดูแลบริษัท เพิ่มพนักงาน และเพิ่มสาขา<br/><br/>
          <b>นโยบายด้านความปลอดภัย</b> ระบบ Thai 7 ทั้งหมด จะเข้าใช้โดยผ่าน ระบบ Login ของ Google หรือ Facebook หรือ Apple ID และระบบจะไม่เก็บรหัสผ่านใดๆ ไว้ในระบบ เพื่อความปลอดภัยสูงสุด เพราะฉะนั้น
          ก่อนใช้งาน จะต้องมีบัญชีของ Google หรือ Facebook หรือ Apple ID<br/><br/>
          <b>พระราชบัญญัติคุ้มครองข้อมูลส่วนบุคคล</b> ลูกค้าจะต้องยินยอมให้ระบบเก็บข้อมูลส่วนตัวของลูกค้า และข้อมูลทั้งหมด ไว้ในระบบ<br/><br/>
          <b>ขั้นตอนการลงทะเบียน</b> มี 3 ขั้นตอนคือ<br/>
          <ul>
            <li>Login ด้วย Google,Facebook,Apple ID อย่างใดอย่างหนึ่ง</li>
            <li>ยินยอมให้ระบบเก็บข้อมูลส่วนตัวของลูกค้า และข้อมูลทั้งหมด ไว้ในระบบ Thai 7 Cloud</li>
            <li>บันทึกหมายเลขโทรศัพท์เพื่อรับ OTP จากเรา และทำการยืนยันตัวตน ด้วย OTP จึงจะสามารถเข้าใช้งานได้</li>
          </ul>""";

  // Future googleSignIn() async {
  //   final GoogleSignInAccount? googleUser = await _googleSignIn.signIn();
  //   final GoogleSignInAuthentication googleAuth =
  //       await googleUser!.authentication;
  //   final credential = GoogleAuthProvider.credential(
  //     accessToken: googleAuth.accessToken,
  //     idToken: googleAuth.idToken,
  //   );
  //   await FirebaseAuth.instance.signInWithCredential(credential);
  //   global.loginName = googleUser.displayName!;
  //   global.loginEmail = googleUser.email;
  //   global.loginPhotoUrl = googleUser.photoUrl!;
  //   return googleUser.email;
  // }

  // Future<UserCredential?> googleSignIn() async {
  //   try {
  //     // Trigger the authentication flow
  //     final GoogleSignInAccount? googleUser = await GoogleSignIn(
  //             scopes: ['https://www.googleapis.com/auth/contacts.readonly'])
  //         .signIn();

  //     // Obtain the auth details from the request
  //     if (googleUser != null) {
  //       final GoogleSignInAuthentication googleAuth =
  //           await googleUser.authentication;

  //       // Create a new credential
  //       final OAuthCredential credential = GoogleAuthProvider.credential(
  //         accessToken: googleAuth.accessToken,
  //         idToken: googleAuth.idToken,
  //       );

  //       // Sign in to Firebase with the Google [UserCredential]

  //       return await _auth.signInWithCredential(credential);
  //     }
  //   } catch (e) {
  //     // print(e);
  //     return null;
  //   }
  // }

  // Future<String?> getCurrentUserIdToken() async {
  //   User? currentUser = _auth.currentUser;

  //   if (currentUser != null) {
  //     String? idToken = await currentUser.getIdToken();
  //     return idToken;
  //   } else {
  //     // No user is signed in.
  //     return null;
  //   }
  // }

  @override
  void initState() {
    _userControl.text = widget.phoneNumber;
    super.initState();
  }

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      home: BlocListener<LoginBloc, LoginState>(
        listener: (context, state) {
          if (state is RegisterSuccess) {
            Future.delayed(const Duration(milliseconds: 500), () {
              Navigator.of(context).pushAndRemoveUntil(
                MaterialPageRoute(builder: (_) => const LoginShopScreen()),
                (route) => false,
              );
            });

            setState(() {
              global.showSnackBar(
                context,
                Icon(Icons.save, color: Colors.white),
                global.language("success_save"),
                Colors.green,
              );
            });
          } else if (state is RegisterFailed) {
            setState(() {
              global.showSnackBar(
                context,
                Icon(Icons.save, color: Colors.white),
                "${global.language("not_success_save")} : ${state.message}",
                Colors.red,
              );
            });
          }
        },
        child: Scaffold(
          appBar: AppBar(
            leading: IconButton(
              icon: const Icon(Icons.arrow_back),
              onPressed: () => Navigator.of(context).pop(),
            ),
            title: const Text('ลงทะเบียนใหม่'),
          ), //
          body: SingleChildScrollView(
            child: Center(
              child: Column(
                children: <Widget>[
                  const SizedBox(height: 20),
                  const Text(
                    "ลงทะเบียนใหม่ เพื่อเป็นเจ้าของกิจการ",
                    style: TextStyle(fontWeight: FontWeight.bold, fontSize: 24),
                  ),
                  const SizedBox(height: 20),
                  Html(data: messages),
                  const SizedBox(height: 20),
                  Column(
                    children: [
                      usernameField(),
                      const SizedBox(height: 10),
                      passwordField(),
                      const SizedBox(height: 10),
                      confirmpasswordField(),
                      SizedBox(height: 10),
                      Container(
                        padding: const EdgeInsets.only(left: 20, right: 20),
                        width: double.infinity,
                        child: ElevatedButton(
                          onPressed: () async {
                            if (_passControl.text == _confirmpassControl.text) {
                              context.read<LoginBloc>().add(
                                RegisterUser(
                                  userName: _userControl.text,
                                  passWord: _passControl.text,
                                  timezonelabel: "",
                                  timezoneoffset: "",
                                  yeartype: "",
                                ),
                              );
                            } else {
                              setState(() {
                                global.showSnackBar(
                                  context,
                                  Icon(Icons.save, color: Colors.white),
                                  "${global.language("not_success_save")} : รหัสผ่านและรหัสยืนยันไม่ตรงกัน",
                                  Colors.red,
                                );
                              });
                            }
                          },
                          child: const Padding(
                            padding: EdgeInsets.only(top: 10, bottom: 10),
                            child: Text(
                              'ลงทะเบียน',
                              style: TextStyle(
                                color: Colors.white,
                                fontSize: 24,
                                fontWeight: FontWeight.w500,
                              ),
                            ),
                          ),
                        ),
                      ),
                      const SizedBox(height: 100),
                    ],
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }

  Widget usernameField() {
    return Container(
      height: 60,
      alignment: Alignment.centerLeft,
      decoration: BoxDecoration(
        color: global.theme.cardColor,
        borderRadius: BorderRadius.circular(10),
      ),
      child: TextFormField(
        readOnly: true,
        controller: _userControl,
        style: const TextStyle(color: Colors.black),
        decoration: const InputDecoration(
          border: InputBorder.none,
          contentPadding: EdgeInsets.only(top: 14),
          prefixIcon: Icon(Icons.person, color: Color(0xff4c5166)),
          hintText: 'เบอร์โทรศัพท์',
          hintStyle: TextStyle(color: Colors.black38),
        ),
      ),
    );
  }

  Widget passwordField() {
    return Container(
      height: 60,
      alignment: Alignment.centerLeft,
      decoration: BoxDecoration(
        color: global.theme.cardColor,
        borderRadius: BorderRadius.circular(10),
      ),
      child: TextFormField(
        controller: _passControl,
        obscureText: true,
        style: const TextStyle(color: Colors.black),
        decoration: const InputDecoration(
          border: InputBorder.none,
          contentPadding: EdgeInsets.only(top: 14),
          prefixIcon: Icon(Icons.security, color: Color(0xff4c5166)),
          hintText: 'รหัสผ่าน',
          hintStyle: TextStyle(color: Colors.black38),
        ),
      ),
    );
  }

  Widget confirmpasswordField() {
    return Container(
      height: 60,
      alignment: Alignment.centerLeft,
      decoration: BoxDecoration(
        color: global.theme.cardColor,
        borderRadius: BorderRadius.circular(10),
      ),
      child: TextFormField(
        controller: _confirmpassControl,
        obscureText: true,
        style: const TextStyle(color: Colors.black),
        decoration: const InputDecoration(
          border: InputBorder.none,
          contentPadding: EdgeInsets.only(top: 14),
          prefixIcon: Icon(Icons.security, color: Color(0xff4c5166)),
          hintText: 'ยืนยันรหัสผ่าน',
          hintStyle: TextStyle(color: Colors.black38),
        ),
      ),
    );
  }
}
