package model

type GitHubPullRequest struct {
	Title     string     `json:"title"`
	Number    int        `json:"number"`
	CreatedAt string     `json:"created_at"`
	User      GitHubUser `json:"user"`
}

type GitHubUser struct {
	Login string `json:"login"`
}

type ItemReport struct {
	Status string
	Count  int
}
