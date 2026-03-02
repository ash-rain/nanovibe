package services

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	gogithub "github.com/google/go-github/v60/github"
	"golang.org/x/oauth2"
)

// GitHubUser holds basic GitHub profile information.
type GitHubUser struct {
	Login       string `json:"login"`
	AvatarURL   string `json:"avatarUrl"`
	PublicRepos int    `json:"publicRepos"`
}

// Repo holds summary information about a GitHub repository.
type Repo struct {
	FullName    string    `json:"fullName"`
	Description string    `json:"description"`
	Language    string    `json:"language"`
	CloneURL    string    `json:"cloneUrl"`
	HTMLURL     string    `json:"htmlUrl"`
	Stars       int       `json:"stars"`
	Private     bool      `json:"private"`
	PushedAt    time.Time `json:"pushedAt"`
}

// PR holds pull request information.
type PR struct {
	Number     int       `json:"number"`
	Title      string    `json:"title"`
	State      string    `json:"state"`
	HeadBranch string    `json:"headBranch"`
	BaseBranch string    `json:"baseBranch"`
	HTMLURL    string    `json:"htmlUrl"`
	CreatedAt  time.Time `json:"createdAt"`
}

// PRCreate holds the parameters for creating a pull request.
type PRCreate struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Head  string `json:"head"`
	Base  string `json:"base"`
}

// GitHubEvent represents a GitHub activity event.
type GitHubEvent struct {
	Type        string    `json:"type"`
	RepoName    string    `json:"repoName"`
	Description string    `json:"description"`
	Branch      string    `json:"branch"`
	CreatedAt   time.Time `json:"createdAt"`
}

// GetClient returns an authenticated GitHub client using the stored token.
func GetClient() (*gogithub.Client, error) {
	token, ok := Get("github_token")
	if !ok || token == "" {
		return nil, fmt.Errorf("github: not authenticated")
	}
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)
	return gogithub.NewClient(tc), nil
}

// IsAuthenticated returns true if a GitHub token is stored.
func IsAuthenticated() bool {
	return Exists("github_token")
}

// GetUser returns the authenticated user's profile.
func GetUser(ctx context.Context) (*GitHubUser, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}
	u, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("github: get user: %w", err)
	}
	return &GitHubUser{
		Login:       u.GetLogin(),
		AvatarURL:   u.GetAvatarURL(),
		PublicRepos: u.GetPublicRepos(),
	}, nil
}

// ListRepos returns a paginated list of the authenticated user's repositories.
// When search is non-empty a simple name filter is applied client-side.
func ListRepos(ctx context.Context, page int, search string) ([]Repo, int, error) {
	client, err := GetClient()
	if err != nil {
		return nil, 0, err
	}
	if page < 1 {
		page = 1
	}
	opts := &gogithub.RepositoryListByAuthenticatedUserOptions{
		Sort: "pushed",
		ListOptions: gogithub.ListOptions{
			Page:    page,
			PerPage: 30,
		},
	}
	repos, resp, err := client.Repositories.ListByAuthenticatedUser(ctx, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("github: list repos: %w", err)
	}

	var total int
	if resp != nil {
		total = resp.LastPage * 30
		if total == 0 {
			total = len(repos)
		}
	}

	var result []Repo
	for _, r := range repos {
		name := r.GetFullName()
		if search != "" && !strings.Contains(strings.ToLower(name), strings.ToLower(search)) {
			continue
		}
		result = append(result, Repo{
			FullName:    name,
			Description: r.GetDescription(),
			Language:    r.GetLanguage(),
			CloneURL:    r.GetCloneURL(),
			HTMLURL:     r.GetHTMLURL(),
			Stars:       r.GetStargazersCount(),
			Private:     r.GetPrivate(),
			PushedAt:    r.GetPushedAt().Time,
		})
	}
	return result, total, nil
}

// ListPRs returns open pull requests for the given repository.
func ListPRs(ctx context.Context, owner, repo string) ([]PR, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}
	prs, _, err := client.PullRequests.List(ctx, owner, repo, &gogithub.PullRequestListOptions{
		State: "open",
		ListOptions: gogithub.ListOptions{PerPage: 50},
	})
	if err != nil {
		return nil, fmt.Errorf("github: list prs: %w", err)
	}
	var result []PR
	for _, pr := range prs {
		result = append(result, PR{
			Number:     pr.GetNumber(),
			Title:      pr.GetTitle(),
			State:      pr.GetState(),
			HeadBranch: pr.GetHead().GetRef(),
			BaseBranch: pr.GetBase().GetRef(),
			HTMLURL:    pr.GetHTMLURL(),
			CreatedAt:  pr.GetCreatedAt().Time,
		})
	}
	return result, nil
}

// CreatePR creates a new pull request.
func CreatePR(ctx context.Context, owner, repo string, params PRCreate) (*PR, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}
	pr, _, err := client.PullRequests.Create(ctx, owner, repo, &gogithub.NewPullRequest{
		Title: gogithub.String(params.Title),
		Body:  gogithub.String(params.Body),
		Head:  gogithub.String(params.Head),
		Base:  gogithub.String(params.Base),
	})
	if err != nil {
		return nil, fmt.Errorf("github: create pr: %w", err)
	}
	result := &PR{
		Number:     pr.GetNumber(),
		Title:      pr.GetTitle(),
		State:      pr.GetState(),
		HeadBranch: pr.GetHead().GetRef(),
		BaseBranch: pr.GetBase().GetRef(),
		HTMLURL:    pr.GetHTMLURL(),
		CreatedAt:  pr.GetCreatedAt().Time,
	}
	return result, nil
}

// GetActivity returns recent public events for the given GitHub login.
func GetActivity(ctx context.Context, login string) ([]GitHubEvent, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}
	events, _, err := client.Activity.ListEventsPerformedByUser(ctx, login, false, &gogithub.ListOptions{PerPage: 30})
	if err != nil {
		return nil, fmt.Errorf("github: get activity: %w", err)
	}

	var result []GitHubEvent
	for _, e := range events {
		ev := GitHubEvent{
			Type:      e.GetType(),
			RepoName:  e.GetRepo().GetName(),
			CreatedAt: e.GetCreatedAt().Time,
		}

		payload, _ := e.ParsePayload()
		switch p := payload.(type) {
		case *gogithub.PushEvent:
			ev.Branch = strings.TrimPrefix(p.GetRef(), "refs/heads/")
			if p.GetSize() > 0 && len(p.Commits) > 0 {
				ev.Description = p.Commits[0].GetMessage()
			}
		case *gogithub.PullRequestEvent:
			ev.Description = p.GetPullRequest().GetTitle()
		case *gogithub.CreateEvent:
			ev.Branch = p.GetRef()
			ev.Description = fmt.Sprintf("Created %s %s", p.GetRefType(), p.GetRef())
		case *gogithub.IssuesEvent:
			ev.Description = p.GetIssue().GetTitle()
		}

		result = append(result, ev)
	}
	return result, nil
}

// ImportRepo clones a GitHub repository and streams progress lines.
func ImportRepo(ctx context.Context, repoURL, destPath string) <-chan string {
	ch := make(chan string, 32)
	go func() {
		defer close(ch)
		gitCfg := gitConfigPath()
		cmd := exec.CommandContext(ctx, "git", "clone", "--progress", repoURL, destPath)
		cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL="+gitCfg)

		pr, pw, err := os.Pipe()
		if err != nil {
			ch <- fmt.Sprintf("error: %v", err)
			return
		}
		cmd.Stdout = pw
		cmd.Stderr = pw

		if err := cmd.Start(); err != nil {
			pw.Close()
			pr.Close()
			ch <- fmt.Sprintf("error: %v", err)
			return
		}
		pw.Close()

		scanner := bufio.NewScanner(pr)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				pr.Close()
				return
			case ch <- scanner.Text():
			}
		}
		pr.Close()

		if err := cmd.Wait(); err != nil {
			ch <- fmt.Sprintf("error: clone failed: %v", err)
		} else {
			ch <- "done"
		}
	}()
	return ch
}
