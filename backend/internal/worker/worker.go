package worker

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/rivalpr/backend/internal/models"
	"github.com/rivalpr/backend/internal/services"
	"gorm.io/gorm"
)

const (
	systemPrompt = `You are an expert PR copywriter. Write a professional, highly personalised pitch email.
Output ONLY the email — subject line first, then the body. No preamble, commentary, or sign-off instructions.`

	workerCount = 4
	queueSize   = 100
)

// PitchJob contains all context needed to generate one pitch version asynchronously.
type PitchJob struct {
	PitchID        uuid.UUID
	VersionNumber  int
	PromptSnapshot string

	// Context fields injected into the Claude prompt.
	JournalistName string
	OutletName     string
	Niche          string
	ClientName     string
	CampaignTitle  string
	ContextText    string
	CRMScore       int // 0 = no relationship on file
}

// Worker is a goroutine pool that processes PitchJobs asynchronously.
type Worker struct {
	jobs      chan PitchJob
	db        *gorm.DB
	anthropic *services.AnthropicClient
}

// New creates a Worker. Call Start() to launch goroutines.
func New(db *gorm.DB, anthropic *services.AnthropicClient) *Worker {
	return &Worker{
		jobs:      make(chan PitchJob, queueSize),
		db:        db,
		anthropic: anthropic,
	}
}

// Start spins up workerCount goroutines. ctx cancellation drains the queue and exits.
func (w *Worker) Start(ctx context.Context) {
	for i := 0; i < workerCount; i++ {
		go w.run(ctx)
	}
	log.Printf("AI worker pool started (%d goroutines, queue capacity %d)", workerCount, queueSize)
}

// Enqueue sends a job to the pool. Returns false if the queue is full.
func (w *Worker) Enqueue(job PitchJob) bool {
	select {
	case w.jobs <- job:
		return true
	default:
		return false
	}
}

// Stop closes the job channel so goroutines exit after draining.
func (w *Worker) Stop() {
	close(w.jobs)
}

func (w *Worker) run(ctx context.Context) {
	for job := range w.jobs {
		if err := w.process(ctx, job); err != nil {
			log.Printf("worker: pitch %s v%d failed: %v", job.PitchID, job.VersionNumber, err)
		}
	}
}

func (w *Worker) process(ctx context.Context, job PitchJob) error {
	var body string

	if w.anthropic.Enabled() {
		// Build the user-facing part of the prompt with all context fields injected.
		userPrompt := fmt.Sprintf(
			"Write a personalised pitch email for journalist %s at %s (niche: %s) on behalf of client %s.",
			job.JournalistName, job.OutletName, job.Niche, job.ClientName,
		)
		if job.CampaignTitle != "" {
			userPrompt += fmt.Sprintf(" Campaign: \"%s\".", job.CampaignTitle)
		}
		if job.CRMScore > 0 {
			userPrompt += fmt.Sprintf(
				" Our relationship score with this journalist is %d/10 — adjust the tone and warmth accordingly.",
				job.CRMScore,
			)
		}
		userPrompt += fmt.Sprintf("\n\nKey context / press release excerpt:\n%s", job.ContextText)

		generated, err := w.anthropic.Generate(ctx, systemPrompt, userPrompt)
		if err != nil {
			return fmt.Errorf("anthropic: %w", err)
		}
		body = generated
	} else {
		// Clearly-labelled fallback when no API key is configured.
		body = fmt.Sprintf(
			"Subject: Story Opportunity — %s\n\nDear %s,\n\nI'm reaching out on behalf of %s with a story tailored for your %s readers at %s.\n\n%s\n\nWould you be open to a brief call this week?\n\nBest,\n[Your Name]\n\n---\n[Mock — set ANTHROPIC_API_KEY for real AI generation | v%d]",
			job.ClientName, job.JournalistName, job.ClientName, job.Niche, job.OutletName,
			job.ContextText, job.VersionNumber,
		)
	}

	version := models.PitchVersion{
		PitchID:         job.PitchID,
		VersionNumber:   job.VersionNumber,
		AIGeneratedBody: body,
		PromptSnapshot:  job.PromptSnapshot,
	}

	if err := w.db.Create(&version).Error; err != nil {
		return fmt.Errorf("db save: %w", err)
	}

	log.Printf("worker: saved pitch version %s v%d", job.PitchID, job.VersionNumber)
	return nil
}
