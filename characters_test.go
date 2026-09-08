package westmarches

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
)

func TestListCharacters(t *testing.T) {
	c, _, reqs := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("method = %s", r.Method)
		}
		if r.URL.Path != "/characters" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("page"); got != "3" {
			t.Errorf("page = %q", got)
		}
		if got := r.URL.Query().Get("pageSize"); got != "100" {
			t.Errorf("pageSize = %q", got)
		}
		writeJSON(t, w, 200, map[string]any{
			"success": true,
			"data": []map[string]any{
				{
					"id": "char_1", "name": "Grog", "level": 5, "experience": 1200,
					"status": "ACTIVE", "isApproved": true,
					"user": map[string]any{"id": "user_1", "discordId": "12345"},
				},
			},
			"pagination": map[string]any{"page": 3, "pageSize": 100, "total": 1, "totalPages": 1},
		})
	})

	page, err := c.ListCharacters(context.Background(), ListOptions{Page: 3, PageSize: 100})
	if err != nil {
		t.Fatalf("ListCharacters: %v", err)
	}
	if len(page.Data) != 1 {
		t.Fatalf("len(data) = %d", len(page.Data))
	}
	if page.Data[0].Name != "Grog" {
		t.Errorf("name = %q", page.Data[0].Name)
	}
	if page.Data[0].User.ID != "user_1" {
		t.Errorf("user.id = %q", page.Data[0].User.ID)
	}
	if page.Pagination.Total != 1 {
		t.Errorf("pagination.total = %d", page.Pagination.Total)
	}
	_ = reqs
}

func TestGetCharacter(t *testing.T) {
	c, _, reqs := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("method = %s", r.Method)
		}
		if r.URL.RawPath != "/characters/char%2Fid" {
			t.Errorf("raw path = %q, want escaped id", r.URL.RawPath)
		}
		writeJSON(t, w, 200, map[string]any{
			"success": true,
			"data": map[string]any{
				"id": "char_1", "name": "Grog", "level": 5, "experience": 1200,
				"status": "ACTIVE", "isApproved": true,
				"user":            map[string]any{"id": "user_1"},
				"characterSheet":  "https://dndbeyond.com/characters/1",
				"currencies":      []any{},
				"attributeValues": []any{},
				"dataTables":      []any{},
				"inventoryItems":  []any{},
				"deletedAt":       nil,
				"deletedById":     nil,
			},
		})
	})

	ch, err := c.GetCharacter(context.Background(), "char/id")
	if err != nil {
		t.Fatalf("GetCharacter: %v", err)
	}
	if ch.ID != "char_1" || ch.Name != "Grog" {
		t.Errorf("unexpected character: %+v", ch)
	}
	if ch.CharacterSheet != "https://dndbeyond.com/characters/1" {
		t.Errorf("characterSheet = %q", ch.CharacterSheet)
	}
	_ = reqs
}

func TestGetCharacterNotFound(t *testing.T) {
	c, _, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, 404, map[string]any{"success": false, "error": "Character not found"})
	})
	_, err := c.GetCharacter(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if !apiErr.IsNotFound() {
		t.Error("expected IsNotFound()")
	}
}

func TestGetCharacterStats(t *testing.T) {
	c, _, reqs := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("method = %s", r.Method)
		}
		if r.URL.Path != "/characters/char_1/stats" {
			t.Errorf("path = %q", r.URL.Path)
		}
		writeJSON(t, w, 200, map[string]any{
			"success": true,
			"data": map[string]any{
				"abilityScores":        map[string]any{"strength": 16, "dexterity": 12, "constitution": 14, "intelligence": 10, "wisdom": 8, "charisma": 13},
				"abilityModifiers":     map[string]any{"strength": 3, "dexterity": 1, "constitution": 2, "intelligence": 0, "wisdom": -1, "charisma": 1},
				"armorClass":           17,
				"maxHitPoints":         40,
				"totalLevel":           5,
				"proficiencyBonus":     3,
				"initiative":           1,
				"initiativeAdvantage":  false,
				"walkingSpeed":         30,
				"darkvision":           nil,
				"passivePerception":    9,
				"passiveInvestigation": 10,
				"passiveInsight":       11,
				"languages":            []any{"Common", "Dwarvish"},
				"race":                 "Dwarf",
				"classes":              []any{map[string]any{"name": "Fighter", "subclass": nil, "level": 5}},
				"savingThrows":         []any{},
				"skills":               []any{},
				"conditionalRolls":     []any{},
				"spellcasting":         []any{},
				"spellSlots":           []any{},
				"pactMagicSlots":       []any{},
			},
		})
	})

	stats, err := c.GetCharacterStats(context.Background(), "char_1")
	if err != nil {
		t.Fatalf("GetCharacterStats: %v", err)
	}
	if stats.ArmorClass != 17 {
		t.Errorf("armorClass = %d", stats.ArmorClass)
	}
	if stats.AbilityScores.Strength != 16 {
		t.Errorf("strength = %d", stats.AbilityScores.Strength)
	}
	if stats.Race != "Dwarf" {
		t.Errorf("race = %q", stats.Race)
	}
	if stats.Darkvision != nil {
		t.Errorf("darkvision should be nil, got %v", *stats.Darkvision)
	}
	_ = reqs
}

func TestDistributeReward(t *testing.T) {
	c, _, reqs := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %s", r.Method)
		}
		if r.URL.Path != "/characters/char_1/rewards" {
			t.Errorf("path = %q", r.URL.Path)
		}
		var body RewardRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decoding body: %v", err)
		}
		if body.Experience != 100 {
			t.Errorf("experience = %d", body.Experience)
		}
		if body.Currencies["cl1"] != 50 {
			t.Errorf("currencies = %v", body.Currencies)
		}
		writeJSON(t, w, 201, map[string]any{
			"success": true,
			"data": map[string]any{
				"characterId": "char_1", "experience": 100,
				"currencies": map[string]any{"cl1": 50}, "reason": "Quest",
				"levelUp": false, "newLevel": 5, "newExperience": 1300,
				"items": []any{},
			},
		})
	})

	resp, err := c.DistributeReward(context.Background(), "char_1", RewardRequest{
		Experience: 100,
		Currencies: map[string]float64{"cl1": 50},
		Reason:     "Quest",
	})
	if err != nil {
		t.Fatalf("DistributeReward: %v", err)
	}
	if resp.CharacterID != "char_1" || resp.Experience != 100 {
		t.Errorf("unexpected response: %+v", resp)
	}
	_ = reqs
}

func TestUpdateCharacterStatus(t *testing.T) {
	c, _, reqs := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("method = %s", r.Method)
		}
		if r.URL.Path != "/characters/char_1/status" {
			t.Errorf("path = %q", r.URL.Path)
		}
		var body UpdateCharacterStatusRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decoding body: %v", err)
		}
		if body.Status != CharacterStatusRetired {
			t.Errorf("status = %q", body.Status)
		}
		writeJSON(t, w, 200, map[string]any{
			"success": true,
			"data":    map[string]any{"characterId": "char_1", "status": "RETIRED", "reason": "Peaceful retirement"},
		})
	})

	resp, err := c.UpdateCharacterStatus(context.Background(), "char_1", UpdateCharacterStatusRequest{
		Status: CharacterStatusRetired,
		Reason: "Peaceful retirement",
	})
	if err != nil {
		t.Fatalf("UpdateCharacterStatus: %v", err)
	}
	if resp.Status != CharacterStatusRetired {
		t.Errorf("status = %q", resp.Status)
	}
	if resp.Reason == nil || *resp.Reason != "Peaceful retirement" {
		t.Errorf("reason = %v", resp.Reason)
	}
	_ = reqs
}

func TestApproveCharacter(t *testing.T) {
	c, _, reqs := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %s", r.Method)
		}
		if r.URL.Path != "/characters/char_1/approve" {
			t.Errorf("path = %q", r.URL.Path)
		}
		writeJSON(t, w, 200, map[string]any{
			"success": true,
			"data":    map[string]any{"characterId": "char_1", "isApproved": true},
		})
	})

	resp, err := c.ApproveCharacter(context.Background(), "char_1")
	if err != nil {
		t.Fatalf("ApproveCharacter: %v", err)
	}
	if !resp.IsApproved || resp.CharacterID != "char_1" {
		t.Errorf("unexpected response: %+v", resp)
	}
	_ = reqs
}

func TestUpdateInventoryItem(t *testing.T) {
	c, _, reqs := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("method = %s", r.Method)
		}
		if r.URL.Path != "/characters/char_1/inventory/item_1" {
			t.Errorf("path = %q", r.URL.Path)
		}
		var body UpdateInventoryItemRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decoding body: %v", err)
		}
		if body.Name == nil || *body.Name != "Greatsword" {
			t.Errorf("name = %v", body.Name)
		}
		writeJSON(t, w, 200, map[string]any{
			"success": true,
			"data": map[string]any{
				"id": "item_1", "name": "Greatsword", "quantity": 1, "remainingQty": 1,
				"isConsumable": false, "purchasedAt": "2024-01-01T00:00:00Z",
				"createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z",
			},
		})
	})

	name := "Greatsword"
	item, err := c.UpdateInventoryItem(context.Background(), "char_1", "item_1", UpdateInventoryItemRequest{Name: &name})
	if err != nil {
		t.Fatalf("UpdateInventoryItem: %v", err)
	}
	if item.Name != "Greatsword" {
		t.Errorf("name = %q", item.Name)
	}
	_ = reqs
}

func TestConsumeInventoryItem(t *testing.T) {
	c, _, reqs := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %s", r.Method)
		}
		if r.URL.Path != "/characters/char_1/inventory/item_1/consume" {
			t.Errorf("path = %q", r.URL.Path)
		}
		var body QuantityRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decoding body: %v", err)
		}
		if body.Quantity != 2 {
			t.Errorf("quantity = %d", body.Quantity)
		}
		writeJSON(t, w, 200, map[string]any{
			"success": true,
			"data": map[string]any{
				"id": "item_1", "name": "Potion", "quantity": 5, "remainingQty": 3,
				"isConsumable": true, "purchasedAt": "2024-01-01T00:00:00Z",
				"createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z",
			},
		})
	})

	item, err := c.ConsumeInventoryItem(context.Background(), "char_1", "item_1", 2)
	if err != nil {
		t.Fatalf("ConsumeInventoryItem: %v", err)
	}
	if item.RemainingQty != 3 {
		t.Errorf("remainingQty = %d", item.RemainingQty)
	}
	_ = reqs
}

func TestSellInventoryItem(t *testing.T) {
	c, _, reqs := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %s", r.Method)
		}
		if r.URL.Path != "/characters/char_1/inventory/item_1/sell" {
			t.Errorf("path = %q", r.URL.Path)
		}
		var body QuantityRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decoding body: %v", err)
		}
		if body.Quantity != 1 {
			t.Errorf("quantity = %d", body.Quantity)
		}
		writeJSON(t, w, 200, map[string]any{
			"success": true,
			"data": map[string]any{
				"item": map[string]any{
					"id": "item_1", "name": "Sword", "quantity": 2, "remainingQty": 1,
					"isConsumable": false, "purchasedAt": "2024-01-01T00:00:00Z",
					"createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z",
				},
				"soldQuantity":      1,
				"currenciesAwarded": map[string]any{"cl1": 10},
			},
		})
	})

	resp, err := c.SellInventoryItem(context.Background(), "char_1", "item_1", 1)
	if err != nil {
		t.Fatalf("SellInventoryItem: %v", err)
	}
	if resp.SoldQuantity != 1 {
		t.Errorf("soldQuantity = %d", resp.SoldQuantity)
	}
	if resp.CurrenciesAwarded["cl1"] != 10 {
		t.Errorf("currenciesAwarded = %v", resp.CurrenciesAwarded)
	}
	_ = reqs
}

func TestTransferInventoryItem(t *testing.T) {
	c, _, reqs := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %s", r.Method)
		}
		if r.URL.Path != "/characters/char_1/inventory/item_1/transfer" {
			t.Errorf("path = %q", r.URL.Path)
		}
		var body TransferItemRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decoding body: %v", err)
		}
		if body.ToCharacterID != "char_2" {
			t.Errorf("toCharacterId = %q", body.ToCharacterID)
		}
		writeJSON(t, w, 200, map[string]any{
			"success": true,
			"data": map[string]any{
				"updatedItem": map[string]any{
					"id": "item_1", "name": "Sword", "quantity": 2, "remainingQty": 1,
					"isConsumable": false, "purchasedAt": "2024-01-01T00:00:00Z",
					"createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z",
				},
				"transferredItem": map[string]any{
					"id": "item_9", "name": "Sword", "quantity": 1, "remainingQty": 1,
					"isConsumable": false, "purchasedAt": "2024-01-01T00:00:00Z",
					"createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z",
				},
				"transferredQuantity": 1,
				"toCharacterId":       "char_2",
			},
		})
	})

	resp, err := c.TransferInventoryItem(context.Background(), "char_1", "item_1", TransferItemRequest{
		ToCharacterID: "char_2",
		Quantity:      1,
	})
	if err != nil {
		t.Fatalf("TransferInventoryItem: %v", err)
	}
	if resp.ToCharacterID != "char_2" || resp.TransferredQuantity != 1 {
		t.Errorf("unexpected response: %+v", resp)
	}
	if resp.TransferredItem == nil || resp.TransferredItem.ID != "item_9" {
		t.Errorf("transferredItem = %+v", resp.TransferredItem)
	}
	_ = reqs
}

func TestUpdateDataTableRows(t *testing.T) {
	c, _, reqs := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("method = %s", r.Method)
		}
		if r.URL.Path != "/characters/char_1/data-tables/dt_1" {
			t.Errorf("path = %q", r.URL.Path)
		}
		var body UpdateDataTableRowsRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decoding body: %v", err)
		}
		if len(body.Rows) != 1 {
			t.Fatalf("len(rows) = %d", len(body.Rows))
		}
		if body.Rows[0].Values["col-1"] != "Longsword" {
			t.Errorf("values = %v", body.Rows[0].Values)
		}
		writeJSON(t, w, 200, map[string]any{
			"success": true,
			"data": map[string]any{
				"dataTableId": "dt_1",
				"characterId": "char_1",
				"rows": []any{
					map[string]any{
						"id": "row_1", "contentHash": "abc",
						"values": map[string]any{"col-1": "Longsword"}, "rowIndex": 0,
					},
				},
			},
		})
	})

	resp, err := c.UpdateDataTableRows(context.Background(), "char_1", "dt_1", []DataTableRowInput{
		{Values: map[string]any{"col-1": "Longsword"}, RowIndex: 0},
	})
	if err != nil {
		t.Fatalf("UpdateDataTableRows: %v", err)
	}
	if len(resp.Rows) != 1 || resp.Rows[0].ID != "row_1" {
		t.Errorf("unexpected response: %+v", resp)
	}
	_ = reqs
}

// TestCollectAllCharacters verifies that CollectAll pages through a paginated
// endpoint until the last page.
func TestCollectAllCharacters(t *testing.T) {
	pages := map[int][]CharacterSummary{
		1: {{ID: "c1", Name: "A"}, {ID: "c2", Name: "B"}},
		2: {{ID: "c3", Name: "C"}},
	}
	c, _, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		pageStr := r.URL.Query().Get("page")
		page := 1
		if pageStr != "" {
			page = atoiOr(pageStr, 1)
		}
		totalPages := len(pages)
		writeJSON(t, w, 200, map[string]any{
			"success":    true,
			"data":       pages[page],
			"pagination": map[string]any{"page": page, "pageSize": 2, "total": 3, "totalPages": totalPages},
		})
	})

	all, err := CollectAll(context.Background(), 2, func(ctx context.Context, page, pageSize int) (*Page[CharacterSummary], error) {
		return c.ListCharacters(ctx, ListOptions{Page: page, PageSize: pageSize})
	})
	if err != nil {
		t.Fatalf("CollectAll: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("len(all) = %d, want 3", len(all))
	}
	if all[0].ID != "c1" || all[2].ID != "c3" {
		t.Errorf("unexpected order: %+v", all)
	}
}

func atoiOr(s string, def int) int {
	v := 0
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return def
		}
		v = v*10 + int(ch-'0')
	}
	if v == 0 {
		return def
	}
	return v
}

var _ = url.PathEscape
