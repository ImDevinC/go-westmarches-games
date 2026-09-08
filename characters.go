package westmarches

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// listCharactersPage fetches one page of characters for CollectAll.
func (c *Client) listCharactersPage(ctx context.Context, page, pageSize int) (*Page[CharacterSummary], error) {
	data, err := c.do(ctx, "GET", "/characters"+queryFromOptions(ListOptions{Page: page, PageSize: pageSize}, nil), nil)
	if err != nil {
		return nil, err
	}
	var env struct {
		Success    bool               `json:"success"`
		Data       []CharacterSummary `json:"data"`
		Pagination Pagination         `json:"pagination"`
	}
	if err := jsonUnmarshal(data, &env); err != nil {
		return nil, err
	}
	return &Page[CharacterSummary]{Data: env.Data, Pagination: env.Pagination}, nil
}

// ListCharacters returns all non-deleted characters in the community tied to
// the API key, ordered by level descending. Pass a *PageFetcher to
// CollectAll to retrieve every page in one call.
func (c *Client) ListCharacters(ctx context.Context, opts ListOptions) (*Page[CharacterSummary], error) {
	return c.listCharactersPage(ctx, opts.Page, opts.PageSize)
}

// GetCharacter returns detailed information for a single character, including
// resolved currencies, attributes, data tables, and inventory.
func (c *Client) GetCharacter(ctx context.Context, characterID string) (*CharacterDetail, error) {
	data, err := c.do(ctx, "GET", "/characters/"+url.PathEscape(characterID), nil)
	if err != nil {
		return nil, err
	}
	var out CharacterDetail
	if err := decodeData(data, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetCharacterStats returns the character's game statistics derived live from
// its linked D&D Beyond sheet. The sheet must be a public D&D Beyond character.
// D&D Beyond responses are cached by the server for 5 minutes.
func (c *Client) GetCharacterStats(ctx context.Context, characterID string) (*DndBeyondStats, error) {
	data, err := c.do(ctx, "GET", "/characters/"+url.PathEscape(characterID)+"/stats", nil)
	if err != nil {
		return nil, err
	}
	var out DndBeyondStats
	if err := decodeData(data, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DistributeReward awards experience and/or currencies to a character.
// Requires write permission.
func (c *Client) DistributeReward(ctx context.Context, characterID string, req RewardRequest) (*RewardResponse, error) {
	data, err := c.do(ctx, "POST", "/characters/"+url.PathEscape(characterID)+"/rewards", req)
	if err != nil {
		return nil, err
	}
	var out RewardResponse
	if err := decodeData(data, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateCharacterStatusRequest changes a character's status.
type UpdateCharacterStatusRequest struct {
	Status    CharacterStatus `json:"status"`
	Reason    string          `json:"reason,omitempty"`
	DiscordID string          `json:"discordId,omitempty"`
}

// UpdateCharacterStatusResponse is the result of changing a character's status.
type UpdateCharacterStatusResponse struct {
	CharacterID string          `json:"characterId"`
	Status      CharacterStatus `json:"status"`
	Reason      *string         `json:"reason"`
}

// UpdateCharacterStatus changes a character's status to ACTIVE, RETIRED, or
// DECEASED. Requires write permission.
func (c *Client) UpdateCharacterStatus(ctx context.Context, characterID string, req UpdateCharacterStatusRequest) (*UpdateCharacterStatusResponse, error) {
	data, err := c.do(ctx, "PATCH", "/characters/"+url.PathEscape(characterID)+"/status", req)
	if err != nil {
		return nil, err
	}
	var out UpdateCharacterStatusResponse
	if err := decodeData(data, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ApproveCharacterResponse is the result of approving a character.
type ApproveCharacterResponse struct {
	CharacterID string `json:"characterId"`
	IsApproved  bool   `json:"isApproved"`
}

// ApproveCharacter approves a character, allowing them to receive rewards and
// participate in adventures. Idempotent. Requires write permission.
func (c *Client) ApproveCharacter(ctx context.Context, characterID string) (*ApproveCharacterResponse, error) {
	data, err := c.do(ctx, "POST", "/characters/"+url.PathEscape(characterID)+"/approve", nil)
	if err != nil {
		return nil, err
	}
	var out ApproveCharacterResponse
	if err := decodeData(data, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateInventoryItemRequest updates an inventory item's metadata. Updating
// Quantity resets RemainingQty to the new value.
type UpdateInventoryItemRequest struct {
	Name         *string            `json:"name,omitempty"`
	Description  *string            `json:"description,omitempty"`
	Notes        *string            `json:"notes,omitempty"`
	IsConsumable *bool              `json:"isConsumable,omitempty"`
	SellValue    map[string]float64 `json:"sellValue,omitempty"`
	Quantity     *int               `json:"quantity,omitempty"`
}

// UpdateInventoryItem updates an inventory item's metadata. Requires write
// permission.
func (c *Client) UpdateInventoryItem(ctx context.Context, characterID, itemID string, req UpdateInventoryItemRequest) (*InventoryItem, error) {
	path := "/characters/" + url.PathEscape(characterID) + "/inventory/" + url.PathEscape(itemID)
	data, err := c.do(ctx, "PATCH", path, req)
	if err != nil {
		return nil, err
	}
	var out InventoryItem
	if err := decodeData(data, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// QuantityRequest is the body used by the consume and sell inventory item
// endpoints.
type QuantityRequest struct {
	Quantity int `json:"quantity"`
}

// ConsumeInventoryItem consumes a quantity of a consumable item, decrementing
// its RemainingQty. Requires write permission.
func (c *Client) ConsumeInventoryItem(ctx context.Context, characterID, itemID string, quantity int) (*InventoryItem, error) {
	path := "/characters/" + url.PathEscape(characterID) + "/inventory/" + url.PathEscape(itemID) + "/consume"
	data, err := c.do(ctx, "POST", path, QuantityRequest{Quantity: quantity})
	if err != nil {
		return nil, err
	}
	var out InventoryItem
	if err := decodeData(data, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SellItemResponse is the result of selling an item.
type SellItemResponse struct {
	Item              InventoryItem      `json:"item"`
	SoldQuantity      int                `json:"soldQuantity"`
	CurrenciesAwarded map[string]float64 `json:"currenciesAwarded"`
}

// SellInventoryItem sells a quantity of an item, crediting its sellValue
// currencies to the character and decrementing RemainingQty. Requires write
// permission.
func (c *Client) SellInventoryItem(ctx context.Context, characterID, itemID string, quantity int) (*SellItemResponse, error) {
	path := "/characters/" + url.PathEscape(characterID) + "/inventory/" + url.PathEscape(itemID) + "/sell"
	data, err := c.do(ctx, "POST", path, QuantityRequest{Quantity: quantity})
	if err != nil {
		return nil, err
	}
	var out SellItemResponse
	if err := decodeData(data, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// TransferItemRequest transfers an item to another character in the same
// community.
type TransferItemRequest struct {
	ToCharacterID string `json:"toCharacterId"`
	Quantity      int    `json:"quantity,omitempty"`
}

// TransferItemResponse is the result of transferring an item.
type TransferItemResponse struct {
	UpdatedItem         InventoryItem  `json:"updatedItem"`
	TransferredItem     *InventoryItem `json:"transferredItem"`
	TransferredQuantity int            `json:"transferredQuantity"`
	ToCharacterID       string         `json:"toCharacterId"`
}

// TransferInventoryItem transfers a quantity of an item to another character
// in the same community. Creates two reward log entries for a full audit
// trail. Requires write permission.
func (c *Client) TransferInventoryItem(ctx context.Context, characterID, itemID string, req TransferItemRequest) (*TransferItemResponse, error) {
	path := "/characters/" + url.PathEscape(characterID) + "/inventory/" + url.PathEscape(itemID) + "/transfer"
	data, err := c.do(ctx, "POST", path, req)
	if err != nil {
		return nil, err
	}
	var out TransferItemResponse
	if err := decodeData(data, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DataTableRowInput is a row to store in a data table. Values are keyed by
// column ID (UUID) and may be string, number, or null.
type DataTableRowInput struct {
	Values   map[string]any `json:"values"`
	RowIndex int            `json:"rowIndex"`
}

// UpdateDataTableRowsRequest replaces all rows for a character in a data table.
type UpdateDataTableRowsRequest struct {
	Rows []DataTableRowInput `json:"rows"`
}

// UpdateDataTableResponse is the result of saving data table rows.
type UpdateDataTableResponse struct {
	DataTableID string         `json:"dataTableId"`
	CharacterID string         `json:"characterId"`
	Rows        []DataTableRow `json:"rows"`
}

// UpdateDataTableRows replaces all rows for a character in a data table. Rows
// absent from the submitted list are soft-deleted. Requires write permission.
func (c *Client) UpdateDataTableRows(ctx context.Context, characterID, dataTableID string, rows []DataTableRowInput) (*UpdateDataTableResponse, error) {
	path := "/characters/" + url.PathEscape(characterID) + "/data-tables/" + url.PathEscape(dataTableID)
	data, err := c.do(ctx, "PUT", path, UpdateDataTableRowsRequest{Rows: rows})
	if err != nil {
		return nil, err
	}
	var out UpdateDataTableResponse
	if err := decodeData(data, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// jsonUnmarshal wraps encoding/json so endpoint code stays terse and error
// messages are consistent.
func jsonUnmarshal(data []byte, v any) error {
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("westmarches: decoding response: %w", err)
	}
	return nil
}
