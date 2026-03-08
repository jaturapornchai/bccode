import 'dart:io';
import 'package:smlaicloud/bloc/image/image_upload_bloc.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:image_picker/image_picker.dart';
import 'package:smlaicloud/global.dart' as global;

class ImportProductImageScreen extends StatefulWidget {
  const ImportProductImageScreen({super.key});

  @override
  ImportProductImageScreenState createState() =>
      ImportProductImageScreenState();
}

class ImportProductImageScreenState extends State<ImportProductImageScreen> {
  final ImagePicker imagePicker = ImagePicker();
  List<ImportImageModel> imageUpload = [];
  String selectedStatus = '';

  @override
  void initState() {
    super.initState();
  }

  Future<void> pickImages() async {
    final List<XFile> images = await imagePicker.pickMultiImage(
      maxHeight: 400,
      maxWidth: 400,
    );

    if (images.isNotEmpty) {
      if (kIsWeb) {
        // Code for web platform
        for (var imageFile in images) {
          Uint8List bytes = await imageFile.readAsBytes();

          String modifiedName = imageFile.name;
          if (modifiedName.startsWith('scaled_')) {
            modifiedName = modifiedName.replaceFirst('scaled_', '');
          }

          imageUpload.add(
            ImportImageModel(
              imageFile: File(''), // File path is not available on web
              imageWeb: bytes,
              imageName: modifiedName,
              status: 'wait',
            ),
          );
        }
      } else if (Platform.isMacOS) {
        // Code for macOS platform
        // Filter out only .jpg and .png files
        List<XFile> filteredImages = images.where((file) {
          String extension = file.path.split('.').last.toLowerCase();
          return extension == 'jpg' || extension == 'png';
        }).toList();

        for (var imageFile in filteredImages) {
          Uint8List bytes = await imageFile.readAsBytes();
          String modifiedName = imageFile.name.replaceFirst('scaled_', '');

          imageUpload.add(
            ImportImageModel(
              imageFile: File(imageFile.path),
              imageWeb: bytes,
              imageName: modifiedName,
              status: 'wait',
            ),
          );
        }
      }

      setState(() {});
    }
  }

  /// uploadImage
  void uploadImage() {
    int lastIndex = 0;

    for (var i = 0; i < imageUpload.length; i++) {
      if (imageUpload[i].status == 'wait') {
        lastIndex = i;
        break;
      }
    }

    context.read<ImageUploadBloc>().add(
      ImportImageProduct(
        imageName: imageUpload[lastIndex].imageName,
        imageFiles: imageUpload[lastIndex].imageFile,
        imageWeb: imageUpload[lastIndex].imageWeb,
        index: lastIndex,
      ),
    );
  }

  // Helper method to build status indicators
  Widget _buildStatusIndicator(
    String label,
    IconData icon,
    Color color,
    String status,
  ) {
    return MouseRegion(
      cursor: SystemMouseCursors.click,
      child: GestureDetector(
        onTap: () {
          setState(() {
            selectedStatus = status;
          });
        },
        child: Container(
          padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
          margin: const EdgeInsets.symmetric(horizontal: 5),
          decoration: BoxDecoration(
            color: color.withOpacity(0.2),
            borderRadius: BorderRadius.circular(10),
          ),
          child: Row(
            children: [
              Icon(icon, color: color),
              const SizedBox(width: 4),
              Text(
                '$label: ${imageUpload.where((item) => item.status == status).length}',
                style: TextStyle(fontSize: 18, color: color),
              ),
            ],
          ),
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return BlocListener<ImageUploadBloc, ImageUploadState>(
      listener: (context, state) {
        if (state is ImportImageProductInProgress) {
          setState(() {
            imageUpload[state.index].status = 'uploading';
          });
        } else if (state is ImportImageProductSuccess) {
          setState(() {
            imageUpload[state.index].status = 'success';
            if (state.index < imageUpload.length - 1) {
              imageUpload[state.index + 1].status = 'uploading';
              context.read<ImageUploadBloc>().add(
                ImportImageProduct(
                  imageName: imageUpload[state.index + 1].imageName,
                  imageFiles: imageUpload[state.index + 1].imageFile,
                  imageWeb: imageUpload[state.index + 1].imageWeb,
                  index: state.index + 1,
                ),
              );
            }
          });
        } else if (state is ImportImageProductFailure) {
          setState(() {
            imageUpload[state.index].status = 'error';
            if (state.index < imageUpload.length - 1) {
              imageUpload[state.index + 1].status = 'uploading';
              context.read<ImageUploadBloc>().add(
                ImportImageProduct(
                  imageName: imageUpload[state.index + 1].imageName,
                  imageFiles: imageUpload[state.index + 1].imageFile,
                  imageWeb: imageUpload[state.index + 1].imageWeb,
                  index: state.index + 1,
                ),
              );
            }
          });
        }
      },
      child: Scaffold(
        appBar: AppBar(
          backgroundColor: global.theme.appBarColor,
          title: Text(global.language("import_product_image")),
          actions: [
            Padding(
              padding: EdgeInsets.only(right: 10.0),
              child: Row(
                children: [
                  IconButton(
                    tooltip: global.language("add_image"),
                    onPressed: () {
                      pickImages();
                    },
                    icon: const Icon(Icons.add),
                  ),
                  IconButton(
                    tooltip: global.language("upload_image"),
                    onPressed: () {
                      /// funtion upload image
                      uploadImage();
                    },
                    icon: const Icon(Icons.upload),
                  ),
                ],
              ),
            ),
          ],
        ),
        body: Center(
          child: (imageUpload.isNotEmpty)
              ? Padding(
                  padding: const EdgeInsets.all(8.0),
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: <Widget>[
                      // Section for displaying totals
                      Padding(
                        padding: const EdgeInsets.only(bottom: 10.0),
                        child: SingleChildScrollView(
                          scrollDirection: Axis.horizontal,
                          child: Row(
                            mainAxisAlignment: MainAxisAlignment.spaceEvenly,
                            children: [
                              // Total Images
                              MouseRegion(
                                cursor: SystemMouseCursors.click,
                                child: GestureDetector(
                                  onTap: () {
                                    setState(() {
                                      selectedStatus = '';
                                    });
                                  },
                                  child: Container(
                                    padding: const EdgeInsets.symmetric(
                                      horizontal: 10,
                                      vertical: 5,
                                    ),
                                    margin: const EdgeInsets.symmetric(
                                      horizontal: 5,
                                    ),
                                    decoration: BoxDecoration(
                                      color: Colors.blue[100],
                                      borderRadius: BorderRadius.circular(10),
                                    ),
                                    child: Text(
                                      'Total: ${imageUpload.length}',
                                      style: const TextStyle(
                                        fontSize: 18,
                                        fontWeight: FontWeight.bold,
                                      ),
                                    ),
                                  ),
                                ),
                              ),

                              // Each Status with Icon
                              _buildStatusIndicator(
                                'Waiting',
                                Icons.hourglass_empty,
                                Colors.grey,
                                'wait',
                              ),
                              _buildStatusIndicator(
                                'Uploading',
                                Icons.cloud_upload,
                                Colors.blue,
                                'uploading',
                              ),
                              _buildStatusIndicator(
                                'Success',
                                Icons.check_circle,
                                Colors.green,
                                'success',
                              ),
                              _buildStatusIndicator(
                                'Error',
                                Icons.error,
                                Colors.red,
                                'error',
                              ),
                            ],
                          ),
                        ),
                      ),

                      Expanded(
                        child: GridView.builder(
                          gridDelegate:
                              const SliverGridDelegateWithFixedCrossAxisCount(
                                crossAxisCount:
                                    4, // Number of columns in the grid
                                crossAxisSpacing:
                                    4.0, // Horizontal space between items
                                mainAxisSpacing:
                                    4.0, // Vertical space between items
                              ),
                          itemCount: selectedStatus.isEmpty
                              ? imageUpload.length
                              : imageUpload
                                    .where(
                                      (item) => item.status == selectedStatus,
                                    )
                                    .length,
                          itemBuilder: (context, index) {
                            final item = selectedStatus.isEmpty
                                ? imageUpload[index]
                                : imageUpload
                                      .where(
                                        (item) => item.status == selectedStatus,
                                      )
                                      .toList()[index];
                            return Container(
                              decoration: BoxDecoration(
                                borderRadius: BorderRadius.circular(10),
                                color: Colors.grey[300],
                                border: Border.all(
                                  color:
                                      item.status == 'wait' ||
                                          item.status == 'uploading'
                                      ? Colors
                                            .grey // Blue border for uploading status
                                      : item.status == 'success'
                                      ? Colors
                                            .blue // Green border for uploaded status
                                      : Colors
                                            .red, // Red border for other statuses (e.g., error)
                                  width: 3,
                                ),
                              ),
                              child: Padding(
                                padding: const EdgeInsets.all(8.0),
                                child: Column(
                                  children: [
                                    Row(
                                      children: <Widget>[
                                        // Conditional Icon
                                        Icon(
                                          item.status == 'error'
                                              ? Icons
                                                    .error // Icon for error
                                              : item.status == 'success'
                                              ? Icons
                                                    .check_circle // Icon for success
                                              : item.status == 'wait'
                                              ? Icons
                                                    .upload // Icon for wait
                                              : item.status == 'uploading'
                                              ? null // Icon for process
                                              : Icons
                                                    .info, // Default icon for other statuses
                                          color: item.status == 'error'
                                              ? Colors.red
                                              : item.status == 'success'
                                              ? Colors.green
                                              : Colors.black,
                                        ),
                                        const SizedBox(width: 5),
                                        // Your Text widget
                                        Expanded(
                                          child: Text(
                                            item.imageName,
                                            style: TextStyle(
                                              fontSize: 18,
                                              fontWeight: FontWeight.bold,
                                              color: (item.status == 'error')
                                                  ? Colors.red
                                                  : (item.status == 'success')
                                                  ? Colors.green
                                                  : Colors.black,
                                            ),
                                          ),
                                        ),

                                        IconButton(
                                          icon: const Icon(
                                            Icons.delete,
                                            color: Colors.red,
                                          ),
                                          onPressed: () {
                                            setState(() {
                                              imageUpload.removeAt(
                                                index,
                                              ); // Remove the image from the list
                                            });
                                          },
                                        ),
                                      ],
                                    ),
                                    (item.status == 'error')
                                        ? Text(
                                            "Error: ${global.language("file_name_barcode_not_found")}",
                                            style: const TextStyle(
                                              color: Colors.red,
                                            ),
                                          )
                                        : Container(),

                                    Text(item.status),

                                    // Image or progress indicator
                                    Expanded(
                                      child: item.status == 'uploading'
                                          ? const Center(
                                              child:
                                                  CircularProgressIndicator(),
                                            )
                                          : kIsWeb
                                          ? Image.memory(
                                              item.imageWeb,
                                            ) // Display image from memory for web
                                          : Image.file(
                                              item.imageFile,
                                            ), // Display image from file for other platforms
                                    ),
                                  ],
                                ),
                              ),
                            );
                          },
                        ),
                      ),
                    ],
                  ),
                )
              : Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    IconButton(
                      iconSize: 150,
                      color: Colors.grey,
                      icon: const Icon(Icons.add_photo_alternate),
                      onPressed: () {
                        pickImages();
                      },
                    ),
                    Text(
                      global.language('import_product_image_file_only_jpg_png'),
                      style: const TextStyle(fontSize: 20, color: Colors.grey),
                    ),
                  ],
                ),
        ),
      ),
    );
  }
}

class ImportImageModel {
  File imageFile;
  Uint8List imageWeb;
  String imageName;
  String status;

  ImportImageModel({
    required this.imageFile,
    required this.imageWeb,
    required this.imageName,
    required this.status,
  });
}
