package db

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/pashagolub/pgxmock/v4"
)

func assertiveError(message string) error {
	return errors.New(message)
}

func formatU1SeedEntries(entries []u1SeedEntry) string {
	var builder strings.Builder

	for index, entry := range entries {
		builder.WriteString(fmt.Sprintf("[%d] ID=%q Sector=%q EligibleOperationType=%q ConditionGuidance=%q\n", index, entry.ID, entry.Sector, entry.EligibleOperationType, entry.ConditionGuidance))
	}

	return builder.String()
}

func TestCanonicalExclusionSeedEntries(t *testing.T) {
	t.Parallel()

	entries := canonicalExclusionSeedEntries()
	if len(entries) != 4 {
		t.Fatalf("len(entries) = %d, want 4", len(entries))
	}

	wantActivityTypes := []string{
		"Mining of thermal coal.",
		"Electricity generation from coal.",
		"Extraction of peat.",
		"Electricity generation from peat.",
	}

	for index, wantActivityType := range wantActivityTypes {
		if entries[index].ActivityType != wantActivityType {
			t.Fatalf("entries[%d].ActivityType = %q, want %q", index, entries[index].ActivityType, wantActivityType)
		}
	}
}

func TestCanonicalU1SeedEntries(t *testing.T) {
	t.Parallel()

	want := []u1SeedEntry{
		{
			ID:                    "01962b8f-aeb2-7e03-a8ff-1edce1301001",
			Sector:                "Energy",
			EligibleOperationType: "Generation of renewable energy (e.g., from wind, solar, wave power, etc.) with negligible lifecycle GHG emissions.",
			ConditionGuidance:     "Includes generation of heat or cooling.",
		},
		{
			ID:                    "01962b8f-aeb2-7e03-a8ff-1edce1301002",
			Sector:                "Energy",
			EligibleOperationType: "Rehabilitation and desilting of existing hydropower plants, including maintenance of the catchment area (for example, a forest management plan).",
			ConditionGuidance:     "Rehabilitation includes work on the water holding capacity of the dam and work on pipes / turbines to increase productivity and bring additional grid stabilization benefits, and for pumped storage.",
		},
		{
			ID:                    "01962b8f-aeb2-7e03-a8ff-1edce1301003",
			Sector:                "Energy",
			EligibleOperationType: "District heating or cooling systems with negligible lifecycle GHG emissions.",
			ConditionGuidance:     "Using significant renewable energy or waste heat or cogenerated heat or a) modifications to lower temperature delta b) advanced pilot systems (control and energy management, etc.).",
		},
		{
			ID:                    "01962b8f-aeb2-7e03-a8ff-1edce1301004",
			Sector:                "Energy",
			EligibleOperationType: "Electricity transmission and distribution, including energy access, energy storage, and demand-side management.",
			ConditionGuidance:     "",
		},
		{
			ID:                    "01962b8f-aeb2-7e03-a8ff-1edce1301005",
			Sector:                "Energy",
			EligibleOperationType: "Cleaner cooking technologies.",
			ConditionGuidance:     "Cleaner cooking technologies substitute the use of traditional solid biomass fuels in open fires; they include sustainable biomass or electric cook stoves.",
		},
		{
			ID:                    "01962b8f-aeb2-7e03-a8ff-1edce1301006",
			Sector:                "Manufacturing",
			EligibleOperationType: "Non-energy-intensive industry (excludes chemicals, iron and steel, cement, pulp and paper, and aluminium).",
			ConditionGuidance:     "Consider the nature of the product produced (carbon content, lifetime, ability to be reused/recycled).",
		},
		{
			ID:                    "01962b8f-aeb2-7e03-a8ff-1edce1301007",
			Sector:                "Manufacturing",
			EligibleOperationType: "Manufacture of electric vehicles; non-motorized vehicles, electric locomotives; non-motorized rolling stock.",
			ConditionGuidance:     "",
		},
		{
			ID:                    "01962b8f-aeb2-7e03-a8ff-1edce1301008",
			Sector:                "Manufacturing",
			EligibleOperationType: "Manufacture of components for renewable energy or energy efficiency.",
			ConditionGuidance:     "",
		},
		{
			ID:                    "01962b8f-aeb2-7e03-a8ff-1edce1301009",
			Sector:                "Agriculture, forestry, land use and fisheries",
			EligibleOperationType: "Afforestation, reforestation, sustainable forest management, forest conservation, soil health improvement.",
			ConditionGuidance:     "With the exception of operations that expand or promote expansion into areas of high carbon stocks or high biodiversity areas.",
		},
		{
			ID:                    "01962b8f-aeb2-7e03-a8ff-1edce1301010",
			Sector:                "Agriculture, forestry, land use and fisheries",
			EligibleOperationType: "Low-GHG agriculture, climate-smart agriculture.",
			ConditionGuidance:     "With the exception of operations that expand and promote expansion into areas of high carbon stocks or high biodiversity areas and taking into account (international) transport.",
		},
		{
			ID:                    "01962b8f-aeb2-7e03-a8ff-1edce1301011",
			Sector:                "Agriculture, forestry, land use and fisheries",
			EligibleOperationType: "Conservation of natural habitats and ecosystems. Fishing and aquaculture. Non-ruminant livestock with negligible lifecycle GHG emissions.",
			ConditionGuidance:     "With the exception of operations that expand or promote expansion into areas of high carbon stocks or high biodiversity areas.",
		},
		{
			ID:                    "01962b8f-aeb2-7e03-a8ff-1edce1301012",
			Sector:                "Agriculture, forestry, land use and fisheries",
			EligibleOperationType: "Flood management and protection, coastal protection, urban drainage.",
			ConditionGuidance:     "",
		},
		{
			ID:                    "01962b8f-aeb2-7e03-a8ff-1edce1301013",
			Sector:                "Waste",
			EligibleOperationType: "Separate waste collection (in preparation for reuse and recycling), composting and anaerobic digestion of biowaste, material recovery, and landfill gas recovery from closed landfills.",
			ConditionGuidance:     "",
		},
		{
			ID:                    "01962b8f-aeb2-7e03-a8ff-1edce1301014",
			Sector:                "Water supply and wastewater",
			EligibleOperationType: "Water supply systems (e.g., expansion, rehabilitation); water quality improvement; water efficiency (e.g., non-revenue water reduction, efficient process in industries); drought management; water management at watershed level.",
			ConditionGuidance:     "Desalination plants need to go through specific assessment",
		},
		{
			ID:                    "01962b8f-aeb2-7e03-a8ff-1edce1301015",
			Sector:                "Water supply and wastewater",
			EligibleOperationType: "Gravity-based or renewable energy-powered irrigation systems.",
			ConditionGuidance:     "",
		},
		{
			ID:                    "01962b8f-aeb2-7e03-a8ff-1edce1301016",
			Sector:                "Water supply and wastewater",
			EligibleOperationType: "Wastewater treatment (domestic or industrial), including treatment and collection of sewage, sludge treatment (e.g., digestion, dewatering, drying, storage), wastewater reuse technology, resource recovery technologies (e.g., biogas into biofuel, phosphorus recovery, sludge as agriculture input, sludge as co-combustion material).",
			ConditionGuidance:     "",
		},
		{
			ID:                    "01962b8f-aeb2-7e03-a8ff-1edce1301017",
			Sector:                "Transport",
			EligibleOperationType: "Electric and non-motorized urban mobility.",
			ConditionGuidance:     "",
		},
		{
			ID:                    "01962b8f-aeb2-7e03-a8ff-1edce1301018",
			Sector:                "Transport",
			EligibleOperationType: "Roads with low traffic volumes providing access to communities which currently do not have all-weather access (for example, connecting farmers to markets or providing access to a rural school, hospital, or better social benefits).",
			ConditionGuidance:     "Except if there is any risk of contributing to deforestation",
		},
		{
			ID:                    "01962b8f-aeb2-7e03-a8ff-1edce1301019",
			Sector:                "Transport",
			EligibleOperationType: "Electric passenger or freight transport. Short sea shipping of passengers and freight ships.",
			ConditionGuidance:     "",
		},
		{
			ID:                    "01962b8f-aeb2-7e03-a8ff-1edce1301020",
			Sector:                "Transport",
			EligibleOperationType: "Inland waterways passenger and freight transport vessels",
			ConditionGuidance:     "",
		},
		{
			ID:                    "01962b8f-aeb2-7e03-a8ff-1edce1301021",
			Sector:                "Transport",
			EligibleOperationType: "Port infrastructure (maritime and inland waterways).",
			ConditionGuidance:     "",
		},
		{
			ID:                    "01962b8f-aeb2-7e03-a8ff-1edce1301022",
			Sector:                "Transport",
			EligibleOperationType: "Rail infrastructure.",
			ConditionGuidance:     "",
		},
		{
			ID:                    "01962b8f-aeb2-7e03-a8ff-1edce1301023",
			Sector:                "Transport",
			EligibleOperationType: "Road upgrading, rehabilitation, reconstruction, and maintenance without capacity expansion.",
			ConditionGuidance:     "",
		},
		{
			ID:                    "01962b8f-aeb2-7e03-a8ff-1edce1301024",
			Sector:                "Buildings and public Installations",
			EligibleOperationType: "Buildings (education, healthcare, housing, offices, retail, etc.).",
			ConditionGuidance:     "Needs to meet green building certification criteria as established by each individual MDB1.",
		},
		{
			ID:                    "01962b8f-aeb2-7e03-a8ff-1edce1301025",
			Sector:                "Buildings and public Installations",
			EligibleOperationType: "LED street lighting.",
			ConditionGuidance:     "",
		},
		{
			ID:                    "01962b8f-aeb2-7e03-a8ff-1edce1301026",
			Sector:                "Buildings and public Installations",
			EligibleOperationType: "Parks and open public spaces.",
			ConditionGuidance:     "Excluding energy-consuming installations2.",
		},
		{
			ID:                    "01962b8f-aeb2-7e03-a8ff-1edce1301027",
			Sector:                "Information and communications technology (ICT) and digital technologies",
			EligibleOperationType: "Information and communication.",
			ConditionGuidance:     "Data centres need to go through specific assessment",
		},
		{
			ID:                    "01962b8f-aeb2-7e03-a8ff-1edce1301028",
			Sector:                "Research, development, and innovation",
			EligibleOperationType: "Professional, scientific, research and development (R&D), and technical activities.",
			ConditionGuidance:     "",
		},
		{
			ID:                    "01962b8f-aeb2-7e03-a8ff-1edce1301029",
			Sector:                "Services",
			EligibleOperationType: "Public administration and compulsory social security.",
			ConditionGuidance:     "",
		},
		{
			ID:                    "01962b8f-aeb2-7e03-a8ff-1edce1301030",
			Sector:                "Services",
			EligibleOperationType: "Education (excluding infrastructure/buildings). Human health and social work activities (excluding infrastructure/buildings).",
			ConditionGuidance:     "",
		},
		{
			ID:                    "01962b8f-aeb2-7e03-a8ff-1edce1301031",
			Sector:                "Services",
			EligibleOperationType: "Social protection, cash transfer schemes.",
			ConditionGuidance:     "",
		},
		{
			ID:                    "01962b8f-aeb2-7e03-a8ff-1edce1301032",
			Sector:                "Services",
			EligibleOperationType: "Arts, entertainment, and recreation (excluding infrastructure/buildings).",
			ConditionGuidance:     "",
		},
		{
			ID:                    "01962b8f-aeb2-7e03-a8ff-1edce1301033",
			Sector:                "Cross-sectoral activities",
			EligibleOperationType: "Conversion to electricity of applications that currently use fossil fuels.",
			ConditionGuidance:     "",
		},
	}

	got := canonicalU1SeedEntries()
	if len(got) != len(want) {
		t.Fatalf("len(canonicalU1SeedEntries()) = %d, want %d", len(got), len(want))
	}

	for index := range want {
		if got[index] != want[index] {
			t.Fatalf(
				"canonicalU1SeedEntries()[%d] = %+v, want %+v\nfull got:\n%s\nfull want:\n%s",
				index,
				got[index],
				want[index],
				formatU1SeedEntries(got),
				formatU1SeedEntries(want),
			)
		}
	}
}

func TestReferenceDataSeederSeedInsertsCanonicalRows(t *testing.T) {
	t.Parallel()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool() error = %v", err)
	}
	defer mock.Close()

	seeder := NewReferenceDataSeeder(mock)

	// The seeder is upsert-only, so unrelated existing rows are left untouched because no delete SQL is issued.
	mock.ExpectBegin()
	for _, entry := range canonicalExclusionSeedEntries() {
		mock.ExpectQuery(regexp.QuoteMeta(upsertExclusionListEntryQuery)).
			WithArgs(entry.ID, entry.ActivityType, seededReferenceDataCreatedBy, pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnRows(pgxmock.NewRows([]string{"inserted"}).AddRow(true))
	}
	for _, entry := range canonicalU1SeedEntries() {
		mock.ExpectQuery(regexp.QuoteMeta(upsertU1ListEntryQuery)).
			WithArgs(entry.ID, entry.Sector, entry.EligibleOperationType, entry.ConditionGuidance, seededReferenceDataCreatedBy, pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnRows(pgxmock.NewRows([]string{"inserted"}).AddRow(true))
	}
	mock.ExpectCommit()

	result, err := seeder.Seed(context.Background())
	if err != nil {
		t.Fatalf("Seed() error = %v", err)
	}

	if result.ExclusionInserted != len(canonicalExclusionSeedEntries()) {
		t.Fatalf("result.ExclusionInserted = %d, want %d", result.ExclusionInserted, len(canonicalExclusionSeedEntries()))
	}

	if result.U1Inserted != len(canonicalU1SeedEntries()) {
		t.Fatalf("result.U1Inserted = %d, want %d", result.U1Inserted, len(canonicalU1SeedEntries()))
	}

	if result.ExclusionUpdated != 0 {
		t.Fatalf("result.ExclusionUpdated = %d, want 0", result.ExclusionUpdated)
	}

	if result.U1Updated != 0 {
		t.Fatalf("result.U1Updated = %d, want 0", result.U1Updated)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet() error = %v", err)
	}
}

func TestReferenceDataSeederSeedIsUpsertOnly(t *testing.T) {
	t.Parallel()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool() error = %v", err)
	}
	defer mock.Close()

	seeder := NewReferenceDataSeeder(mock)

	mock.ExpectBegin()
	for _, entry := range canonicalExclusionSeedEntries() {
		mock.ExpectQuery(regexp.QuoteMeta(upsertExclusionListEntryQuery)).
			WithArgs(entry.ID, entry.ActivityType, seededReferenceDataCreatedBy, pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnRows(pgxmock.NewRows([]string{"inserted"}).AddRow(true))
	}
	for _, entry := range canonicalU1SeedEntries() {
		mock.ExpectQuery(regexp.QuoteMeta(upsertU1ListEntryQuery)).
			WithArgs(entry.ID, entry.Sector, entry.EligibleOperationType, entry.ConditionGuidance, seededReferenceDataCreatedBy, pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnRows(pgxmock.NewRows([]string{"inserted"}).AddRow(true))
	}
	mock.ExpectCommit()

	if _, err := seeder.Seed(context.Background()); err != nil {
		t.Fatalf("Seed() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet() error = %v", err)
	}
}

func TestReferenceDataSeederSeedRollsBackOnExecError(t *testing.T) {
	t.Parallel()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool() error = %v", err)
	}
	defer mock.Close()

	entry := canonicalExclusionSeedEntries()[0]
	seedErr := errors.New("boom")

	seeder := NewReferenceDataSeeder(mock)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(upsertExclusionListEntryQuery)).
		WithArgs(entry.ID, entry.ActivityType, seededReferenceDataCreatedBy, pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnError(seedErr)
	mock.ExpectRollback()

	_, err = seeder.Seed(context.Background())
	if !errors.Is(err, seedErr) {
		t.Fatalf("Seed() error = %v, want wrapped %v", err, seedErr)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet() error = %v", err)
	}
}

func TestReferenceDataSeederSeedCountsUpdatedRows(t *testing.T) {
	t.Parallel()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool() error = %v", err)
	}
	defer mock.Close()

	seeder := NewReferenceDataSeeder(mock)

	mock.ExpectBegin()
	for _, entry := range canonicalExclusionSeedEntries() {
		mock.ExpectQuery(regexp.QuoteMeta(upsertExclusionListEntryQuery)).
			WithArgs(entry.ID, entry.ActivityType, seededReferenceDataCreatedBy, pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnRows(pgxmock.NewRows([]string{"inserted"}).AddRow(false))
	}
	for _, entry := range canonicalU1SeedEntries() {
		mock.ExpectQuery(regexp.QuoteMeta(upsertU1ListEntryQuery)).
			WithArgs(entry.ID, entry.Sector, entry.EligibleOperationType, entry.ConditionGuidance, seededReferenceDataCreatedBy, pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnRows(pgxmock.NewRows([]string{"inserted"}).AddRow(false))
	}
	mock.ExpectCommit()

	result, err := seeder.Seed(context.Background())
	if err != nil {
		t.Fatalf("Seed() error = %v", err)
	}

	if result.ExclusionInserted != 0 {
		t.Fatalf("result.ExclusionInserted = %d, want 0", result.ExclusionInserted)
	}

	if result.ExclusionUpdated != len(canonicalExclusionSeedEntries()) {
		t.Fatalf("result.ExclusionUpdated = %d, want %d", result.ExclusionUpdated, len(canonicalExclusionSeedEntries()))
	}

	if result.U1Inserted != 0 {
		t.Fatalf("result.U1Inserted = %d, want 0", result.U1Inserted)
	}

	if result.U1Updated != len(canonicalU1SeedEntries()) {
		t.Fatalf("result.U1Updated = %d, want %d", result.U1Updated, len(canonicalU1SeedEntries()))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet() error = %v", err)
	}
}

func TestReferenceDataSeederSeedPreservesExecErrorWhenRollbackFails(t *testing.T) {
	t.Parallel()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool() error = %v", err)
	}
	defer mock.Close()

	entry := canonicalExclusionSeedEntries()[0]
	seedErr := assertiveError("seed failed")
	rollbackErr := assertiveError("rollback failed")

	seeder := NewReferenceDataSeeder(mock)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(upsertExclusionListEntryQuery)).
		WithArgs(entry.ID, entry.ActivityType, seededReferenceDataCreatedBy, pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnError(seedErr)
	mock.ExpectRollback().WillReturnError(rollbackErr)

	_, err = seeder.Seed(context.Background())
	if err == nil {
		t.Fatal("Seed() error = nil, want error")
	}

	if !errors.Is(err, seedErr) {
		t.Fatalf("Seed() error = %v, want wrapped %v", err, seedErr)
	}

	if !errors.Is(err, rollbackErr) {
		t.Fatalf("Seed() error = %v, want wrapped %v", err, rollbackErr)
	}

	if !strings.Contains(err.Error(), "rolling back reference data seed transaction after") {
		t.Fatalf("Seed() error = %q, want rollback context", err.Error())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet() error = %v", err)
	}
}
