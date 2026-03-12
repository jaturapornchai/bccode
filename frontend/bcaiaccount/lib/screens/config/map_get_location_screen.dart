import 'package:flutter/material.dart';
import 'package:flutter_map/flutter_map.dart';
import 'package:latlong2/latlong.dart';
import 'package:geolocator/geolocator.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/utils/logger/app_logger.dart';

class MapGetLocationScreen extends StatefulWidget {
  const MapGetLocationScreen({
    super.key,
    required this.latitude,
    required this.longitude,
  });
  final double latitude;
  final double longitude;

  @override
  State<MapGetLocationScreen> createState() => _MapGetLocationScreenState();
}

class _MapGetLocationScreenState extends State<MapGetLocationScreen>
    with global.ThemeRefreshMixin {
  double latitude = 0.0;
  double longitude = 0.0;
  double zoommap = 0.0;
  bool _isLocating = false;

  LatLng defaultLocation = const LatLng(0.0, 0.0);

  List<Marker> customMarkers = [];

  final MapController _mapController = MapController();

  @override
  void initState() {
    super.initState();

    if (widget.latitude == 0 && widget.longitude == 0) {
      latitude = 13.827700395475112;
      longitude = 100.525890413137;
      zoommap = 7;
      customMarkers = [];
    } else {
      latitude = widget.latitude;
      longitude = widget.longitude;
      zoommap = 15;
      defaultLocation = LatLng(latitude, longitude);
      customMarkers = [buildPin(defaultLocation)];
    }
  }

  Marker buildPin(LatLng point) => Marker(
    point: point,
    child: Icon(Icons.location_pin, size: 60, color: global.theme.negativeHighlightTextColor),
    width: 60,
    height: 60,
  );

  /// อัพเดทตำแหน่งบนแผนที่
  void _updateLocation(double newLat, double newLng) {
    setState(() {
      latitude = newLat;
      longitude = newLng;
      defaultLocation = LatLng(newLat, newLng);
      if (customMarkers.isNotEmpty) {
        customMarkers[0] = buildPin(defaultLocation);
      } else {
        customMarkers.add(buildPin(defaultLocation));
      }
    });
    _mapController.move(defaultLocation, 15);
  }

  /// ดึงตำแหน่งปัจจุบันผ่าน geolocator (รองรับ Android, iOS, macOS, Windows, Web)
  Future<void> getCurrentLatLng() async {
    if (_isLocating) return;

    setState(() => _isLocating = true);

    try {
      // ตรวจสอบว่า location service เปิดอยู่หรือไม่
      bool serviceEnabled = await Geolocator.isLocationServiceEnabled();
      if (!serviceEnabled) {
        if (mounted) {
          global.showWarningSnackBar(context, global.language('please_enable_location_service'));
        }
        return;
      }

      // ตรวจสอบ permission
      LocationPermission permission = await Geolocator.checkPermission();
      if (permission == LocationPermission.denied) {
        permission = await Geolocator.requestPermission();
        if (permission == LocationPermission.denied) {
          if (mounted) {
            global.showWarningSnackBar(context, global.language('location_permission_denied'));
          }
          return;
        }
      }

      if (permission == LocationPermission.deniedForever) {
        if (mounted) {
          global.showErrorSnackBar(context, global.language('location_permission_denied_forever'));
        }
        return;
      }

      // ดึงตำแหน่งปัจจุบัน
      final position = await Geolocator.getCurrentPosition(
        locationSettings: const LocationSettings(
          accuracy: LocationAccuracy.high,
          timeLimit: Duration(seconds: 15),
        ),
      );

      _updateLocation(position.latitude, position.longitude);
      AppLogger.info('[Map] ตำแหน่งปัจจุบัน: ${position.latitude}, ${position.longitude}');
    } catch (e) {
      AppLogger.warning('[Map] ไม่สามารถหาตำแหน่งปัจจุบันได้: $e');
      if (mounted) {
        global.showErrorSnackBar(context, '${global.language("cannot_get_current_location")}: $e');
      }
    } finally {
      if (mounted) setState(() => _isLocating = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        title: Text(global.language('select_map_location')),
        actions: [
          IconButton(
            icon: Icon(Icons.save),
            onPressed: () {
              if (defaultLocation.latitude == 0 && defaultLocation.longitude == 0) {
                global.showWarningSnackBar(context, global.language('please_select_location'));
                return;
              }
              Navigator.pop(context, defaultLocation);
            },
          ),
        ],
      ),
      body: Column(
        children: [
          // แสดงตำแหน่งที่เลือกไว้
          if (defaultLocation.latitude != 0 || defaultLocation.longitude != 0)
            Container(
              padding: EdgeInsets.all(8),
              color: global.theme.cardColor,
              child: Row(
                children: [
                  Icon(Icons.location_on, color: global.theme.negativeHighlightTextColor, size: 20),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      'Lat: ${defaultLocation.latitude.toStringAsFixed(6)}, Lng: ${defaultLocation.longitude.toStringAsFixed(6)}',
                      style: TextStyle(fontSize: 12),
                    ),
                  ),
                ],
              ),
            ),
          Expanded(
            child: FlutterMap(
              mapController: _mapController,
              options: MapOptions(
                initialCenter: LatLng(latitude, longitude),
                initialZoom: zoommap,
                onTap: (_, point) {
                  setState(() {
                    if (customMarkers.isNotEmpty) {
                      customMarkers[0] = buildPin(point);
                    } else {
                      customMarkers.add(buildPin(point));
                    }
                    defaultLocation = point;
                  });
                },
                interactionOptions: const InteractionOptions(
                  flags: ~InteractiveFlag.doubleTapZoom,
                ),
              ),
              children: [
                TileLayer(
                  urlTemplate: 'https://tile.openstreetmap.org/{z}/{x}/{y}.png',
                  userAgentPackageName: 'com.bcaicloud.bcaiaccount',
                ),
                MarkerLayer(
                  markers: customMarkers,
                ),
              ],
            ),
          ),
        ],
      ),
      // ปุ่ม FAB — แสดงทุก platform
      floatingActionButton: Column(
        mainAxisAlignment: MainAxisAlignment.end,
        children: [
          // ปุ่มตำแหน่งปัจจุบัน — แสดงเสมอทุก platform
          FloatingActionButton(
            heroTag: 'currentLocation',
            onPressed: _isLocating ? null : getCurrentLatLng,
            backgroundColor: global.theme.cardColor,
            child: _isLocating
                ? const SizedBox(
                    width: 24,
                    height: 24,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : Icon(Icons.my_location, color: global.theme.infoHighlightTextColor),
          ),
          const SizedBox(height: 10),
          // ปุ่มไปตำแหน่งที่บันทึกไว้ (ถ้ามี)
          if (widget.latitude != 0 || widget.longitude != 0)
            FloatingActionButton(
              heroTag: 'savedLocation',
              onPressed: () {
                final savedLocation = LatLng(widget.latitude, widget.longitude);
                _mapController.move(savedLocation, 15);
                setState(() {
                  defaultLocation = savedLocation;
                  if (customMarkers.isNotEmpty) {
                    customMarkers[0] = buildPin(savedLocation);
                  } else {
                    customMarkers.add(buildPin(savedLocation));
                  }
                });
              },
              backgroundColor: global.theme.cardColor,
              child: Icon(Icons.location_on, color: global.theme.negativeHighlightTextColor),
            ),
        ],
      ),
    );
  }
}
