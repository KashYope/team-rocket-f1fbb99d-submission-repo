package domain

type RouteCode string

const (
	RoutePrague RouteCode = "PRAGUE" // ES 453 / 452 (Brussels - Prague)
	RouteMilan  RouteCode = "MILAN"  // ES 401 / 400 (Brussels - Milan Porta Garibaldi)
	RouteParis  RouteCode = "PARIS"  // ES 475 / 474 (Paris Nord - Brussels - Berlin)
)

type IncidentRequest struct {
	PresetID             string    `json:"presetId"`
	TrainID              string    `json:"trainId"`              // "453", "452", "401", "400", "475", "474"
	Route                RouteCode `json:"route"`                // "PRAGUE", "MILAN", "PARIS"
	IssueType            string    `json:"issueType"`            // "FLAT_WHEEL", "CANCELED", "NO_HV", "NO_HEATING", "SHUNT_BLOCKED"
	CauseCategory        string    `json:"causeCategory"`        // "OPERATOR_TECH", "INFRA_MANAGER", "THIRD_PARTY_EXTREME"
	ExpectedDelayMin     int       `json:"expectedDelayMin"`     // e.g., 45, 75, 95, 180, 999 (canceled)
	AffectedCar          string    `json:"affectedCar"`          // e.g., "56 (RDC WL33)" or "ALL"
	AffectedPax          int       `json:"affectedPax"`
	AvgTicketEUR         float64   `json:"avgTicketEur"`
	HasPRMPax            bool      `json:"hasPrmPax"`
	HasPetPrivateComp    bool      `json:"hasPetPrivateComp"`
	HasWomenOnlyComp     bool      `json:"hasWomenOnlyComp"`
	SimulateUnfilteredAI bool      `json:"simulateUnfilteredAi"` // Demo toggle: pass unfiltered yard pool to AI so Go post-clamp fires visibly
	ExtraInfo            string    `json:"extraInfo"`
}

type DemoPreset struct {
	ID       string          `json:"id"`
	Label    string          `json:"label"`
	Subtitle string          `json:"subtitle"`
	Badge    string          `json:"badge"`
	Request  IncidentRequest `json:"request"`
}

type ConsistCar struct {
	Position               int    `json:"position"`
	CarNumber              string `json:"carNumber"`
	UIC                    string `json:"uic"`
	Owner                  string `json:"owner"`
	CarType                string `json:"carType"`
	ProductClass           string `json:"productClass"`
	BookedPax              int    `json:"bookedPax"`
	Capacity               int    `json:"capacity"`
	IsCommercial           bool   `json:"isCommercial"`
	IsFaulty               bool   `json:"isFaulty"`
	ExceedsSingleConductor bool   `json:"exceedsSingleConductor"`
	SpecialNotes           string `json:"specialNotes,omitempty"`
}

type CorridorInfo struct {
	Route              RouteCode    `json:"route"`
	TrainIDs           []string     `json:"trainIds"`
	Title              string       `json:"title"`
	ScheduleSummary    string       `json:"scheduleSummary"`
	InfraManagers      []string     `json:"infraManagers"`
	KeyNodes           string       `json:"keyNodes"`
	MaxCarsCorridor    int          `json:"maxCarsCorridor"`
	MaxCarsDetail      string       `json:"maxCarsDetail"`
	SecondConductorEUR int          `json:"secondConductorEur"`
	WorkshopRoute      string       `json:"workshopRoute"`
	NextDepartureHours int          `json:"nextDepartureHours"`
	DefaultConsist     []ConsistCar `json:"defaultConsist"`
}

type SpareCar struct {
	UIC            string   `json:"uic"`
	Owner          string   `json:"owner"`          // "RDC", "EE", "GFF"
	CarType        string   `json:"carType"`        // "WL33", "Bvcmz5", "Bvcmz6", "Bdcm36", "Bimz"
	ProductClass   string   `json:"productClass"`   // "Comfort Plus", "Comfort Standard", "Classic", "Budget"
	Capacity       int      `json:"capacity"`
	Yard           string   `json:"yard"`           // "ddl" (Denderleeuw), "schaarbeek", "zot"
	StatusCode     string   `json:"statusCode"`     // "X" (parked/reserve), "Rev"
	ApprovedRoutes []string `json:"approvedRoutes"` // e.g. ["PRAGUE","PARIS"] or ["PRAGUE","MILAN","PARIS"]
	BlockedReason  string   `json:"blockedReason,omitempty"`
}

type FleetState struct {
	TimetableBaseline string         `json:"timetableBaseline"`
	Corridors         []CorridorInfo `json:"corridors"`
	SpareCars         []SpareCar     `json:"spareCars"`
}

type RuleCheck struct {
	ID       string `json:"id"`
	Category string `json:"category"` // "HOMOLOGATION", "SAFETY", "ECONOMICS", "EU_2021_782", "PRODUCT"
	Severity string `json:"severity"` // "PASS", "WARN", "BLOCK"
	Summary  string `json:"summary"`
	Detail   string `json:"detail"`
}

type EU2021782Assessment struct {
	CompensationBandPct int     `json:"compensationBandPct"` // 0, 25, or 50
	EstimatedRefundEUR  float64 `json:"estimatedRefundEur"`
	ForceMajeureExempt  bool    `json:"forceMajeureExempt"`
	RerouteObligation   bool    `json:"rerouteObligation"`
	SelfReroute100Min   bool    `json:"selfReroute100Min"`
	HotelRequired       bool    `json:"hotelRequired"`
	PRMPriorityRequired bool    `json:"prmPriorityRequired"`
	LegalBasisNotes     string  `json:"legalBasisNotes"`
}

type LayaChoiceAnswer struct {
	Choice           string             `json:"choice"`
	Probabilities    map[string]float64 `json:"probabilities"`
	Confidence       float64            `json:"confidence"`
	AnswerConfidence float64            `json:"answerConfidence"`
}

type LayaScoreAnswer struct {
	Score            float64            `json:"score"`
	NearestLabel     string             `json:"nearestLabel"`
	Probabilities    map[string]float64 `json:"probabilities"`
	Confidence       float64            `json:"confidence"`
	AnswerConfidence float64            `json:"answerConfidence"`
}

type LayaTriage struct {
	Source              string           `json:"source"` // "laya-live" or "laya-mock"
	Model               string           `json:"model"`
	LatencyMs           int64            `json:"latencyMs"`
	RecoveryStrategy    LayaChoiceAnswer `json:"recoveryStrategy"`
	OperationalSeverity LayaScoreAnswer  `json:"operationalSeverity"`
	RefundCliffRisk     float64          `json:"refundCliffRisk"`  // noul P(true)
	TurnaroundBreach    float64          `json:"turnaroundBreach"` // noul P(true)
	RoutingGate         string           `json:"routingGate"`      // "AUTO_CANDIDATE" or "HUMAN_REVIEW"
	RoutingReason       string           `json:"routingReason"`
}

type ScenarioCard struct {
	ID             string   `json:"id"` // "LOWEST_COST", "PAX_FIRST", "PROTECT_NEXT_TRAIN"
	Title          string   `json:"title"`
	EstimatedCost  int      `json:"estimatedCostEur"`
	Downgrades     int      `json:"downgrades"`
	Cancellations  int      `json:"cancellations"`
	NextTrainRisk  string   `json:"nextTrainRisk"` // "LOW", "MEDIUM", "HIGH"
	RecommendedCar string   `json:"recommendedCar"`
	ActionSteps    []string `json:"actionSteps"`
	GuardrailNotes []string `json:"guardrailNotes"`
	WasClampedByGo bool     `json:"wasClampedByGo"`
	TcsGermanOrder string   `json:"tcsGermanOrder"`
	PassengerAlert string   `json:"passengerAlert"`
}

type IncidentResponse struct {
	Request            IncidentRequest     `json:"request"`
	TrainSummary       string              `json:"trainSummary"`
	CorridorNotes      string              `json:"corridorNotes"`
	NextDepartureHours int                 `json:"nextDepartureHours"`
	MaxCorridorCars    int                 `json:"maxCorridorCars"`
	ActiveConsist      []ConsistCar        `json:"activeConsist"`
	RuleChecks         []RuleCheck         `json:"ruleChecks"`
	EligibleSpares     []SpareCar          `json:"eligibleSpares"`
	RejectedSpares     []SpareCar          `json:"rejectedSpares"`
	EU2021782          EU2021782Assessment `json:"eu2021782"`
	Laya               LayaTriage          `json:"laya"`
	GeminiSource       string              `json:"geminiSource"` // "gemini-live" or "gemini-mock"
	GeminiModel        string              `json:"geminiModel"`
	Scenarios          []ScenarioCard      `json:"scenarios"`
	TcsGermanOrder     string              `json:"tcsGermanOrder"`
	PassengerAlert     string              `json:"passengerAlert"`
	AuditLog           []string            `json:"auditLog"`
}
