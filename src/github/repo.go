package github

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"

	"github.com/coopstools-homebrew/I-am-zuul/src/config"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

type LoremIpsumAppender struct {
	client      *github.Client
	accessToken string
}

func NewLoremIpsumAppender(accessToken string) *LoremIpsumAppender {
	return &LoremIpsumAppender{
		accessToken: accessToken,
	}
}

var funnyCommitMessages = []string{
	"Making the code winter solstice compliant",
	"Major breaking change that may take down google",
	"Fixing quantum entanglement issues",
	"Removing excess cosmic radiation",
	"Teaching AI systems interpretive dance",
	"Aligning code with planetary movements",
	"Upgrading hamster-powered servers",
	"Implementing time travel safeguards",
	"Patching dimensional rifts",
	"Adding support for parallel universe compatibility",
	"Recalibrating the flux capacitor",
	"Debugging the space-time continuum",
	"Optimizing code for wormhole traversal",
	"Implementing unicorn-based authentication",
	"Refactoring according to astrology signs",
	"Adding support for quantum computing via rubber ducks",
	"Upgrading to Web 5.0 compliance",
	"Patching security vulnerability in space-time fabric",
	"Implementing telepathic user interface",
	"Converting codebase to ancient hieroglyphics",
	"Migrating data to cloud nine",
	"Fixing bugs reported by time travelers",
}

var loremIpsumSentences = []string{
	"Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
	"Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
	"Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris.",
	"Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.",
	"Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",
	"Sed ut perspiciatis unde omnis iste natus error sit voluptatem accusantium doloremque laudantium.",
	"Totam rem aperiam, eaque ipsa quae ab illo inventore veritatis et quasi architecto beatae vitae dicta sunt explicabo.",
	"Nemo enim ipsam voluptatem quia voluptas sit aspernatur aut odit aut fugit.",
	"Sed quia consequuntur magni dolores eos qui ratione voluptatem sequi nesciunt?",
	"At vero eos et accusamus et iusto odio dignissimos ducimus qui blanditiis praesentium voluptatum.",
	"Deleniti atque corrupti quos dolores et quas molestias excepturi sint occaecati cupiditate non provident.",
	"Quis autem vel eum iure reprehenderit qui in ea voluptate velit esse quam nihil molestiae consequatur, vel illum qui dolorem eum fugiat quo voluptas nulla pariatur?",
	"Temporibus autem quibusdam et aut officiis debitis aut rerum necessitatibus saepe eveniet ut et voluptates.",
	"Nam libero tempore, cum soluta nobis est eligendi optio cumque nihil impedit quo minus id quod maxime placeat.",
	"Itaque earum rerum hic tenetur a sapiente delectus, ut aut reiciendis voluptatibus maiores alias consequatur.",
	"Et harum quidem rerum facilis est et expedita distinctio nam libero tempore?",
	"Omnis voluptas assumenda est, omnis dolor repellendus temporibus autem quibusdam.",
	"Nisi ut aliquid ex ea commodi consequatur quis autem vel eum iure reprehenderit.",
	"Neque porro quisquam est, qui dolorem ipsum quia dolor sit amet, consectetur, adipisci velit.",
	"Ut enim ad minima veniam, quis nostrum exercitationem ullam corporis suscipit laboriosam.",
	"Quis autem vel eum iure reprehenderit qui in ea voluptate velit esse quam nihil molestiae.",
	"At vero eos et accusamus et iusto odio dignissimos ducimus qui blanditiis praesentium.",
	"Nam libero tempore, cum soluta nobis est eligendi optio cumque nihil impedit.",
	"Temporibus autem quibusdam et aut officiis debitis aut rerum necessitatibus saepe eveniet.",
	"Minim veniam nostrud exercitation!",
	"Dolor sit amet consectetur!",
}

func (l *LoremIpsumAppender) appendLoremIpsum(repo, branch, path string) error {
	repo_parts := strings.Split(repo, "/")
	repo_name, org_name := repo_parts[1], repo_parts[0]
	ctx := context.Background()

	// Initialize GitHub client with personal access token
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: l.accessToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	l.client = github.NewClient(tc)

	// Get current file content
	fileContent, _, _, err := l.client.Repositories.GetContents(ctx, org_name, repo_name, path, &github.RepositoryContentGetOptions{
		Ref: branch,
	})
	if err != nil {
		log.Printf("Failed to get file contents: %v", err)
		return err
	}

	content, err := fileContent.GetContent()
	if err != nil {
		log.Printf("Failed to decode content: %v", err)
		return err
	}

	// Add random lorem ipsum paragraph
	newParagraph := generateLoremIpsum()
	updatedContent := fmt.Sprintf("%s\n%s", content, newParagraph)

	// Create commit
	commitMessage := funnyCommitMessages[rand.Intn(len(funnyCommitMessages))]
	opts := &github.RepositoryContentFileOptions{
		Message: &commitMessage,
		Content: []byte(updatedContent),
		SHA:     fileContent.SHA,
		Branch:  &branch,
	}

	_, _, err = l.client.Repositories.UpdateFile(ctx, org_name, repo_name, path, opts)
	if err != nil {
		log.Printf("Failed to update file: %v", err)
		return err
	}

	return nil
}

func generateLoremIpsum() string {
	sentenceCount := rand.Intn(10) + 1
	paragraphs := make([]string, sentenceCount)

	for i := 0; i < sentenceCount; i++ {
		paragraphs[i] = loremIpsumSentences[rand.Intn(len(loremIpsumSentences))]
	}

	content := strings.Join(paragraphs, " ")
	//add new line every 80 characters, but do not break words
	words := append([]string{"  "}, strings.Split(content, " ")...)
	curLine, curLineLength := "", 0
	lines := []string{}
	for _, word := range words {
		if curLineLength+len(word) > 81 {
			lines = append(lines, curLine+"\n")
			curLine = " " + word
			curLineLength = len(word)
		} else {
			curLine += " " + word
			curLineLength += len(word) + 1
		}
	}
	lines = append(lines, curLine)
	return strings.Join(lines, "")
}

func HandleGenerateLoremIpsum(config *config.Config, l *LoremIpsumAppender) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		newParagraph := generateLoremIpsum()
		w.Write([]byte(newParagraph))
	}
}

func HandleAppendLoremIpsum(config *config.Config, l *LoremIpsumAppender) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := l.appendLoremIpsum(config.LoremIpsumRepo, config.LoremIpsumBranch, config.LoremIpsumPath)
		if err != nil {
			log.Printf("Failed to append lorem ipsum: %v", err)
		}
	}
}
