package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github_report/logger"
	"github_report/model"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/noirbizarre/gonja"
	gomail "gopkg.in/mail.v2"
)

type Config struct {
	apiUrl                string
	apiToken              string
	smtpServer            string
	smtpPort              string
	smtpUser              string
	smtpPassword          string
	emailRecipientsReport string
	githubUser            string
	githubRepo            string
	emailSender           string
	dryRun                string
}

type Program struct {
	config Config
}

func main() {
	logger.Info("about to start the app...")

	logger.Info("Initializing env ...")
	configInfo := Config{}
	configInfo.dryRun = "true"
	configInfo.apiUrl = readManatoryConfig("apiUrl")
	configInfo.apiToken = readManatoryConfig("apiToken")
	configInfo.githubUser = readManatoryConfig("githubUser")
	configInfo.githubRepo = readManatoryConfig("githubRepo")
	configInfo.emailSender = readManatoryConfig("emailSender")
	configInfo.smtpServer = readConfig("smtpServer")
	configInfo.smtpPort = readConfig("smtpPort")
	configInfo.smtpUser = readConfig("smtpUser")
	configInfo.smtpPassword = readConfig("smtpPassword")
	configInfo.dryRun = readManatoryConfig("dryRun")
	configInfo.emailRecipientsReport = readManatoryConfig("emailRecipientsReport")

	logger.Info("Env initialized succesfully")

	program := &Program{
		config: configInfo,
	}
	openRequests, err := program.queryGitHubPullRequests("open")
	closedRequests, err := program.queryGitHubPullRequests("closed")
	inProgressRequests, err := program.queryGitHubPullRequests("in-progress")

	itemReport1 := model.ItemReport{Status: "Open", Count: len(openRequests)}
	itemReport2 := model.ItemReport{Status: "Closed", Count: len(closedRequests)}
	itemReport3 := model.ItemReport{Status: "In-Progress", Count: len(inProgressRequests)}

	var reportItems []model.ItemReport
	reportItems = append(reportItems, itemReport1)
	reportItems = append(reportItems, itemReport2)
	reportItems = append(reportItems, itemReport3)

	var reportItems2 []model.ItemReport

	count1, err := countRequestsMoreThanNDays(60, openRequests)
	count2, err := countRequestsMoreThanNDays(60, inProgressRequests)

	itemReport2_1 := model.ItemReport{Status: "Openened >= 60 ago and not closed", Count: count1 + count2}
	reportItems2 = append(reportItems2, itemReport2_1)

	templateContent, err := os.ReadFile("templates/report.html.j2")

	if err != nil {
		logger.Error("Cannot open jinja template", err)
	}
	tpl, err := gonja.FromString(string(templateContent))
	if err != nil {
		logger.Error("Error reading jinja content", err)

	}
	gitRepo := "https://github.com/" + program.config.githubUser + "/" + program.config.githubRepo + ".git"
	outputPath := "temp/report.html"
	out, err := tpl.Execute(gonja.Context{"pullRequests": reportItems, "gitrepo": gitRepo, "pullRequests2": reportItems2})
	file, err := os.Create(outputPath)
	if err != nil {
		logger.Error("Error creating email", err)
	}

	_, err = file.WriteString(out)
	if err != nil {
		logger.Error("Error writing temp file email", err)
	}
	file.Close()

	printOutput(outputPath, program.config.emailRecipientsReport, program.config.emailSender, gitRepo)

	/*err = os.Remove(outputPath)
	if err != nil {
		logger.Error("Error cleaning temp file", err)
	}*/

	if program.config.dryRun != "true" {
		smtpPort, err := strconv.Atoi(program.config.smtpPort)
		content, err := os.ReadFile("temp/report.html")
		if err != nil {
			logger.Error("Error reading temporary report file", err)
		}
		emailBdy := string(content)
		smtpSender := program.config.emailSender
		currentTime := time.Now()
		dateString := currentTime.Format("2006-01-02")
		subject := "Pull requests from " + gitRepo + " " + dateString
		dialer := gomail.NewDialer(program.config.smtpServer, smtpPort, smtpSender, program.config.smtpPassword)
		tos := strings.Split(program.config.emailRecipientsReport, ",")
		sendEmail(dialer, subject, emailBdy, smtpSender, tos, []string{}, []string{}, []string{})
	}

}

func readManatoryConfig(config string) string {
	value := os.Getenv(config)
	if value != "" {
		return value
	} else {
		err := fmt.Errorf("Configuration error")
		logger.Fatal(config+" Configuration value could not be empty", err)
	}
	return ""
}

func readConfig(config string) string {
	return os.Getenv(config)
}

func (program *Program) queryGitHubPullRequests(filter string) ([]model.GitHubPullRequest, error) {
	var pullRequests []model.GitHubPullRequest
	urlAPI := program.config.apiUrl + "/repos" + "/" + program.config.githubUser + "/" + program.config.githubRepo + "/pulls?state=" + filter
	request, err := http.NewRequest("GET", urlAPI, nil)
	if err != nil {
		logger.Error("Error creating get request to github", err)
		return pullRequests, err
	}
	request.Header.Set("Authorization", "Bearer "+program.config.apiToken)
	request.Header.Set("Content-Type", "application/vnd.github+json")
	request.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	response, err := client.Do(request)
	if err != nil {
		return pullRequests, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return pullRequests, fmt.Errorf("Error reading response: %d", err.Error())
	}
	err = json.Unmarshal(body, &pullRequests)
	if err != nil {
		logger.Error("Error parsing respons ", err)
	}
	return pullRequests, nil
}

func countRequestsMoreThanNDays(nDays int, pullRequests []model.GitHubPullRequest) (int, error) {
	now := time.Now()
	nDaysAgo := now.AddDate(0, 0, -nDays)
	count := 0
	for _, req := range pullRequests {
		date, err := time.Parse(time.RFC3339, req.CreatedAt)
		if err != nil {
			return -1, err
		}
		if date.Before(nDaysAgo) {
			count++
		}
	}
	return count, nil
}

func printOutput(path string, to string, from string, repo string) {
	content, err := os.ReadFile("temp/report.html")
	if err != nil {
		logger.Error("Error reading temporary report file", err)
	}
	printBody := string(content)
	currentTime := time.Now()
	dateString := currentTime.Format("2006-01-02")
	fmt.Println("From : " + from)
	fmt.Println("To: " + to)
	fmt.Println("Subject: Pull requests from " + repo + " " + dateString)
	fmt.Println("Body: " + printBody)
}

func sendEmail(dialer *gomail.Dialer, subject string, body string, from string, tos []string, cc []string, bcc []string, attachmentsPaths []string) {
	m := gomail.NewMessage()

	m.SetHeader("From", from)
	m.SetHeader("To", tos...)
	m.SetHeader("Cc", cc...)
	m.SetHeader("Bcc", bcc...)
	m.SetHeader("Subject", subject)

	m.SetBody("text/html", body)

	for _, filePath := range attachmentsPaths {
		m.Attach(filePath)
	}

	if err := dialer.DialAndSend(m); err != nil {
		fmt.Println(err)
		logger.Error("Error sending email ", err)
	}
}
