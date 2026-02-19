package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"CONVERDA/internal/messaging/application/service"
	"CONVERDA/internal/messaging/controller/dto"
	"CONVERDA/internal/messaging/domain"
	"CONVERDA/internal/messaging/domain/model/entity"
	"CONVERDA/internal/messaging/infrastructure/gateway"
	"CONVERDA/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ConversationController struct {
	svc service.ConversationService
	Hub *gateway.Hub
}

func NewConversationController(svc service.ConversationService, hub *gateway.Hub) *ConversationController {
	return &ConversationController{svc: svc, Hub: hub}
}

// InboundMessage godoc
// @Summary Receive Inbound Message
// @Description Receive a message from external subscriber
// @Tags Messaging
// @Accept json
// @Produce json
// @Param request body dto.InboundMessageRequest true "Message Data"
// @Success 201 {object} dto.MessageResponse
// @Router /conversations/inbound [post]
func (c *ConversationController) InboundMessage(ctx *gin.Context) (interface{}, error) {
	var req dto.InboundMessageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewBadRequestError(err.Error())
	}

	// Context Extraction (from EnvKeyAuth middleware)
	tenantID, _ := ctx.Get("tenant_id")
	envID, _ := ctx.Get("environment_id")

	msg, err := c.svc.ReceiveMessage(ctx.Request.Context(), tenantID.(uuid.UUID), envID.(uuid.UUID), req.SubscriberKey, req.Channel, req.Content, nil)
	if err != nil {
		return nil, response.NewInternalServerError(err.Error())
	}

	go c.broadcastMessage(context.Background(), msg)

	return dto.ToMessageResponse(msg), nil
}

// ReplyMessage godoc
// @Summary Reply to Conversation
// @Description Agent replies to a conversation pool
// @Tags Messaging
// @Accept json
// @Produce json
// @Param id path string true "Pool ID"
// @Param request body dto.ReplyMessageRequest true "Message Data"
// @Success 201 {object} dto.MessageResponse
// @Security BearerAuth
// @Router /conversations/{id}/messages [post]
func (c *ConversationController) ReplyMessage(ctx *gin.Context) (interface{}, error) {
	threadID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return nil, response.NewBadRequestError("Invalid Thread ID")
	}

	var req dto.ReplyMessageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewBadRequestError(err.Error())
	}

	// Context Extraction
	tenantID := uuid.Nil
	agentID := uuid.Nil

	if tID, ok := ctx.Get("tenant_id"); ok {
		tenantID = tID.(uuid.UUID)
	}
	if uID, ok := ctx.Get("user_id"); ok {
		agentID = uID.(uuid.UUID)
	}

	// threadID is resolved by service, but we can pass uuid.Nil or from context if available
	msg, err := c.svc.ReplyMessage(ctx.Request.Context(), tenantID, uuid.Nil, threadID, agentID, req.Content)
	if err != nil {
		if err == domain.ErrThreadResolved {
			return nil, response.NewBadRequestError(err.Error())
		}
		return nil, response.NewInternalServerError(err.Error())
	}

	go c.broadcastMessage(context.Background(), msg)

	return dto.ToMessageResponse(msg), nil
}

// AddInternalNote godoc
// @Summary Add Internal Note
// @Description Add an internal note to a conversation
// @Tags Messaging
// @Accept json
// @Produce json
// @Param id path string true "Thread ID"
// @Param request body dto.InternalNoteRequest true "Note Content"
// @Success 201 {object} dto.MessageResponse
// @Security BearerAuth
// @Router /conversations/{id}/notes [post]
func (c *ConversationController) AddInternalNote(ctx *gin.Context) (interface{}, error) {
	threadID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return nil, response.NewBadRequestError("Invalid Thread ID")
	}

	var req dto.InternalNoteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewBadRequestError(err.Error())
	}

	// Context Extraction
	tenantID := uuid.Nil
	agentID := uuid.Nil

	if tID, ok := ctx.Get("tenant_id"); ok {
		tenantID = tID.(uuid.UUID)
	}
	if uID, ok := ctx.Get("user_id"); ok {
		agentID = uID.(uuid.UUID)
	}

	// envID is optional/extracted from context if needed for validation, but service handles it via uuid.Nil if not passed,
	// checking if thread belongs to environment is good practice if environment_id is in context.
	envID := uuid.Nil
	if eID, ok := ctx.Get("environment_id"); ok {
		envID = eID.(uuid.UUID)
	}

	msg, err := c.svc.AddInternalNote(ctx.Request.Context(), tenantID, envID, threadID, agentID, req.Content)
	if err != nil {
		return nil, response.NewInternalServerError(err.Error())
	}

	go c.broadcastMessage(context.Background(), msg)

	return dto.ToMessageResponse(msg), nil
}

// CreateDirectChat godoc
// @Summary Create or Get Direct Chat
// @Description Create a 1-1 chat between 2 members or return existing one
// @Tags Messaging
// @Accept json
// @Produce json
// @Param request body dto.CreateDirectChatRequest true "Data"
// @Success 200 {object} dto.ThreadResponse
// @Security BearerAuth
// @Router /conversations/direct [post]
func (c *ConversationController) CreateDirectChat(ctx *gin.Context) (interface{}, error) {
	var req dto.CreateDirectChatRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewBadRequestError(err.Error())
	}

	envID := uuid.Nil
	if eID, ok := ctx.Get("environment_id"); ok {
		envID = eID.(uuid.UUID)
	}

	agentID := uuid.Nil
	if uID, ok := ctx.Get("user_id"); ok {
		agentID = uID.(uuid.UUID)
	}

	thread, err := c.svc.GetOrCreateDirectThread(ctx.Request.Context(), envID, agentID, req.Target.ID, req.Target.Type)
	if err != nil {
		return nil, response.NewInternalServerError(err.Error())
	}

	return dto.ToThreadResponse(thread), nil
}

// CreateGroupChat godoc
// @Summary Create Group Chat
// @Description Create a global group chat
// @Tags Messaging
// @Accept json
// @Produce json
// @Param request body dto.CreateGroupChatRequest true "Data"
// @Success 201 {object} dto.ThreadResponse
// @Security BearerAuth
// @Router /conversations/group [post]
func (c *ConversationController) CreateGroupChat(ctx *gin.Context) (interface{}, error) {
	var req dto.CreateGroupChatRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewBadRequestError(err.Error())
	}

	envID := uuid.Nil
	if eID, ok := ctx.Get("environment_id"); ok {
		envID = eID.(uuid.UUID)
	}

	agentID := uuid.Nil
	if uID, ok := ctx.Get("user_id"); ok {
		agentID = uID.(uuid.UUID)
	}

	// Prepare participants lists
	var participants []*entity.ThreadParticipant

	// Add creator (agent) if not exists
	creatorAdded := false
	for _, p := range req.Participants {
		participants = append(participants, &entity.ThreadParticipant{
			EntityID:   p.ID,
			EntityType: p.Type,
		})
		if p.ID == agentID && p.Type == "user" {
			creatorAdded = true
		}
	}

	if !creatorAdded && agentID != uuid.Nil {
		participants = append(participants, &entity.ThreadParticipant{
			EntityID:   agentID,
			EntityType: "user",
		})
	}

	thread, err := c.svc.CreateGroupThread(ctx.Request.Context(), envID, req.Name, participants)
	if err != nil {
		return nil, response.NewInternalServerError(err.Error())
	}

	return dto.ToThreadResponse(thread), nil
}

// UpdateGroupChat godoc
// @Summary Update Group Chat
// @Description Update group name
// @Tags Messaging
// @Accept json
// @Produce json
// @Param id path string true "Thread ID"
// @Param request body dto.UpdateGroupChatRequest true "Data"
// @Success 200
// @Security BearerAuth
// @Router /conversations/group/{id} [patch]
func (c *ConversationController) UpdateGroupChat(ctx *gin.Context) (interface{}, error) {
	threadID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return nil, response.NewBadRequestError("Invalid Thread ID")
	}

	var req dto.UpdateGroupChatRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewBadRequestError(err.Error())
	}

	envID := uuid.Nil
	if eID, ok := ctx.Get("environment_id"); ok {
		envID = eID.(uuid.UUID)
	}

	if err := c.svc.UpdateGroupThread(ctx.Request.Context(), envID, threadID, req.Name); err != nil {
		if err == domain.ErrNotGroupThread {
			return nil, response.NewBadRequestError(err.Error())
		}
		return nil, response.NewInternalServerError(err.Error())
	}

	return map[string]string{"status": "updated"}, nil
}

// AddGroupParticipants godoc
// @Summary Add Group Participants
// @Description Add members to group
// @Tags Messaging
// @Accept json
// @Produce json
// @Param id path string true "Thread ID"
// @Param request body dto.AddParticipantsRequest true "Data"
// @Success 200
// @Security BearerAuth
// @Router /conversations/group/{id}/participants [post]
func (c *ConversationController) AddGroupParticipants(ctx *gin.Context) (interface{}, error) {
	threadID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return nil, response.NewBadRequestError("Invalid Thread ID")
	}

	var req dto.AddParticipantsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewBadRequestError(err.Error())
	}

	var participants []*entity.ThreadParticipant
	for _, p := range req.Participants {
		participants = append(participants, &entity.ThreadParticipant{
			EntityID:   p.ID,
			EntityType: p.Type,
		})
	}

	envID := uuid.Nil
	if eID, ok := ctx.Get("environment_id"); ok {
		envID = eID.(uuid.UUID)
	}

	if err := c.svc.AddGroupParticipants(ctx.Request.Context(), envID, threadID, participants); err != nil {
		return nil, response.NewInternalServerError(err.Error())
	}

	return map[string]string{"status": "participants_added"}, nil
}

// RemoveGroupParticipant godoc
// @Summary Remove Group Participant
// @Description Remove a member from group
// @Tags Messaging
// @Produce json
// @Param id path string true "Thread ID"
// @Param memberID path string true "Member ID"
// @Success 200
// @Security BearerAuth
// @Router /conversations/group/{id}/participants/{memberID} [delete]
func (c *ConversationController) RemoveGroupParticipant(ctx *gin.Context) (interface{}, error) {
	threadID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return nil, response.NewBadRequestError("Invalid Thread ID")
	}

	memberID, err := uuid.Parse(ctx.Param("memberID"))
	if err != nil {
		return nil, response.NewBadRequestError("Invalid Member ID")
	}

	envID := uuid.Nil
	if eID, ok := ctx.Get("environment_id"); ok {
		envID = eID.(uuid.UUID)
	}

	if err := c.svc.RemoveGroupParticipant(ctx.Request.Context(), envID, threadID, "user", memberID); err != nil {
		return nil, response.NewInternalServerError(err.Error())
	}

	return map[string]string{"status": "participant_removed"}, nil
}

// MarkAsRead godoc
// @Summary Mark Thread Read
// @Description Mark all messages in thread as read
// @Tags Messaging
// @Produce json
// @Param id path string true "Thread ID"
// @Success 200
// @Security BearerAuth
// @Router /conversations/{id}/read [post]
func (c *ConversationController) MarkAsRead(ctx *gin.Context) (interface{}, error) {
	threadID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return nil, response.NewBadRequestError("Invalid Thread ID")
	}

	agentID := uuid.Nil
	if uID, ok := ctx.Get("user_id"); ok {
		agentID = uID.(uuid.UUID)
	}

	envID := uuid.Nil
	if eID, ok := ctx.Get("environment_id"); ok {
		envID = eID.(uuid.UUID)
	}

	if err := c.svc.MarkThreadRead(ctx.Request.Context(), envID, threadID, agentID); err != nil {
		if err == domain.ErrParticipantNotFound {
			return nil, response.NewNotFoundError("Participant not found")
		}
		return nil, response.NewInternalServerError(err.Error())
	}

	return map[string]string{"status": "read"}, nil
}

// GetThread godoc
// @Summary Get Conversation Thread
// @Description List messages in a conversation
// @Tags Messaging
// @Produce json
// @Param id path string true "Pool ID"
// @Param cursor query string false "Pagination Cursor (Base64)"
// @Param direction query string false "Direction: 'before' (older) or 'after' (newer)"
// @Param limit query int false "Limit (default 20)"
// @Success 200 {object} map[string]interface{}
// @Router /conversations/{id} [get]
func (c *ConversationController) GetThread(ctx *gin.Context) (interface{}, error) {
	threadID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return nil, response.NewBadRequestError("Invalid Pool ID")
	}

	envID := uuid.Nil
	if eID, ok := ctx.Get("environment_id"); ok {
		envID = eID.(uuid.UUID)
	}

	cursorStr := ctx.Query("cursor")
	if cursorStr != "" || ctx.Query("direction") != "" {
		// New Cursor Pagination Logic
		direction := ctx.DefaultQuery("direction", "before")
		limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))

		msgs, nextCursor, prevCursor, err := c.svc.GetMessagesByCursor(ctx.Request.Context(), envID, threadID, cursorStr, direction, limit)
		if err != nil {
			return nil, response.NewInternalServerError("failed to get messages")
		}

		// Use map for response structure (or create DTO later)
		return map[string]interface{}{
			"data": dto.ToMessageResponseList(msgs),
			"meta": map[string]interface{}{
				"next_cursor": nextCursor,
				"prev_cursor": prevCursor,
				"has_more":    nextCursor != "" || prevCursor != "",
			},
		}, nil
	}

	// Fallback to old Offset Pagination
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	offset := (page - 1) * limit

	msgs, total, err := c.svc.GetMessagesByThread(ctx.Request.Context(), envID, threadID, limit, offset)
	if err != nil {
		return nil, response.NewInternalServerError("failed to get messages")
	}

	return dto.NewPaginatedResponse(dto.ToMessageResponseList(msgs), page, limit, total), nil
}

// ListConversations godoc
// @Summary List Conversations
// @Description List conversations with filters
// @Tags Messaging
// @Produce json
// @Param status query string false "Status (unassigned, assigned, resolved)"
// @Param assigned_to query string false "filter: 'me' or 'all'"
// @Param page query int false "Page"
// @Param limit query int false "Limit"
// @Success 200 {object} dto.PaginatedResponse
// @Security BearerAuth
// @Router /conversations [get]
func (c *ConversationController) ListConversations(ctx *gin.Context) (interface{}, error) {
	status := ctx.Query("status")
	assignedTo := ctx.Query("assigned_to")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	tenantID := uuid.Nil
	agentID := uuid.Nil
	envID := uuid.Nil // Assuming envID might be needed for ListThreads, if not, it can be uuid.Nil
	if tID, ok := ctx.Get("tenant_id"); ok {
		tenantID = tID.(uuid.UUID)
	}
	if uID, ok := ctx.Get("user_id"); ok {
		agentID = uID.(uuid.UUID)
	}
	if eID, ok := ctx.Get("environment_id"); ok { // Added envID extraction if available
		envID = eID.(uuid.UUID)
	}

	assignedToMe := (assignedTo == "me")

	threads, total, err := c.svc.ListThreads(ctx.Request.Context(), tenantID, envID, status, assignedToMe, agentID, limit, offset)
	if err != nil {
		return nil, response.NewInternalServerError("failed to list threads")
	}

	return dto.NewPaginatedResponse(dto.ToThreadResponseList(threads), page, limit, total), nil
}

// AssignConversation godoc
// @Summary Assign Conversation
// @Description Assign a conversation to an agent
// @Tags Messaging
// @Accept json
// @Produce json
// @Param id path string true "Pool ID"
// @Param request body dto.AssignConversationRequest false "Assignee Data"
// @Success 200
// @Security BearerAuth
// @Router /conversations/{id}/assign [patch]
func (c *ConversationController) AssignThread(ctx *gin.Context) (interface{}, error) {
	threadID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return nil, response.NewBadRequestError("Invalid Thread ID")
	}

	var req dto.AssignConversationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// Optional body, ignore error? No, ShouldBindJSON returns EOF if empty sometimes, logic depends on framework version.
		// If body is optional, use Bind checks. For now assume strict JSON if provided.
	}

	tenantID := uuid.Nil
	agentID := uuid.Nil
	if tID, ok := ctx.Get("tenant_id"); ok {
		tenantID = tID.(uuid.UUID)
	}
	if uID, ok := ctx.Get("user_id"); ok {
		agentID = uID.(uuid.UUID)
	}

	assigneeID := agentID
	if req.MemberID != nil {
		assigneeID = *req.MemberID
	}

	envID := uuid.Nil
	if eID, ok := ctx.Get("environment_id"); ok {
		envID = eID.(uuid.UUID)
	}

	if err := c.svc.AssignThread(ctx.Request.Context(), tenantID, envID, threadID, assigneeID); err != nil {
		if err == domain.ErrInvalidStatusTransition {
			return nil, response.NewBadRequestError(err.Error())
		}
		return nil, response.NewInternalServerError(err.Error())
	}

	return map[string]string{"status": "assigned"}, nil
}

// UnassignConversation godoc
// @Summary Unassign Conversation
// @Description Return conversation to the pool (unassigned)
// @Tags Messaging
// @Produce json
// @Param id path string true "Thread ID"
// @Success 200
// @Security BearerAuth
// @Router /conversations/{id}/unassign [post]
func (c *ConversationController) UnassignThread(ctx *gin.Context) (interface{}, error) {
	threadID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return nil, response.NewBadRequestError("Invalid Thread ID")
	}

	tenantID := uuid.Nil
	if tID, ok := ctx.Get("tenant_id"); ok {
		tenantID = tID.(uuid.UUID)
	}

	envID := uuid.Nil
	if eID, ok := ctx.Get("environment_id"); ok {
		envID = eID.(uuid.UUID)
	}

	if err := c.svc.UnassignThread(ctx.Request.Context(), tenantID, envID, threadID); err != nil {
		if err == domain.ErrThreadResolved {
			return nil, response.NewBadRequestError(err.Error())
		}
		return nil, response.NewInternalServerError(err.Error())
	}

	return map[string]string{"status": "unassigned"}, nil
}

// BulkAssignConversations godoc
// @Summary Bulk Assign Conversations
// @Description Assign multiple conversations efficiently
// @Tags Messaging
// @Accept json
// @Produce json
// @Param request body dto.BulkAssignRequest true "Data"
// @Success 200
// @Security BearerAuth
// @Router /conversations/assign [patch]
func (c *ConversationController) BulkAssignThreads(ctx *gin.Context) (interface{}, error) {
	var req dto.BulkAssignRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewBadRequestError(err.Error())
	}

	tenantID := uuid.Nil
	agentID := uuid.Nil
	if tID, ok := ctx.Get("tenant_id"); ok {
		tenantID = tID.(uuid.UUID)
	}
	if uID, ok := ctx.Get("user_id"); ok {
		agentID = uID.(uuid.UUID)
	}

	assigneeID := agentID
	if req.MemberID != nil {
		assigneeID = *req.MemberID
	}

	envID := uuid.Nil
	if eID, ok := ctx.Get("environment_id"); ok {
		envID = eID.(uuid.UUID)
	}

	if err := c.svc.BulkAssignThreads(ctx.Request.Context(), tenantID, envID, req.ThreadIDs, assigneeID); err != nil {
		// Could carry partial success info if we changed service return, but simpler for now
		return nil, response.NewInternalServerError(err.Error())
	}

	return map[string]string{"status": "bulk_assigned"}, nil
}

// ResolveConversation godoc
// @Summary Resolve Conversation
// @Description Mark conversation as resolved
// @Tags Messaging
// @Produce json
// @Param id path string true "Pool ID"
// @Success 200
// @Security BearerAuth
// @Router /conversations/{id}/resolve [post]
func (c *ConversationController) ResolveThread(ctx *gin.Context) (interface{}, error) {
	threadID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return nil, response.NewBadRequestError("Invalid Thread ID")
	}

	tenantID := uuid.Nil
	if tID, ok := ctx.Get("tenant_id"); ok {
		tenantID = tID.(uuid.UUID)
	}

	envID := uuid.Nil
	if eID, ok := ctx.Get("environment_id"); ok {
		envID = eID.(uuid.UUID)
	}

	if err := c.svc.ResolveThread(ctx.Request.Context(), tenantID, envID, threadID); err != nil {
		return nil, response.NewInternalServerError(err.Error())
	}

	return map[string]string{"status": "resolved"}, nil
}

// GetThreadAuditTrail godoc
// @Summary Get Conversation Audit Trail
// @Description Fetch paginated history of assignments and resolutions
// @Tags Messaging
// @Produce json
// @Param id path string true "Thread ID"
// @Param page query int false "Page number (default 1)"
// @Param page_size query int false "Items per page (default 20, max 100)"
// @Param from query string false "Filter from date (RFC3339)"
// @Param to query string false "Filter to date (RFC3339)"
// @Success 200 {object} dto.AuditTrailResponse
// @Security BearerAuth
// @Router /conversations/{id}/audit-trail [get]
func (c *ConversationController) GetThreadAuditTrail(ctx *gin.Context) (interface{}, error) {
	threadID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return nil, response.NewBadRequestError("Invalid Thread ID")
	}

	envID := uuid.Nil
	if eID, ok := ctx.Get("environment_id"); ok {
		envID = eID.(uuid.UUID)
	}

	// Pagination defaults
	page := 1
	pageSize := 20
	if p := ctx.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if ps := ctx.Query("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 {
			pageSize = v
			if pageSize > 100 {
				pageSize = 100
			}
		}
	}

	// Optional date-range filter
	var from, to *time.Time
	if f := ctx.Query("from"); f != "" {
		if t, err := time.Parse(time.RFC3339, f); err == nil {
			from = &t
		}
	}
	if t := ctx.Query("to"); t != "" {
		if parsed, err := time.Parse(time.RFC3339, t); err == nil {
			to = &parsed
		}
	}

	trail, err := c.svc.GetThreadAuditTrail(ctx.Request.Context(), envID, threadID, page, pageSize, from, to)
	if err != nil {
		return nil, response.NewInternalServerError(err.Error())
	}

	return trail, nil
}

// GetTeamStats godoc
// @Summary Get Team Analytics
// @Description Get messaging performance metrics for the team
// @Tags Analytics
// @Produce json
// @Param from query string false "From Date (YYYY-MM-DD)"
// @Param to query string false "To Date (YYYY-MM-DD)"
// @Success 200 {object} dto.TeamStatsResponse
// @Security BearerAuth
// @Router /analytics/conversations/team [get]
func (c *ConversationController) GetTeamStats(ctx *gin.Context) (interface{}, error) {
	// Parse Dates
	now := time.Now()
	fromStr := ctx.DefaultQuery("from", now.AddDate(0, 0, -30).Format("2006-01-02"))
	toStr := ctx.DefaultQuery("to", now.Format("2006-01-02"))

	from, _ := time.Parse("2006-01-02", fromStr)
	to, _ := time.Parse("2006-01-02", toStr)
	// Adjust 'to' to end of day
	to = to.Add(24 * time.Hour).Add(-1 * time.Second)

	tenantID := uuid.Nil
	envID := uuid.Nil
	if tID, ok := ctx.Get("tenant_id"); ok {
		tenantID = tID.(uuid.UUID)
	}
	if eID, ok := ctx.Get("environment_id"); ok {
		envID = eID.(uuid.UUID)
	}

	stats, err := c.svc.GetTeamStats(ctx.Request.Context(), tenantID, envID, from, to)
	if err != nil {
		return nil, response.NewInternalServerError(err.Error())
	}

	return dto.TeamStatsResponse{
		TotalConversations: stats.TotalConversations,
		AvgResponseTime:    stats.AvgResponseTime,
		ResolvedCount:      stats.ResolvedCount,
		SLAComplianceRate:  stats.SLAComplianceRate,
	}, nil
}

// GetMyStats godoc
// @Summary Get My Analytics
// @Description Get messaging performance metrics for the current agent
// @Tags Analytics
// @Produce json
// @Param from query string false "From Date (YYYY-MM-DD)"
// @Param to query string false "To Date (YYYY-MM-DD)"
// @Success 200 {object} dto.AgentStatsResponse
// @Security BearerAuth
// @Router /analytics/conversations/me [get]
func (c *ConversationController) GetMyStats(ctx *gin.Context) (interface{}, error) {
	// Parse Dates
	now := time.Now()
	fromStr := ctx.DefaultQuery("from", now.AddDate(0, 0, -30).Format("2006-01-02"))
	toStr := ctx.DefaultQuery("to", now.Format("2006-01-02"))

	from, _ := time.Parse("2006-01-02", fromStr)
	to, _ := time.Parse("2006-01-02", toStr)
	to = to.Add(24 * time.Hour).Add(-1 * time.Second)

	tenantID := uuid.Nil
	envID := uuid.Nil
	agentID := uuid.Nil
	if tID, ok := ctx.Get("tenant_id"); ok {
		tenantID = tID.(uuid.UUID)
	}
	if eID, ok := ctx.Get("environment_id"); ok {
		envID = eID.(uuid.UUID)
	}
	if uID, ok := ctx.Get("user_id"); ok {
		agentID = uID.(uuid.UUID)
	}

	stats, err := c.svc.GetAgentStats(ctx.Request.Context(), tenantID, envID, agentID, from, to)
	if err != nil {
		return nil, response.NewInternalServerError(err.Error())
	}

	return dto.AgentStatsResponse{
		MemberID:           stats.MemberID,
		TotalAssigned:      stats.TotalAssigned,
		TotalResolved:      stats.TotalResolved,
		AvgResponseTime:    stats.AvgResponseTime,
		CurrentOpenThreads: stats.CurrentOpenThreads,
	}, nil
}

// ServeWS godoc
// @Summary WebSocket Connection
// @Description Connect to WebSocket for real-time messaging updates (requires auth)
// @Tags Messaging
// @Router /conversations/ws [get]
func (c *ConversationController) ServeWS(ctx *gin.Context) {
	// Auth middleware sets user_id in context (JWT bearer or query token)
	userID := uuid.Nil
	if uID, ok := ctx.Get("user_id"); ok {
		userID = uID.(uuid.UUID)
	}

	if userID == uuid.Nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	envID := uuid.Nil
	if eID, ok := ctx.Get("environment_id"); ok {
		envID = eID.(uuid.UUID)
	}

	gateway.ServeWs(c.Hub, ctx, userID, envID)
}

// GetPersonalDashboard godoc
// @Summary Get Personal Dashboard
// @Description Get comprehensive metrics and activity for the current agent
// @Tags Analytics
// @Produce json
// @Param environment_id query string true "Environment ID"
// @Param from query string false "From Date (YYYY-MM-DD)"
// @Param to query string false "To Date (YYYY-MM-DD)"
// @Param sla_threshold query int false "SLA Threshold override (seconds)"
// @Success 200 {object} dto.PersonalDashboardResponse
// @Security BearerAuth
// @Router /dashboards/personal [get]
func (c *ConversationController) GetPersonalDashboard(ctx *gin.Context) (interface{}, error) {
	// Parse Environment ID
	envID, err := uuid.Parse(ctx.Query("environment_id"))
	if err != nil {
		return nil, response.NewBadRequestError("Invalid Environment ID")
	}

	// Parse Dates
	now := time.Now()
	fromStr := ctx.DefaultQuery("from", now.AddDate(0, 0, -1).Format("2006-01-02")) // Default to last 24h
	toStr := ctx.DefaultQuery("to", now.Format("2006-01-02"))

	from, _ := time.Parse("2006-01-02", fromStr)
	to, _ := time.Parse("2006-01-02", toStr)
	to = to.Add(24 * time.Hour).Add(-1 * time.Second)

	sla, _ := strconv.Atoi(ctx.Query("sla_threshold"))

	agentID := uuid.Nil
	if uID, ok := ctx.Get("user_id"); ok {
		agentID = uID.(uuid.UUID)
	}

	req := dto.DashboardStatsRequest{
		EnvironmentID: envID,
		From:          from,
		To:            to,
		SLAThreshold:  sla,
	}

	dashboard, err := c.svc.GetPersonalDashboard(ctx.Request.Context(), agentID, req)
	if err != nil {
		return nil, response.NewInternalServerError(err.Error())
	}

	return dashboard, nil
}

func (c *ConversationController) broadcastMessage(ctx context.Context, msg *entity.Message) {
	if c.Hub == nil {
		return
	}

	// 1. Fetch thread participants (to know who to notify)
	thread, err := c.svc.GetThread(ctx, msg.EnvironmentID, msg.ThreadID)
	if err != nil {
		// Log error (we need a logger injected, but for now just skip)
		return
	}

	// 2. Extract User IDs
	var userIDs []uuid.UUID
	for _, p := range thread.Participants {
		if p.EntityType == "user" {
			userIDs = append(userIDs, p.EntityID)
		}
	}

	// 3. Prepare payload
	// We wrap it in an event structure
	event := map[string]interface{}{
		"event": "message_created",
		"data":  dto.ToMessageResponse(msg),
	}

	jsonBytes, err := json.Marshal(event)
	if err != nil {
		return
	}

	// 4. Broadcast
	c.Hub.BroadcastToUsers(jsonBytes, userIDs)
}

// GetTeamDashboard godoc
// @Summary Get Team Dashboard
// @Description Get comprehensive team performance metrics
// @Tags Dashboards
// @Produce json
// @Param environment_id query string true "Environment ID"
// @Param from query string false "From Time (RFC3339)"
// @Param to query string false "To Time (RFC3339)"
// @Param sla_threshold query int false "SLA Threshold in seconds"
// @Success 200 {object} dto.TeamDashboardResponse
// @Security BearerAuth
// @Router /dashboards/team [get]
func (c *ConversationController) GetTeamDashboard(ctx *gin.Context) (interface{}, error) {
	var req dto.DashboardStatsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		return nil, response.NewBadRequestError(err.Error())
	}

	res, err := c.svc.GetTeamDashboard(ctx.Request.Context(), req)
	if err != nil {
		return nil, response.NewInternalServerError(err.Error())
	}

	return res, nil
}

// GetPartnerDashboard godoc
// @Summary Get Partner Dashboard
// @Description Get comprehensive partner performance metrics
// @Tags Dashboards
// @Produce json
// @Param environment_id query string true "Environment ID"
// @Param from query string false "From Time (RFC3339)"
// @Param to query string false "To Time (RFC3339)"
// @Param sla_threshold query int false "SLA Threshold in seconds"
// @Success 200 {object} dto.PartnerDashboardResponse
// @Security BearerAuth
// @Router /dashboards/partner [get]
func (c *ConversationController) GetPartnerDashboard(ctx *gin.Context) (interface{}, error) {
	var req dto.DashboardStatsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		return nil, response.NewBadRequestError(err.Error())
	}

	res, err := c.svc.GetPartnerDashboard(ctx.Request.Context(), req)
	if err != nil {
		return nil, response.NewInternalServerError(err.Error())
	}

	return res, nil
}

// GetAgentDashboard godoc
// @Summary Get Agent Dashboard
// @Description Get comprehensive agent performance metrics
// @Tags Dashboards
// @Produce json
// @Param member_id path string true "Member ID"
// @Param environment_id query string true "Environment ID"
// @Param from query string false "From Time (RFC3339)"
// @Param to query string false "To Time (RFC3339)"
// @Param sla_threshold query int false "SLA Threshold in seconds"
// @Success 200 {object} dto.AgentDashboardResponse
// @Security BearerAuth
// @Router /dashboards/agent/{member_id} [get]
func (c *ConversationController) GetAgentDashboard(ctx *gin.Context) (interface{}, error) {
	memberID, err := uuid.Parse(ctx.Param("member_id"))
	if err != nil {
		return nil, response.NewBadRequestError("Invalid Member ID")
	}

	var req dto.DashboardStatsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		return nil, response.NewBadRequestError(err.Error())
	}

	res, err := c.svc.GetAgentDashboard(ctx.Request.Context(), memberID, req)
	if err != nil {
		return nil, response.NewInternalServerError(err.Error())
	}

	return res, nil
}
