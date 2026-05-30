package db

const seededReferenceDataCreatedBy = "01962b8f-aeb2-7e03-a8ff-1edce1300002"

type exclusionSeedEntry struct {
	ID           string
	ActivityType string
}

type u1SeedEntry struct {
	ID                    string
	Sector                string
	EligibleOperationType string
	ConditionGuidance     string
}

func canonicalExclusionSeedEntries() []exclusionSeedEntry {
	return []exclusionSeedEntry{
		{ID: "01962b8f-aeb2-7e03-a8ff-1edce1302001", ActivityType: "Mining of thermal coal."},
		{ID: "01962b8f-aeb2-7e03-a8ff-1edce1302002", ActivityType: "Electricity generation from coal."},
		{ID: "01962b8f-aeb2-7e03-a8ff-1edce1302003", ActivityType: "Extraction of peat."},
		{ID: "01962b8f-aeb2-7e03-a8ff-1edce1302004", ActivityType: "Electricity generation from peat."},
	}
}

func canonicalU1SeedEntries() []u1SeedEntry {
	return []u1SeedEntry{
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
}
