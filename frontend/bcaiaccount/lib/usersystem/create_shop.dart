import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../global.dart' as global;

class CreateShop extends StatefulWidget {
  const CreateShop({super.key});

  @override
  State<CreateShop> createState() => _CreateShopState();
}

class _CreateShopState extends State<CreateShop> {
  final _formKey = GlobalKey<FormState>();
  final TextEditingController _shopname = TextEditingController();
  final TextEditingController _tel = TextEditingController();
  final TextEditingController _user = TextEditingController();
  final TextEditingController _pass = TextEditingController();
  final TextEditingController _confirmPass = TextEditingController();
  bool _isObscurePass = true;
  bool _isObscureConfirmPass = true;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Center(
        child: Form(
          key: _formKey,
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: 30),
            child: SingleChildScrollView(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                mainAxisSize: MainAxisSize.max,
                children: [
                  const FlutterLogo(size: 100),
                  const Text(
                    "DEDE",
                    style: TextStyle(fontWeight: FontWeight.bold, fontSize: 48),
                  ),
                  const Text(
                    "Merchant",
                    style: TextStyle(fontWeight: FontWeight.bold, fontSize: 96),
                  ),
                  Container(
                    margin: const EdgeInsets.only(top: 20.0),
                    child: TextFormField(
                      controller: _shopname,
                      decoration: const InputDecoration(
                        border: OutlineInputBorder(),
                        labelText: 'ชื่อร้าน',
                      ),
                      validator: (value) {
                        if (value == null || value.isEmpty) {
                          return '*กรุณาป้อน ชื่อร้าน';
                        }
                        return null;
                      },
                    ),
                  ),
                  Container(
                    margin: const EdgeInsets.only(top: 20.0),
                    child: TextFormField(
                      controller: _tel,
                      keyboardType: TextInputType.number,
                      inputFormatters: <TextInputFormatter>[
                        FilteringTextInputFormatter.allow(RegExp(r'[0-9]')),
                      ],
                      decoration: const InputDecoration(
                        border: OutlineInputBorder(),
                        labelText: 'หมายเลขโทรศัทพ์',
                      ),
                      validator: (value) {
                        if (value == null || value.isEmpty) {
                          return '*กรุณาป้อน หมายเลขโทรศัทพ์';
                        }
                        return null;
                      },
                    ),
                  ),
                  Container(
                    margin: const EdgeInsets.only(top: 20.0),
                    child: TextFormField(
                      controller: _user,
                      decoration: const InputDecoration(
                        border: OutlineInputBorder(),
                        labelText: 'รหัสผู้จัดการ',
                      ),
                      validator: (value) {
                        if (value == null || value.isEmpty) {
                          return '*กรุณาป้อน รหัสผู้จัดการ';
                        }
                        return null;
                      },
                    ),
                  ),

                  Container(
                    margin: EdgeInsets.only(top: 20.0),
                    child: TextFormField(
                      controller: _pass,
                      obscureText: _isObscurePass,
                      decoration: InputDecoration(
                        border: OutlineInputBorder(),
                        labelText: global.language('password'),
                        suffixIcon: IconButton(
                          icon: Icon(
                            _isObscurePass
                                ? Icons.visibility
                                : Icons.visibility_off,
                          ),
                          onPressed: () {
                            setState(() {
                              _isObscurePass = !_isObscurePass;
                            });
                          },
                        ),
                      ),
                      validator: (value) {
                        if (value == null || value.isEmpty) {
                          return '*กรุณาป้อน รหัสผ่าน';
                        }
                        return null;
                      },
                    ),
                  ),
                  Container(
                    margin: EdgeInsets.only(top: 20.0),
                    child: TextFormField(
                      controller: _confirmPass,
                      obscureText: _isObscureConfirmPass,
                      decoration: InputDecoration(
                        border: OutlineInputBorder(),
                        labelText: global.language('con_password'),
                        suffixIcon: IconButton(
                          icon: Icon(
                            _isObscureConfirmPass
                                ? Icons.visibility
                                : Icons.visibility_off,
                          ),
                          onPressed: () {
                            setState(() {
                              _isObscureConfirmPass = !_isObscureConfirmPass;
                            });
                          },
                        ),
                      ),
                      validator: (value) {
                        if (value == null || value.isEmpty) {
                          return '*กรุณาป้อน ยืนยันรหัสผ่าน';
                        }
                        if (value != _pass.text) {
                          return '*Password ไม่ตรงกัน';
                        }
                        return null;
                      },
                    ),
                  ),
                  Container(
                    margin: const EdgeInsets.only(top: 20.0, bottom: 20.0),
                    width: double.infinity,
                    child: ElevatedButton(
                      style: ElevatedButton.styleFrom(
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(30.0),
                        ),
                        backgroundColor: Colors.redAccent,
                        padding: const EdgeInsets.symmetric(
                          horizontal: 50,
                          vertical: 20,
                        ),
                        textStyle: const TextStyle(
                          fontSize: 30,
                          fontWeight: FontWeight.bold,
                          color: Colors.white,
                        ),
                      ),
                      child: Text(
                        global.language('next'),
                        style: TextStyle(fontSize: 16.0),
                      ),
                      onPressed: () {
                        // Navigator.push(
                        //   context,
                        //   MaterialPageRoute(
                        //     builder: (context) => const OtpScreen(),
                        //   ),
                        // );
                        // check from validate ---
                        // if (_formKey.currentState!.validate()) {
                        //   Navigator.push(
                        //     context,
                        //     MaterialPageRoute(
                        //       builder: (context) => const OtpScreen(),
                        //     ),
                        //   );
                        // }
                      },
                    ),
                  ), //button: login
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}
