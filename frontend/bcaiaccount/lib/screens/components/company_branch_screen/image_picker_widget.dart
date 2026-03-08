import 'dart:io';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';

class ImagePickerWidget extends StatelessWidget {
  final String title;
  final List<File> imageFiles;
  final List<Uint8List> imageWeb;
  final String? imageUri;
  final bool enabled;
  final VoidCallback onUpload;
  final ImagePicker picker;

  const ImagePickerWidget({
    super.key,
    required this.title,
    required this.imageFiles,
    required this.imageWeb,
    this.imageUri,
    required this.enabled,
    required this.onUpload,
    required this.picker,
  });

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        const SizedBox(height: 10),
        Center(
          child: Text(
            title,
            style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
          ),
        ),
        GridView.builder(
          primary: true,
          shrinkWrap: true,
          physics: const NeverScrollableScrollPhysics(),
          itemCount: imageFiles.length,
          gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
            crossAxisCount: 1,
            childAspectRatio: 1.5,
          ),
          itemBuilder: (BuildContext context, int index) {
            return Column(
              children: [
                if (enabled)
                  Row(
                    children: [
                      Expanded(
                        child: IconButton(
                          focusNode: FocusNode(skipTraversal: true),
                          onPressed: () {
                            imageFiles[index] = File('');
                            imageWeb[index] = Uint8List(0);
                          },
                          icon: const Icon(Icons.delete),
                        ),
                      ),
                      const SizedBox(width: 5),
                      Expanded(
                        child: IconButton(
                          focusNode: FocusNode(skipTraversal: true),
                          onPressed: () async {
                            final XFile? image = await picker.pickImage(
                              source: ImageSource.gallery,
                              maxHeight: 400,
                              maxWidth: 400,
                            );
                            if (image != null) {
                              var f = await image.readAsBytes();
                              imageWeb[index] = f;
                              imageFiles[index] = File(image.path);
                              onUpload();
                            }
                          },
                          icon: const Icon(Icons.folder),
                        ),
                      ),
                      const SizedBox(width: 5),
                      if (kIsWeb == false)
                        Expanded(
                          child: IconButton(
                            focusNode: FocusNode(skipTraversal: true),
                            onPressed: () async {
                              final XFile? photo = await picker.pickImage(
                                source: ImageSource.camera,
                                maxHeight: 400,
                                maxWidth: 400,
                                imageQuality: 60,
                              );
                              if (photo != null) {
                                var f = await photo.readAsBytes();
                                imageWeb[index] = f;
                                imageFiles[index] = File(photo.path);
                                onUpload();
                              }
                            },
                            icon: const Icon(Icons.camera_alt),
                          ),
                        ),
                    ],
                  ),
                Expanded(
                  child: Padding(
                    padding: const EdgeInsets.all(10.0),
                    child: Container(
                      decoration: BoxDecoration(
                        color: Colors.white,
                        border: Border.all(color: Colors.black),
                        borderRadius: BorderRadius.circular(5),
                        image: (imageFiles[index].path != '')
                            ? DecorationImage(
                                image: MemoryImage(imageWeb[index]),
                                fit: BoxFit.fill,
                              )
                            : (imageUri != null && imageUri!.isNotEmpty)
                            ? DecorationImage(
                                image: NetworkImage(imageUri!),
                                fit: BoxFit.fill,
                              )
                            : const DecorationImage(
                                image: AssetImage('assets/img/noimage.png'),
                              ),
                      ),
                    ),
                  ),
                ),
              ],
            );
          },
        ),
      ],
    );
  }
}
