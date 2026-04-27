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
	baseSystemPrompt = `You are an expert PR copywriter working for a PR agency. Write a professional, highly personalised pitch email.
Rules:
- Output ONLY the email — subject line first ("Subject: ..."), then the body.
- NEVER use placeholder text such as [Your Name], [Your Title], [Company], [Link], [Date], or any bracketed/angled substitutions. Use only the real values provided.
- Sign off with the sender's real email address provided in the prompt. Do not invent a name or title.
- No preamble, meta-commentary, or instructions outside the email itself.`

	workerCount = 4
	queueSize   = 100
)

// Tone instructions appended to system prompt.
var toneInstructions = map[string]string{
	"formal":         "Write in a formal, polished, and professional tone.",
	"conversational": "Write in a conversational, warm, and approachable tone — as if writing to a colleague.",
	"urgent":         "Write with a sense of urgency and timeliness — emphasise why this story matters right now.",
	"enthusiastic":   "Write with genuine enthusiasm and energy — convey excitement about the story.",
}

// Length instructions appended to system prompt.
var lengthInstructions = map[string]string{
	"concise":  "Keep the email very concise — under 100 words total.",
	"standard": "Keep the email a standard length — 150 to 200 words.",
	"detailed": "Write a detailed, comprehensive pitch — 300 to 350 words. Include more context and supporting points.",
}

// Angle instructions appended to user prompt.
var angleInstructions = map[string]string{
	"news_hook":          "Frame the pitch around a timely news hook — tie it to a current event, trend, or breaking story.",
	"exclusive":          "Position this as an exclusive opportunity — the journalist gets first access or an exclusive story angle.",
	"follow_up":          "Write this as a follow-up to a previous conversation — reference an existing relationship and build on prior contact.",
	"thought_leadership": "Frame this as a thought leadership piece — position the client as an expert voice on the topic.",
	"event":              "Frame this as an event invitation or event-related pitch — focus on the event details and why it matters to the journalist's audience.",
}

// PitchJob contains all context needed to generate one pitch version asynchronously.
type PitchJob struct {
	PitchID          uuid.UUID
	VersionNumber    int
	PromptSnapshot   string
	GenerationParams string // JSON of generation parameters for audit

	// Context fields injected into the Gemini prompt.
	JournalistName string
	OutletName     string
	Niche          string
	ClientName     string
	CampaignTitle  string
	ContextText    string
	CRMScore       int // 0 = no relationship on file
	SenderEmail    string

	// AI generation controls.
	Tone               string
	Length             string
	Angle              string
	CustomInstructions string
	RefinementNote     string
	ReferenceBody      string // full text of the version being refined
}

// Worker is a goroutine pool that processes PitchJobs asynchronously.
type Worker struct {
	jobs   chan PitchJob
	db     *gorm.DB
	gemini *services.GeminiClient
}

// New creates a Worker. Call Start() to launch goroutines.
func New(db *gorm.DB, gemini *services.GeminiClient) *Worker {
	return &Worker{
		jobs:   make(chan PitchJob, queueSize),
		db:     db,
		gemini: gemini,
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

	if w.gemini.Enabled() {
		// Build the system prompt dynamically based on generation controls.
		sysPrompt := baseSystemPrompt
		if inst, ok := toneInstructions[job.Tone]; ok {
			sysPrompt += "\n- " + inst
		}
		if inst, ok := lengthInstructions[job.Length]; ok {
			sysPrompt += "\n- " + inst
		} else {
			// Default length instruction when none specified.
			sysPrompt += "\n- Keep the email concise (under 200 words), specific, and compelling."
		}

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
				" Our relationship score with this journalist is %d/10 — adjust tone and warmth accordingly.",
				job.CRMScore,
			)
		}

		// Apply angle instruction.
		if inst, ok := angleInstructions[job.Angle]; ok {
			userPrompt += "\n\nAngle: " + inst
		}

		userPrompt += fmt.Sprintf("\n\nKey context / press release excerpt:\n%s", job.ContextText)

		// Apply custom instructions from user.
		if job.CustomInstructions != "" {
			userPrompt += fmt.Sprintf("\n\nAdditional instructions from the sender:\n%s", job.CustomInstructions)
		}

		// Apply refinement context if refining a previous version.
		if job.ReferenceBody != "" {
			userPrompt += "\n\n--- REFINEMENT REQUEST ---"
			userPrompt += "\nBelow is a previous version of this pitch. Rewrite and improve it based on the feedback provided."
			if job.RefinementNote != "" {
				userPrompt += fmt.Sprintf("\n\nFeedback / refinement instructions:\n%s", job.RefinementNote)
			}
			userPrompt += fmt.Sprintf("\n\nPrevious version:\n%s", job.ReferenceBody)
		}

		if job.SenderEmail != "" {
			userPrompt += fmt.Sprintf("\n\nSender (sign off with this email address, no title or company needed): %s", job.SenderEmail)
		}

		generated, err := w.gemini.Generate(ctx, sysPrompt, userPrompt)
		if err != nil {
			return fmt.Errorf("gemini: %w", err)
		}
		body = generated
	} else {
		// Clearly-labelled fallback when no API key is configured.
		body = fmt.Sprintf(
			"Subject: Story Opportunity — %s\n\nDear %s,\n\nI'm reaching out on behalf of %s with a story tailored for your %s readers at %s.\n\n%s\n\nWould you be open to a brief call this week?\n\nBest,\n[Your Name]\n\n---\n[Mock — set GEMINI_API_KEY for real AI generation | v%d]",
			job.ClientName, job.JournalistName, job.ClientName, job.Niche, job.OutletName,
			job.ContextText, job.VersionNumber,
		)
	}

	version := models.PitchVersion{
		PitchID:          job.PitchID,
		VersionNumber:    job.VersionNumber,
		AIGeneratedBody:  body,
		PromptSnapshot:   job.PromptSnapshot,
		GenerationParams: job.GenerationParams,
	}

	if err := w.db.Create(&version).Error; err != nil {
		return fmt.Errorf("db save: %w", err)
	}

	log.Printf("worker: saved pitch version %s v%d", job.PitchID, job.VersionNumber)
	return nil
}
