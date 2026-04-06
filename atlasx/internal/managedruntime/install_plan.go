package managedruntime

import (
	"errors"
	"fmt"
	"net/url"
)

var (
	ErrInstallPlanVersionRequired       = errors.New("managed runtime install plan version is required")
	ErrInstallPlanChannelRequired       = errors.New("managed runtime install plan channel is required")
	ErrInstallPlanSHA256Required        = errors.New("managed runtime install plan sha256 is required")
	ErrInstallPlanArchivePathRequired   = errors.New("managed runtime install plan archive path is required")
	ErrInstallPlanBundlePathRequired    = errors.New("managed runtime install plan bundle path is required")
	ErrInstallPlanSourceURLRequired     = errors.New("managed runtime install plan source url is required")
	ErrInstallPlanSourceURLInvalid      = errors.New("managed runtime install plan source url must be https and include host")
	ErrInstallPlanFailureReasonRequired = errors.New("managed runtime install plan failure reason is required")
	ErrInstallPlanTransitionInvalid     = errors.New("managed runtime install plan transition is invalid")
)

type InstallPhase string

const (
	InstallPhasePlanned     InstallPhase = "planned"
	InstallPhaseDownloading InstallPhase = "downloading"
	InstallPhaseDownloaded  InstallPhase = "downloaded"
	InstallPhaseVerifying   InstallPhase = "verifying"
	InstallPhaseVerified    InstallPhase = "verified"
	InstallPhaseStaging     InstallPhase = "staging"
	InstallPhaseStaged      InstallPhase = "staged"
	InstallPhaseRollback    InstallPhase = "rollback"
	InstallPhaseRolledBack  InstallPhase = "rolled_back"
	InstallPhaseFailed      InstallPhase = "failed"
)

type InstallEvent string

const (
	InstallEventStartDownload  InstallEvent = "start_download"
	InstallEventFinishDownload InstallEvent = "finish_download"
	InstallEventStartVerify    InstallEvent = "start_verify"
	InstallEventFinishVerify   InstallEvent = "finish_verify"
	InstallEventStartStage     InstallEvent = "start_stage"
	InstallEventFinishStage    InstallEvent = "finish_stage"
	InstallEventFail           InstallEvent = "fail"
	InstallEventStartRollback  InstallEvent = "start_rollback"
	InstallEventFinishRollback InstallEvent = "finish_rollback"
)

type InstallPlan struct {
	Version          string       `json:"version"`
	Channel          string       `json:"channel"`
	SourceURL        string       `json:"source_url"`
	ExpectedSHA256   string       `json:"expected_sha256"`
	ArchivePath      string       `json:"archive_path"`
	StagedBundlePath string       `json:"staged_bundle_path"`
	CurrentPhase     InstallPhase `json:"current_phase"`
	LastError        string       `json:"last_error"`
}

type InstallPlanOptions struct {
	Version          string
	Channel          string
	SourceURL        string
	ExpectedSHA256   string
	ArchivePath      string
	StagedBundlePath string
}

func NewInstallPlan(opts InstallPlanOptions) (InstallPlan, error) {
	if opts.Version == "" {
		return InstallPlan{}, ErrInstallPlanVersionRequired
	}
	if opts.Channel == "" {
		return InstallPlan{}, ErrInstallPlanChannelRequired
	}
	if opts.ExpectedSHA256 == "" {
		return InstallPlan{}, ErrInstallPlanSHA256Required
	}
	if opts.ArchivePath == "" {
		return InstallPlan{}, ErrInstallPlanArchivePathRequired
	}
	if opts.StagedBundlePath == "" {
		return InstallPlan{}, ErrInstallPlanBundlePathRequired
	}
	if opts.SourceURL == "" {
		return InstallPlan{}, ErrInstallPlanSourceURLRequired
	}
	if err := validateInstallPlanURL(opts.SourceURL); err != nil {
		return InstallPlan{}, err
	}

	return InstallPlan{
		Version:          opts.Version,
		Channel:          opts.Channel,
		SourceURL:        opts.SourceURL,
		ExpectedSHA256:   opts.ExpectedSHA256,
		ArchivePath:      opts.ArchivePath,
		StagedBundlePath: opts.StagedBundlePath,
		CurrentPhase:     InstallPhasePlanned,
	}, nil
}

func AdvanceInstallPlan(plan InstallPlan, event InstallEvent, failureReason string) (InstallPlan, error) {
	switch event {
	case InstallEventStartDownload:
		return advanceInstallPhase(plan, event, failureReason, InstallPhasePlanned, InstallPhaseDownloading)
	case InstallEventFinishDownload:
		return advanceInstallPhase(plan, event, failureReason, InstallPhaseDownloading, InstallPhaseDownloaded)
	case InstallEventStartVerify:
		return advanceInstallPhase(plan, event, failureReason, InstallPhaseDownloaded, InstallPhaseVerifying)
	case InstallEventFinishVerify:
		return advanceInstallPhase(plan, event, failureReason, InstallPhaseVerifying, InstallPhaseVerified)
	case InstallEventStartStage:
		return advanceInstallPhase(plan, event, failureReason, InstallPhaseVerified, InstallPhaseStaging)
	case InstallEventFinishStage:
		return advanceInstallPhase(plan, event, failureReason, InstallPhaseStaging, InstallPhaseStaged)
	case InstallEventStartRollback:
		if !canStartRollback(plan.CurrentPhase) {
			return plan, invalidInstallTransition(plan.CurrentPhase, event)
		}
		plan.CurrentPhase = InstallPhaseRollback
		return plan, nil
	case InstallEventFinishRollback:
		return advanceInstallPhase(plan, event, failureReason, InstallPhaseRollback, InstallPhaseRolledBack)
	case InstallEventFail:
		if failureReason == "" {
			return plan, ErrInstallPlanFailureReasonRequired
		}
		if plan.CurrentPhase == InstallPhaseStaged || plan.CurrentPhase == InstallPhaseRolledBack {
			return plan, invalidInstallTransition(plan.CurrentPhase, event)
		}
		plan.CurrentPhase = InstallPhaseFailed
		plan.LastError = failureReason
		return plan, nil
	default:
		return plan, invalidInstallTransition(plan.CurrentPhase, event)
	}
}

func validateInstallPlanURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInstallPlanSourceURLInvalid, err)
	}
	if parsed.Scheme != "https" || parsed.Host == "" {
		return ErrInstallPlanSourceURLInvalid
	}
	return nil
}

func advanceInstallPhase(plan InstallPlan, event InstallEvent, failureReason string, from InstallPhase, to InstallPhase) (InstallPlan, error) {
	if failureReason != "" {
		return plan, fmt.Errorf("%w: event %s does not accept failure reason", ErrInstallPlanTransitionInvalid, event)
	}
	if plan.CurrentPhase != from {
		return plan, invalidInstallTransition(plan.CurrentPhase, event)
	}
	plan.CurrentPhase = to
	if to != InstallPhaseFailed && to != InstallPhaseRollback && to != InstallPhaseRolledBack {
		plan.LastError = ""
	}
	return plan, nil
}

func canStartRollback(phase InstallPhase) bool {
	switch phase {
	case InstallPhaseDownloading, InstallPhaseDownloaded, InstallPhaseVerifying, InstallPhaseVerified, InstallPhaseStaging, InstallPhaseFailed:
		return true
	default:
		return false
	}
}

func invalidInstallTransition(phase InstallPhase, event InstallEvent) error {
	return fmt.Errorf("%w: phase=%s event=%s", ErrInstallPlanTransitionInvalid, phase, event)
}
