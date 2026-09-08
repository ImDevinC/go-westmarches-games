package westmarches

import (
	"context"
	"net/http"
	"testing"
)

func TestListAdventures(t *testing.T) {
	c, _, reqs := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("method = %s", r.Method)
		}
		if r.URL.Path != "/adventures" {
			t.Errorf("path = %q", r.URL.Path)
		}
		writeJSON(t, w, 200, map[string]any{
			"success": true,
			"data": []map[string]any{
				{
					"id": "adv_1", "title": "The Lost Mine", "startTime": "2024-01-01T00:00:00Z",
					"endTime": "2024-01-02T00:00:00Z", "minPlayers": 3, "maxPlayers": 6,
					"isCancelled": false,
					"gm":          map[string]any{"id": "user_1"},
					"participants": []any{
						map[string]any{"characterId": "char_1", "status": "APPROVED"},
					},
				},
			},
			"pagination": map[string]any{"page": 1, "pageSize": 500, "total": 1, "totalPages": 1},
		})
	})

	page, err := c.ListAdventures(context.Background(), ListOptions{})
	if err != nil {
		t.Fatalf("ListAdventures: %v", err)
	}
	if len(page.Data) != 1 {
		t.Fatalf("len(data) = %d", len(page.Data))
	}
	adv := page.Data[0]
	if adv.Title != "The Lost Mine" {
		t.Errorf("title = %q", adv.Title)
	}
	if adv.GM == nil || adv.GM.ID != "user_1" {
		t.Errorf("gm = %+v", adv.GM)
	}
	if len(adv.Participants) != 1 || adv.Participants[0].Status != ParticipantStatusApproved {
		t.Errorf("participants = %+v", adv.Participants)
	}
	_ = reqs
}

func TestGetAdventure(t *testing.T) {
	c, _, reqs := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("method = %s", r.Method)
		}
		if r.URL.Path != "/adventures/adv_1" {
			t.Errorf("path = %q", r.URL.Path)
		}
		writeJSON(t, w, 200, map[string]any{
			"success": true,
			"data": map[string]any{
				"id": "adv_1", "title": "The Lost Mine", "startTime": "2024-01-01T00:00:00Z",
				"endTime": "2024-01-02T00:00:00Z", "minPlayers": 3, "maxPlayers": 6,
				"isCancelled": false, "participants": []any{},
				"minLevel": 1, "maxLevel": 3,
				"attributeValues": []any{},
				"notes": []any{
					map[string]any{"id": "note_1", "publicNotes": "Bring torches", "createdAt": "2024-01-01T00:00:00Z"},
				},
			},
		})
	})

	adv, err := c.GetAdventure(context.Background(), "adv_1")
	if err != nil {
		t.Fatalf("GetAdventure: %v", err)
	}
	if adv.MinLevel != 1 || adv.MaxLevel != 3 {
		t.Errorf("levels = %d-%d", adv.MinLevel, adv.MaxLevel)
	}
	if len(adv.Notes) != 1 || adv.Notes[0].PublicNotes == nil || *adv.Notes[0].PublicNotes != "Bring torches" {
		t.Errorf("notes = %+v", adv.Notes)
	}
	_ = reqs
}

func TestListWikiArticlesWithFilters(t *testing.T) {
	c, _, reqs := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("method = %s", r.Method)
		}
		if r.URL.Path != "/wiki/articles" {
			t.Errorf("path = %q", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("search") != "Riverrun" {
			t.Errorf("search = %q", q.Get("search"))
		}
		if q.Get("category") != "LOCATION" {
			t.Errorf("category = %q", q.Get("category"))
		}
		if q.Get("page") != "2" {
			t.Errorf("page = %q", q.Get("page"))
		}
		writeJSON(t, w, 200, map[string]any{
			"success": true,
			"data": []map[string]any{
				{
					"id": "art_1", "title": "Riverrun", "slug": "riverrun",
					"articleCategory": "LOCATION",
					"createdAt":       "2024-01-01T00:00:00Z",
					"updatedAt":       "2024-01-01T00:00:00Z",
				},
			},
			"pagination": map[string]any{"page": 2, "pageSize": 100, "total": 1, "totalPages": 1},
		})
	})

	page, err := c.ListArticles(context.Background(), ListArticlesOptions{
		ListOptions: ListOptions{Page: 2},
		Search:      "Riverrun",
		Category:    ArticleCategoryLocation,
	})
	if err != nil {
		t.Fatalf("ListArticles: %v", err)
	}
	if len(page.Data) != 1 {
		t.Fatalf("len(data) = %d", len(page.Data))
	}
	if page.Data[0].ArticleCategory == nil || *page.Data[0].ArticleCategory != ArticleCategoryLocation {
		t.Errorf("category = %v", page.Data[0].ArticleCategory)
	}
	_ = reqs
}

func TestGetWikiArticle(t *testing.T) {
	c, _, reqs := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("method = %s", r.Method)
		}
		if r.URL.Path != "/wiki/articles/art_1" {
			t.Errorf("path = %q", r.URL.Path)
		}
		writeJSON(t, w, 200, map[string]any{
			"success": true,
			"data": map[string]any{
				"id": "art_1", "title": "Riverrun", "slug": "riverrun",
				"createdAt": "2024-01-01T00:00:00Z",
				"updatedAt": "2024-01-01T00:00:00Z",
				"content": map[string]any{
					"type": "doc",
					"content": []any{
						map[string]any{"type": "paragraph", "content": []any{map[string]any{"type": "text", "text": "Hello world"}}},
					},
				},
			},
		})
	})

	art, err := c.GetArticle(context.Background(), "art_1")
	if err != nil {
		t.Fatalf("GetArticle: %v", err)
	}
	if art.Title != "Riverrun" {
		t.Errorf("title = %q", art.Title)
	}
	if art.Content["type"] != "doc" {
		t.Errorf("content = %v", art.Content)
	}
	_ = reqs
}

func TestDistributeBulkRewards(t *testing.T) {
	c, _, reqs := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %s", r.Method)
		}
		if r.URL.Path != "/rewards" {
			t.Errorf("path = %q", r.URL.Path)
		}
		var body BulkRewardsRequest
		if err := decodeJSONBody(r, &body); err != nil {
			t.Errorf("decoding body: %v", err)
		}
		if len(body.Rewards) != 2 {
			t.Errorf("len(rewards) = %d", len(body.Rewards))
		}
		if body.Rewards[0].CharacterID != "char_1" {
			t.Errorf("rewards[0].characterId = %q", body.Rewards[0].CharacterID)
		}
		writeJSON(t, w, 201, map[string]any{
			"success": true,
			"data": []any{
				map[string]any{
					"characterId": "char_1", "experience": 100, "currencies": map[string]any{},
					"reason": "Quest", "levelUp": false, "newLevel": 5, "newExperience": 1300, "items": []any{},
				},
				map[string]any{
					"characterId": "char_2", "experience": 100, "currencies": map[string]any{},
					"reason": "Quest", "levelUp": false, "newLevel": 3, "newExperience": 800, "items": []any{},
				},
			},
		})
	})

	rewards, err := c.DistributeBulkRewards(context.Background(), BulkRewardsRequest{
		Rewards: []BulkRewardEntry{
			{CharacterID: "char_1", Experience: 100},
			{CharacterID: "char_2", Experience: 100},
		},
	})
	if err != nil {
		t.Fatalf("DistributeBulkRewards: %v", err)
	}
	if len(rewards) != 2 {
		t.Fatalf("len(rewards) = %d", len(rewards))
	}
	if rewards[1].CharacterID != "char_2" {
		t.Errorf("rewards[1].characterId = %q", rewards[1].CharacterID)
	}
	_ = reqs
}

func TestListMarketplaces(t *testing.T) {
	c, _, reqs := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("method = %s", r.Method)
		}
		if r.URL.Path != "/marketplaces" {
			t.Errorf("path = %q", r.URL.Path)
		}
		writeJSON(t, w, 200, map[string]any{
			"success": true,
			"data": []map[string]any{
				{
					"id": "mkt_1", "name": "Adventurer's Emporium", "description": nil, "order": 1,
					"restrictToRoles": false, "allowedRoles": []any{},
					"tagLabels": []any{"Type", "Rarity"}, "addPurchasesToInventory": true,
					"items": []any{
						map[string]any{
							"name": "Healing Potion", "description": "Restores 2d4+2 HP",
							"tags":             []any{map[string]any{"label": "Type", "value": "Consumable"}},
							"levelRequirement": nil,
							"costs":            []any{map[string]any{"currency": "cl1", "value": 50}},
							"experienceCost":   nil, "sellsFor": []any{},
							"isDisabled": false, "isConsumable": true, "notesRequired": false,
							"link": "", "lootTableIds": []any{},
						},
					},
				},
			},
		})
	})

	mkts, err := c.ListMarketplaces(context.Background())
	if err != nil {
		t.Fatalf("ListMarketplaces: %v", err)
	}
	if len(mkts) != 1 {
		t.Fatalf("len(mkts) = %d", len(mkts))
	}
	if mkts[0].Name != "Adventurer's Emporium" {
		t.Errorf("name = %q", mkts[0].Name)
	}
	if len(mkts[0].Items) != 1 {
		t.Fatalf("len(items) = %d", len(mkts[0].Items))
	}
	item := mkts[0].Items[0]
	if item.Name != "Healing Potion" || !item.IsConsumable {
		t.Errorf("unexpected item: %+v", item)
	}
	if len(item.Tags) != 1 || item.Tags[0].Label != "Type" {
		t.Errorf("tags = %+v", item.Tags)
	}
	_ = reqs
}

func TestListCurrencies(t *testing.T) {
	c, _, reqs := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("method = %s", r.Method)
		}
		if r.URL.Path != "/currencies" {
			t.Errorf("path = %q", r.URL.Path)
		}
		writeJSON(t, w, 200, map[string]any{
			"success": true,
			"data": []map[string]any{
				{"id": "cl1", "name": "Gold", "order": 1},
				{"id": "cl2", "name": "Silver", "order": 2},
			},
		})
	})

	currencies, err := c.ListCurrencies(context.Background())
	if err != nil {
		t.Fatalf("ListCurrencies: %v", err)
	}
	if len(currencies) != 2 {
		t.Fatalf("len(currencies) = %d", len(currencies))
	}
	if currencies[0].ID != "cl1" || currencies[0].Name != "Gold" {
		t.Errorf("unexpected currency: %+v", currencies[0])
	}
	_ = reqs
}
