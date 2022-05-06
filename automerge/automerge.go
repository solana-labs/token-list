package main

import (
	"bufio"
	"bytes"
	"context"
	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	cueerrors "cuelang.org/go/cue/errors"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/google/go-github/v40/github"
	"github.com/solana-labs/token-list/automerge/auth"
	"github.com/solana-labs/token-list/automerge/parser"
	"github.com/sourcegraph/go-diff/diff"
	"golang.org/x/net/context/ctxhttp"
	"golang.org/x/oauth2"
	"io"
	"k8s.io/klog/v2"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

//go:embed schema.cue
var schema []byte

type knownEntry struct {
	ChainID int
	Entry   string
}

type Automerger struct {
	client     *github.Client
	owner      string
	repo       string
	cuer       *cue.Context
	cues       cue.Value
	r          *git.Repository
	tl         parser.TokenList
	fs         billy.Filesystem
	dryRun     bool
	knownAddrs map[knownEntry]bool
	knownNames map[knownEntry]bool
}

const (
	tokenlistPath = "src/tokens/solana.tokenlist.json"
	appId         = int64(152533)
)

type ErrInvalidSchema error
type ErrManualReviewNeeded error

func loadCueSchema(r *cue.Context, schema []byte, topLevel string) (*cue.Value, error) {
	v := r.CompileBytes(schema)
	if v.Err() != nil {
		return nil, v.Err()
	}
	v = v.LookupPath(cue.MakePath(cue.Def(topLevel)))
	if v.Err() != nil {
		return nil, v.Err()
	}
	return &v, nil
}

func NewAutomerger(owner string, repo string, token string, dryRun bool) *Automerger {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)

	tc := oauth2.NewClient(context.Background(), ts)

	r := cuecontext.New()
	s, err := loadCueSchema(r, schema, "StrictTokenInfo")
	if err != nil {
		panic(err)
	}

	return &Automerger{
		client:     github.NewClient(tc),
		owner:      owner,
		repo:       repo,
		cuer:       r,
		cues:       *s,
		dryRun:     dryRun,
		knownAddrs: map[knownEntry]bool{},
		knownNames: map[knownEntry]bool{},
	}
}

func (m *Automerger) GetCurrentUser(ctx context.Context) (*github.User, error) {
	acc, _, err := m.client.Users.Get(ctx, "")
	if err != nil {
		return nil, err
	}

	return acc, nil
}

func (m *Automerger) GetOpenPRs(ctx context.Context, max int) ([]*github.PullRequest, error) {
	// paginate through all open PRs
	var allPRs []*github.PullRequest
	opt := &github.PullRequestListOptions{
		State: "open",
	}
	for {
		klog.V(1).Infof("page %d", opt.Page)
		prs, resp, err := m.client.PullRequests.List(ctx, m.owner, m.repo, opt)
		if err != nil {
			return nil, err
		}
		allPRs = append(allPRs, prs...)
		if max > 0 && len(allPRs) >= max {
			break
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allPRs, nil
}

func (m *Automerger) InitRepo() error {
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %v", err)
	}

	var fs billy.Filesystem
	fs = memfs.New()

	r, err := git.Clone(memory.NewStorage(), fs, &git.CloneOptions{
		URL:      "file://" + pwd,
		Progress: os.Stderr,
		Depth:    1,
	})
	if err != nil {
		return err
	}

	m.r = r
	m.fs = fs
	return nil
}

func (m *Automerger) InitTokenlist() error {
	f, err := m.fs.Open(tokenlistPath)
	if err != nil {
		return fmt.Errorf("failed to open tokenlist: %v", err)
	}
	defer f.Close()

	var tl parser.TokenList
	dec := json.NewDecoder(f)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&tl); err != nil {
		return fmt.Errorf("failed to decode tokenlist: %v", err)
	}

	m.tl = tl

	head, err := m.r.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD: %v", err)
	}

	for _, t := range tl.Tokens {
		m.storeKnownToken(&t)
	}

	klog.Infof("current tokenlist loaded from %s (%s) with %d tokens",
		head.Hash(),
		head.Name(),
		len(m.tl.Tokens))

	return nil
}

func (m *Automerger) storeKnownToken(t *parser.Token) {
	m.knownAddrs[knownEntry{t.ChainId, t.Address}] = true
	m.knownNames[knownEntry{t.ChainId, strings.ToLower(t.Name)}] = true
}

func (m *Automerger) IsBlacklistedToken(t *parser.Token) error {
	// see #10163: user is spamming token-list by listing individual NFTs
	// TODO: move blacklist to list
	if strings.HasPrefix(t.Name, "SOLKITTY NFT") {
		return fmt.Errorf("name %s blacklisted; NFT in fungible token repo", t.Name)
	}
	return nil
}

func (m *Automerger) IsKnownToken(t *parser.Token) error {
	if _, ok := m.knownAddrs[knownEntry{t.ChainId, t.Address}]; ok {
		return fmt.Errorf("token address %s is already used", t.Address)
	}
	if _, ok := m.knownNames[knownEntry{t.ChainId, strings.ToLower(t.Name)}]; ok {
		return fmt.Errorf("token name %s is already used", t.Name)
	}
	return nil
}

func (m *Automerger) ProcessPR(ctx context.Context, pr *github.PullRequest) error {
	klog.Infof("processing PR %s", pr.GetHTMLURL())

	var hasErrorLabel bool
	for _, l := range pr.Labels {
		if l.GetName() == "automerge-error" {
			hasErrorLabel = true
		}
	}

	if hasErrorLabel {
		lastRun, err := m.getLastCheckTimestamp(ctx, pr)
		if err != nil {
			return fmt.Errorf("failed to get last check timestamp: %v", err)
		}

		lastChange := pr.GetUpdatedAt()
		if lastRun != nil && lastRun.After(lastChange.Add(-15*time.Second)) {
			klog.Infof("last check was after last change, skipping")
			return nil
		}
	}

	// Get diff
	d, resp, err := m.client.PullRequests.GetRaw(ctx, m.owner, m.repo, pr.GetNumber(),
		github.RawOptions{Type: github.Diff})
	if err != nil {
		return fmt.Errorf("failed to get diff: %w", err)
	}

	klog.V(1).Infof("rate limit: %v", resp.Rate)

	md, err := diff.ParseMultiFileDiff([]byte(d))
	if err != nil {
		return fmt.Errorf("failed to parse diff: %w", err)
	}

	assets, tl, err := m.parseDiff(md)
	if err != nil {
		klog.Warningf("failed to parse diff: %v", err)
		if err := m.reportError(ctx, pr, err); err != nil {
			return fmt.Errorf("failed to report error: %v", err)
		}
		return nil
	}

	tt, err := m.processTokenlist(ctx, tl, assets)
	if err != nil {
		klog.Warningf("failed to process tokenlist: %v", err)
		if err := m.reportError(ctx, pr, err); err != nil {
			return fmt.Errorf("failed to report error: %v", err)
		}
		return nil
	}

	if err := m.commitTokenDiff(tt, pr, assets); err != nil {
		klog.Warningf("failed to commit token diff: %v", err)
		if err := m.reportError(ctx, pr, err); err != nil {
			return fmt.Errorf("failed to report error: %v", err)
		}
		return nil
	}

	if err := m.reportSuccess(ctx, pr); err != nil {
		return fmt.Errorf("failed to report success: %v", err)
	}

	return nil
}

func (m *Automerger) Push(remote string, remoteHead string, force bool) error {
	klog.Infof("pushing to %s/%s (force: %v)", remote, remoteHead, force)

	if force && (remoteHead == "main" || remoteHead == "master") {
		return fmt.Errorf("refusing to force push to main branch")
	}

	// get current HEAD
	head, err := m.r.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD: %v", err)
	}

	refspec := fmt.Sprintf("%s:refs/heads/%s", head.Name(), remoteHead)
	klog.V(1).Infof("refspec: %s", refspec)

	if err := m.r.Push(&git.PushOptions{
		RemoteName: remote,
		Progress:   os.Stderr,
		RefSpecs: []config.RefSpec{
			config.RefSpec(refspec),
		},
		Force: force,
	}); err != nil {
		return fmt.Errorf("failed to push: %v", err)
	}

	return nil
}

func (m *Automerger) parseDiff(md []*diff.FileDiff) ([]string, *diff.FileDiff, error) {
	assets := make([]string, 0)
	var tlDiff *diff.FileDiff

	for _, z := range md {
		newFile := strings.TrimPrefix(z.NewName, "b/")
		klog.V(1).Infof("found file: %s", newFile)

		switch {
		case strings.HasPrefix(newFile, "assets/"):
			if z.OrigName != "/dev/null" {
				return nil, nil, fmt.Errorf("found modified asset file %s - only new assets are allowed", newFile)
			}
			p := strings.Split(newFile, "/")
			if len(p) != 4 || p[1] != "mainnet" {
				return nil, nil, fmt.Errorf("invalid asset path: %s", newFile)
			}

			switch path.Ext(p[3]) {
			case ".png", ".jpg", ".svg", ".PNG", ".JPG", ".SVG":
			default:
				return nil, nil, fmt.Errorf("invalid asset extension: %s (wants png, jpg, svg)", newFile)
			}

			assets = append(assets, newFile)
		case newFile == "src/tokens/solana.tokenlist.json":
			if tlDiff != nil {
				return nil, nil, fmt.Errorf("found multiple tokenlist diffs")
			}
			tlDiff = z
			klog.V(1).Infof("found solana.tokenlist.json")
		case newFile == "CHANGELOG.md" || newFile == "package.json":
			klog.V(1).Infof("ignoring spurious %s change", newFile)
			continue
		default:
			// Unknown file modified - fail
			return nil, nil, fmt.Errorf("unsupported file modified: %s", newFile)
		}
	}

	if tlDiff == nil {
		return nil, nil, fmt.Errorf("no tokenlist diff found")
	}

	return assets, tlDiff, nil
}

func (m *Automerger) commitTokenDiff(tt []parser.Token, pr *github.PullRequest, assets []string) error {
	klog.Infof("committing change for %d", pr.GetNumber())

	tl := m.tl
	tl.Tokens = append(m.tl.Tokens, tt...)

	// marshal tl
	b, err := json.MarshalIndent(tl, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tokenlist: %v", err)
	}

	w, err := m.r.Worktree()
	if err != nil {
		panic(err)
	}

	// request and write out assets to file
	for _, asset := range assets {
		uri := fmt.Sprintf(`https://raw.githubusercontent.com/solana-labs/token-list/%s/%s`, pr.GetHead().GetSHA(), asset)
		klog.V(1).Infof("downloading asset %s", uri)
		resp, err := http.Get(uri)
		if err != nil {
			resp.Body.Close()
			return fmt.Errorf("failed to get asset: %v", err)
		}

		// fail if resp is larger than 200KiB
		if resp.ContentLength > 200*1024 {
			resp.Body.Close()
			return fmt.Errorf("asset too large: %s is %d KiB (must be less than 200 KiB)", asset, resp.ContentLength/1024)
		}

		if err := m.fs.MkdirAll(path.Dir(asset), 0755); err != nil {
			resp.Body.Close()
			return fmt.Errorf("failed to create asset directory: %v", err)
		}
		f, err := m.fs.Create(asset)
		if err != nil {
			resp.Body.Close()
			return fmt.Errorf("failed to create asset file: %v", err)
		}
		if _, err := io.Copy(f, resp.Body); err != nil {
			resp.Body.Close()
			return fmt.Errorf("failed to write asset file: %v", err)
		}
		resp.Body.Close()
		f.Close()
		w.Add(asset)
	}

	// write to file
	f, err := m.fs.Create(tokenlistPath)
	if err != nil {
		return fmt.Errorf("failed to create tokenlist file: %v", err)
	}
	defer f.Close()

	if _, err := f.Write(b); err != nil {
		return fmt.Errorf("failed to write tokenlist file: %v", err)
	}

	title := pr.GetTitle()
	if title == "" {
		title = fmt.Sprintf("Merge #%d", pr.GetNumber())
	}

	_, err = w.Add(tokenlistPath)
	if err != nil {
		return fmt.Errorf("failed to add tokenlist file: %v", err)
	}

	user := pr.GetUser()
	if user == nil || user.Login == nil {
		return fmt.Errorf("failed to get user")
	}

	author := &object.Signature{
		Name:  *user.Login,
		Email: fmt.Sprintf("%s@users.noreply.github.com", *user.Login),
		When:  *pr.UpdatedAt,
	}
	h, err := w.Commit(
		fmt.Sprintf("%s\n\nCloses #%d", title, pr.GetNumber()),
		&git.CommitOptions{
			Author:    author,
			Committer: author,
		})
	if err != nil {
		return fmt.Errorf("failed to commit: %v", err)
	}

	m.tl = tl
	for _, t := range tt {
		m.storeKnownToken(&t)
	}

	klog.V(1).Infof("committed %s (%s)", h, title)

	return nil
}

func (m *Automerger) reportError(ctx context.Context, pr *github.PullRequest, err error) error {
	if m.dryRun {
		klog.V(1).Infof("dry run: not marking automerge check as failed")
		return nil
	}

	ts := &github.Timestamp{Time: time.Now()}
	msg := err.Error()
	title := "Not eligible for automerge (click details for more information)"
	conclFailure := "failure"
	statusCompleted := "completed"

	if _, _, err := m.client.Checks.CreateCheckRun(ctx, m.owner, m.repo, github.CreateCheckRunOptions{
		Name:        "New automerge",
		HeadSHA:     pr.Head.GetSHA(),
		Status:      &statusCompleted,
		Conclusion:  &conclFailure,
		StartedAt:   ts,
		CompletedAt: ts,
		Output: &github.CheckRunOutput{
			Title:   &title,
			Summary: &msg,
		},
	}); err != nil {
		return fmt.Errorf("failed to create check run: %v", err)
	}

	if err := m.markOldChecksCompleted(ctx, pr); err != nil {
		return err
	}

	// Remove automerge label
	if resp, err := m.client.Issues.RemoveLabelForIssue(ctx, m.owner, m.repo, pr.GetNumber(), "automerge"); err != nil {
		if resp.StatusCode != 404 {
			return fmt.Errorf("failed to remove automerge label: %v", err)
		}
	}
	// Add automerge-error label
	if _, _, err := m.client.Issues.AddLabelsToIssue(ctx, m.owner, m.repo, pr.GetNumber(), []string{"automerge-error"}); err != nil {
		return fmt.Errorf("failed to add automerge-error label: %v", err)
	}

	return nil
}

func (m *Automerger) reportSuccess(ctx context.Context, pr *github.PullRequest) error {
	if m.dryRun {
		klog.V(1).Infof("dry run: not marking automerge check as completed")
		return nil
	}

	ts := &github.Timestamp{Time: time.Now()}
	title := "Eligible for automerge \U0001F973"
	msg := "Your PR will be auto-merged shortly"
	conclSuccess := "success"
	statusCompleted := "completed"

	if _, _, err := m.client.Checks.CreateCheckRun(ctx, m.owner, m.repo, github.CreateCheckRunOptions{
		Name:        "New automerge",
		HeadSHA:     pr.Head.GetSHA(),
		Status:      &statusCompleted,
		Conclusion:  &conclSuccess,
		StartedAt:   ts,
		CompletedAt: ts,
		Output: &github.CheckRunOutput{
			Title:   &title,
			Summary: &msg,
		},
	}); err != nil {
		return fmt.Errorf("failed to create check run: %v", err)
	}

	if err := m.markOldChecksCompleted(ctx, pr); err != nil {
		return err
	}

	// Remove automerge-error label
	if resp, err := m.client.Issues.RemoveLabelForIssue(ctx, m.owner, m.repo, pr.GetNumber(), "automerge-error"); err != nil {
		if resp.StatusCode != 404 {
			return fmt.Errorf("failed to remove automerge-error label: %v", err)
		}
	}
	// Add automerge label
	if _, _, err := m.client.Issues.AddLabelsToIssue(ctx, m.owner, m.repo, pr.GetNumber(), []string{"automerge"}); err != nil {
		return fmt.Errorf("failed to add automerge label: %v", err)
	}

	return nil
}

func (m *Automerger) markOldChecksCompleted(ctx context.Context, pr *github.PullRequest) error {
	ts := &github.Timestamp{Time: time.Now()}
	titleSuperseded := "Disregard - please see new automerge"
	conclNeutral := "neutral"
	statusCompleted := "completed"

	list, _, err := m.client.Checks.ListCheckRunsForRef(ctx, m.owner, m.repo, pr.Head.GetSHA(), nil)
	if err != nil {
		return fmt.Errorf("failed to list check runs: %v", err)
	}

	for _, c := range list.CheckRuns {
		name := c.GetName()
		if name == "Automatic merge" ||
			name == "Bulletproof Automerge" ||
			name == "auto-merge" {
			if _, _, err := m.client.Checks.CreateCheckRun(ctx, m.owner, m.repo, github.CreateCheckRunOptions{
				Name:        name,
				HeadSHA:     pr.Head.GetSHA(),
				Status:      &statusCompleted,
				Conclusion:  &conclNeutral,
				StartedAt:   ts,
				CompletedAt: ts,
				Output: &github.CheckRunOutput{
					Title:   &titleSuperseded,
					Summary: &titleSuperseded,
				},
			}); err != nil {
				return fmt.Errorf("failed to update check run: %v", err)
			}
		}
	}
	return nil
}

func (m *Automerger) getLastCheckTimestamp(ctx context.Context, pr *github.PullRequest) (*time.Time, error) {
	list, _, err := m.client.Checks.ListCheckRunsForRef(ctx, m.owner, m.repo, pr.Head.GetSHA(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list check runs: %v", err)
	}

	for _, c := range list.CheckRuns {
		if c.GetName() == "New automerge" && *c.GetApp().ID == appId {
			t := c.GetCompletedAt().Time
			return &t, nil
		}
	}
	return nil, nil
}

func (m *Automerger) processTokenlist(ctx context.Context, d *diff.FileDiff, assets []string) ([]parser.Token, error) {
	assetAddrs := make([]string, len(assets))
	for i, a := range assets {
		assetAddrs[i] = strings.Split(a, "/")[2]
	}

	// log assets
	klog.V(1).Infof("found %d image assets", len(assetAddrs))
	for _, a := range assetAddrs {
		klog.V(1).Infof("  %s", a)
	}

	var res []parser.Token

	knownAddrs := map[knownEntry]bool{}
	knownNames := map[knownEntry]bool{}
	for _, h := range d.Hunks {
		body := string(h.Body)
		body = strings.Trim(body, "\n")

		var plain bytes.Buffer

		// Extract added lines
		scanner := bufio.NewScanner(strings.NewReader(body))
		i := 1
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "-") {
				// exception - ignore single control character deletions
				// (harmless, typically introduced by JSON reformatting)
				content := strings.Replace(strings.TrimPrefix(line, "-"), " ", "", -1)
				if content == "{" || content == "}" || content == "[" || content == "]" {
					klog.V(1).Infof("ignoring deletion of line %s (%s)", line, content)
					continue
				}

				return nil, fmt.Errorf("found removed line: %s", line)
			}
			if strings.HasPrefix(line, "!") {
				return nil, fmt.Errorf("found modified line: %s", line)
			}
			if strings.HasPrefix(line, " ") {
				continue
			}
			if !strings.HasPrefix(line, "+") {
				return nil, fmt.Errorf("unknown diff op: %s", line)
			}
			line = strings.TrimPrefix(line, "+")
			klog.V(2).Infof("ADD: %d: %s", i, line)
			plain.Write([]byte(line + "\n"))
			i++
		}

		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("failed to scan hunks: %w", err)
		}

		s := plain.String()
		tt, err := parser.NormalizeWhatever(s)
		if err != nil {
			return nil, fmt.Errorf("failed to normalize: %w", err)
		}

		for _, t := range tt {
			if knownAddrs[knownEntry{t.ChainId, t.Address}] {
				return nil, fmt.Errorf("duplicate address within PR")
			}
			if knownNames[knownEntry{t.ChainId, strings.ToLower(t.Name)}] {
				return nil, fmt.Errorf("duplicate name within PR")
			}
			knownAddrs[knownEntry{t.ChainId, t.Address}] = true
			knownNames[knownEntry{t.ChainId, strings.ToLower(t.Name)}] = true

			if err := m.IsKnownToken(&t); err != nil {
				return nil, fmt.Errorf("duplicate token: %v", err)
			}
			if err := m.IsBlacklistedToken(&t); err != nil {
				return nil, fmt.Errorf("blacklisted token: %v", err)
			}

			v := m.cuer.Encode(t)
			if v.Err() != nil {
				return nil, fmt.Errorf("error encoding to Cue: %v", v.Err())
			}

			u := v.Unify(m.cues)
			if v.Err() != nil {
				return nil, fmt.Errorf("failed to unify with schema: %v", err)
			}

			if err := u.Validate(cue.Final(), cue.Concrete(true)); err != nil {
				// Print last error encountered (which is usually the regex conflict).
				// The full list of errors may be confusing to users who do not understand Cue unification.
				errs := cueerrors.Errors(u.Err())
				last := errs[len(errs)-1]
				return nil, fmt.Errorf("error validating schema: %v", last.Error())
			}

			if strings.Trim(t.Name, " ") == "" {
				return nil, fmt.Errorf("empty token name for %v", t)
			}

			if err := verifyLogoURI(t.LogoURI, assets); err != nil {
				return nil, fmt.Errorf("failed verifying image URI (make sure the filename matches your JSON!): %v", err)
			}

			if id, ok := t.Extensions["coingeckoId"]; ok {
				if err := verifyCoingeckoId(id); err != nil {
					return nil, fmt.Errorf("failed to verify coingecko ID: %v", err)
				}
			}

			if website, ok := t.Extensions["website"]; ok {
				if err := tryHEADRequest(website); err != nil {
					return nil, fmt.Errorf("failed to verify website: %s: %v", website, err)
				}
				// see #10163: user is spamming token-list by listing individual NFTs
				// TODO: move blacklist to list
				if strings.HasPrefix(t.Extensions["website"], "https://solkitty.io/nft") {
					return nil, fmt.Errorf("blacklisted solkitty: %s:", website)
				}
			}

			if twitter, ok := t.Extensions["twitter"]; ok {
				if err := verifyTwitterHandle(twitter); err != nil {
					return nil, fmt.Errorf("failed to verify Twitter handle: %s: %v", twitter, err)
				}
			}

			klog.V(1).Infof("found valid JSON for token %s", t.Name)
		}

		for _, asset := range assetAddrs {
			var found bool
			for _, t := range tt {
				if t.Address == asset {
					found = true
				}
			}
			if !found {
				return nil, fmt.Errorf("asset file for unknown token found: %s", asset)
			}
		}

		res = append(res, tt...)
	}

	return res, nil
}

func verifyLogoURI(uri string, file []string) error {
	prefix := "https://raw.githubusercontent.com/solana-labs/token-list/main/"
	if strings.HasPrefix(uri, prefix+"assets/") {
		// When a local asset is specified, check if it's part of the PR before checking remotely
		for _, f := range file {
			if uri == (prefix + f) {
				return nil
			}
		}
	}

	klog.V(1).Infof("verifying external image URI %s", uri)

	err2 := tryHEADRequest(uri)
	if err2 != nil {
		return err2
	}

	return nil
}

func tryHEADRequest(uri string) error {
	// Send HEAD request to verify external URIs
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	resp, err := ctxhttp.Head(ctx, &http.Client{
		Timeout: 5 * time.Second,
	}, uri)
	if err != nil {
		return fmt.Errorf("failed to verify %s using HEAD request: %v", uri, err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("non-200 response code for URL %s: %d", uri, resp.StatusCode)
	}

	return nil
}

func verifyCoingeckoId(id string) error {
	uri := "https://www.coingecko.com/en/coins/" + id

	klog.V(1).Infof("verifying coingeckoId %s", uri)

	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	resp, err := ctxhttp.Head(ctx, &http.Client{
		Timeout: 5 * time.Second,
	}, uri)
	if err != nil {
		return fmt.Errorf("failed to verify %s using HEAD request: %v", uri, err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("non-200 response code for URL %s: %d", uri, resp.StatusCode)
	}

	return nil

}

func verifyTwitterHandle(uri string) error {
	return nil
}

var (
	flagDryRun    = flag.Bool("dryRun", false, "Simulate only")
	flagMax       = flag.Int("max", 0, "Maximum number of tokens to process")
	flagSetRemote = flag.Bool("setRemoteForCI", false, "[FOR CI] add app origin to local repo")
)

// Helper function for GitHub actions, allowing it to push to main using
// the app's identity rather than GITHUB_TOKEN, allowing subsequent
// workflows to trigger.
func configureLocalGitRemoteToken(token string) {
	// git remote add --push origin https://your_username:${token}@github.com/solana-labs/token-list.git
	remote := fmt.Sprintf("https://token-list-automerger:%s@github.com/solana-labs/token-list.git", token)
	cmd := exec.Command("git", "remote", "add", "app", remote)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		klog.Fatalf("failed to add remote: %v", err)
	}
}

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	var token string
	if t := os.Getenv("GITHUB_TOKEN"); t != "" {
		token = t
	} else if b := os.Getenv("GITHUB_APP_PEM"); b != "" {
		key, err := base64.StdEncoding.DecodeString(b)
		if err != nil {
			klog.Exitf("failed to decode GITHUB_APP_PEM as base64: %v", err)
		}

		t, err := auth.GetInstallationToken([]byte(key), appId, "solana-labs")
		if err != nil {
			klog.Exitf("failed to get installation token: %v", err)
		}
		token = t
	} else {
		klog.Exit("GITHUB_TOKEN or GITHUB_APP_PEM environment variable is not set")
	}

	if *flagSetRemote {
		configureLocalGitRemoteToken(token)
	}

	klog.Info("starting automerge")

	m := NewAutomerger("solana-labs", "token-list", token, *flagDryRun)

	klog.Info("initializing virtual Git worktree")
	if err := m.InitRepo(); err != nil {
		klog.Exitf("failed to initialize virtual Git worktree: %v", err)
	}

	klog.Info("loading tokenlist from worktree")
	if err := m.InitTokenlist(); err != nil {
		klog.Exitf("failed to load tokenlist from worktree: %v", err)
	}

	user, err := m.GetCurrentUser(context.TODO())
	if err != nil {
		klog.Warningf("failed to get current user: %v", err)
	}
	klog.Infof("running as: %s", user.GetLogin())

	r, err := m.GetOpenPRs(context.TODO(), *flagMax)
	if err != nil {
		klog.Errorf("error getting open prs: %v", err)
		return
	}

	klog.Infof("processing %d PRs", len(r))

	i := 0
	for _, pr := range r {
		err := m.ProcessPR(context.TODO(), pr)
		if err != nil {
			klog.Warningf("error processing pr, retrying: %v", err)
			// retry once
			err = m.ProcessPR(context.TODO(), pr)
			if err != nil {
				klog.Exitf("error processing pr: %v", err)
			}
		}

		i++
		if *flagMax > 0 && i >= *flagMax {
			break
		}
	}

	klog.Info("pushing")
	if err := m.Push("origin", "automerge-pending", true); err != nil {
		klog.Exitf("failed to push: %v", err)
	}

	klog.Info("done")
}
