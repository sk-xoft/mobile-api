package email

import (
	"fmt"
	"net/smtp"
	"strings"

	"github.com/smartseaman/mobile-api/internal/config"
)

type Sender struct{ config config.Email }

func NewSender(cfg config.Email) *Sender { return &Sender{config: cfg} }

func (s *Sender) SendRegistration(fullName, recipient, uid string) error {
	link := strings.TrimRight(s.config.RegisterURL, "/") + "/?uid=" + uid
	body := "ขอต้อนรับ คุณ " + fullName + " เข้าสู่แอปพลิเคชั่น Smart Seaman" +
		"<br/>การลงทะเบียนใกล้เสร็จสมบูรณ์ กรุณาคลิกลิงก์เพื่อยืนยันอีเมล " + "<br/>" + link
	return s.send(recipient, "กรุณายืนยันอีเมลเพื่อเข้าสู่ระบบ Smart Seaman", body)
}

func (s *Sender) SendPasswordReset(fullName, recipient, uid string) error {
	link := strings.TrimRight(s.config.ResetPasswordURL, "/") + "/?uid=" + uid
	body := "เรียนคุณ " + fullName + "<br/>คุณได้ส่งคำขอเพื่อตั้งรหัสผ่านใหม่ กรุณาคลิกลิงก์เพื่อดำเนินการ<br/>" + link
	return s.send(recipient, "ขอตั้งรหัสผ่านใหม่", body)
}

func (s *Sender) send(recipient, subject, body string) error {
	if s.config.Username == "" || s.config.Password == "" {
		return fmt.Errorf("SMTP credentials are not configured")
	}
	address := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
	message := []byte("From: Smart Seaman <" + s.config.Username + ">\r\n" +
		"To: " + recipient + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n" + body)
	return smtp.SendMail(address, auth, s.config.Username, []string{recipient}, message)
}
