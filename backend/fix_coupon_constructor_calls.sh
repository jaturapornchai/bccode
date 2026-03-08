#!/bin/bash

echo "=== Fix CouponHttpService Constructor Calls ==="
echo

echo "The following files need to be updated to add the new dependencies:"
echo "1. internal/transaction/saleinvoice/saleinvoice_http.go"
echo "2. internal/transaction/pickandpack/pickandpack_http.go"
echo "3. internal/product/eorder/eorder_http.go"
echo "4. internal/slipimage/slipimage_http.go"
echo "5. internal/pos/shift/shift_http.go"

echo
echo "Required changes for each file:"
echo

echo "OLD Constructor Call:"
echo "svc := services.NewCouponHttpService(repo, reservationRepo, usageHistoryRepo)"

echo
echo "NEW Constructor Call:"
echo "// Add these repositories"
echo "productBarcodeRepo := productbarcode_repositories.NewProductBarcodeRepository(pst, cache)"
echo "masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)"
echo "svc := services.NewCouponHttpService(repo, reservationRepo, usageHistoryRepo, productBarcodeRepo, masterSyncCacheRepo)"

echo
echo "Required imports to add:"
echo 'mastersync "smlcloudplatform/internal/mastersync/repositories"'
echo 'productbarcode_repositories "smlcloudplatform/internal/product/productbarcode/repositories"'

echo
echo "=== Example Implementation ==="
echo

cat << 'EOF'
// Add to imports
import (
    // ... existing imports ...
    mastersync "smlcloudplatform/internal/mastersync/repositories"
    productbarcode_repositories "smlcloudplatform/internal/product/productbarcode/repositories"
)

// Update constructor
func NewXXXHttp(ms *microservice.Microservice, cfg config.IConfig) XXXHttp {
    pst := ms.MongoPersister(cfg.MongoPersisterConfig())
    cache := ms.Cacher(cfg.CacherConfig()) // Make sure cache is available

    // ... existing repos ...
    repo := repositories.NewCouponRepository(pst)
    reservationRepo := repositories.NewCouponReservationRepository(pst)
    usageHistoryRepo := repositories.NewCouponUsageHistoryRepository(pst)
    
    // Add new dependencies
    productBarcodeRepo := productbarcode_repositories.NewProductBarcodeRepository(pst, cache)
    masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
    
    svc := services.NewCouponHttpService(repo, reservationRepo, usageHistoryRepo, productBarcodeRepo, masterSyncCacheRepo)
    
    // ... rest of function ...
}
EOF

echo
echo "After fixing all files, run: go build -v ."
echo "All compilation errors should be resolved."