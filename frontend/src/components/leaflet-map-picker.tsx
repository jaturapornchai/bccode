"use client";

import "leaflet/dist/leaflet.css";

import L from "leaflet";
import { useEffect } from "react";
import {
  MapContainer,
  Marker,
  TileLayer,
  useMap,
  useMapEvents,
} from "react-leaflet";

const ICON_RETINA = "/leaflet/marker-icon-2x.png";
const ICON = "/leaflet/marker-icon.png";
const SHADOW = "/leaflet/marker-shadow.png";

const defaultIcon = L.icon({
  iconUrl: ICON,
  iconRetinaUrl: ICON_RETINA,
  shadowUrl: SHADOW,
  iconSize: [25, 41],
  iconAnchor: [12, 41],
  popupAnchor: [1, -34],
  shadowSize: [41, 41],
});

export type LeafletMapPickerProps = {
  lat: number;
  lng: number;
  zoom: number;
  onChange: (lat: number, lng: number) => void;
};

export default function LeafletMapPicker({
  lat,
  lng,
  zoom,
  onChange,
}: LeafletMapPickerProps) {
  return (
    <MapContainer
      center={[lat, lng]}
      zoom={zoom}
      scrollWheelZoom
      style={{ height: "100%", width: "100%" }}
    >
      <TileLayer
        attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
        url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
      />
      <ClickToMoveMarker lat={lat} lng={lng} onChange={onChange} />
      <RecenterOnExternalChange lat={lat} lng={lng} />
    </MapContainer>
  );
}

function ClickToMoveMarker({
  lat,
  lng,
  onChange,
}: {
  lat: number;
  lng: number;
  onChange: (lat: number, lng: number) => void;
}) {
  useMapEvents({
    click(event) {
      onChange(event.latlng.lat, event.latlng.lng);
    },
  });
  return (
    <Marker
      position={[lat, lng]}
      icon={defaultIcon}
      draggable
      eventHandlers={{
        dragend(event) {
          const next = event.target.getLatLng();
          onChange(next.lat, next.lng);
        },
      }}
    />
  );
}

function RecenterOnExternalChange({ lat, lng }: { lat: number; lng: number }) {
  const map = useMap();
  useEffect(() => {
    const current = map.getCenter();
    const moved =
      Math.abs(current.lat - lat) > 0.0001 ||
      Math.abs(current.lng - lng) > 0.0001;
    if (!moved) return;
    map.panTo([lat, lng], { animate: true });
  }, [lat, lng, map]);
  return null;
}
