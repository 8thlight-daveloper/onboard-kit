package starter

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	reviewapi "onboard-kit/contracts/invitation/reviewer"
	api "onboard-kit/contracts/invitation/starter"
	"onboard-kit/internal/invitation/reviewer"
	"onboard-kit/internal/model"
)

type command struct {
	runtime  api.Runtime
	reviewer reviewapi.Reviewer
}

func New(facts api.Runtime) api.Command {
	if facts.Clock == nil {
		facts.Clock = func() model.Timestamp { return model.Timestamp(time.Now().UTC().Format(time.RFC3339Nano)) }
	}
	if facts.LookupEnvironment == nil {
		facts.LookupEnvironment = os.Getenv
	}
	if facts.WorkingDirectory == nil {
		facts.WorkingDirectory = os.Getwd
	}
	if facts.UserHomeDirectory == nil {
		facts.UserHomeDirectory = os.UserHomeDir
	}
	if facts.UserConfigDirectory == nil {
		facts.UserConfigDirectory = os.UserConfigDir
	}
	return &command{runtime: facts, reviewer: reviewer.New()}
}

type options struct {
	archive, token, root, home string
	explicitHome               bool
}

const helpText = "Usage: onboard-kit start (--invitation FILE | --resume-import TOKEN) --root DIR [--home DIR]\n\nInspect an invitation, confirm identity, review, and explicitly begin a read-only session.\nExisting work is retained or reconciled; no checks, answers, or outcomes are executed or saved.\n"

func parse(args []string) (options, error) {
	var result options
	seen := map[string]bool{}
	for i := 0; i < len(args); i++ {
		key, value, equals := strings.Cut(args[i], "=")
		if key != "--invitation" && key != "--resume-import" && key != "--root" && key != "--home" || seen[key] {
			return result, errInput
		}
		seen[key] = true
		if !equals {
			i++
			if i == len(args) || strings.HasPrefix(args[i], "--") {
				return result, errInput
			}
			value = args[i]
		}
		if value == "" {
			return result, errInput
		}
		switch key {
		case "--invitation":
			result.archive = value
		case "--resume-import":
			result.token = value
		case "--root":
			result.root = value
		case "--home":
			result.home = value
			result.explicitHome = true
		}
	}
	if result.root == "" || (result.archive == "") == (result.token == "") || result.token != "" && !hex64(result.token) {
		return result, errInput
	}
	return result, nil
}
func (c *command) paths(o options) (options, error) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		return o, errContext
	}
	if !o.explicitHome {
		o.home = c.runtime.LookupEnvironment("OBK_HOME")
		if o.home == "" {
			if runtime.GOOS == "darwin" {
				d, err := c.runtime.UserConfigDirectory()
				if err != nil || !safePath(d) {
					return o, errContext
				}
				o.home = filepath.Join(d, "obk")
			} else {
				d := c.runtime.LookupEnvironment("XDG_STATE_HOME")
				if d != "" {
					if !safePath(d) {
						return o, errContext
					}
					o.home = filepath.Join(d, "obk")
				} else {
					d, err := c.runtime.UserHomeDirectory()
					if err != nil || !safePath(d) {
						return o, errContext
					}
					o.home = filepath.Join(d, ".local/state/obk")
				}
			}
		}
	}
	if !safePath(o.root) || !safePath(o.home) || o.archive != "" && (!safePath(o.archive) || strings.HasSuffix(o.archive, "/")) {
		return o, errContext
	}
	if !filepath.IsAbs(o.root) || !filepath.IsAbs(o.home) || o.archive != "" && !filepath.IsAbs(o.archive) {
		wd, err := c.runtime.WorkingDirectory()
		if err != nil || !safePath(wd) || !filepath.IsAbs(wd) {
			return o, errContext
		}
		if !filepath.IsAbs(o.root) {
			o.root = wd + "/" + o.root
		}
		if !filepath.IsAbs(o.home) {
			o.home = wd + "/" + o.home
		}
		if o.archive != "" && !filepath.IsAbs(o.archive) {
			o.archive = wd + "/" + o.archive
		}
	}
	if !safePath(o.root) || !safePath(o.home) || o.archive != "" && !safePath(o.archive) {
		return o, errContext
	}
	return o, nil
}

type report struct {
	Status      string `json:"status"`
	Code        string `json:"code"`
	Message     string `json:"message"`
	Transaction string `json:"transaction"`
}

func reportFor(code, token string) (report, int) {
	r := report{Code: "start." + code, Transaction: token}
	exit := 1
	switch code {
	case "cancelled":
		r.Status = "cancelled"
		r.Message = "Opening cancelled before begin; no files were changed."
		exit = 0
	case "recovery-retained":
		r.Status = "retained"
		r.Message = "Existing recovery state was retained."
		exit = 0
	case "opened":
		r.Status = "opened"
		r.Message = "Opening completed and the read-only session was displayed."
		exit = 0
	case "cleaned":
		r.Status = "cleaned"
		r.Message = "Completed recovery data was cleaned; no session was displayed."
		exit = 0
	case "staging-cleaned":
		r.Status = "cleaned"
		r.Message = "Owned incomplete staging was cleaned; reopen the invitation to begin."
		exit = 0
	case "usage":
		r.Status = "refused"
		r.Message = "Use one invitation or resume token, an explicit root, and an optional home."
		exit = 2
	case "source":
		r.Status = "refused"
		r.Message = "Read a bounded regular invitation or selected local file."
	case "context":
		r.Status = "refused"
		r.Message = "Select a safe supported root, state home, and unchanged session context."
	case "review":
		r.Status = "refused"
		r.Message = "Resolve the displayed invitation review problem before beginning."
	case "input":
		r.Status = "refused"
		r.Message = "Use complete valid input lines within the displayed limits."
	case "output":
		r.Status = "refused"
		r.Message = "Required output could not be completed; opening did not begin."
	case "limit":
		r.Status = "refused"
		r.Message = "The opening or frozen display exceeds the supported limit."
	case "commit":
		r.Status = "refused"
		r.Message = "Artifacts were not published; approved empty directories or owned staging may need cleanup."
	case "recovery-required":
		r.Status = "pending"
		r.Message = "An approved opening is pending; resume its transaction before starting another."
	case "cleanup-required":
		r.Status = "opened"
		r.Message = "Opening is complete; clean its recovery state before starting another."
	case "pending":
		r.Status = "pending"
		r.Message = "Approved opening is incomplete; retain recovery data and resume this transaction."
	case "cleanup-pending":
		r.Status = "pending"
		r.Message = "Opening is incomplete and owned recovery cleanup needs repair."
	case "cleanup-opened":
		r.Status = "opened"
		r.Message = "Opening is complete, but owned recovery cleanup needs repair."
	case "display-opened":
		r.Status = "opened"
		r.Message = "Opening is complete, but the read-only session could not be fully displayed."
	case "input-opened":
		r.Status = "opened"
		r.Message = "Opening is complete; the read-only session ended with an input error."
	case "limit-opened":
		r.Status = "opened"
		r.Message = "Opening is complete; the read-only session reached its input limit."
	case "report-opened":
		r.Status = "opened"
		r.Message = "Opening is complete, but its final report could not be written."
	case "recovery-input":
		r.Message = "Existing recovery state was retained after an input error."
	case "recovery-limit":
		r.Message = "Existing recovery state was retained after reaching the input limit."
	case "recovery-output":
		r.Message = "Existing recovery state was retained; required output could not be completed."
	case "report-cleaned":
		r.Status = "cleaned"
		r.Message = "Owned incomplete staging was cleaned, but its final report could not be written."
	}
	return r, exit
}
func finish(w io.Writer, code, token, state string) int {
	r, exit := reportFor(code, token)
	if state != "" {
		r.Status = state
	}
	if writeAll(w, string(encode(r))) == nil {
		return exit
	}
	if r.Status == "opened" {
		fallback := "report-opened"
		if code == "cleanup-opened" || code == "cleanup-required" {
			fallback = code
		}
		r, _ = reportFor(fallback, token)
		_ = writeAll(w, string(encode(r)))
	} else if code == "staging-cleaned" {
		r, _ = reportFor("report-cleaned", token)
		_ = writeAll(w, string(encode(r)))
	}
	return 1
}
func errorCode(err error) string {
	switch {
	case errors.Is(err, errLimit):
		return "limit"
	case errors.Is(err, errOutput):
		return "output"
	case errors.Is(err, errInput) || errors.Is(err, errInputLimit):
		return "input"
	case errors.Is(err, errSource):
		return "source"
	default:
		return "context"
	}
}
func diagnostics(w io.Writer, items []reviewapi.Diagnostic) error {
	b := textBuilder{limit: displayLimit}
	b.diagnostics("Opening problems", items)
	if b.failed {
		return errLimit
	}
	return writeAll(w, b.String())
}
func identityDisplay(incoming reviewapi.Incoming, o options, selection contextSelection) string {
	b := textBuilder{limit: displayLimit}
	b.line("Invitation opening is explicit.")
	b.line("Organization plan and preparation may be shared. Personal records and recovery copies contain personal data.")
	b.line("Terminal output may remain in scrollback. Viewing accepts no defaults and saves no answers or outcomes.")
	b.line("Input limit: 1048576 bytes total; 65536 bytes per logical line.")
	b.field("Account", incoming.Files.Account)
	b.field("Plan", incoming.PlanID)
	b.field("Selected root", selection.root)
	b.field("State home", o.home)
	paths := targetPaths(o.home, selection.root, selection.key, incoming.Paths.Record, incoming.Paths.Preparation)
	b.field("Local plan", paths[0])
	b.field("Preparation", paths[1])
	b.field("Record", paths[3])
	b.field("Recovery", filepath.Join(o.home, "import-pending"))
	return b.String()
}
func reviewDisplay(r reviewapi.Review, o options, s contextSelection, metadataAction string) (string, error) {
	b := textBuilder{limit: displayLimit}
	b.line("Opening review — not begun.")
	b.field("Account", r.Account)
	b.field("Plan", r.PlanID)
	paths := targetPaths(o.home, s.root, s.key, r.Paths.Record, r.Paths.Preparation)
	actions := []string{string(r.Artifacts[0].Action), string(r.Artifacts[1].Action), metadataAction, string(r.Artifacts[2].Action)}
	for i, role := range roleNames() {
		b.line("Role: " + quote(role) + " Action: " + quote(actions[i]) + " Path: " + quote(paths[i]))
	}
	b.count("Other matching contexts", len(s.others))
	for _, m := range s.others {
		b.line("  - Root: " + quote(m.Root) + " Record: " + quote(m.Record))
	}
	if s.matching == nil && len(s.others) > 0 {
		b.line("This opening registers an independent local context; other contexts are not copied or moved.")
		if r.NeedsFirstOpenedAt {
			b.line("Begin creates an independent record and session here; other matching work remains untouched.")
		}
	}
	b.field("Previous comparison", r.Comparison.Reason)
	b.diagnostics("Previous comparison problems", r.Comparison.Problems)
	lists := r.Reconciliation
	b.list("New questions", lists.NewQuestions)
	b.list("Removed questions", lists.RemovedQuestions)
	b.list("Answers to review", lists.AnswersToReview)
	b.list("New activities", lists.NewSections)
	b.list("Removed activities", lists.RemovedSections)
	b.list("Different recorded definitions", lists.ChangedOutcomes)
	b.diagnostics("Incoming warnings", r.IncomingWarnings)
	b.diagnostics("Effective plan warnings", r.EffectivePlanWarnings)
	b.diagnostics("Support problems", r.PreparationProblems)
	b.diagnostics("Session problems", r.SessionProblems)
	b.flag("Support available", r.Support.Available)
	if r.NeedsFirstOpenedAt {
		b.line("First use: a new begin event will be sampled only after begin.")
	} else {
		b.line("First use: the existing event is preserved; an empty event stays unknown.")
	}
	b.line("Private recovery copies may temporarily exist in the state home and selected artifact parents.")
	b.field("Recovery", filepath.Join(o.home, "import-pending"))
	b.line("Begin enters a read-only onboarding view. It does not verify activities, run commands, or accept answers.")
	if b.failed {
		return "", errLimit
	}
	return b.String(), nil
}
func session(input *lineInput, w io.Writer, event, display string) string {
	if writeAll(w, "First opened at: "+quote(event)+"\n"+display) != nil {
		return "display-opened"
	}
	_, err := choose(input, w, "Session: read only. Choose exit:\n", "exit")
	if errors.Is(err, errOutput) {
		return "display-opened"
	}
	if errors.Is(err, errInputLimit) {
		return "limit-opened"
	}
	if err != nil {
		return "input-opened"
	}
	return "opened"
}

func (c *command) Run(args []string, input io.Reader, stdout, stderr io.Writer) int {
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		if writeAll(stdout, helpText) != nil {
			return finish(stderr, "output", "", "")
		}
		return 0
	}
	o, err := parse(args)
	if err != nil {
		return finish(stderr, "usage", "", "")
	}
	o, err = c.paths(o)
	if err != nil {
		return finish(stderr, errorCode(err), o.token, "")
	}
	root, err := openChain(o.root)
	if err != nil {
		return finish(stderr, "context", o.token, "")
	}
	defer root.close()
	if len(root.missing) > 0 {
		return finish(stderr, "context", o.token, "")
	}
	home, err := openChain(o.home)
	if err != nil {
		return finish(stderr, "context", o.token, "")
	}
	defer home.close()
	o.root = root.absolute
	o.home = home.absolute
	if len(home.missing) == 0 && !private(home.last().original) {
		return finish(stderr, "context", o.token, "")
	}
	locks, err := lockDirectories(root.last(), home.last())
	if err != nil {
		return finish(stderr, "context", o.token, "")
	}
	defer func() { locks.release() }()
	lines := &lineInput{reader: input}
	tx, err := inspectTransaction(root, home)
	if err != nil {
		return finish(stderr, errorCode(err), o.token, "")
	}
	if tx != nil {
		defer tx.close()
		token := tx.owner.Transaction
		if tx.receipt != nil {
			token = tx.receipt.Transaction
		}
		if o.token == "" {
			code := "recovery-required"
			if tx.receipt != nil {
				code = "cleanup-required"
			}
			return finish(stderr, code, token, "")
		}
		if o.token != token {
			return finish(stderr, "context", token, "")
		}
		return c.recover(tx, lines, stdout, stderr, locks)
	}
	if o.token != "" {
		return finish(stderr, "context", o.token, "")
	}
	archiveParent, err := openChain(filepath.Dir(o.archive))
	if err != nil {
		return finish(stderr, "context", "", "")
	}
	defer archiveParent.close()
	if len(archiveParent.missing) > 0 {
		return finish(stderr, "source", "", "")
	}
	source, err := readAt(archiveParent.last().fd(), filepath.Base(o.archive), 2097152, false)
	if err != nil {
		return finish(stderr, errorCode(err), "", "")
	}
	if !source.present {
		return finish(stderr, "source", "", "")
	}
	incoming, problems := c.reviewer.Inspect(source.data)
	if len(problems) > 0 {
		if err := diagnostics(stdout, problems); err != nil {
			return finish(stderr, errorCode(err), "", "")
		}
		return finish(stderr, "review", "", "")
	}
	registered, err := scanRegistry(home)
	if err != nil {
		return finish(stderr, errorCode(err), "", "")
	}
	selection, err := chooseContext(incoming, root, home, registered)
	if err != nil {
		return finish(stderr, "context", "", "")
	}
	if err := namespaceChecks(selection.root, o.home, selection.key, incoming.Paths.Record, incoming.Paths.Preparation, o.archive, make([]snapshot, 4), source, registered); err != nil {
		return finish(stderr, errorCode(err), "", "")
	}
	if err := writeAll(stdout, identityDisplay(incoming, o, selection)); err != nil {
		return finish(stderr, "output", "", "")
	}
	action, err := choose(lines, stdout, "Identity: choose confirm "+incoming.Files.Account+" or cancel:\n", "confirm "+incoming.Files.Account, "cancel")
	if err != nil {
		return finish(stderr, errorCode(err), "", "")
	}
	if action == "" || action == "cancel" {
		return finish(stderr, "cancelled", "", "")
	}
	selected, err := snapshots(root, home, selection.key, incoming.Paths.Record, incoming.Paths.Preparation)
	if err != nil {
		return finish(stderr, errorCode(err), "", "")
	}
	if err := namespaceChecks(selection.root, o.home, selection.key, incoming.Paths.Record, incoming.Paths.Preparation, o.archive, selected, source, registered); err != nil {
		return finish(stderr, errorCode(err), "", "")
	}
	local := reviewapi.Local{Plan: reviewapi.Body{Present: selected[0].present, Bytes: selected[0].data}, Preparation: reviewapi.Body{Present: selected[1].present, Bytes: selected[1].data}, Record: reviewapi.Body{Present: selected[3].present, Bytes: selected[3].data}, KnownRecordMissing: selection.matching != nil && !selected[3].present}
	reviewed, problems := c.reviewer.Review(incoming, local, reviewapi.PlanChoiceUndecided)
	if len(problems) > 0 {
		if err := diagnostics(stdout, problems); err != nil {
			return finish(stderr, errorCode(err), "", "")
		}
		return finish(stderr, "review", "", "")
	}
	if reviewed.PlanChoiceNeeded {
		if writeAll(stdout, "Local plan bytes differ from the invitation. No recency or merge is assumed.\n") != nil {
			return finish(stderr, "output", "", "")
		}
		action, err = choose(lines, stdout, "Plan copy: choose use-incoming or cancel:\n", "use-incoming", "cancel")
		if err != nil {
			return finish(stderr, errorCode(err), "", "")
		}
		if action == "" || action == "cancel" {
			return finish(stderr, "cancelled", "", "")
		}
		reviewed, problems = c.reviewer.Review(incoming, local, reviewapi.PlanChoiceUseIncoming)
		if len(problems) > 0 {
			if err := diagnostics(stdout, problems); err != nil {
				return finish(stderr, errorCode(err), "", "")
			}
			return finish(stderr, "review", "", "")
		}
	}
	if reviewed.Proposal == nil {
		return finish(stderr, "review", "", "")
	}
	proposal := *reviewed.Proposal
	display, err := freeze(proposal.Review)
	if err != nil {
		return finish(stderr, errorCode(err), "", "")
	}
	_, metadataAction, err := metadataProposal(incoming, selection, root.last().original, strings.Repeat("0", 64), selected[2])
	if err != nil {
		return finish(stderr, errorCode(err), "", "")
	}
	reviewText, err := reviewDisplay(proposal.Review, o, selection, metadataAction)
	if err != nil {
		return finish(stderr, errorCode(err), "", "")
	}
	if err := preflight(proposal.Review, selected, display); err != nil {
		return finish(stderr, errorCode(err), "", "")
	}
	if writeAll(stdout, reviewText) != nil {
		return finish(stderr, "output", "", "")
	}
	action, err = choose(lines, stdout, "Review complete. Choose begin or cancel:\n", "begin", "cancel")
	if err != nil {
		return finish(stderr, errorCode(err), "", "")
	}
	if action == "" || action == "cancel" {
		return finish(stderr, "cancelled", "", "")
	}
	recheck := func() error {
		if root.recheck() != nil || home.recheck() != nil || archiveParent.recheck() != nil {
			return errContext
		}
		now, err := readAt(archiveParent.last().fd(), filepath.Base(o.archive), 2097152, false)
		if err != nil || !sourceEqual(source, now) {
			return errContext
		}
		ss, err := snapshots(root, home, selection.key, incoming.Paths.Record, incoming.Paths.Preparation)
		if err != nil || !sameSources(selected, ss) {
			return errContext
		}
		rr, err := scanRegistry(home)
		if err != nil || !sameRegistry(registered, rr) {
			return errContext
		}
		currentSelection, err := chooseContext(incoming, root, home, rr)
		if err != nil || !sameSelection(selection, currentSelection) {
			return errContext
		}
		existing, err := inspectTransaction(root, home)
		if err != nil {
			return err
		}
		if existing != nil {
			existing.close()
			return errContext
		}
		return namespaceChecks(selection.root, o.home, selection.key, incoming.Paths.Record, incoming.Paths.Preparation, o.archive, ss, now, rr)
	}
	if err := recheck(); err != nil {
		return finish(stderr, "context", "", "")
	}
	if err := home.bootstrap(); err != nil {
		return finish(stderr, "commit", "", "")
	}
	if err := locks.release(); err != nil {
		return finish(stderr, "commit", "", "")
	}
	locks, err = lockDirectories(root.last(), home.last())
	if err != nil {
		return finish(stderr, "context", "", "")
	}
	if err := recheck(); err != nil {
		return finish(stderr, "context", "", "")
	}
	event := model.Timestamp("")
	if proposal.Review.NeedsFirstOpenedAt {
		event = c.runtime.Clock()
	}
	publication, problems := c.reviewer.Begin(proposal, event)
	if len(problems) > 0 {
		if err := diagnostics(stdout, problems); err != nil {
			return finish(stderr, errorCode(err), "", "")
		}
		return finish(stderr, "review", "", "")
	}
	var random [32]byte
	if _, err := rand.Read(random[:]); err != nil {
		return finish(stderr, "commit", "", "")
	}
	token := hex.EncodeToString(random[:])
	metadataBytes, metadataAction, err := metadataProposal(incoming, selection, root.last().original, token, selected[2])
	if err != nil {
		return finish(stderr, "commit", token, "")
	}
	p := makePending(incoming, selection, o, root.last().original, home.last().original, publication, selected, metadataBytes, metadataAction, token, display)
	tx, err = startTransaction(root, home, p)
	if tx != nil {
		defer tx.close()
	}
	if err != nil {
		code := "commit"
		if tx != nil && tx.states["pending.json"].present {
			code = "pending"
		}
		return finish(stderr, code, token, "")
	}
	if err := tx.rollForward(); err != nil {
		return finish(stderr, "pending", token, "")
	}
	if err := tx.cleanup(); err != nil {
		return finish(stderr, "cleanup-opened", token, "")
	}
	if err := tx.close(); err != nil {
		return finish(stderr, "cleanup-opened", token, "")
	}
	if locks.release() != nil || archiveParent.close() != nil || root.close() != nil || home.close() != nil {
		return finish(stderr, "cleanup-opened", token, "")
	}
	return finish(stderr, session(lines, stdout, p.FirstOpenedAt, display), token, "")
}
func preflight(review reviewapi.Review, ss []snapshot, display string) error {
	total := 0
	for _, s := range ss {
		total += len(s.data)
	}
	for _, a := range review.Artifacts {
		total += len(a.Bytes)
	}
	if total > 10485760 || total+len(display) > 18874368 {
		return errLimit
	}
	claim := emptyClaim(strings.Repeat("0", 64), strings.Repeat("0", 64))
	for i, s := range ss {
		for _, p := range s.parents {
			if !p.Present {
				g := objectGuard{Kind: "directory", Device: strings.Repeat("9", 20), Inode: strings.Repeat("9", 20), Security: security{UID: strings.Repeat("9", 20), GID: strings.Repeat("9", 20), Mode: "0700", ACLFormat: "darwin-filesec-v1"}}
				claim.Roles[i].NewParents = append(claim.Roles[i].NewParents, createdParent{Ordinal: p.Ordinal, Guard: g})
			}
		}
	}
	if len(encode(claim))+4096 > 16384 {
		return errLimit
	}
	return nil
}
func makePending(incoming reviewapi.Incoming, selection contextSelection, o options, root, home objectGuard, pub reviewapi.Publication, ss []snapshot, meta []byte, metaAction, token, display string) pending {
	p := pending{Version: 1, Transaction: token, SessionKey: selection.key, PlanIDBase64: base64.StdEncoding.EncodeToString([]byte(incoming.PlanID)), Account: incoming.Files.Account, Root: selection.root, RootGuard: root, Home: o.home, HomeGuard: home, Record: incoming.Paths.Record, Preparation: incoming.Paths.Preparation, FirstOpenedAt: string(pub.FirstOpenedAt), SessionDisplayBase64: base64.StdEncoding.EncodeToString([]byte(display)), Roles: []pendingRole{}}
	after := [][]byte{pub.Artifacts[0].Bytes, pub.Artifacts[1].Bytes, meta, pub.Artifacts[2].Bytes}
	actions := []string{string(pub.Artifacts[0].Action), string(pub.Artifacts[1].Action), metaAction, string(pub.Artifacts[2].Action)}
	for i, name := range roleNames() {
		s := ss[i]
		r := pendingRole{Role: name, Action: actions[i], BeforePresent: s.present, BeforeBase64: base64.StdEncoding.EncodeToString(s.data), BeforeGuard: s.guard, AfterBase64: base64.StdEncoding.EncodeToString(after[i]), AfterSHA256: hash(after[i]), Parents: s.parents}
		if r.Parents == nil {
			r.Parents = []parentGuard{}
		}
		if s.present {
			r.BeforeSHA256 = hash(s.data)
		}
		p.Roles = append(p.Roles, r)
	}
	return p
}
func (c *command) recover(t *transaction, input *lineInput, stdout, stderr io.Writer, locks *lockSet) int {
	token := t.owner.Transaction
	state := "retained"
	if t.pending != nil {
		token = t.pending.Transaction
		state = "pending"
	}
	if t.receipt != nil {
		token = t.receipt.Transaction
		state = "opened"
	}
	b := textBuilder{limit: displayLimit}
	prompt := ""
	accepted := ""
	if t.pending != nil {
		p := t.pending
		b.field("Pending opening", token)
		b.field("Account", p.Account)
		plan, _ := base64Bytes(p.PlanIDBase64, 1048576)
		b.field("Plan", string(plan))
		paths := targetPaths(p.Home, p.Root, p.SessionKey, p.Record, p.Preparation)
		for i, r := range p.Roles {
			b.line("Role: " + quote(r.Role) + " Action: " + quote(r.Action) + " Path: " + quote(paths[i]))
		}
		prompt = "Recovery: choose resume or cancel:\n"
		accepted = "resume"
		if t.receipt != nil {
			prompt = "Completed opening: choose clean or cancel:\n"
			accepted = "clean"
		}
	} else if t.receipt != nil {
		b.field("Completed opening", token)
		b.line("Session display is no longer staged. Cleanup will not reopen the session.")
		prompt = "Completed opening: choose clean or cancel:\n"
		accepted = "clean"
	} else {
		b.field("Incomplete staging", token)
		b.line("Reopen the invitation after cleanup.")
		prompt = "Staging: choose clean or cancel:\n"
		accepted = "clean"
	}
	if writeAll(stdout, b.String()) != nil {
		return finish(stderr, "recovery-output", token, state)
	}
	action, err := choose(input, stdout, prompt, accepted, "cancel")
	if err != nil {
		code := "recovery-input"
		if errors.Is(err, errInputLimit) {
			code = "recovery-limit"
		}
		if errors.Is(err, errOutput) {
			code = "recovery-output"
		}
		return finish(stderr, code, token, state)
	}
	if action == "" || action == "cancel" {
		return finish(stderr, "recovery-retained", token, "")
	}
	if t.pending != nil && t.receipt == nil {
		if err := t.rollForward(); err != nil {
			return finish(stderr, "pending", token, "")
		}
	}
	completed := t.receipt != nil
	var display, event string
	if t.pending != nil {
		data, _ := base64Bytes(t.pending.SessionDisplayBase64, displayLimit)
		display = string(data)
		event = t.pending.FirstOpenedAt
	}
	if err := t.cleanup(); err != nil {
		code := "cleanup-pending"
		if completed {
			code = "cleanup-opened"
		}
		return finish(stderr, code, token, "")
	}
	if err := t.close(); err != nil {
		code := "cleanup-pending"
		if completed {
			code = "cleanup-opened"
		}
		return finish(stderr, code, token, "")
	}
	if locks.release() != nil || t.root.close() != nil || t.home.close() != nil {
		code := "cleanup-pending"
		if completed {
			code = "cleanup-opened"
		}
		return finish(stderr, code, token, "")
	}
	if !completed {
		return finish(stderr, "staging-cleaned", token, "")
	}
	if display == "" {
		return finish(stderr, "cleaned", token, "")
	}
	return finish(stderr, session(input, stdout, event, display), token, "")
}
