package services

import (
	"context"
	"log"
	"smlcloudplatform/internal/coupon/repositories"
	"time"
)

type ICouponCleanupService interface {
	CleanupExpiredReservations(holdingCode string) error
	StartCleanupScheduler(ctx context.Context, interval time.Duration, holdingCode string)
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
func (s *CouponCleanupService) CleanupExpiredReservations(holdingCode string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := s.reservationRepo.CleanupExpiredReservations(ctx, holdingCode)
	if err != nil {
		if holdingCode != "" {
			log.Printf("Error cleaning up expired reservations for shop %s: %v", holdingCode, err)
		} else {
			log.Printf("Error cleaning up expired reservations for all shops: %v", err)
		}
		return err
	}

	if holdingCode != "" {
		log.Printf("Cleaned up expired reservations for shop %s", holdingCode)
	} else {
		log.Printf("Cleaned up expired reservations for all shops")
	}
	return nil
}

// เริ่มต้น scheduler สำหรับ cleanup อัตโนมัติ
func (s *CouponCleanupService) StartCleanupScheduler(ctx context.Context, interval time.Duration, holdingCode string) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	if holdingCode != "" {
		log.Printf("Starting coupon reservation cleanup scheduler for shop %s with interval %v", holdingCode, interval)
	} else {
		log.Printf("Starting coupon reservation cleanup scheduler for all shops with interval %v", interval)
	}

	for {
		select {
		case <-ctx.Done():
			if holdingCode != "" {
				log.Printf("Coupon cleanup scheduler stopped for shop %s", holdingCode)
			} else {
				log.Printf("Coupon cleanup scheduler stopped for all shops")
			}
			return
		case <-ticker.C:
			err := s.CleanupExpiredReservations(holdingCode)
			if err != nil {
				if holdingCode != "" {
					log.Printf("Cleanup scheduler error for shop %s: %v", holdingCode, err)
				} else {
					log.Printf("Cleanup scheduler error for all shops: %v", err)
				}
			}
		}
	}
}
