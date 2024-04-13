package mail

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"github_report/logger"
	"log"
	"net"
	"net/smtp"
	"strconv"
	"strings"
)

type SMTPClientData struct {
	Host     string
	Port     int
	UserName string
	Password string
}

type SMTPAttachments struct {
	Name     string
	MimeType string
	Content  []byte
}

type smtpClient struct{}

var (
	SMTPClient smtpClientInterface = &smtpClient{}
)

type smtpClientInterface interface {
	SendEmail(smtpClientData SMTPClientData, subject string, body string, from string, tos []string, cc []string, bcc []string, attachments []SMTPAttachments) error
}

func (smtpClient *smtpClient) SendEmail(smtpClientData SMTPClientData, subject string, body string, from string, tos []string, cc []string, bcc []string, attachments []SMTPAttachments) error {

	servername := smtpClientData.Host + ":" + strconv.Itoa(smtpClientData.Port)
	host, _, _ := net.SplitHostPort(servername)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}

	conn, err := tls.Dial("tcp", servername, tlsConfig)
	if err != nil {
		logger.Error("Error connecting to SMTP server:", err)
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		logger.Error("Error creating SMTP client:", err)
		return err
	}
	defer client.Quit()

	auth := smtp.PlainAuth("", smtpClientData.UserName, smtpClientData.Password, host)
	if err := client.Auth(auth); err != nil {
		logger.Error("Error authenticating with SMTP server:", err)
		return err
	}

	if err := client.Mail(smtpClientData.UserName); err != nil {
		fmt.Println("Error setting sender:", err)
		return nil
	}

	if err := client.Rcpt(tos[0]); err != nil {
		fmt.Println("Error setting recipient:", err)
		return nil
	}

	writer, writerErr := client.Data()
	if writerErr != nil {
		logger.Error("Error writing data to email:", err)
	}
	sampleMsg := fmt.Sprintf("From: %s\r\n", from)
	sampleMsg += fmt.Sprintf("To: %s\r\n", strings.Join(tos, ";"))
	if len(cc) > 0 {
		sampleMsg += fmt.Sprintf("Cc: %s\r\n", strings.Join(cc, ";"))
	}
	if len(bcc) > 0 {
		sampleMsg += fmt.Sprintf("Bcc: %s\r\n", strings.Join(cc, ";"))
	}
	sampleMsg += "Subject: " + subject + "\r\n"

	delimeter := "boundary123456"

	log.Println("Mark content to accept multiple contents")
	sampleMsg += "MIME-Version: 1.0\r\n"
	sampleMsg += fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", delimeter)

	log.Println("Put HTML message")
	sampleMsg += fmt.Sprintf("\r\n--%s\r\n", delimeter)
	sampleMsg += "Content-Type: text/html; charset=\"utf-8\"\r\n"
	sampleMsg += "Content-Transfer-Encoding: 7bit\r\n"
	sampleMsg += fmt.Sprintf("\r\n%s", body+"\r\n")

	for _, attachment := range attachments {
		log.Println("Put file attachment")
		sampleMsg += fmt.Sprintf("\r\n--%s\r\n", delimeter)
		sampleMsg += "Content-Type: " + attachment.MimeType + "; charset=\"utf-8\"\r\n"
		sampleMsg += "Content-Transfer-Encoding: base64\r\n"
		sampleMsg += "Content-Disposition: attachment;filename=\"" + attachment.Name + "\"\r\n"
		rawFile := attachment.Content
		sampleMsg += "\r\n" + base64.StdEncoding.EncodeToString(rawFile)

	}

	log.Println("Write content into client writter I/O")
	if _, err := writer.Write([]byte(sampleMsg)); err != nil {
		log.Panic(err)
	}
	if closeErr := writer.Close(); closeErr != nil {
		log.Panic(closeErr)
	}

	client.Quit()
	return nil
}
