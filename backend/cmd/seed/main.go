package main

import (
	"flag"
	"log"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/rivalpr/backend/internal/config"
	"github.com/rivalpr/backend/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	action := flag.String("action", "seed", "Action to perform: seed | clear")
	flag.Parse()

	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}

	switch *action {
	case "seed":
		runSeed(db, cfg)
	case "clear":
		runClear(db)
	default:
		log.Fatalf("unknown action %q — use: seed | clear", *action)
	}
}

// ── Seed ─────────────────────────────────────────────────────────────────────

func runSeed(db *gorm.DB, cfg *config.Config) {
	log.Println("Seeding demo data...")

	// ── User ──────────────────────────────────────────────────────────────────
	hash, err := bcrypt.GenerateFromPassword([]byte("Demo1234!"), cfg.BcryptCost)
	if err != nil {
		log.Fatalf("bcrypt error: %v", err)
	}
	user := models.User{
		Email:        "alice@rivalpr.com",
		PasswordHash: string(hash),
	}
	if err := db.FirstOrCreate(&user, models.User{Email: user.Email}).Error; err != nil {
		log.Fatalf("create user: %v", err)
	}
	log.Printf("  user:    %s (password: Demo1234!)", user.Email)

	// ── Outlets ───────────────────────────────────────────────────────────────
	outlets := []models.Outlet{
		{UserID: user.ID, Name: "TechCrunch", Website: "https://techcrunch.com", Country: "US"},
		{UserID: user.ID, Name: "Forbes Romania", Website: "https://forbes.ro", Country: "RO"},
		{UserID: user.ID, Name: "The Verge", Website: "https://theverge.com", Country: "US"},
	}
	for i := range outlets {
		if err := db.Where("user_id = ? AND name = ?", user.ID, outlets[i].Name).
			FirstOrCreate(&outlets[i]).Error; err != nil {
			log.Fatalf("create outlet %s: %v", outlets[i].Name, err)
		}
		log.Printf("  outlet:  %s", outlets[i].Name)
	}

	// ── Clients ───────────────────────────────────────────────────────────────
	clients := []models.Client{
		{UserID: user.ID, Name: "NovaSpark AI", Industry: "Artificial Intelligence / SaaS"},
		{UserID: user.ID, Name: "GreenRoute Logistics", Industry: "Sustainability / Logistics"},
	}
	for i := range clients {
		if err := db.Where("user_id = ? AND name = ?", user.ID, clients[i].Name).
			FirstOrCreate(&clients[i]).Error; err != nil {
			log.Fatalf("create client %s: %v", clients[i].Name, err)
		}
		log.Printf("  client:  %s", clients[i].Name)
	}

	// ── Journalists ───────────────────────────────────────────────────────────
	journalists := []models.Journalist{
		{
			UserID: user.ID, OutletID: outlets[0].ID,
			Name: "Sarah Chen", Email: "sarah.chen@techcrunch.com",
			Niche: "Enterprise SaaS",
		},
		{
			UserID: user.ID, OutletID: outlets[0].ID,
			Name: "Marcus Webb", Email: "marcus.webb@techcrunch.com",
			Niche: "AI / Machine Learning",
		},
		{
			UserID: user.ID, OutletID: outlets[1].ID,
			Name: "Irina Popescu", Email: "irina.popescu@forbes.ro",
			Niche: "Business Technology",
		},
		{
			UserID: user.ID, OutletID: outlets[2].ID,
			Name: "James Park", Email: "james.park@theverge.com",
			Niche: "Consumer Technology",
		},
	}
	for i := range journalists {
		if err := db.Where("user_id = ? AND email = ?", user.ID, journalists[i].Email).
			FirstOrCreate(&journalists[i]).Error; err != nil {
			log.Fatalf("create journalist %s: %v", journalists[i].Name, err)
		}
		log.Printf("  journalist: %s (%s)", journalists[i].Name, journalists[i].Niche)
	}

	// ── CRM Relationships ─────────────────────────────────────────────────────
	relationships := []models.CrmRelationship{
		{
			UserID: user.ID, JournalistID: journalists[0].ID,
			RelationshipScore: 8,
			PrivateNotes:      "Warm contact. Met at TechCrunch Disrupt 2025. Responds quickly. Prefers enterprise data angles and exclusive first asks.",
		},
		{
			UserID: user.ID, JournalistID: journalists[1].ID,
			RelationshipScore: 5,
			PrivateNotes:      "Cold-ish. One email exchange in 2024. Strong AI beat — worth re-engaging with a solid demo or benchmark data.",
		},
		{
			UserID: user.ID, JournalistID: journalists[2].ID,
			RelationshipScore: 9,
			PrivateNotes:      "Best contact in the book. Covered our last 3 client launches. Always give her the exclusive.",
		},
	}
	for i := range relationships {
		if err := db.Where("user_id = ? AND journalist_id = ?", user.ID, relationships[i].JournalistID).
			FirstOrCreate(&relationships[i]).Error; err != nil {
			log.Fatalf("create crm relationship: %v", err)
		}
		log.Printf("  crm: %s → score %d", journalists[i].Name, relationships[i].RelationshipScore)
	}

	// ── Campaigns ─────────────────────────────────────────────────────────────
	campaigns := []models.Campaign{
		{
			UserID:   user.ID,
			ClientID: clients[0].ID,
			Title:    "NovaSpark AI Platform — Q2 2026 Launch",
			PressReleaseText: `FOR IMMEDIATE RELEASE

NovaSpark AI today announced the general availability of its enterprise AI platform, NovaSpark Core v3.0 — enabling mid-market companies to deploy production-grade AI workflows in under 48 hours without specialised ML expertise.

KEY FACTS
• Reduces operational costs by an average of 40% within 90 days (based on 50-company beta cohort)
• Native integrations with Salesforce, SAP, and Microsoft 365
• Priced at $2,500/month for up to 500 employees; enterprise pricing available
• SOC 2 Type II certified; GDPR compliant

QUOTES
"We built NovaSpark Core for operations teams that can't afford a 6-month AI implementation cycle," said Elena Novak, CEO. "Our beta customers are seeing 3x ROI in the first quarter."

Early customers include RetailCo (500-person UK retailer, 38% logistics cost reduction) and ManuTech GmbH (German manufacturer, 52% reduction in defect-detection time).

NovaSpark AI was founded in 2022 and is backed by $18M in Series A funding. The platform is available immediately at novaspark.ai.`,
		},
		{
			UserID:   user.ID,
			ClientID: clients[1].ID,
			Title:    "GreenRoute Carbon Offset Partnership",
			PressReleaseText: `FOR IMMEDIATE RELEASE

GreenRoute Logistics and CarbonBridge today announced a strategic partnership to offer verified carbon-neutral last-mile delivery across 12 European markets.

KEY FACTS
• 100% carbon-neutral delivery for e-commerce orders under 5kg
• Verified offsets via Gold Standard-certified reforestation projects in Romania and Portugal
• No price premium for end consumers — GreenRoute absorbs offset costs as part of its ESG commitment
• Available: Q3 2026; initial rollout in Romania, Germany, and the Netherlands

This partnership positions GreenRoute as the first logistics provider in CEE to offer certified carbon-neutral delivery at scale.`,
		},
	}
	for i := range campaigns {
		if err := db.Where("user_id = ? AND title = ?", user.ID, campaigns[i].Title).
			FirstOrCreate(&campaigns[i]).Error; err != nil {
			log.Fatalf("create campaign %s: %v", campaigns[i].Title, err)
		}
		log.Printf("  campaign: %s", campaigns[i].Title)
	}

	// ── Pitches ───────────────────────────────────────────────────────────────
	type pitchSeed struct {
		model    models.Pitch
		versions []string // one string per version body to create
	}

	pitches := []pitchSeed{
		{
			model: models.Pitch{
				UserID:       user.ID,
				ClientID:     clients[0].ID,
				CampaignID:   &campaigns[0].ID,
				JournalistID: journalists[0].ID,
				ContextBrief: "Sarah covers enterprise SaaS extensively. Lead with the 48-hour deployment angle and the ROI data. She prefers exclusive access — offer her the embargo.",
				Status:       "replied",
			},
			versions: []string{
				`Subject: Exclusive: NovaSpark AI cuts enterprise deployment from months to 48 hours

Dear Sarah,

I hope this finds you well. I'm reaching out with an exclusive on a story I think your TechCrunch Enterprise SaaS readers will find highly relevant.

NovaSpark AI is launching NovaSpark Core v3.0 today — the first enterprise AI platform that gets mid-market companies to production in under 48 hours, with no ML expertise required.

The headline number: beta customers are averaging 40% operational cost reduction within 90 days. RetailCo (UK, 500 employees) cut logistics costs by 38%; ManuTech GmbH reduced defect-detection time by 52%.

I'd love to offer you an embargo briefing with CEO Elena Novak before the public announcement goes out. She's available Thursday or Friday this week.

Would that work for you?

Best,
Alice
RivalPR Intel`,
				`Subject: Exclusive briefing: How NovaSpark AI gets enterprises to production AI in 48 hours

Dear Sarah,

Quick follow-up on NovaSpark AI's Q2 launch — I wanted to lead with the customer data this time, since I know you appreciate concrete numbers over vendor claims.

50-company beta cohort results:
• Average time-to-production: 47 hours (industry average: 4–6 months)
• 90-day ROI: 3.2x average
• Cost reduction: 40% operational costs within first quarter

The platform integrates natively with Salesforce, SAP, and Microsoft 365 — no API work required. SOC 2 Type II certified.

CEO Elena Novak is available for an exclusive 30-minute briefing this week. Embargo available until 9am ET Tuesday.

Let me know if you'd like the full data pack.

Alice`,
			},
		},
		{
			model: models.Pitch{
				UserID:       user.ID,
				ClientID:     clients[0].ID,
				CampaignID:   &campaigns[0].ID,
				JournalistID: journalists[1].ID,
				ContextBrief: "Marcus focuses on AI research and benchmarks. Lead with the technical architecture — specifically how NovaSpark handles model fine-tuning without requiring MLOps teams.",
				Status:       "sent",
			},
			versions: []string{
				`Subject: NovaSpark AI: production ML without MLOps — benchmark data inside

Hi Marcus,

I'm reaching out about NovaSpark AI's v3.0 launch — specifically because of your recent coverage of enterprise ML deployment friction.

The core technical claim: NovaSpark Core abstracts the entire MLOps stack so domain experts (not data scientists) can deploy fine-tuned models directly from business data. No Python, no infrastructure, no model serving layer to manage.

Under the hood: proprietary AutoFinetune pipeline built on top of open-weight models, with a workflow DSL that non-technical users can configure via UI.

Happy to arrange a technical deep-dive with their ML team if you'd like to stress-test the claims.

Alice`,
			},
		},
		{
			model: models.Pitch{
				UserID:       user.ID,
				ClientID:     clients[0].ID,
				CampaignID:   &campaigns[0].ID,
				JournalistID: journalists[2].ID,
				ContextBrief: "Irina covers business technology for a Romanian audience. Angle: NovaSpark has Romanian customers and is expanding into CEE. Local story with global context.",
				Status:       "draft",
			},
			versions: []string{},
		},
		{
			model: models.Pitch{
				UserID:       user.ID,
				ClientID:     clients[1].ID,
				CampaignID:   &campaigns[1].ID,
				JournalistID: journalists[3].ID,
				ContextBrief: "James covers consumer tech and sustainability at The Verge. Angle: carbon-neutral delivery as a consumer expectation, not a premium feature.",
				Status:       "draft",
			},
			versions: []string{},
		},
	}

	for i := range pitches {
		p := &pitches[i].model
		// Use a unique check: user + journalist + campaign
		query := db.Where("user_id = ? AND journalist_id = ?", p.UserID, p.JournalistID)
		if p.CampaignID != nil {
			query = query.Where("campaign_id = ?", *p.CampaignID)
		}
		if err := query.FirstOrCreate(p).Error; err != nil {
			log.Fatalf("create pitch: %v", err)
		}
		log.Printf("  pitch:  journalist=%s status=%s", journalists[i].Name, p.Status)

		// Create pitch versions
		for vNum, body := range pitches[i].versions {
			ver := models.PitchVersion{
				PitchID:         p.ID,
				VersionNumber:   vNum + 1,
				AIGeneratedBody: body,
				PromptSnapshot: generatePromptSnapshot(
					journalists[i].Name,
					journalists[i].Outlet.Name,
					journalists[i].Niche,
					clients[i%len(clients)].Name,
					campaigns[i%len(campaigns)].Title,
				),
			}
			// Pre-populate outlet name for snapshot (not preloaded)
			var outlet models.Outlet
			db.First(&outlet, journalists[i].OutletID)
			ver.PromptSnapshot = generatePromptSnapshot(
				journalists[i].Name, outlet.Name, journalists[i].Niche,
				clients[i%len(clients)].Name, campaigns[i%len(campaigns)].Title,
			)

			existing := models.PitchVersion{}
			if err := db.Where("pitch_id = ? AND version_number = ?", p.ID, ver.VersionNumber).
				FirstOrCreate(&existing, ver).Error; err != nil {
				log.Fatalf("create pitch version: %v", err)
			}
			log.Printf("    version v%d created", ver.VersionNumber)
		}
	}

	log.Println("")
	log.Println("Seed complete. Demo credentials:")
	log.Println("  Email:    alice@rivalpr.com")
	log.Println("  Password: Demo1234!")
}

func generatePromptSnapshot(journalistName, outletName, niche, clientName, campaignTitle string) string {
	return "Generate a personalized pitch email for journalist " + journalistName +
		" at " + outletName + " (niche: " + niche + ") on behalf of client " + clientName +
		". Campaign context: " + campaignTitle + "."
}

// ── Clear ─────────────────────────────────────────────────────────────────────

func runClear(db *gorm.DB) {
	log.Println("Clearing all data from the database...")

	// Truncate in FK-safe order using CASCADE.
	sql := `TRUNCATE TABLE pitch_versions, pitches, crm_relationships, campaigns, journalists, outlets, clients, users RESTART IDENTITY CASCADE`
	if err := db.Exec(sql).Error; err != nil {
		log.Fatalf("truncate failed: %v", err)
	}

	log.Println("All tables cleared. Schema preserved.")
	log.Println("Run with --action=seed to repopulate.")
}

// ptr is a convenience helper for getting a pointer to a uuid.UUID value.
func ptr(id uuid.UUID) *uuid.UUID { return &id }
