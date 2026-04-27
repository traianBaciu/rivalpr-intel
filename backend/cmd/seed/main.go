package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"

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
	log.Println("Seeding demo data (large-scale)…")
	rng := rand.New(rand.NewSource(42)) // deterministic

	// ── Users ─────────────────────────────────────────────────────────────────
	userDefs := []struct{ name, email string }{
		{"Alice Morgan", "alice@rivalpr.com"},
		{"Bob Singh", "bob@rivalpr.com"},
		{"Carol Tanaka", "carol@rivalpr.com"},
		{"Dave Okoye", "dave@rivalpr.com"},
		{"Eve Lindström", "eve@rivalpr.com"},
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("Demo1234!"), cfg.BcryptCost)
	if err != nil {
		log.Fatalf("bcrypt: %v", err)
	}
	users := make([]models.User, len(userDefs))
	for i, u := range userDefs {
		users[i] = models.User{Email: u.email, PasswordHash: string(hash)}
		if err := db.Where("email = ?", u.email).FirstOrCreate(&users[i]).Error; err != nil {
			log.Fatalf("create user %s: %v", u.email, err)
		}
		log.Printf("  user: %s (%s)", u.name, u.email)
	}

	// ── Outlets (30 agency-wide) ──────────────────────────────────────────────
	outletDefs := []struct{ name, website, country string }{
		{"TechCrunch", "https://techcrunch.com", "US"},
		{"The Verge", "https://theverge.com", "US"},
		{"Wired", "https://wired.com", "US"},
		{"Forbes", "https://forbes.com", "US"},
		{"Bloomberg Technology", "https://bloomberg.com/technology", "US"},
		{"Fast Company", "https://fastcompany.com", "US"},
		{"VentureBeat", "https://venturebeat.com", "US"},
		{"Ars Technica", "https://arstechnica.com", "US"},
		{"MIT Technology Review", "https://technologyreview.com", "US"},
		{"The Information", "https://theinformation.com", "US"},
		{"Forbes Romania", "https://forbes.ro", "RO"},
		{"Ziarul Financiar", "https://zf.ro", "RO"},
		{"G4Media", "https://g4media.ro", "RO"},
		{"Der Spiegel", "https://spiegel.de", "DE"},
		{"Handelsblatt", "https://handelsblatt.com", "DE"},
		{"t3n", "https://t3n.de", "DE"},
		{"Le Monde", "https://lemonde.fr", "FR"},
		{"Challenges", "https://challenges.fr", "FR"},
		{"El País", "https://elpais.com", "ES"},
		{"Expansión", "https://expansion.com", "ES"},
		{"The Guardian", "https://theguardian.com", "UK"},
		{"Financial Times", "https://ft.com", "UK"},
		{"The Times", "https://thetimes.co.uk", "UK"},
		{"CNBC", "https://cnbc.com", "US"},
		{"Reuters", "https://reuters.com", "UK"},
		{"AP News", "https://apnews.com", "US"},
		{"Protocol", "https://protocol.com", "US"},
		{"Axios", "https://axios.com", "US"},
		{"The Economist", "https://economist.com", "UK"},
		{"Business Insider", "https://businessinsider.com", "US"},
	}
	outlets := make([]models.Outlet, len(outletDefs))
	for i, o := range outletDefs {
		addedBy := &users[i%len(users)].ID
		outlets[i] = models.Outlet{
			AddedBy: addedBy,
			Name:    o.name,
			Website: o.website,
			Country: o.country,
		}
		if err := db.Where("name = ?", o.name).FirstOrCreate(&outlets[i]).Error; err != nil {
			log.Fatalf("create outlet %s: %v", o.name, err)
		}
	}
	log.Printf("  outlets: %d created/found", len(outlets))

	// ── Journalists (200 agency-wide) ─────────────────────────────────────────
	niches := []string{
		"Enterprise SaaS", "AI / Machine Learning", "Cybersecurity", "FinTech",
		"Consumer Technology", "Climate Tech", "Business Technology", "Startup / VC",
		"Health Tech", "E-commerce", "Future of Work", "Sustainability",
		"Data & Analytics", "Developer Tools", "Cloud Infrastructure", "Mobility Tech",
	}
	firstNames := []string{
		"Sarah", "Marcus", "Irina", "James", "Emily", "Luca", "Priya", "Tomás",
		"Nadia", "Kevin", "Amelia", "Ravi", "Sophie", "Omar", "Lin", "David",
		"Elena", "Patrick", "Mei", "Carlos", "Zara", "Finn", "Aisha", "Hugo",
		"Rosa", "Sam", "Valentina", "Chris", "Fatima", "Ethan",
	}
	lastNames := []string{
		"Chen", "Webb", "Popescu", "Park", "Thornton", "Ferrari", "Sharma", "Reyes",
		"Koval", "Obi", "Walsh", "Kapoor", "Laurent", "Hassan", "Zhang", "Brooks",
		"Novak", "O'Brien", "Huang", "Vargas", "Khan", "Andersen", "Diallo", "Müller",
		"García", "Taylor", "Rossi", "Kim", "Ibrahim", "Jensen",
	}

	journalists := make([]models.Journalist, 0, 200)
	emailSet := map[string]bool{}
	for i := 0; i < 200; i++ {
		outlet := outlets[i%len(outlets)]
		first := firstNames[i%len(firstNames)]
		last := lastNames[(i*7+3)%len(lastNames)]
		niche := niches[i%len(niches)]
		addedBy := &users[i%len(users)].ID

		// Build unique email
		base := fmt.Sprintf("%s.%s@%d.rivalpr-seed.io",
			slugify(first), slugify(last), i)
		for emailSet[base] {
			base = fmt.Sprintf("%s.%s%d@%d.rivalpr-seed.io",
				slugify(first), slugify(last), rng.Intn(99), i)
		}
		emailSet[base] = true

		j := models.Journalist{
			AddedBy:  addedBy,
			OutletID: outlet.ID,
			Name:     fmt.Sprintf("%s %s", first, last),
			Email:    base,
			Niche:    niche,
		}
		if err := db.Where("email = ?", base).FirstOrCreate(&j).Error; err != nil {
			log.Fatalf("create journalist %d: %v", i, err)
		}
		journalists = append(journalists, j)
	}
	log.Printf("  journalists: %d created/found", len(journalists))

	// ── Clients (20 user-scoped) ──────────────────────────────────────────────
	clientDefs := []struct{ name, industry string }{
		{"NovaSpark AI", "Artificial Intelligence / SaaS"},
		{"GreenRoute Logistics", "Sustainability / Logistics"},
		{"QuantumPay", "FinTech / Payments"},
		{"DataMesh Labs", "Data Infrastructure"},
		{"Skyborne Mobility", "Urban Air Mobility"},
		{"CarbonTrace", "Climate Tech"},
		{"MedVault", "Health Tech / Security"},
		{"RetailOS", "E-commerce / Retail Tech"},
		{"ClearSignal Security", "Cybersecurity"},
		{"FlowOps", "Future of Work / SaaS"},
		{"NexusChain", "Supply Chain Tech"},
		{"BrightClass", "EdTech"},
		{"SolarGrid Analytics", "Energy Tech"},
		{"RoboFarm", "AgriTech"},
		{"CloudNest", "Cloud Infrastructure"},
		{"PulseCare", "Digital Health"},
		{"TradeFlow", "FinTech / Trade Finance"},
		{"UrbanSense", "Smart Cities / IoT"},
		{"DevLens", "Developer Tools"},
		{"IntelliFleet", "Mobility / Fleet Tech"},
	}
	clients := make([]models.Client, len(clientDefs))
	for i, c := range clientDefs {
		owner := users[i%len(users)]
		clients[i] = models.Client{
			UserID:   owner.ID,
			Name:     c.name,
			Industry: c.industry,
		}
		if err := db.Where("user_id = ? AND name = ?", owner.ID, c.name).
			FirstOrCreate(&clients[i]).Error; err != nil {
			log.Fatalf("create client %s: %v", c.name, err)
		}
	}
	log.Printf("  clients: %d created/found", len(clients))

	// ── CRM Relationships (~200, one per user/journalist pair) ────────────────
	crmCount := 0
	for ui, u := range users {
		// Each user gets ~40 journalists rated
		start := ui * 40
		for ji := start; ji < start+40 && ji < len(journalists); ji++ {
			score := rng.Intn(10) + 1
			notes := fmt.Sprintf("Auto-seeded note for %s. Score: %d/10.",
				journalists[ji].Name, score)
			rel := models.CrmRelationship{
				UserID:            u.ID,
				JournalistID:      journalists[ji].ID,
				RelationshipScore: score,
				PrivateNotes:      notes,
			}
			if err := db.Where("user_id = ? AND journalist_id = ?", u.ID, journalists[ji].ID).
				FirstOrCreate(&rel).Error; err != nil {
				log.Fatalf("crm: %v", err)
			}
			crmCount++
		}
	}
	log.Printf("  crm relationships: %d created/found", crmCount)

	// ── Campaigns (20) ────────────────────────────────────────────────────────
	campaigns := make([]models.Campaign, len(clients))
	for i, client := range clients {
		title := fmt.Sprintf("%s — Q%d %d Launch", client.Name, (i%4)+1, 2026)
		pressRelease := fmt.Sprintf(
			"FOR IMMEDIATE RELEASE\n\n%s today announced a major product milestone "+
				"that significantly advances the %s sector. This development is expected "+
				"to impact thousands of customers globally.\n\nThe company has seen "+
				"consistent 40%% year-over-year growth and is backed by leading investors.",
			client.Name, client.Industry,
		)
		campaigns[i] = models.Campaign{
			UserID:           client.UserID,
			ClientID:         client.ID,
			Title:            title,
			PressReleaseText: pressRelease,
		}
		if err := db.Where("user_id = ? AND title = ?", client.UserID, title).
			FirstOrCreate(&campaigns[i]).Error; err != nil {
			log.Fatalf("create campaign %s: %v", title, err)
		}
	}
	log.Printf("  campaigns: %d created/found", len(campaigns))

	// ── Pitches (1000+) and Pitch Versions (~1500) ────────────────────────────
	statuses := []string{"draft", "sent", "opened", "replied"}
	pitchCount := 0
	versionCount := 0

	type pitchKey struct {
		userID       uuid.UUID
		journalistID uuid.UUID
		campaignID   uuid.UUID
	}
	pitchSeen := map[pitchKey]bool{}

	// ~1040 pitches: each user pitches ~208 times across their journalists/campaigns
	for ui, u := range users {
		userClients := make([]models.Client, 0)
		userCampaigns := make([]models.Campaign, 0)
		for _, c := range clients {
			if c.UserID == u.ID {
				userClients = append(userClients, c)
			}
		}
		for _, cam := range campaigns {
			if cam.UserID == u.ID {
				userCampaigns = append(userCampaigns, cam)
			}
		}

		// Each user gets ~40 tracked journalists
		trackedStart := ui * 40
		for ji := trackedStart; ji < trackedStart+40 && ji < len(journalists); ji++ {
			j := journalists[ji]
			// Each journalist gets ~5 pitches from different campaigns
			numPitches := 3 + rng.Intn(4) // 3–6
			for pi := 0; pi < numPitches && pi < len(userCampaigns); pi++ {
				cam := userCampaigns[pi%len(userCampaigns)]
				key := pitchKey{u.ID, j.ID, cam.ID}
				if pitchSeen[key] {
					continue
				}
				pitchSeen[key] = true

				status := statuses[rng.Intn(len(statuses))]
				context := fmt.Sprintf(
					"%s covers %s. Angle: tie %s's announcement to their recent coverage. "+
						"Offer exclusive data.",
					j.Name, j.Niche, cam.Title,
				)

				pitch := models.Pitch{
					UserID:       u.ID,
					ClientID:     cam.ClientID,
					CampaignID:   &cam.ID,
					JournalistID: j.ID,
					ContextBrief: context,
					Status:       status,
				}
				if err := db.Where("user_id = ? AND journalist_id = ? AND campaign_id = ?",
					u.ID, j.ID, cam.ID).FirstOrCreate(&pitch).Error; err != nil {
					log.Fatalf("create pitch: %v", err)
				}
				pitchCount++

				// 0–2 versions per pitch
				numVersions := rng.Intn(3)
				for v := 1; v <= numVersions; v++ {
					body := fmt.Sprintf(
						"Subject: Exclusive: %s reaches key milestone\n\nDear %s,\n\n"+
							"I'm reaching out about %s, which I believe aligns with your "+
							"coverage of %s at %s.\n\n"+
							"Key metrics: 40%% efficiency gains, 3x ROI within 90 days.\n\n"+
							"Would you be open to a 20-minute briefing this week?\n\nBest,\n%s",
						cam.Title, j.Name, cam.Title, j.Niche,
						j.Outlet.Name, u.Email,
					)
					ver := models.PitchVersion{
						PitchID:         pitch.ID,
						VersionNumber:   v,
						AIGeneratedBody: body,
						PromptSnapshot:  generatePromptSnapshot(j.Name, j.Outlet.Name, j.Niche, cam.ClientID.String(), cam.Title),
					}
					// Load outlet name if not preloaded
					if j.Outlet.Name == "" {
						var outlet models.Outlet
						db.First(&outlet, "id = ?", j.OutletID)
						ver.PromptSnapshot = generatePromptSnapshot(j.Name, outlet.Name, j.Niche, cam.ClientID.String(), cam.Title)
						ver.AIGeneratedBody = fmt.Sprintf(
							"Subject: Exclusive: %s reaches key milestone\n\nDear %s,\n\n"+
								"I'm reaching out about %s, which I believe aligns with your "+
								"coverage of %s at %s.\n\n"+
								"Key metrics: 40%% efficiency gains, 3x ROI within 90 days.\n\n"+
								"Would you be open to a 20-minute briefing this week?\n\nBest,\n%s",
							cam.Title, j.Name, cam.Title, j.Niche,
							outlet.Name, u.Email,
						)
					}
					existing := models.PitchVersion{}
					if err := db.Where("pitch_id = ? AND version_number = ?",
						pitch.ID, v).FirstOrCreate(&existing, ver).Error; err != nil {
						log.Fatalf("create version: %v", err)
					}
					versionCount++
				}
			}
		}
	}
	log.Printf("  pitches: %d created/found", pitchCount)
	log.Printf("  pitch versions: %d created/found", versionCount)

	log.Println("")
	log.Println("Seed complete. Demo credentials (password: Demo1234!):")
	for _, u := range userDefs {
		log.Printf("  %s  (%s)", u.email, u.name)
	}
}

func generatePromptSnapshot(journalistName, outletName, niche, clientName, campaignTitle string) string {
	return "Generate a personalized pitch email for journalist " + journalistName +
		" at " + outletName + " (niche: " + niche + ") on behalf of client " + clientName +
		". Campaign context: " + campaignTitle + "."
}

func slugify(s string) string {
	result := make([]byte, 0, len(s))
	for _, c := range []byte(s) {
		if c >= 'A' && c <= 'Z' {
			result = append(result, c+32)
		} else if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			result = append(result, c)
		}
	}
	if len(result) == 0 {
		return "x"
	}
	return string(result)
}

// ── Clear ─────────────────────────────────────────────────────────────────────

func runClear(db *gorm.DB) {
	log.Println("Clearing all data from the database...")

	sql := `TRUNCATE TABLE pitch_versions, pitches, crm_relationships, campaigns, journalists, outlets, clients, users RESTART IDENTITY CASCADE`
	if err := db.Exec(sql).Error; err != nil {
		log.Fatalf("truncate failed: %v", err)
	}

	log.Println("All tables cleared. Schema preserved.")
	log.Println("Run with --action=seed to repopulate.")
}
