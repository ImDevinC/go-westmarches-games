package westmarches

// Error is the standard error envelope returned for non-2xx responses.
type Error struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// Pagination describes the current page of a paginated list response.
type Pagination struct {
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

// Page is a generic paginated response. Data holds the current page's items
// and Pagination describes how to navigate the full result set.
type Page[T any] struct {
	Data       []T
	Pagination Pagination
}

// UserRef is a minimal user reference (no PII).
type UserRef struct {
	ID        string  `json:"id"`
	DiscordID *string `json:"discordId"`
}

// RegionRef references a region and its realm.
type RegionRef struct {
	ID    string    `json:"id"`
	Name  string    `json:"name"`
	Realm *RealmRef `json:"realm"`
}

// RealmRef references a realm.
type RealmRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// AttributeValue is a character/adventure attribute value.
type AttributeValue struct {
	ValueTexts []string  `json:"valueTexts"`
	Attribute  Attribute `json:"attribute"`
}

// Attribute describes an attribute's name and type.
type Attribute struct {
	Name string `json:"name"`
	Type string `json:"type"` // TEXT, OPTION, MULTI_OPTION
}

// DataTableColumn is a column definition within a data table. Use ID as the
// key in row values.
type DataTableColumn struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"` // text, numeric
	Decimals *int   `json:"decimals"`
	Required *bool  `json:"required"`
	Order    int    `json:"order"`
}

// DataTableRow is a row in a data table. ContentHash identifies the row across
// saves. Values are keyed by column ID and may be string, number, or null.
type DataTableRow struct {
	ID          string         `json:"id"`
	ContentHash string         `json:"contentHash"`
	Values      map[string]any `json:"values"`
	RowIndex    int            `json:"rowIndex"`
}

// DataTable is a data table with all non-deleted rows for a character.
type DataTable struct {
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	Order   int               `json:"order"`
	Columns []DataTableColumn `json:"columns"`
	Rows    []DataTableRow    `json:"rows"`
}

// CharacterStatus is the lifecycle status of a character.
type CharacterStatus string

const (
	CharacterStatusActive   CharacterStatus = "ACTIVE"
	CharacterStatusRetired  CharacterStatus = "RETIRED"
	CharacterStatusDeceased CharacterStatus = "DECEASED"
)

// CharacterSummary is the list view of a character.
type CharacterSummary struct {
	ID         string          `json:"id"`
	Name       string          `json:"name"`
	Level      int             `json:"level"`
	Experience int             `json:"experience"`
	Status     CharacterStatus `json:"status"`
	IsApproved bool            `json:"isApproved"`
	Image      *string         `json:"image"`
	User       UserRef         `json:"user"`
}

// CharacterCurrency is a character's balance in one currency.
type CharacterCurrency struct {
	CurrencyID   string  `json:"currencyId"`
	CurrencyName string  `json:"currencyName"`
	Amount       float64 `json:"amount"`
}

// CharacterDetail is the full detail view of a character.
type CharacterDetail struct {
	CharacterSummary
	Description     *string             `json:"description"`
	CharacterSheet  string              `json:"characterSheet"`
	Currencies      []CharacterCurrency `json:"currencies"`
	CurrentRegion   *RegionRef          `json:"currentRegion"`
	AttributeValues []AttributeValue    `json:"attributeValues"`
	DataTables      []DataTable         `json:"dataTables"`
	DeletedAt       *string             `json:"deletedAt"`
	DeletedByID     *string             `json:"deletedById"`
	InventoryItems  []InventoryItem     `json:"inventoryItems"`
}

// ParticipantStatus is the participation status of a character in an adventure.
type ParticipantStatus string

const (
	ParticipantStatusPending  ParticipantStatus = "PENDING"
	ParticipantStatusApproved ParticipantStatus = "APPROVED"
	ParticipantStatusRejected ParticipantStatus = "REJECTED"
)

// Participant links a character to an adventure.
type Participant struct {
	CharacterID string            `json:"characterId"`
	Status      ParticipantStatus `json:"status"`
}

// AdventureSummary is the list view of an adventure.
type AdventureSummary struct {
	ID           string        `json:"id"`
	Title        string        `json:"title"`
	StartTime    string        `json:"startTime"`
	EndTime      string        `json:"endTime"`
	MinPlayers   int           `json:"minPlayers"`
	MaxPlayers   int           `json:"maxPlayers"`
	IsCancelled  bool          `json:"isCancelled"`
	GM           *UserRef      `json:"gm"`
	Participants []Participant `json:"participants"`
}

// AdventureNote is a public note on an adventure.
type AdventureNote struct {
	ID            string  `json:"id"`
	PublicNotes   *string `json:"publicNotes"`
	UserName      *string `json:"userName"`
	CharacterName *string `json:"characterName"`
	CreatedAt     string  `json:"createdAt"`
}

// AdventureDetail is the full detail view of an adventure.
type AdventureDetail struct {
	AdventureSummary
	Description     *string          `json:"description"`
	MinLevel        int              `json:"minLevel"`
	MaxLevel        int              `json:"maxLevel"`
	Region          *RegionRef       `json:"region"`
	AttributeValues []AttributeValue `json:"attributeValues"`
	Notes           []AdventureNote  `json:"notes"`
}

// InventoryItem is an item in a character's inventory with remaining
// quantity > 0.
type InventoryItem struct {
	ID           string             `json:"id"`
	Name         string             `json:"name"`
	Description  *string            `json:"description"`
	Quantity     int                `json:"quantity"`
	RemainingQty int                `json:"remainingQty"`
	IsConsumable bool               `json:"isConsumable"`
	SellValue    map[string]float64 `json:"sellValue"`
	Notes        *string            `json:"notes"`
	Source       *string            `json:"source"`
	PurchasedAt  string             `json:"purchasedAt"`
	CreatedAt    string             `json:"createdAt"`
	UpdatedAt    string             `json:"updatedAt"`
}

// RewardItem is an item granted as part of a reward.
type RewardItem struct {
	Name         string             `json:"name"`
	Description  *string            `json:"description"`
	Quantity     int                `json:"quantity"`
	IsConsumable bool               `json:"isConsumable"`
	SellValue    map[string]float64 `json:"sellValue"`
	Notes        *string            `json:"notes"`
}

// ArticleCategory is the category of a wiki article.
type ArticleCategory string

const (
	ArticleCategoryNPC             ArticleCategory = "NPC"
	ArticleCategoryLocation        ArticleCategory = "LOCATION"
	ArticleCategoryFaction         ArticleCategory = "FACTION"
	ArticleCategoryDeity           ArticleCategory = "DEITY"
	ArticleCategoryLore            ArticleCategory = "LORE"
	ArticleCategoryEvent           ArticleCategory = "EVENT"
	ArticleCategoryArtifact        ArticleCategory = "ARTIFACT"
	ArticleCategoryCreature        ArticleCategory = "CREATURE"
	ArticleCategoryRules           ArticleCategory = "RULES"
	ArticleCategoryAdventureReport ArticleCategory = "ADVENTURE_REPORT"
)

// ArticleImage is the image attached to a wiki article.
type ArticleImage struct {
	URL string `json:"url"`
}

// ArticleSummary is the list view of a published wiki article.
type ArticleSummary struct {
	ID              string           `json:"id"`
	Title           string           `json:"title"`
	Slug            *string          `json:"slug"`
	ArticleCategory *ArticleCategory `json:"articleCategory"`
	Image           *ArticleImage    `json:"image"`
	URL             *string          `json:"url"`
	CreatedAt       string           `json:"createdAt"`
	UpdatedAt       string           `json:"updatedAt"`
}

// ArticleDetail is a wiki article including its full content body. Content is
// a ProseMirror/TipTap JSON document suitable for AI processing; extract text
// from the "text" fields in leaf nodes.
type ArticleDetail struct {
	ArticleSummary
	Content map[string]any `json:"content"`
}

// Currency is a community currency definition.
type Currency struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Order int    `json:"order"`
}

// MarketplaceItemTag is a tag on a marketplace item.
type MarketplaceItemTag struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// MarketplaceItemCost is a currency cost for purchasing a marketplace item.
// Negative values mean the buyer receives that currency.
type MarketplaceItemCost struct {
	Currency string  `json:"currency"`
	Value    float64 `json:"value"`
}

// MarketplaceItemSellValue is what a seller receives when selling an item back.
type MarketplaceItemSellValue struct {
	Currency string  `json:"currency"`
	Value    float64 `json:"value"`
}

// MarketplaceItemEntry is a single item listed in a marketplace. Item UUIDs
// are intentionally omitted and must not be used as stable identifiers.
type MarketplaceItemEntry struct {
	Name             string                     `json:"name"`
	Description      string                     `json:"description"`
	Tags             []MarketplaceItemTag       `json:"tags"`
	LevelRequirement *int                       `json:"levelRequirement"`
	Costs            []MarketplaceItemCost      `json:"costs"`
	ExperienceCost   *float64                   `json:"experienceCost"`
	SellsFor         []MarketplaceItemSellValue `json:"sellsFor"`
	IsDisabled       bool                       `json:"isDisabled"`
	IsConsumable     bool                       `json:"isConsumable"`
	NotesRequired    bool                       `json:"notesRequired"`
	Link             string                     `json:"link"`
	LootTableIDs     []string                   `json:"lootTableIds"`
}

// Marketplace is a community marketplace with all its enabled items.
type Marketplace struct {
	ID                      string                 `json:"id"`
	Name                    string                 `json:"name"`
	Description             *string                `json:"description"`
	Order                   int                    `json:"order"`
	RestrictToRoles         bool                   `json:"restrictToRoles"`
	AllowedRoles            []string               `json:"allowedRoles"`
	TagLabels               []string               `json:"tagLabels"`
	AddPurchasesToInventory bool                   `json:"addPurchasesToInventory"`
	Items                   []MarketplaceItemEntry `json:"items"`
}

// BulkRewardEntry is a per-character reward entry within a bulk reward
// request. Entries with identical experience, currencies, and reason are
// merged into a single Discord notification.
type BulkRewardEntry struct {
	CharacterID string             `json:"characterId"`
	Experience  int                `json:"experience,omitempty"`
	Currencies  map[string]float64 `json:"currencies,omitempty"`
	Reason      string             `json:"reason,omitempty"`
}

// BulkRewardsRequest rewards multiple characters in a single request.
type BulkRewardsRequest struct {
	Rewards     []BulkRewardEntry `json:"rewards"`
	DiscordID   string            `json:"discordId,omitempty"`
	AdventureID string            `json:"adventureId,omitempty"`
}

// RewardRequest awards experience and/or currencies (and optionally items) to
// a single character. At least one of Experience, Currencies, or Items must be
// non-empty.
type RewardRequest struct {
	Experience int                `json:"experience,omitempty"`
	Currencies map[string]float64 `json:"currencies,omitempty"`
	Reason     string             `json:"reason,omitempty"`
	DiscordID  string             `json:"discordId,omitempty"`
	Items      []RewardItem       `json:"items,omitempty"`
}

// RewardResponse is the result of distributing a reward to a character.
type RewardResponse struct {
	CharacterID   string             `json:"characterId"`
	Experience    int                `json:"experience"`
	Currencies    map[string]float64 `json:"currencies"`
	Reason        string             `json:"reason"`
	LevelUp       bool               `json:"levelUp"`
	NewLevel      int                `json:"newLevel"`
	NewExperience int                `json:"newExperience"`
	Items         []RewardItem       `json:"items"`
}

// AbilityScores holds the six D&D ability scores (or modifiers).
type AbilityScores struct {
	Strength     int `json:"strength"`
	Dexterity    int `json:"dexterity"`
	Constitution int `json:"constitution"`
	Intelligence int `json:"intelligence"`
	Wisdom       int `json:"wisdom"`
	Charisma     int `json:"charisma"`
}

// ClassEntry is a character class and level from a D&D Beyond sheet.
type ClassEntry struct {
	Name     string  `json:"name"`
	Subclass *string `json:"subclass"`
	Level    int     `json:"level"`
}

// SavingThrow is a D&D saving throw.
type SavingThrow struct {
	Ability      string `json:"ability"`
	Modifier     int    `json:"modifier"`
	Proficient   bool   `json:"proficient"`
	Advantage    bool   `json:"advantage"`
	Disadvantage bool   `json:"disadvantage"`
}

// Skill is a D&D skill.
type Skill struct {
	Key          string `json:"key"`
	Label        string `json:"label"`
	Ability      string `json:"ability"`
	Modifier     int    `json:"modifier"`
	Proficient   bool   `json:"proficient"`
	Expertise    bool   `json:"expertise"`
	Advantage    bool   `json:"advantage"`
	Disadvantage bool   `json:"disadvantage"`
}

// ConditionalRoll is an advantage/disadvantage that applies only in specific
// situations.
type ConditionalRoll struct {
	Advantage   bool   `json:"advantage"`
	Subject     string `json:"subject"`
	Restriction string `json:"restriction"`
}

// Spellcasting describes a character's spellcasting class.
type Spellcasting struct {
	ClassName   string `json:"className"`
	Ability     string `json:"ability"`
	Modifier    int    `json:"modifier"`
	SpellAttack int    `json:"spellAttack"`
	SaveDC      int    `json:"saveDc"`
}

// SpellSlot is a spell slot level and maximum count.
type SpellSlot struct {
	Level int `json:"level"`
	Max   int `json:"max"`
}

// DndBeyondStats is the game statistics derived live from a character's linked
// D&D Beyond sheet — the same values shown on the DM screen.
type DndBeyondStats struct {
	AbilityScores        AbilityScores     `json:"abilityScores"`
	AbilityModifiers     AbilityScores     `json:"abilityModifiers"`
	ArmorClass           int               `json:"armorClass"`
	MaxHitPoints         int               `json:"maxHitPoints"`
	TotalLevel           int               `json:"totalLevel"`
	ProficiencyBonus     int               `json:"proficiencyBonus"`
	Initiative           int               `json:"initiative"`
	InitiativeAdvantage  bool              `json:"initiativeAdvantage"`
	WalkingSpeed         int               `json:"walkingSpeed"`
	Darkvision           *int              `json:"darkvision"`
	PassivePerception    int               `json:"passivePerception"`
	PassiveInvestigation int               `json:"passiveInvestigation"`
	PassiveInsight       int               `json:"passiveInsight"`
	Languages            []string          `json:"languages"`
	Race                 string            `json:"race"`
	Classes              []ClassEntry      `json:"classes"`
	SavingThrows         []SavingThrow     `json:"savingThrows"`
	Skills               []Skill           `json:"skills"`
	ConditionalRolls     []ConditionalRoll `json:"conditionalRolls"`
	Spellcasting         []Spellcasting    `json:"spellcasting"`
	SpellSlots           []SpellSlot       `json:"spellSlots"`
	PactMagicSlots       []SpellSlot       `json:"pactMagicSlots"`
}
