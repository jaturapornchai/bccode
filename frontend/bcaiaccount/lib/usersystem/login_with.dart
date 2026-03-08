import 'package:smlaicloud/bloc/login_bloc/login_bloc.dart';
import 'package:smlaicloud/components/singin_button.dart';
import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:google_sign_in/google_sign_in.dart';
import 'package:smlaicloud/usersystem/otp/telephone_screen.dart';
import 'package:firebase_auth/firebase_auth.dart';

class LoginWithScreen extends StatefulWidget {
  const LoginWithScreen({super.key});

  @override
  State<LoginWithScreen> createState() => _LoginWithScreenState();
}

class _LoginWithScreenState extends State<LoginWithScreen> {
  final FirebaseAuth _auth = FirebaseAuth.instance;
  global.LoginEnum loginType = global.LoginEnum.none;

  Future<UserCredential?> googleSignIn() async {
    try {
      // Use GoogleSignIn.instance instead of constructor
      final GoogleSignIn googleSignIn = GoogleSignIn.instance;

      // Trigger the authentication flow using authenticate() method
      final GoogleSignInAccount googleUser = await googleSignIn.authenticate();

      // Obtain the auth details from the request
      // authentication is synchronous in v7.x, not a Future
      final GoogleSignInAuthentication googleAuth = googleUser.authentication;

      // Create a new credential
      final OAuthCredential credential = GoogleAuthProvider.credential(
        // accessToken: googleAuth.accessToken, // Not available in v7.x
        idToken: googleAuth.idToken,
      );

      // Sign in to Firebase with the Google [UserCredential]
      return await _auth.signInWithCredential(credential);
        } catch (e) {
      // print(e);
      return null;
    }
    return null;
  }

  Future<String?> getCurrentUserIdToken() async {
    User? currentUser = _auth.currentUser;

    if (currentUser != null) {
      String? idToken = await currentUser.getIdToken();
      return idToken;
    } else {
      // No user is signed in.
      return null;
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Center(
        child: Column(
          children: <Widget>[
            Expanded(
              child: Container(
                padding: const EdgeInsets.all(20),
                width: double.infinity,
                child: const FittedBox(
                  fit: BoxFit.fitWidth,
                  child: Text(
                    "BC AiCloud",
                    style: TextStyle(
                      fontWeight: FontWeight.bold,
                      inherit: true,
                      fontSize: 10.0,
                      color: Colors.blue,
                      shadows: [
                        Shadow(
                          // bottomLeft
                          offset: Offset(-1.5, -1.5),
                          color: Colors.white,
                        ),
                        Shadow(
                          // bottomRight
                          offset: Offset(1.5, -1.5),
                          color: Colors.white,
                        ),
                        Shadow(
                          // topRight
                          offset: Offset(1.5, 1.5),
                          color: Colors.white,
                        ),
                        Shadow(
                          // topLeft
                          offset: Offset(-1.5, 1.5),
                          color: Colors.white,
                        ),
                      ],
                    ),
                  ),
                ),
              ),
            ),
            Container(
              margin: const EdgeInsets.only(
                left: 20,
                right: 20,
                bottom: 10,
                top: 10,
              ),
              child: SingInButton(
                labelText: 'Sign in with google',
                press: () {
                  setState(() {
                    loginType = global.LoginEnum.google;
                    googleSignIn().then((value) async {
                      if (value != null) {
                        // เก็บ Google profile data ไว้ใน global
                        final user = value.user;
                        if (user != null) {
                          global.loginName = user.displayName ?? '';
                          global.loginEmail = user.email ?? '';
                          global.loginPhotoUrl = user.photoURL ?? '';
                        }

                        String? userIdToken = await getCurrentUserIdToken();
                        if (userIdToken != null) {
                          context.read<LoginBloc>().add(
                            TokenLogin(token: userIdToken),
                          );
                        }
                      }
                    });
                  });
                },
                img: const AssetImage("assets/img/google_logo.png"),
              ),
            ),
            Container(
              margin: const EdgeInsets.only(
                left: 20,
                right: 20,
                bottom: 10,
                top: 10,
              ),
              child: SingInButton(
                labelText: 'Sing in with Apple ID',
                press: () {},
                img: const AssetImage("assets/img/apple_logo.png"),
              ),
            ),
            Container(
              margin: const EdgeInsets.only(
                left: 20,
                right: 20,
                bottom: 10,
                top: 10,
              ),
              child: SingInButton(
                labelText: 'Sing in with Phone Number',
                press: () {
                  Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (context) => const TelephoneScreen(),
                    ),
                  );
                },
                img: const AssetImage("assets/img/apple_logo.png"),
              ),
            ),
            // Container(
            //   margin:
            //       const EdgeInsets.only(left: 20, right: 20, bottom: 10, top: 10),
            //   child: SingInButton(
            //     labelText: 'Sing in with Facebook',
            //     press: () {},
            //     img: const AssetImage("assets/img/facebook_logo.png"),
            //   ),
            // ),
            // Container(
            //   margin:
            //       const EdgeInsets.only(left: 20, right: 20, bottom: 10, top: 10),
            //   child: SingInButton(
            //     labelText: 'Sing in with Phone Number',
            //     press: () {},
            //     img: const AssetImage("assets/img/apple_logo.png"),
            //   ),
            // ),
            // Container(
            //   margin:
            //       const EdgeInsets.only(left: 20, right: 20, bottom: 10, top: 10),
            //   child: SingInButton(
            //     labelText: 'Sing in with Line',
            //     press: () {
            //       Navigator.push(
            //         context,
            //         MaterialPageRoute(
            //           builder: (context) => const LoginShop(),
            //         ),
            //       );
            //     },
            //     img: const AssetImage("assets/img/line_logo.png"),
            //   ),
            // ),
            // Container(
            //   margin:
            //       const EdgeInsets.only(left: 20, right: 20, bottom: 20, top: 10),
            //   child: SingInButton(
            //     labelText: 'ลงทะเบียนใหม่',
            //     press: () {
            //       Navigator.push(
            //         context,
            //         MaterialPageRoute(
            //           builder: (context) => const RegistrationScreen(),
            //         ),
            //       );
            //     },
            //     img: const AssetImage("assets/img/avatar.png"),
            //   ),
            // ),
          ],
        ),
      ),
    );
  }
}
