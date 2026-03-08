#!/bin/bash

# Simple API test
response=$(curl -s -X POST "http://localhost:8080/product/barcode/import-refbarcode" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer 70eef34a9a83d9f39d634810b427340285a75b0f0eb4dd1f20687578d6d1ecfc" \
  -d '[{"barcode": "995002", "standvalue": 30, "dividevalue": 1, "barcoderef": "995001"}]')

echo "API Response:"
echo "$response"