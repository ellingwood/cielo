package mcp

import (
	"context"
	"encoding/json"
	"fmt"
)

type ToolDef struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	InputSchema any    `json:"inputSchema"`
}

func strArg(args map[string]any, key string) string {
	if v, ok := args[key].(string); ok {
		return v
	}
	return ""
}

func intArg(args map[string]any, key string) int {
	if v, ok := args[key].(float64); ok {
		return int(v)
	}
	return 0
}

func (s *Server) callTool(ctx context.Context, reqID any, name string, args map[string]any) JSONRPCResponse {
	result, err := s.executeTool(ctx, name, args)
	if err != nil {
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      reqID,
			Result: map[string]any{
				"content": []map[string]any{
					{"type": "text", "text": fmt.Sprintf("Error: %s", err.Error())},
				},
				"isError": true,
			},
		}
	}
	text, _ := json.Marshal(result)
	return JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      reqID,
		Result: map[string]any{
			"content": []map[string]any{
				{"type": "text", "text": string(text)},
			},
		},
	}
}

func (s *Server) executeTool(ctx context.Context, name string, args map[string]any) (any, error) {
	actor := strArg(args, "actor")
	if actor == "" {
		actor = "agent"
	}

	switch name {
	case "list_boards":
		return s.svc.ListBoards(ctx)

	case "get_board":
		return s.svc.GetBoard(ctx, strArg(args, "board_id"))

	case "list_lists":
		return s.svc.ListListsByBoard(ctx, strArg(args, "board_id"))

	case "get_card":
		return s.svc.GetCard(ctx, strArg(args, "card_id"))

	case "search_cards":
		return s.svc.SearchCards(ctx, strArg(args, "board_id"), strArg(args, "query"), strArg(args, "assignee"), strArg(args, "status"), strArg(args, "label"))

	case "get_card_dependencies":
		deps, err := s.svc.GetDependencies(ctx, strArg(args, "card_id"))
		if err != nil {
			return nil, err
		}
		dependents, err := s.svc.GetDependents(ctx, strArg(args, "card_id"))
		if err != nil {
			return nil, err
		}
		return map[string]any{"blockers": deps, "dependents": dependents}, nil

	case "get_activity_log":
		cardID := strArg(args, "card_id")
		boardID := strArg(args, "board_id")
		limit := intArg(args, "limit")
		if limit == 0 {
			limit = 50
		}
		if cardID != "" {
			return s.svc.ListActivityByCard(ctx, cardID, limit)
		}
		return s.svc.ListActivityByBoard(ctx, boardID, limit)

	case "create_board":
		return s.svc.CreateBoard(ctx, strArg(args, "name"), strArg(args, "description"), actor)

	case "create_list":
		return s.svc.CreateList(ctx, strArg(args, "board_id"), strArg(args, "name"), intArg(args, "position"), actor)

	case "create_card":
		return s.svc.CreateCard(ctx, strArg(args, "list_id"), strArg(args, "title"), strArg(args, "description"), strArg(args, "assignee"), strArg(args, "priority"), actor, intArg(args, "position"))

	case "move_card":
		return s.svc.MoveCard(ctx, strArg(args, "card_id"), strArg(args, "list_id"), intArg(args, "position"), actor)

	case "update_card":
		return s.svc.UpdateCard(ctx, strArg(args, "card_id"), args, actor)

	case "assign_card":
		return s.svc.AssignCard(ctx, strArg(args, "card_id"), strArg(args, "assignee"), actor)

	case "add_comment":
		return nil, s.svc.AddComment(ctx, strArg(args, "card_id"), actor, strArg(args, "text"))

	case "add_dependency":
		return nil, s.svc.AddDependency(ctx, strArg(args, "card_id"), strArg(args, "depends_on_card_id"), actor)

	case "remove_dependency":
		return nil, s.svc.RemoveDependency(ctx, strArg(args, "card_id"), strArg(args, "depends_on_card_id"), actor)

	case "add_label_to_card":
		return nil, s.svc.AddLabelToCard(ctx, strArg(args, "card_id"), strArg(args, "label_id"), actor)

	case "remove_label_from_card":
		return nil, s.svc.RemoveLabelFromCard(ctx, strArg(args, "card_id"), strArg(args, "label_id"), actor)

	case "delete_card":
		return nil, s.svc.DeleteCard(ctx, strArg(args, "card_id"), actor)

	case "delete_list":
		return nil, s.svc.DeleteList(ctx, strArg(args, "card_id"))

	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

func (s *Server) buildToolDefs() []ToolDef {
	return []ToolDef{
		{Name: "list_boards", Description: "List all boards", InputSchema: obj()},
		{Name: "get_board", Description: "Get board with lists and card counts", InputSchema: obj(prop("board_id", "string", "Board ID"))},
		{Name: "list_lists", Description: "Get all lists for a board with their cards", InputSchema: obj(prop("board_id", "string", "Board ID"))},
		{Name: "get_card", Description: "Get full card detail including labels, dependencies, and activity", InputSchema: obj(prop("card_id", "string", "Card ID"))},
		{Name: "search_cards", Description: "Search cards by title, assignee, status, or label", InputSchema: obj(prop("board_id", "string", "Board ID"), optProp("query", "string", "Search text"), optProp("assignee", "string", "Filter by assignee"), optProp("status", "string", "Filter by status"), optProp("label", "string", "Filter by label name"))},
		{Name: "get_card_dependencies", Description: "Get blockers and dependents for a card", InputSchema: obj(prop("card_id", "string", "Card ID"))},
		{Name: "get_activity_log", Description: "Get activity history for a card or board", InputSchema: obj(optProp("card_id", "string", "Card ID"), optProp("board_id", "string", "Board ID"), optProp("limit", "integer", "Max entries to return"))},
		{Name: "create_board", Description: "Create a new board", InputSchema: obj(prop("name", "string", "Board name"), optProp("description", "string", "Board description"))},
		{Name: "create_list", Description: "Add a list to a board", InputSchema: obj(prop("board_id", "string", "Board ID"), prop("name", "string", "List name"), optProp("position", "integer", "Position in board"))},
		{Name: "create_card", Description: "Create a card in a list", InputSchema: obj(prop("list_id", "string", "List ID"), prop("title", "string", "Card title"), optProp("description", "string", "Card description"), optProp("assignee", "string", "Assignee name"), optProp("priority", "string", "Priority: low, medium, high, critical"))},
		{Name: "move_card", Description: "Move a card to a different list and/or position", InputSchema: obj(prop("card_id", "string", "Card ID"), prop("list_id", "string", "Target list ID"), optProp("position", "integer", "Position in target list"))},
		{Name: "update_card", Description: "Update card fields", InputSchema: obj(prop("card_id", "string", "Card ID"), optProp("title", "string", "New title"), optProp("description", "string", "New description"), optProp("assignee", "string", "New assignee"), optProp("status", "string", "New status"), optProp("priority", "string", "New priority"))},
		{Name: "assign_card", Description: "Assign or unassign a card to an agent", InputSchema: obj(prop("card_id", "string", "Card ID"), prop("assignee", "string", "Agent name (empty to unassign)"))},
		{Name: "add_comment", Description: "Add a comment to a card's activity log", InputSchema: obj(prop("card_id", "string", "Card ID"), prop("text", "string", "Comment text"))},
		{Name: "add_dependency", Description: "Create a dependency between cards", InputSchema: obj(prop("card_id", "string", "The blocked card ID"), prop("depends_on_card_id", "string", "The blocking card ID"))},
		{Name: "remove_dependency", Description: "Remove a dependency between cards", InputSchema: obj(prop("card_id", "string", "The blocked card ID"), prop("depends_on_card_id", "string", "The blocking card ID"))},
		{Name: "add_label_to_card", Description: "Tag a card with a label", InputSchema: obj(prop("card_id", "string", "Card ID"), prop("label_id", "string", "Label ID"))},
		{Name: "remove_label_from_card", Description: "Remove a label from a card", InputSchema: obj(prop("card_id", "string", "Card ID"), prop("label_id", "string", "Label ID"))},
		{Name: "delete_card", Description: "Delete a card", InputSchema: obj(prop("card_id", "string", "Card ID"))},
		{Name: "delete_list", Description: "Delete a list and its cards", InputSchema: obj(prop("list_id", "string", "List ID"))},
	}
}

type schemaProp struct {
	name     string
	typ      string
	desc     string
	required bool
}

func prop(name, typ, desc string) schemaProp {
	return schemaProp{name: name, typ: typ, desc: desc, required: true}
}

func optProp(name, typ, desc string) schemaProp {
	return schemaProp{name: name, typ: typ, desc: desc, required: false}
}

func obj(props ...schemaProp) map[string]any {
	properties := map[string]any{}
	var required []string
	for _, p := range props {
		properties[p.name] = map[string]any{"type": p.typ, "description": p.desc}
		if p.required {
			required = append(required, p.name)
		}
	}
	schema := map[string]any{
		"type":       "object",
		"properties": properties,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}
