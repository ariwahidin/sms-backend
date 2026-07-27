package email

import (
	"fmt"
	"log"
	"strings"

	"gopkg.in/gomail.v2"

	"sms-backend/config"
	"sms-backend/model"
	"sms-backend/repository"
)

type EmailService struct {
	userRepo *repository.UserRepository
}

func NewEmailService(userRepo *repository.UserRepository) *EmailService {
	return &EmailService{userRepo: userRepo}
}

// func (e *EmailService) send(to []string, subject, body string) {
// 	if config.Cfg.SMTP.Host == "" {
// 		log.Printf("[EMAIL] SMTP not configured, skipping email to %v | subject: %s", to, subject)
// 		return
// 	}

// 	m := gomail.NewMessage()
// 	m.SetHeader("From", config.Cfg.SMTP.From)
// 	m.SetHeader("To", to...)
// 	m.SetHeader("Subject", subject)
// 	m.SetBody("text/html", body)

// 	d := gomail.NewDialer(
// 		config.Cfg.SMTP.Host,
// 		config.Cfg.SMTP.Port,
// 		config.Cfg.SMTP.User,
// 		config.Cfg.SMTP.Password,
// 	)

// 	if err := d.DialAndSend(m); err != nil {
// 		log.Printf("[EMAIL] Failed to send to %v: %v", to, err)
// 		return
// 	}
// 	log.Printf("[EMAIL] Sent '%s' to %v", subject, to)
// }

func (e *EmailService) send(to, cc []string, subject, body string) {
	if config.Cfg.SMTP.Host == "" {
		log.Printf("[EMAIL] SMTP not configured, skipping email to %v cc %v | subject: %s", to, cc, subject)
		return
	}

	m := gomail.NewMessage()
	m.SetHeader("From", config.Cfg.SMTP.From)
	m.SetHeader("To", to...)
	if len(cc) > 0 {
		m.SetHeader("Cc", cc...)
	}
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(
		config.Cfg.SMTP.Host,
		config.Cfg.SMTP.Port,
		config.Cfg.SMTP.User,
		config.Cfg.SMTP.Password,
	)

	if err := d.DialAndSend(m); err != nil {
		log.Printf("[EMAIL] Failed to send to %v cc %v: %v", to, cc, err)
		return
	}
	log.Printf("[EMAIL] Sent '%s' to %v cc %v", subject, to, cc)
}

// collectEmails extracts unique non-empty emails from a user slice.
func collectEmails(users []model.User) []string {
	seen := make(map[string]bool)
	var out []string
	for _, u := range users {
		if u.Email != "" && !seen[u.Email] {
			seen[u.Email] = true
			out = append(out, u.Email)
		}
	}
	return out
}

type Recipient struct {
	Name       string `json:"name"`
	Email      string `json:"email"`
	Role       string `json:"role"`
	Department string `json:"department"`
}

// tidak berubah — dipakai SendDeclarationNotice
// func (e *EmailService) GetRecipients(report *model.RiskReport) []string {
// 	all := e.getRecipientUsers(report)
// 	return collectEmails(all)
// }

// tetap dipakai SendDeclarationNotice, tapi sekarang return (to, cc)
func (e *EmailService) GetRecipients(report *model.RiskReport) (to []string, cc []string) {
	pics, hods := e.getRecipientUsersSplit(report)
	to = collectEmails(pics)
	cc = collectEmails(hods)
	return to, cc
}

// baru — dipakai ReportHandler.GetByID buat ditampilkan di UI
func (e *EmailService) GetRecipientDetails(report *model.RiskReport) []Recipient {
	pics, hods := e.getRecipientUsersSplit(report)

	var recipients []Recipient
	for _, u := range pics {
		dept := ""
		if u.Department != nil {
			dept = u.Department.Name
		}
		recipients = append(recipients, Recipient{Name: u.Name, Email: u.Email, Role: "pic", Department: dept})
	}
	for _, u := range hods {
		dept := ""
		if u.Department != nil {
			dept = u.Department.Name
		}
		recipients = append(recipients, Recipient{Name: u.Name, Email: u.Email, Role: "hod", Department: dept})
	}
	return recipients
}

func (e *EmailService) getRecipientUsersSplit(report *model.RiskReport) (pics, hods []model.User) {
	var deptID uint
	if report.DepartmentID != nil {
		deptID = *report.DepartmentID
	}
	var err error
	pics, err = e.userRepo.FindActivePICs(false, deptID)
	if err != nil {
		log.Printf("[EMAIL] Failed to query PICs: %v", err)
	}
	hods, err = e.userRepo.FindActiveHODsByDepartment(deptID)
	if err != nil {
		log.Printf("[EMAIL] Failed to query HODs: %v", err)
	}
	return pics, hods
}

func (e *EmailService) getRecipientUsers(report *model.RiskReport) []model.User {
	pics, hods := e.getRecipientUsersSplit(report)
	return append(pics, hods...)
}

// func (e *EmailService) GetRecipients(report *model.RiskReport) []string {
// 	var deptID uint
// 	if report.DepartmentID != nil {
// 		deptID = *report.DepartmentID
// 	}

// 	log.Printf("[EMAIL] Declaring recipients for %s (deptID: %d)", report.Code, deptID)

// 	// PIC: all active PICs across all departments.
// 	// To restrict to department: change false → true.
// 	pics, err := e.userRepo.FindActivePICs(false, deptID)
// 	if err != nil {
// 		log.Printf("[EMAIL] Failed to query PICs: %v", err)
// 	}

// 	// HOD: active HODs in the report's department only.
// 	hods, err := e.userRepo.FindActiveHODsByDepartment(deptID)
// 	if err != nil {
// 		log.Printf("[EMAIL] Failed to query HODs: %v", err)
// 	}

// 	all := append(pics, hods...)

// 	log.Printf("[EMAIL] Declared %d recipients for %s", len(all), report.Code)
// 	return collectEmails(all)
// }

func (e *EmailService) SendDeclarationNotice(report *model.RiskReport) {
	subject := fmt.Sprintf("New Safety Risk Report %s", report.Code)
	body := buildEmailBody(subject, []emailField{
		{"Report Code", report.Code},
		{"Description", report.Description},
		{"Risk Level", string(report.RiskLevel)},
		{"Status", string(report.Status)},
		{"Reporter", safeStr(report.ReporterName, "Anonymous")},
		{"Reported At", report.CreatedAt.Format("02 Jan 2006, 15:04 WIB")},
		{"View Report", fmt.Sprintf("%s/reports/%d", config.Cfg.App.AppURL, report.ID)},
	})
	to, cc := e.GetRecipients(report)
	if len(to) == 0 && len(cc) == 0 {
		log.Printf("[EMAIL] No recipients for declaration notice %s", report.Code)
		return
	}
	e.send(to, cc, subject, body)
}

func (e *EmailService) SendStatusUpdateNotice(report *model.RiskReport, action string) {
	subject := fmt.Sprintf("Update on Risk Report %s - %s", report.Code, action)
	body := buildEmailBody(subject, []emailField{
		{"Report Code", report.Code},
		{"Action", action},
		{"Risk Level", string(report.RiskLevel)},
		{"Priority", string(report.Priority)},
		{"Status", string(report.Status)},
		{"Reported At", report.CreatedAt.Format("02 Jan 2006, 15:04 WIB")},
		{"Updated At", report.UpdatedAt.Format("02 Jan 2006, 15:04 WIB")},
		{"View Report", fmt.Sprintf("%s/reports/%d", config.Cfg.App.AppURL, report.ID)},
	})
	to, cc := e.GetRecipients(report)
	if len(to) == 0 && len(cc) == 0 {
		log.Printf("[EMAIL] No recipients for declaration notice %s", report.Code)
		return
	}
	e.send(to, cc, subject, body)
}

func (e *EmailService) SendReviewNotice(report *model.RiskReport) {
	subject := fmt.Sprintf("Review Required: Risk Report %s", report.Code)
	body := buildEmailBody(subject, []emailField{
		{"Report Code", report.Code},
		{"Description", report.Description},
		{"Root Cause", report.RootCause},
		{"Countermeasures", report.Countermeasures},
		{"Status", "Pending HOD Approval"},
		{"Reported At", report.CreatedAt.Format("02 Jan 2006, 15:04 WIB")},
		{"Investigated At", report.InvestigatedAt.Format("02 Jan 2006, 15:04 WIB")},
		{"View Report", fmt.Sprintf("%s/reports/%d", config.Cfg.App.AppURL, report.ID)},
	})
	to, cc := e.GetRecipients(report)
	if len(to) == 0 && len(cc) == 0 {
		log.Printf("[EMAIL] No recipients for declaration notice %s", report.Code)
		return
	}
	e.send(to, cc, subject, body)
}

func (e *EmailService) SendClosingNotice(report *model.RiskReport) {
	subject := fmt.Sprintf("Closed: Risk Report %s", report.Code)
	body := buildEmailBody(subject, []emailField{
		{"Report Code", report.Code},
		{"Description", report.Description},
		{"Status", "CLOSED"},
		{"HOD Comment", report.HODComment},
		{"Reported At", report.CreatedAt.Format("02 Jan 2006, 15:04 WIB")},
		{"Approved At", report.ApprovedAt.Format("02 Jan 2006, 15:04 WIB")},
		{"View Report", fmt.Sprintf("%s/reports/%d", config.Cfg.App.AppURL, report.ID)},
	})
	to, cc := e.GetRecipients(report)
	if len(to) == 0 && len(cc) == 0 {
		log.Printf("[EMAIL] No recipients for declaration notice %s", report.Code)
		return
	}
	e.send(to, cc, subject, body)
}

func (e *EmailService) SendRejectionNotice(report *model.RiskReport) {
	subject := fmt.Sprintf("Rejected: Risk Report %s - Revision Required", report.Code)
	body := buildEmailBody(subject, []emailField{
		{"Report Code", report.Code},
		{"Description", report.Description},
		{"Status", "Returned to Investigation"},
		{"Rejection Reason", report.HODComment},
		{"Reported At", report.CreatedAt.Format("02 Jan 2006, 15:04 WIB")},
		{"Rejected At", report.RejectedAt.Format("02 Jan 2006, 15:04 WIB")},
		{"View Report", fmt.Sprintf("%s/reports/%d", config.Cfg.App.AppURL, report.ID)},
	})
	to, cc := e.GetRecipients(report)
	if len(to) == 0 && len(cc) == 0 {
		log.Printf("[EMAIL] No recipients for declaration notice %s", report.Code)
		return
	}
	e.send(to, cc, subject, body)
}

func (e *EmailService) SendSLAOverdueAlert(report *model.RiskReport) {
	subject := fmt.Sprintf("OVERDUE ALERT: Risk Report %s", report.Code)
	body := buildEmailBody(subject, []emailField{
		{"Report Code", report.Code},
		{"Description", report.Description},
		{"Current Status", string(report.Status)},
		{"Priority", string(report.Priority)},
		{"Alert", "This report has exceeded its SLA deadline. Immediate action required."},
		{"Reported At", report.CreatedAt.Format("02 Jan 2006, 15:04 WIB")},
		{"View Report", fmt.Sprintf("%s/reports/%d", config.Cfg.App.AppURL, report.ID)},
	})
	to, cc := e.GetRecipients(report)
	if len(to) == 0 && len(cc) == 0 {
		log.Printf("[EMAIL] No recipients for declaration notice %s", report.Code)
		return
	}
	e.send(to, cc, subject, body)
}

type emailField struct {
	Label string
	Value string
}

func buildEmailBody(title string, fields []emailField) string {
	var rows strings.Builder
	for _, f := range fields {
		rows.WriteString(fmt.Sprintf(`
			<tr>
				<td style="padding:8px 12px;font-weight:600;color:#444;width:180px;border-bottom:1px solid #eee;">%s</td>
				<td style="padding:8px 12px;color:#222;border-bottom:1px solid #eee;">%s</td>
			</tr>`, f.Label, f.Value))
	}

	return fmt.Sprintf(`
	<div style="font-family:Arial,sans-serif;max-width:600px;margin:0 auto;border:1px solid #ddd;border-radius:8px;overflow:hidden;">
		<div style="background:#1a3a5c;padding:20px 24px;">
			<h2 style="color:#fff;margin:0;font-size:18px;">Yusen Logistics Interlink Indonesia — SMS Safety Management System</h2>
		</div>
		<div style="padding:20px 24px;">
			<h3 style="color:#1a3a5c;margin-top:0;">%s</h3>
			<table style="width:100%%;border-collapse:collapse;background:#f9f9f9;border-radius:6px;">
				%s
			</table>
		</div>
		<div style="background:#f5f5f5;padding:12px 24px;font-size:12px;color:#888;">
			This is an automated message from the SMS system. Please do not reply.
		</div>
	</div>`, title, rows.String())
}

func safeStr(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}
