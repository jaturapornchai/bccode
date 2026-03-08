package services

import (
	"context"
	"log"
	"smlcloudplatform/internal/coupon/repositories"
	"time"
)

type ICouponCleanupService interface {
	CleanupExpiredReservations(shopID string) error
	StartCleanupScheduler(ctx context.Context, interval time.Duration, shopID string)
}

type CouponCleanupService struct {
	reservationRepo repositories.CouponReservationRepository
}

func NewCouponCleanupService(reservationRepo repositories.CouponReservationRepository) ICouponCleanupService {
	return &CouponCleanupService{
		reservationRepo: reservationRepo,
	}
}

// ยกเลิกการจองที่หมดอายุ
func (s *CouponCleanupService) CleanupExpiredReservations(shopID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := s.reservationRepo.CleanupExpiredReservations(ctx, shopID)
	if err != nil {
		if shopID != "" {
			log.Printf("Error cleaning up expired reservations for shop %s: %v", shopID, err)
		} else {
			log.Printf("Error cleaning up expired reservations for all shops: %v", err)
		}
		return err
	}

	if shopID != "" {
		log.Printf("Cleaned up expired reservations for shop %s", shopID)
	} else {
		log.Printf("Cleaned up expired reservations for all shops")
	}
	return nil
}

// เริ่มต้น scheduler สำหรับ cleanup อัตโนมัติ
func (s *CouponCleanupService) StartCleanupScheduler(ctx context.Context, interval time.Duration, shopID string) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	if shopID != "" {
		log.Printf("Starting coupon reservation cleanup scheduler for shop %s with interval %v", shopID, interval)
	} else {
		log.Printf("Starting coupon reservation cleanup scheduler for all shops with interval %v", interval)
	}

	for {
		select {
		case <-ctx.Done():
			if shopID != "" {
				log.Printf("Coupon cleanup scheduler stopped for shop %s", shopID)
			} else {
				log.Printf("Coupon cleanup scheduler stopped for all shops")
			}
			return
		case <-ticker.C:
			err := s.CleanupExpiredReservations(shopID)
			if err != nil {
				if shopID != "" {
					log.Printf("Cleanup scheduler error for shop %s: %v", shopID, err)
				} else {
					log.Printf("Cleanup scheduler error for all shops: %v", err)
				}
			}
		}
	}
}
