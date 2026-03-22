import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:flutter_map/flutter_map.dart';
import 'package:latlong2/latlong.dart';
import 'package:geolocator/geolocator.dart';
import 'package:http/http.dart' as http;
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
  int _mapType = 0; // 0 = OSM, 1 = Google

  // Search
  final _searchCtrl = TextEditingController();
  List<Map<String, dynamic>> _searchResults = [];
  bool _isSearching = false;
  bool _showSearchResults = false;

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

  @override
  void dispose() {
    _searchCtrl.dispose();
    super.dispose();
  }

  /// ค้นหาสถานที่ด้วย Nominatim (OpenStreetMap geocoding — รองรับภาษาไทย)
  Future<void> _searchLocation(String query) async {
    if (query.trim().isEmpty) {
      setState(() { _searchResults.clear(); _showSearchResults = false; });
      return;
    }
    setState(() => _isSearching = true);
    try {
      final uri = Uri.https('nominatim.openstreetmap.org', '/search', {
        'q': query.trim(),
        'format': 'json',
        'accept-language': 'th,en',
        'countrycodes': 'th',
        'limit': '8',
      });
      final resp = await http.get(uri, headers: {
        'User-Agent': 'BCAccountApp/1.0 (support@bcaicloud.com)',
        'Accept-Language': 'th,en;q=0.9',
        'Accept': 'application/json',
      });
      if (resp.statusCode == 200) {
        final body = utf8.decode(resp.bodyBytes);
        final list = json.decode(body) as List<dynamic>;
        setState(() {
          _searchResults = list.cast<Map<String, dynamic>>();
          _showSearchResults = true;
        });
      }
    } catch (e) {
      AppLogger.warning('[Map] ค้นหาสถานที่ไม่สำเร็จ: $e');
    } finally {
      if (mounted) setState(() => _isSearching = false);
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
            child: Stack(
              children: [
                FlutterMap(
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
                        _showSearchResults = false;
                      });
                    },
                    interactionOptions: const InteractionOptions(
                      flags: ~InteractiveFlag.doubleTapZoom,
                    ),
                  ),
                  children: [
                    if (_mapType == 0)
                      TileLayer(
                        urlTemplate: 'https://tile.openstreetmap.org/{z}/{x}/{y}.png',
                        userAgentPackageName: 'com.bcaicloud.bcaiaccount',
                      )
                    else
                      TileLayer(
                        urlTemplate: 'https://mt{s}.google.com/vt/lyrs=y&x={x}&y={y}&z={z}',
                        subdomains: const ['0', '1', '2', '3'],
                        userAgentPackageName: 'com.bcaicloud.bcaiaccount',
                      ),
                    MarkerLayer(markers: customMarkers),
                  ],
                ),
                // Search bar overlay
                Positioned(
                  top: 8,
                  left: 8,
                  right: 8,
                  child: Column(
                    children: [
                      Material(
                        elevation: 4,
                        borderRadius: BorderRadius.circular(8),
                        color: global.theme.cardColor,
                        child: TextField(
                          controller: _searchCtrl,
                          decoration: InputDecoration(
                            hintText: global.language('search_location'),
                            hintStyle: TextStyle(color: global.theme.textSecondaryColor),
                            prefixIcon: _isSearching
                                ? const Padding(
                                    padding: EdgeInsets.all(12),
                                    child: SizedBox(width: 20, height: 20, child: CircularProgressIndicator(strokeWidth: 2)),
                                  )
                                : Icon(Icons.search, color: global.theme.iconColor),
                            suffixIcon: _searchCtrl.text.isNotEmpty
                                ? IconButton(
                                    icon: Icon(Icons.clear, size: 18, color: global.theme.textSecondaryColor),
                                    onPressed: () => setState(() {
                                      _searchCtrl.clear();
                                      _searchResults.clear();
                                      _showSearchResults = false;
                                    }),
                                  )
                                : null,
                            border: OutlineInputBorder(borderRadius: BorderRadius.circular(8), borderSide: BorderSide.none),
                            filled: true,
                            fillColor: global.theme.cardColor,
                            contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 12),
                          ),
                          style: TextStyle(color: global.theme.textColor),
                          onSubmitted: _searchLocation,
                          onChanged: (v) {
                            if (v.isEmpty) setState(() { _searchResults.clear(); _showSearchResults = false; });
                          },
                        ),
                      ),
                      if (_showSearchResults && _searchResults.isNotEmpty)
                        Material(
                          elevation: 4,
                          borderRadius: BorderRadius.circular(8),
                          color: global.theme.cardColor,
                          child: ConstrainedBox(
                            constraints: const BoxConstraints(maxHeight: 220),
                            child: ListView.separated(
                              shrinkWrap: true,
                              padding: EdgeInsets.zero,
                              itemCount: _searchResults.length,
                              separatorBuilder: (context, i) => Divider(height: 1, color: global.theme.textSecondaryColor.withValues(alpha: 0.2)),
                              itemBuilder: (_, i) {
                                final r = _searchResults[i];
                                final lat = double.tryParse(r['lat']?.toString() ?? '') ?? 0;
                                final lng = double.tryParse(r['lon']?.toString() ?? '') ?? 0;
                                return ListTile(
                                  dense: true,
                                  leading: Icon(Icons.place_outlined, size: 18, color: global.theme.iconColor),
                                  title: Text(
                                    r['display_name']?.toString() ?? '',
                                    maxLines: 2,
                                    overflow: TextOverflow.ellipsis,
                                    style: TextStyle(fontSize: 13, color: global.theme.textColor),
                                  ),
                                  onTap: () {
                                    _updateLocation(lat, lng);
                                    setState(() {
                                      _searchCtrl.clear();
                                      _searchResults.clear();
                                      _showSearchResults = false;
                                    });
                                  },
                                );
                              },
                            ),
                          ),
                        ),
                    ],
                  ),
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
          // ปุ่มสลับ map type
          FloatingActionButton.small(
            heroTag: 'mapType',
            onPressed: () => setState(() => _mapType = _mapType == 0 ? 1 : 0),
            backgroundColor: _mapType == 1 ? global.theme.primaryColor : global.theme.cardColor,
            child: Icon(Icons.satellite_alt, size: 18, color: _mapType == 1 ? global.theme.cardColor : global.theme.textColor),
          ),
          const SizedBox(height: 10),
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
