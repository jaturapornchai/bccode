import 'package:flutter/foundation.dart';
import 'dart:io';
import 'dart:async';
// import 'package:smlaicloud/screens/Product/product_add.dart';
import 'package:smlaicloud/bloc/image/image_upload_bloc.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:image_cropper/image_cropper.dart'; // แทน image_crop deprecated
import 'package:image_picker/image_picker.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';
import 'package:smlaicloud/global.dart' as global;

class ProductEditImageScreen extends StatefulWidget {
  final bool useCamera;
  const ProductEditImageScreen(this.useCamera, {super.key});

  @override
  _ProductEditImageState createState() => _ProductEditImageState();
}

class _ProductEditImageState extends State<ProductEditImageScreen> with global.ThemeRefreshMixin {
  File? _file;
  File? _sample;
  late File lastCropped;

  final ImagePicker _picker = ImagePicker();

  @override
  void initState() {
    super.initState();
    _openImage();
  }

  @override
  void dispose() {
    super.dispose();
    /*_file?.delete();
    _sample?.delete();
    _lastCropped?.delete();*/
  }

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      debugShowCheckedModeBanner: false,
      home: BlocListener<ImageUploadBloc, ImageUploadState>(
        listener: (context, state) {
          if (state is ImageUploadSaveSuccess) {
            // // print('Url Response' + state.imageUpload.uri);
            Navigator.of(context, rootNavigator: true).pop(state.imageUpload);

            // Navigator.of(context).pushReplacement(
            //   MaterialPageRoute(
            //     builder: (_) => ProductAdd(url: state.imageUpload.uri),
            //   ),
            // );
          }
        },
        child: SafeArea(
          child: Container(
            color: global.theme.backgroundColor,
            padding: const EdgeInsets.symmetric(
              vertical: 40.0,
              horizontal: 20.0,
            ),
            child: _sample == null
                ? _buildOpeningImage()
                : _buildCroppingImage(),
          ),
        ),
      ),
    );
  }

  Widget _buildOpeningImage() {
    return Center(child: _buildOpenImage());
  }

  Widget _buildOpenImage() {
    return TextButton(onPressed: () {}, child: Container());
  }

  Widget _buildCroppingImage() {
    return Column(
      children: <Widget>[
        Expanded(child: _sample != null ? Image.file(_sample!) : Container()),
        Container(
          padding: const EdgeInsets.only(top: 20.0),
          alignment: AlignmentDirectional.center,
          child: Row(
            mainAxisAlignment: MainAxisAlignment.spaceAround,
            children: <Widget>[
              TextButton(
                child: Text(
                  global.language('save'),
                  style: Theme.of(
                    context,
                  ).textTheme.labelLarge?.copyWith(color: global.theme.onPrimaryColor),
                ),
                onPressed: () => _cropImage(),
              ),
              TextButton(
                child: Text(
                  global.language(
                    widget.useCamera ? 'take_new_photo' : 'select_new_image',
                  ),
                  style: Theme.of(
                    context,
                  ).textTheme.labelLarge?.copyWith(color: global.theme.onPrimaryColor),
                ),
                onPressed: () => _openImage(),
              ),
            ],
          ),
        ),
      ],
    );
  }

  Future<void> _openImage() async {
    if (widget.useCamera) {
      // _file =
      //     (await ImagePicker().pickImage(source: ImageSource.camera)) as File;
      final XFile? image = await _picker.pickImage(source: ImageSource.camera);
      setState(() {
        if (image != null) {
          final File file = File(image.path);
          _file = file;
        } else {
          // print('No image selected.');
        }
      });

      // final File? file = File(image!.path);
    } else {
      // _file =
      //     (await ImagePicker().pickImage(source: ImageSource.gallery)) as File;
      final XFile? image = await _picker.pickImage(source: ImageSource.gallery);
      setState(() {
        if (image != null) {
          final File file = File(image.path);
          _file = file;
        } else {
          // print('No image selected.');
        }
      });
    }
    if (_file != null) {
      _sample = _file; // ใช้ไฟล์ต้นฉบับสำหรับตอนนี้
      setState(() {});
    } else {
      Navigator.pop(context);
    }

    if (widget.useCamera) {
      /*final file = await ImagePicker.pickImage(source: ImageSource.camera);
      final sample = await ImageCrop.sampleImage(
        file: file,
        preferredSize: context.size.longestSide.ceil(),
      );*/

      //_sample?.delete();
      //_file?.delete();

      setState(() {
        /*_sample = sample;
        _file = file;*/
      });
    } else {
      /*final file = await ImagePicker.pickImage(source: ImageSource.gallery);
      final sample = await ImageCrop.sampleImage(
        file: file,
        preferredSize: context.size.longestSide.ceil(),
      );*/

      //_sample?.delete();
      //_file?.delete();

      setState(() {
        /*_sample = sample;
        _file = file;*/
      });
    }
  }

  Future<void> _cropImage() async {
    if (_sample == null) return;

    try {
      CroppedFile? croppedFile = await ImageCropper().cropImage(
        sourcePath: _sample!.path,
        uiSettings: [
          AndroidUiSettings(
            toolbarTitle: global.language('crop_image'),
            toolbarColor: Colors.deepOrange,
            toolbarWidgetColor: global.theme.onPrimaryColor,
            aspectRatioPresets: [
              CropAspectRatioPreset.original,
              CropAspectRatioPreset.square,
            ],
          ),
          IOSUiSettings(
            title: global.language('crop_image'),
            aspectRatioPresets: [
              CropAspectRatioPreset.original,
              CropAspectRatioPreset.square,
            ],
          ),
          WebUiSettings(context: context),
        ],
      );

      if (croppedFile != null) {
        File file = File(croppedFile.path);

        ImageUpload url = ImageUpload(uri: file.path.toString());

        context.read<ImageUploadBloc>().add(ImageUploadSaved(imageUpload: url));
      }
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('Error cropping image: $e');
      }
    }
  }
}

class ImageResponse {
  final String? path;
  final Image? image;

  const ImageResponse(this.path, this.image);
}
