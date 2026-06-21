package scheduler

import (
	"fmt"
	"log"

	"github.com/robfig/cron/v3"

	"sms-backend/config"
	"sms-backend/email"
	"sms-backend/repository"
)

type SLAChecker struct {
	reportRepo *repository.ReportRepository
	emailSvc   *email.EmailService
	cron       *cron.Cron
}

func NewSLAChecker(reportRepo *repository.ReportRepository, emailSvc *email.EmailService) *SLAChecker {
	return &SLAChecker{
		reportRepo: reportRepo,
		emailSvc:   emailSvc,
		cron:       cron.New(),
	}
}

func (s *SLAChecker) Start() {
	interval := config.Cfg.Scheduler.SLACheckIntervalMinutes
	spec := fmt.Sprintf("@every %dm", interval)

	s.cron.AddFunc(spec, s.check)
	s.cron.Start()

	log.Printf("[SCHEDULER] SLA checker started — runs every %d minutes", interval)
}

func (s *SLAChecker) Stop() {
	s.cron.Stop()
	log.Println("[SCHEDULER] SLA checker stopped")
}

func (s *SLAChecker) check() {
	log.Println("[SCHEDULER] Running SLA overdue check...")

	reports, err := s.reportRepo.FindOverdue()
	if err != nil {
		log.Printf("[SCHEDULER] Error fetching overdue reports: %v", err)
		return
	}

	if len(reports) == 0 {
		log.Println("[SCHEDULER] No overdue reports found")
		return
	}

	log.Printf("[SCHEDULER] Found %d overdue report(s)", len(reports))

	for i := range reports {
		report := &reports[i]
		report.IsOverdue = true
		s.reportRepo.Update(report)
		go s.emailSvc.SendSLAOverdueAlert(report)
		log.Printf("[SCHEDULER] Overdue alert sent for %s (status: %s)", report.Code, report.Status)
	}
}
